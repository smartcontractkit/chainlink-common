package flags

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/config/configdoc"
)

// newRoot returns a root command wired for testing: a --config flag and silenced output. Each
// command gets its own Viper instance (see commandMetaData.v), created fresh the first time it's
// registered, so tests can't leak keys or a consumed config-file-load guard into each other the
// way a single process-global viper would.
func newRoot(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "app", RunE: func(*cobra.Command, []string) error { return nil }}
	cmd.PersistentFlags().String("config", "", "path to config file")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd
}

// run registers target on a fresh root command and executes it with args, returning the error
// from the decode/validate step (nil if the config was accepted).
func run(t *testing.T, target any, args ...string) error {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, target, DefaultTOMLOptions("TEST")))
	root.SetArgs(args)
	return root.Execute()
}

// runWithOptions is like run, but lets the caller supply Options directly instead of always
// going through DefaultTOMLOptions - for exercising decoding under bare/custom Options.
func runWithOptions(t *testing.T, target any, opts Options, args ...string) error {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, target, opts))
	root.SetArgs(args)
	return root.Execute()
}

// writeConfig writes a TOML config file and returns its path, for passing via --config.
func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestRequiredLeaf(t *testing.T) {
	type cfg struct {
		Name string `toml:"name" validate:"required"`
	}

	t.Run("missing", func(t *testing.T) {
		var c cfg
		require.ErrorContains(t, run(t, &c), "'required'")
	})

	t.Run("provided", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--name", "x"))
		assert.Equal(t, "x", c.Name)
	})
}

func TestUntaggedFieldUsesItsGoNameInKebabCase(t *testing.T) {
	type inner struct {
		PollInterval config.Duration
	}
	type cfg struct {
		URL              string
		ChainID          uint32
		UseRealDBForFake bool
		Chain            inner
	}

	// No tags at all: the key, the flag and the env var all come from the field name, so a struct
	// only carries a tag where its key differs from it.
	var c cfg
	require.NoError(t, run(t, &c,
		"--url", "postgres://x",
		"--chain-id", "137",
		"--use-real-db-for-fake",
		"--chain.poll-interval", "7s",
	))
	assert.Equal(t, "postgres://x", c.URL)
	assert.Equal(t, uint32(137), c.ChainID)
	assert.True(t, c.UseRealDBForFake)
	assert.Equal(t, 7*time.Second, c.Chain.PollInterval.Duration())
}

func TestUntaggedFieldReadsItsEnvVar(t *testing.T) {
	type cfg struct {
		FinalityTagEnabled bool
	}

	t.Setenv("TEST_FINALITY_TAG_ENABLED", "true")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.True(t, c.FinalityTagEnabled)
}

func TestTagWinsOverTheGoName(t *testing.T) {
	type cfg struct {
		// The plural field is bound to a singular key, which the field name cannot produce.
		HTTPURLs []string `toml:"http-url"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--http-url", "https://one"))
	assert.Equal(t, []string{"https://one"}, c.HTTPURLs)
}

func TestNestedStructDecodes(t *testing.T) {
	type inner struct {
		Host string `toml:"host"`
	}
	type cfg struct {
		Chain inner `toml:"chain"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--chain.host", "example.com"))
	assert.Equal(t, "example.com", c.Chain.Host)
}

func TestSquashedStructDecodes(t *testing.T) {
	type inner struct {
		Host string `toml:"host"`
	}
	type cfg struct {
		// DefaultTOMLOptions names the squash option "inline", matching TOML's inline table.
		Inner inner `toml:",inline"`
	}

	// Squashed fields contribute no key segment, so the flag is --host, not --inner.host.
	var c cfg
	require.NoError(t, run(t, &c, "--host", "example.com"))
	assert.Equal(t, "example.com", c.Inner.Host)
}

func TestSquashOptionIsConfigurable(t *testing.T) {
	type inner struct {
		Host string `toml:"host"`
	}
	type cfg struct {
		Inner inner `toml:",squash"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.DecoderConfig.SquashTagOption = "squash"

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, opts))

	root.SetArgs([]string{"--host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Inner.Host)
}

// Exported so the embedded field itself is exported (an embedded unexported type is an
// unexported field, which the walker skips).
type EmbeddedInner struct {
	Host string `toml:"host" mapstructure:"host"`
}

func TestEmbeddedStructIsSquashedWhenDecoderSquashes(t *testing.T) {
	type cfg struct {
		EmbeddedInner // no tag; DefaultTOMLOptions sets DecoderConfig.Squash
	}

	var c cfg
	require.NoError(t, run(t, &c, "--host", "example.com"))
	assert.Equal(t, "example.com", c.Host)
}

func TestEmbeddedStructIsNestedWhenDecoderDoesNot(t *testing.T) {
	type cfg struct {
		EmbeddedInner
	}

	// Squash off: mapstructure treats the embedded struct as a field named after its type,
	// so the flag must be namespaced to match, not flattened to --host.
	opts := DefaultTOMLOptions("TEST")
	opts.DecoderConfig.Squash = false

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, opts))

	root.SetArgs([]string{"--embedded-inner.host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Host)
}

func TestNamedEmbeddedStructIsRejectedWhenSquashing(t *testing.T) {
	type cfg struct {
		// The name is a lie under squashing: these fields flatten into the parent, but
		// encoding/json would nest them under "inner".
		EmbeddedInner `toml:"inner"`
	}

	err := RegisterCommandFlags(newRoot(t), &cfg{}, DefaultTOMLOptions("TEST"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be named")
}

func TestNamedEmbeddedStructIsAllowedWhenNotSquashing(t *testing.T) {
	type cfg struct {
		EmbeddedInner `toml:"inner"`
	}

	// Squash off, so the name is honoured rather than ignored - and agrees with json.
	opts := DefaultTOMLOptions("TEST")
	opts.DecoderConfig.Squash = false

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, opts))

	root.SetArgs([]string{"--inner.host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Host)
}

func TestUnnamedEmbeddedStructMayCarryTagOptions(t *testing.T) {
	type cfg struct {
		// Options without a name are fine: nothing is being contradicted.
		EmbeddedInner `toml:",inline"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--host", "example.com"))
	assert.Equal(t, "example.com", c.Host)
}

// EmbeddedShared stands in for a config struct owned elsewhere (the package that consumes it),
// which a binary embeds by pointer to add its own settings alongside without copying it.
type EmbeddedShared struct {
	Host string `toml:"host" usage:"remote host"`
}

func TestEmbeddedPointerStructIsSquashed(t *testing.T) {
	type cfg struct {
		*EmbeddedShared `toml:",inline"`

		Mine string `toml:"mine" usage:"this binary's own setting"`
	}

	// Non-nil, the way a caller supplies the instance the shared defaults were set on.
	c := cfg{EmbeddedShared: &EmbeddedShared{}}
	require.NoError(t, run(t, &c, "--host", "example.com", "--mine", "x"))
	assert.Equal(t, "example.com", c.Host, "the embedded fields flatten into the parent")
	assert.Equal(t, "x", c.Mine)
}

func TestEmbeddedPointerStructIsSquashedUnderNamespace(t *testing.T) {
	type cfg struct {
		*EmbeddedShared `toml:",inline"`

		Mine string `toml:"mine" usage:"this binary's own setting"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "remote"

	root := newRoot(t)
	c := cfg{EmbeddedShared: &EmbeddedShared{}}
	require.NoError(t, RegisterCommandFlags(root, &c, opts))

	root.SetArgs([]string{"--remote.host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Host)

	// One table, not an empty "[remote]" plus a "[remote.]" holding the embedded fields.
	toml, err := structToDocs(&c, opts.Namespace, opts)
	require.NoError(t, err)
	assert.Contains(t, toml, "[remote]")
	assert.NotContains(t, toml, "[remote.]")
	assert.Contains(t, toml, "host = ")
	assert.Contains(t, toml, "mine = ")
}

// A cross-field rule can only name fields of its own struct, so with settings split across an
// embedded struct the rule has to sit on the outer field - naming a promoted sibling.
func TestExcludedWithNamesPromotedSibling(t *testing.T) {
	type cfg struct {
		*EmbeddedShared `toml:",inline"`

		Proxy string `toml:"proxy" usage:"use a proxy instead of a host" validate:"required_without=Host,excluded_with=Host"`
	}

	t.Run("neither set", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		require.ErrorContains(t, run(t, &c), "'required_without'")
	})

	t.Run("both set", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		require.ErrorContains(t, run(t, &c, "--host", "example.com", "--proxy", "localhost:1"), "'excluded_with'")
	})

	t.Run("only the promoted one set", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		require.NoError(t, run(t, &c, "--host", "example.com"))
	})

	t.Run("only the outer one set", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		require.NoError(t, run(t, &c, "--proxy", "localhost:1"))
	})

	t.Run("docs name it by its config key", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		toml, err := structToDocs(&c, "", DefaultTOMLOptions("TEST"))
		require.NoError(t, err)
		assert.Contains(t, toml, "must not be set when host is set")
	})

	t.Run("example config picks one", func(t *testing.T) {
		c := cfg{EmbeddedShared: &EmbeddedShared{}}
		example, err := exampleDoc(&c, "", DefaultTOMLOptions("TEST"))
		require.NoError(t, err)
		// The promoted field is declared first and wins, even though the rule ruling the other
		// one out lives in a different struct.
		assert.Contains(t, example, "host = ")
		assert.NotContains(t, example, "proxy = ")
	})
}

func TestTagNameIsConfigurable(t *testing.T) {
	type cfg struct {
		Host string `mapstructure:"host"`
	}

	// The zero Options falls back to mapstructure's own tag name and squash option.
	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, Options{Prefixes: []string{"TEST"}}))

	root.SetArgs([]string{"--host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Host)
}

func TestOptionalNestedStructStaysNil(t *testing.T) {
	type inner struct {
		Host string `toml:"host" validate:"required"`
	}
	type cfg struct {
		// Nothing under it was supplied, so it must stay nil rather than being allocated
		// with defaults - otherwise its inner `required` would fire for an absent section.
		Inner *inner `toml:"inner"`
	}

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Nil(t, c.Inner)
}

func TestOptionalNestedStructAllocatedWhenSet(t *testing.T) {
	type inner struct {
		Host string `toml:"host" validate:"required"`
	}
	type cfg struct {
		Inner *inner `toml:"inner"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--inner.host", "example.com"))
	require.NotNil(t, c.Inner)
	assert.Equal(t, "example.com", c.Inner.Host)
}

// The exactly-one-of shape: two mutually exclusive nested sections, each required when the
// other is absent. Pointers make "absent" representable, which is what stops the unselected
// section's own `required` fields from being reported.
type modeB struct {
	X int32 `toml:"x" validate:"required"`
	Z int32 `toml:"z" validate:"required"`
	W int32 `toml:"w"` // deliberately not required
}

type modeC struct {
	Y int32 `toml:"y"`
	Q int32 `toml:"q" validate:"required"`
}

type modesCfg struct {
	B *modeB `toml:"b" validate:"required_without=C,excluded_with=C"`
	C *modeC `toml:"c" validate:"required_without=B,excluded_with=B"`
}

func TestExclusiveModes_NeitherSet(t *testing.T) {
	var c modesCfg
	err := run(t, &c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'required_without'")
}

func TestExclusiveModes_OnlyBSet(t *testing.T) {
	var c modesCfg
	// C is absent, so none of C's own required fields should be reported.
	require.NoError(t, run(t, &c, "--b.x", "1", "--b.z", "2"))
	require.NotNil(t, c.B)
	assert.Nil(t, c.C)
	assert.Equal(t, int32(1), c.B.X)
}

func TestExclusiveModes_OnlyCSet(t *testing.T) {
	var c modesCfg
	require.NoError(t, run(t, &c, "--c.q", "5"))
	require.NotNil(t, c.C)
	assert.Nil(t, c.B)
}

func TestExclusiveModes_BothSetIsRejected(t *testing.T) {
	var c modesCfg
	err := run(t, &c, "--b.x", "1", "--b.z", "2", "--c.q", "5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'excluded_with'")
}

func TestExclusiveModes_PartialSectionReportsItsOwnRequired(t *testing.T) {
	var c modesCfg
	// B is present but incomplete: its missing required field is what should be reported.
	err := run(t, &c, "--b.x", "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'Z'")
}

func TestExclusiveModes_OnlyNonRequiredFieldSet(t *testing.T) {
	var c modesCfg
	// Setting only W allocates B, so B's unset required fields are reported.
	err := run(t, &c, "--b.w", "9")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'X'")
	assert.Contains(t, err.Error(), "'Z'")
}

func TestValueStructWithCrossFieldRuleIsRejected(t *testing.T) {
	type inner struct {
		Host string `toml:"host" validate:"required"`
	}
	type other struct {
		Addr string `toml:"addr"`
	}
	type cfg struct {
		// Not a pointer, so it can never be absent - registration should refuse it rather
		// than silently mis-validating at run time.
		A inner `toml:"a" validate:"excluded_with=B"`
		B other `toml:"b"`
	}

	cmd := &cobra.Command{Use: "app"}
	err := RegisterCommandFlags(cmd, &cfg{}, DefaultTOMLOptions())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a pointer field")
}

func TestExcludedWithoutLeaf(t *testing.T) {
	type cfg struct {
		URL    string `toml:"url"`
		RealDB bool   `toml:"real-db" validate:"excluded_without=URL"`
	}

	t.Run("set without its dependency", func(t *testing.T) {
		var c cfg
		require.ErrorContains(t, run(t, &c, "--real-db"), "'excluded_without'")
	})

	t.Run("set with its dependency", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--url", "postgres://x", "--real-db"))
	})
}

func TestRequiredWithLeaf(t *testing.T) {
	type cfg struct {
		Enable bool   `toml:"enable"`
		Token  string `toml:"token" validate:"required_with=Enable"`
	}

	t.Run("dependency set, field missing", func(t *testing.T) {
		var c cfg
		require.ErrorContains(t, run(t, &c, "--enable"), "'required_with'")
	})

	t.Run("dependency unset", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c))
	})
}

// --- precedence: flag > env > config file > compiled-in default ---

type precedenceCfg struct {
	Value string `toml:"value"`
}

func newPrecedenceCfg() *precedenceCfg { return &precedenceCfg{Value: "from-default"} }

func TestPrecedence_DefaultWhenNothingSet(t *testing.T) {
	c := newPrecedenceCfg()
	require.NoError(t, run(t, c))
	assert.Equal(t, "from-default", c.Value)
}

func TestPrecedence_ConfigFileBeatsDefault(t *testing.T) {
	path := writeConfig(t, "value = 'from-file'\n")

	c := newPrecedenceCfg()
	require.NoError(t, run(t, c, "--config", path))
	assert.Equal(t, "from-file", c.Value)
}

func TestPrecedence_EnvBeatsConfigFile(t *testing.T) {
	path := writeConfig(t, "value = 'from-file'\n")
	t.Setenv("TEST_VALUE", "from-env")

	c := newPrecedenceCfg()
	require.NoError(t, run(t, c, "--config", path))
	assert.Equal(t, "from-env", c.Value)
}

func TestPrecedence_FlagBeatsEnv(t *testing.T) {
	path := writeConfig(t, "value = 'from-file'\n")
	t.Setenv("TEST_VALUE", "from-env")

	c := newPrecedenceCfg()
	require.NoError(t, run(t, c, "--config", path, "--value", "from-flag"))
	assert.Equal(t, "from-flag", c.Value)
}

func TestEnvPrefixesTriedInOrder(t *testing.T) {
	type cfg struct {
		Value string `toml:"value"`
	}

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("FIRST", "SECOND")))

	// Both are bound; the earlier prefix wins.
	t.Setenv("FIRST_VALUE", "first")
	t.Setenv("SECOND_VALUE", "second")

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	assert.Equal(t, "first", c.Value)
}

// --- subcommands ---

// newSubcommand registers rootTarget on a root and subTarget under namespace on a child
// command, returning both so a test can execute the child.
func newSubcommand(t *testing.T, rootTarget, subTarget any, namespace string) *cobra.Command {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, rootTarget, DefaultTOMLOptions("TEST")))

	sub := &cobra.Command{Use: namespace, RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, RegisterSubcommandFlags(sub, namespace, subTarget, DefaultTOMLOptions()))
	root.AddCommand(sub)
	return root
}

func TestSubcommand_FlagIsNotNamespaced(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host"`
	}
	type subCfg struct {
		Retries int `toml:"retries"`
	}

	var r rootCfg
	var s subCfg
	root := newSubcommand(t, &r, &s, "sub")

	// The viper key is "sub.retries", but the flag stays --retries.
	root.SetArgs([]string{"sub", "--retries", "9"})
	require.NoError(t, root.Execute())
	assert.Equal(t, 9, s.Retries)
}

func TestSubcommand_EnvIsNamespaced(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host"`
	}
	type subCfg struct {
		Retries int `toml:"retries"`
	}

	var r rootCfg
	var s subCfg
	root := newSubcommand(t, &r, &s, "sub")

	// Namespace appears in the env var (and the root's prefix is inherited).
	t.Setenv("TEST_SUB_RETRIES", "11")

	root.SetArgs([]string{"sub"})
	require.NoError(t, root.Execute())
	assert.Equal(t, 11, s.Retries)
}

func TestSubcommand_ConfigFileUsesNamespacedTable(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host"`
	}
	type subCfg struct {
		Retries int `toml:"retries"`
	}

	var r rootCfg
	var s subCfg
	root := newSubcommand(t, &r, &s, "sub")
	path := writeConfig(t, "host = 'example.com'\n\n[sub]\nretries = 4\n")

	root.SetArgs([]string{"sub", "--config", path})
	require.NoError(t, root.Execute())
	assert.Equal(t, 4, s.Retries)
	assert.Equal(t, "example.com", r.Host, "root config decodes too when a subcommand runs")
}

func TestSubcommand_RootValidationStillApplies(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host" validate:"required"`
	}
	type subCfg struct {
		Retries int `toml:"retries"`
	}

	var r rootCfg
	var s subCfg
	root := newSubcommand(t, &r, &s, "sub")

	root.SetArgs([]string{"sub"})
	require.ErrorContains(t, root.Execute(), "'required'")
}

// --- hook chaining ---

func TestCallersPreRunESeesDecodedConfig(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	root := newRoot(t)
	var c cfg
	var seen string
	root.PreRunE = func(*cobra.Command, []string) error {
		seen = c.Host // must already be populated
		return nil
	}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))

	root.SetArgs([]string{"--host", "example.com"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", seen)
}

func TestCallersPersistentPreRunEStillRuns(t *testing.T) {
	type cfg struct {
		Host string `toml:"host"`
	}

	root := newRoot(t)
	var c cfg
	called := false
	root.PersistentPreRunE = func(*cobra.Command, []string) error {
		called = true
		return nil
	}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	assert.True(t, called)
}

// --- several independent targets on one command ---

func TestMultipleTargetsDecodeIndependently(t *testing.T) {
	type dbCfg struct {
		URL string `toml:"db-url" validate:"required"`
	}
	type evmCfg struct {
		ChainID string `toml:"evm-chain-id" validate:"required"`
	}

	root := newRoot(t)
	var db dbCfg
	var evm evmCfg
	require.NoError(t, RegisterCommandFlags(root, &db, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterCommandFlags(root, &evm, DefaultTOMLOptions("TEST")))

	root.SetArgs([]string{"--db-url", "postgres://x", "--evm-chain-id", "1"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "postgres://x", db.URL)
	assert.Equal(t, "1", evm.ChainID)
}

func TestMultipleTargetsBothReportTheirOwnErrors(t *testing.T) {
	type dbCfg struct {
		URL string `toml:"db-url" validate:"required"`
	}
	type evmCfg struct {
		ChainID string `toml:"evm-chain-id" validate:"required"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &dbCfg{}, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterCommandFlags(root, &evmCfg{}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'URL'")
	assert.Contains(t, err.Error(), "'ChainID'", "one target's failure must not hide the other's")
}

// --- cross-field rules naming a nested field ---

// validator resolves a rule's parameter from the struct carrying the rule, so a parameter may
// reach *down* into a nested struct ("Mid.Deep.Bar") but never up out of its own struct. These
// cover the reaching-down form end to end: flag, env, docs and example config.

type deepCfg struct {
	Bar string `usage:"the deep setting"`
}

type midCfg struct {
	Deep deepCfg
}

type nestedRuleCfg struct {
	Mid midCfg
	Val string `usage:"needed alongside the deep setting" validate:"required_with=Mid.Deep.Bar"`
}

func TestNestedFieldRuleFiresFromFlag(t *testing.T) {
	var c nestedRuleCfg
	require.ErrorContains(t, run(t, &c, "--mid.deep.bar", "x"), "'required_with'")
}

func TestNestedFieldRuleFiresFromEnv(t *testing.T) {
	t.Setenv("TEST_MID_DEEP_BAR", "x")

	var c nestedRuleCfg
	require.ErrorContains(t, run(t, &c), "'required_with'")
}

func TestNestedFieldRuleSatisfied(t *testing.T) {
	var c nestedRuleCfg
	require.NoError(t, run(t, &c, "--mid.deep.bar", "x", "--val", "y"))
	assert.Equal(t, "x", c.Mid.Deep.Bar)
	assert.Equal(t, "y", c.Val)

	// Trigger absent, so the rule doesn't fire.
	var d nestedRuleCfg
	require.NoError(t, run(t, &d))
}

func TestNestedFieldRuleUnderNamespace(t *testing.T) {
	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "app"

	root := newRoot(t)
	var c nestedRuleCfg
	require.NoError(t, RegisterCommandFlags(root, &c, opts))

	root.SetArgs([]string{"--app.mid.deep.bar", "x"})
	require.ErrorContains(t, root.Execute(), "'required_with'")
}

// The section form: the rule sits on the outer pointer field, since a rule inside Foo could not
// name Baz. Whether Foo's own fields are then required is Foo's business.
type sectionRuleCfg struct {
	Baz deepCfg
	Foo *struct {
		Name string `usage:"foo's name" validate:"required"`
	} `usage:"the foo section" validate:"required_with=Baz.Bar"`
}

func TestNestedFieldRuleOnPointerSection(t *testing.T) {
	t.Run("section missing", func(t *testing.T) {
		var c sectionRuleCfg
		require.ErrorContains(t, run(t, &c, "--baz.bar", "x"), "'required_with'")
	})

	t.Run("section present but incomplete", func(t *testing.T) {
		var c sectionRuleCfg
		// Naming any of the section's keys allocates it, and then its own `required` applies.
		require.ErrorContains(t, run(t, &c, "--baz.bar", "x", "--foo.name", ""), "'required'")
	})

	t.Run("section complete", func(t *testing.T) {
		var c sectionRuleCfg
		require.NoError(t, run(t, &c, "--baz.bar", "x", "--foo.name", "n"))
		require.NotNil(t, c.Foo)
		assert.Equal(t, "n", c.Foo.Name)
	})
}

func TestNestedFieldRuleIsDocumentedByItsKey(t *testing.T) {
	toml, err := structToDocs(&nestedRuleCfg{}, "", DefaultTOMLOptions("TEST"))
	require.NoError(t, err)
	// The dotted path is reported as the config key it resolves to, not as Go field names.
	assert.Contains(t, toml, "required when mid.deep.bar is set")
	assert.NotContains(t, toml, "Mid.Deep.Bar")
}

func TestNestedFieldExclusiveRuleIsResolvedInTheExample(t *testing.T) {
	type cfg struct {
		Mid midCfg
		Val string `usage:"the alternative to the deep setting" validate:"excluded_with=Mid.Deep.Bar"`
	}

	opts := DefaultTOMLOptions("TEST")

	example, err := exampleDoc(&cfg{Mid: midCfg{Deep: deepCfg{Bar: "shown"}}}, "", opts)
	require.NoError(t, err)
	assert.Contains(t, example, "bar = ")
	assert.NotContains(t, example, "val = ", "the example must show one of the two, not both")

	// Still documented, with the rule spelled out.
	toml, err := structToDocs(&cfg{}, "", opts)
	require.NoError(t, err)
	assert.Contains(t, toml, "must not be set when mid.deep.bar is set")
}

// --- profiles ---

type profileChain struct {
	ID  uint32 `toml:"id"`
	RPC string `toml:"rpc"`
}

type profileCfg struct {
	Chain profileChain `toml:"chain"`
}

// runWithProfile registers c plus a profile map keyed on Chain.ID, then executes with args.
func runWithProfile(t *testing.T, c *profileCfg, args ...string) error {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "Chain.ID", map[uint32]profileCfg{
		1:   {Chain: profileChain{RPC: "https://one"}},
		137: {Chain: profileChain{RPC: "https://one-thirty-seven"}},
	}, DefaultTOMLOptions("TEST")))
	root.SetArgs(args)
	return root.Execute()
}

func TestProfileFillsDefaultsForSelectedKey(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c))
	assert.Equal(t, "https://one", c.Chain.RPC)
}

func TestProfileFollowsSelector(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.id", "137"))
	assert.Equal(t, uint32(137), c.Chain.ID)
	assert.Equal(t, "https://one-thirty-seven", c.Chain.RPC)
}

func TestProfileDoesNotOverrideExplicitValue(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.rpc", "https://mine"))
	assert.Equal(t, "https://mine", c.Chain.RPC)
}

func TestProfileUnknownSelectorAppliesNothing(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.id", "999"))
	assert.Empty(t, c.Chain.RPC, "no matching profile means no defaults, not an error")
}

func TestProfileRequiresARegisteredTargetOfItsType(t *testing.T) {
	root := newRoot(t)
	err := RegisterProfile(root, "Chain.ID", map[uint32]profileCfg{1: {}}, DefaultTOMLOptions())
	require.ErrorContains(t, err, "no registered target")
}

// --- slices ---

func TestStringSliceFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		URLs []string `toml:"url"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--url", "a", "--url", "b"))
	assert.Equal(t, []string{"a", "b"}, c.URLs)
}

func TestStringSliceFromCommaSeparatedEnv(t *testing.T) {
	type cfg struct {
		URLs []string `toml:"url"`
	}

	t.Setenv("TEST_URL", "a,b")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, []string{"a", "b"}, c.URLs)
}

// A list of a non-string primitive splits on commas the way a []string does - mapstructure's
// own StringToSliceHookFunc only ever splits into []string, so this needs the wider hook that
// map-of-list values needed anyway.
func TestNonStringSliceFromCommaSeparatedFlag(t *testing.T) {
	type cfg struct {
		Ports []int `toml:"ports"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--ports", "1,2,3"))
	assert.Equal(t, []int{1, 2, 3}, c.Ports)
}

// --- maps ---

func TestStringMapFromFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=prod;region=us"))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=prod", "--labels", "region=us"))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromEnv(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	t.Setenv("TEST_LABELS", "env=prod;region=us")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

func TestStringMapFromConfigFile(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	path := writeConfig(t, "[labels]\nenv = 'prod'\nregion = 'us'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path))
	assert.Equal(t, map[string]string{"env": "prod", "region": "us"}, c.Labels)
}

// A map's values are converted to the map's own value type, not left as the strings a flag or
// env var delivers them as.
func TestMapOfPrimitiveValues(t *testing.T) {
	type cfg struct {
		Counts  map[string]int     `toml:"counts"`
		Enabled map[string]bool    `toml:"enabled"`
		Weights map[string]float64 `toml:"weights"`
	}

	var c cfg
	require.NoError(t, run(t, &c,
		"--counts", "a=3;b=4",
		"--enabled", "a=true;b=false",
		"--weights", "a=1.5",
	))
	assert.Equal(t, map[string]int{"a": 3, "b": 4}, c.Counts)
	assert.Equal(t, map[string]bool{"a": true, "b": false}, c.Enabled)
	assert.Equal(t, map[string]float64{"a": 1.5}, c.Weights)
}

// Non-string keys are converted the same way the values are.
func TestMapWithPrimitiveKeys(t *testing.T) {
	type cfg struct {
		Chains map[uint32]string `toml:"chains"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--chains", "1=mainnet;11155111=sepolia"))
	assert.Equal(t, map[uint32]string{1: "mainnet", 11155111: "sepolia"}, c.Chains)
}

// A value that unmarshals itself from text goes through its own UnmarshalText, from every
// source - so config.Duration keeps parsing "45s" and keeps rejecting a negative duration.
func TestMapOfTextUnmarshalerValues(t *testing.T) {
	type cfg struct {
		Timeouts map[string]config.Duration `toml:"timeouts"`
	}

	t.Run("flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--timeouts", "read=45s;write=1m30s"))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
		assert.Equal(t, 90*time.Second, c.Timeouts["write"].Duration())
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TEST_TIMEOUTS", "read=45s")

		var c cfg
		require.NoError(t, run(t, &c))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
	})

	t.Run("config file", func(t *testing.T) {
		path := writeConfig(t, "[timeouts]\nread = '45s'\n")

		var c cfg
		require.NoError(t, run(t, &c, "--config", path))
		assert.Equal(t, 45*time.Second, c.Timeouts["read"].Duration())
	})

	t.Run("invalid", func(t *testing.T) {
		var c cfg
		require.ErrorContains(t, run(t, &c, "--timeouts", "read=-45s"), "cannot make negative time duration")
	})
}

func TestMapOfDurationValues(t *testing.T) {
	type cfg struct {
		Timeouts map[string]time.Duration `toml:"timeouts"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--timeouts", "read=45s"))
	assert.Equal(t, 45*time.Second, c.Timeouts["read"])
}

// A map left alone keeps whatever the caller's struct was constructed with, and its compiled-in
// entries show up as the flag's default rather than as a Go-printed "map[...]".
func TestMapDefaultUntouched(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	c := cfg{Labels: map[string]string{"env": "dev"}}
	require.NoError(t, run(t, &c))
	assert.Equal(t, map[string]string{"env": "dev"}, c.Labels)

	root := newRoot(t)
	c2 := cfg{Labels: map[string]string{"env": "dev"}}
	require.NoError(t, RegisterCommandFlags(root, &c2, DefaultTOMLOptions("TEST")))
	f := root.PersistentFlags().Lookup("labels")
	require.NotNil(t, f)
	assert.Equal(t, "key=value;...", f.Value.Type())
	assert.Equal(t, "[env=dev]", f.DefValue)
}

// A flag overrides the config file per key, not per map: setting --labels replaces the table.
func TestMapFlagOverridesConfigFile(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	path := writeConfig(t, "[labels]\nenv = 'prod'\nregion = 'us'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path, "--labels", "env=staging"))
	assert.Equal(t, map[string]string{"env": "staging"}, c.Labels)
}

// A map whose values can't be written as a single string has no flag form, so no flag is
// registered for it - one bound anyway would default to the string "map[]" and fail to decode.
// The config file still fills it in.
func TestMapOfStructsIsConfigFileOnly(t *testing.T) {
	type chain struct {
		RPC     string `toml:"rpc"`
		Enabled bool   `toml:"enabled"`
	}
	type cfg struct {
		Chains map[string]chain `toml:"chains"`
	}

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Nil(t, root.PersistentFlags().Lookup("chains"))

	path := writeConfig(t, "[chains.eth]\nrpc = 'http://x'\nenabled = true\n")
	require.NoError(t, run(t, &c, "--config", path))
	assert.Equal(t, map[string]chain{"eth": {RPC: "http://x", Enabled: true}}, c.Chains)
}

// A list value's elements are separated by a comma, which is why entries are separated by a
// semicolon rather than by the comma pflag's own StringToString uses.
func TestMapOfLists(t *testing.T) {
	type cfg struct {
		URLs map[string][]string `toml:"urls"`
	}

	want := map[string][]string{"primary": {"a", "b"}, "backup": {"c"}}

	t.Run("flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=a,b;backup=c"))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("repeated flag", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=a,b", "--urls", "backup=c"))
		assert.Equal(t, want, c.URLs)
	})

	// AWS CLI shorthand brackets its list values; accepted so a list reads as one.
	t.Run("bracketed", func(t *testing.T) {
		var c cfg
		require.NoError(t, run(t, &c, "--urls", "primary=[a,b];backup=[c]"))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TEST_URLS", "primary=a,b;backup=c")

		var c cfg
		require.NoError(t, run(t, &c))
		assert.Equal(t, want, c.URLs)
	})

	t.Run("config file", func(t *testing.T) {
		path := writeConfig(t, "[urls]\nprimary = ['a', 'b']\nbackup = ['c']\n")

		var c cfg
		require.NoError(t, run(t, &c, "--config", path))
		assert.Equal(t, want, c.URLs)
	})
}

// A list of a non-string primitive converts element by element, the same as a scalar value does.
func TestMapOfPrimitiveLists(t *testing.T) {
	type cfg struct {
		Ports    map[string][]int             `toml:"ports"`
		Timeouts map[string][]config.Duration `toml:"timeouts"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--ports", "a=1,2;b=3", "--timeouts", "a=1s,2s"))
	assert.Equal(t, map[string][]int{"a": {1, 2}, "b": {3}}, c.Ports)
	require.Len(t, c.Timeouts["a"], 2)
	assert.Equal(t, time.Second, c.Timeouts["a"][0].Duration())
	assert.Equal(t, 2*time.Second, c.Timeouts["a"][1].Duration())
}

func TestMapOfListsDefault(t *testing.T) {
	type cfg struct {
		URLs map[string][]string `toml:"urls"`
	}

	root := newRoot(t)
	c := cfg{URLs: map[string][]string{"primary": {"a", "b"}}}
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	f := root.PersistentFlags().Lookup("urls")
	require.NotNil(t, f)
	assert.Equal(t, "key=v1,v2;...", f.Value.Type())
	assert.Equal(t, "[primary=a,b]", f.DefValue)
}

// A map of []byte is text, not a list of numbers, so it gets no flag rather than one that
// decodes its value into bytes one comma-separated element at a time.
func TestMapOfBytesIsConfigFileOnly(t *testing.T) {
	type cfg struct {
		Keys map[string][]byte `toml:"keys"`
	}

	root := newRoot(t)
	var c cfg
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Nil(t, root.PersistentFlags().Lookup("keys"))
}

// The comma-separated form pflag's StringToString and kubectl use is the likeliest mistake
// here, and it would otherwise be stored whole as one value.
func TestMapFlagRejectsCommaSeparatedPairs(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.ErrorContains(t, run(t, &c, "--labels", "env=prod,region=us"), `separate map entries with ";"`)
}

// Quoting is the way out for a scalar value that really does contain a comma.
func TestMapFlagQuotedValueKeepsComma(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", `env="a,b=c";region=us`))
	assert.Equal(t, map[string]string{"env": "a,b=c", "region": "us"}, c.Labels)
}

// A quoted value survives the round trip out through the flag's own text and back, which is
// the trip viper makes for every flag type it doesn't parse itself.
func TestMapFlagQuotedValueKeepsSeparator(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", `q="a;b";region=us`))
	assert.Equal(t, map[string]string{"q": "a;b", "region": "us"}, c.Labels)
}

// A comma in a scalar value is only rejected when it looks like another pair.
func TestMapFlagAllowsPlainCommaValue(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--labels", "env=a,b"))
	assert.Equal(t, map[string]string{"env": "a,b"}, c.Labels)
}

func TestMapFlagRejectsMalformedPair(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	var c cfg
	require.Error(t, run(t, &c, "--labels", "env"))
}

func TestMapEnvRejectsMalformedPair(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	t.Setenv("TEST_LABELS", "env")

	var c cfg
	require.ErrorContains(t, run(t, &c), "key=value")
}

func TestNamespacedMapFlag(t *testing.T) {
	type cfg struct {
		Labels map[string]string `toml:"labels"`
	}

	root := newRoot(t)
	var c cfg
	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "db"
	require.NoError(t, RegisterCommandFlags(root, &c, opts))
	root.SetArgs([]string{"--db.labels", "env=prod"})
	require.NoError(t, root.Execute())
	assert.Equal(t, map[string]string{"env": "prod"}, c.Labels)
}

// --- config.Duration ---

func TestConfigDurationDecodes(t *testing.T) {
	type cfg struct {
		Timeout config.Duration `toml:"timeout"`
	}

	c := cfg{Timeout: *config.MustNewDuration(5 * time.Second)}
	require.NoError(t, run(t, &c, "--timeout", "45s"))
	assert.Equal(t, 45*time.Second, c.Timeout.Duration())
}

func TestConfigDurationRejectsNegative(t *testing.T) {
	type cfg struct {
		Timeout config.Duration `toml:"timeout"`
	}

	// pflag accepts -5s as a duration; config.Duration is what refuses it.
	t.Setenv("TEST_TIMEOUT", "-5s")

	var c cfg
	require.ErrorContains(t, run(t, &c), "negative")
}

// --- string coercion applies to every Options, not just DefaultTOMLOptions ---

// upperString is a minimal encoding.TextUnmarshaler, standing in for any caller-defined type
// with its own text decoding - distinct from config.Duration, which the rest of the suite
// already covers.
type upperString string

func (u *upperString) UnmarshalText(text []byte) error {
	*u = upperString(strings.ToUpper(string(text)))
	return nil
}

// coercionCfg exercises every field kind bindLeafFlag special-cases: plain string, an int and a
// uint kind, bool, time.Duration, config.Duration (a TextUnmarshaler struct), a string slice, and
// a caller-defined TextUnmarshaler.
type coercionCfg struct {
	Str     string
	Count   int
	Limit   uint
	Enabled bool
	Timeout time.Duration
	Retry   config.Duration
	Tags    []string
	Label   upperString
}

func assertCoercionCfg(t *testing.T, c *coercionCfg) {
	t.Helper()
	assert.Equal(t, "hello", c.Str)
	assert.Equal(t, 42, c.Count)
	assert.Equal(t, uint(7), c.Limit)
	assert.True(t, c.Enabled)
	assert.Equal(t, 45*time.Second, c.Timeout)
	assert.Equal(t, 5*time.Second, c.Retry.Duration())
	assert.Equal(t, []string{"a", "b"}, c.Tags)
	assert.Equal(t, upperString("SHOUT"), c.Label)
}

func TestCoercion_EveryFieldKindFromEnv_WithBareOptions(t *testing.T) {
	t.Setenv("TEST_STR", "hello")
	t.Setenv("TEST_COUNT", "42")
	t.Setenv("TEST_LIMIT", "7")
	t.Setenv("TEST_ENABLED", "true")
	t.Setenv("TEST_TIMEOUT", "45s")
	t.Setenv("TEST_RETRY", "5s")
	t.Setenv("TEST_TAGS", "a,b")
	t.Setenv("TEST_LABEL", "shout")

	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}}))
	assertCoercionCfg(t, &c)
}

func TestCoercion_EveryFieldKindFromFlag_WithBareOptions(t *testing.T) {
	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}},
		"--str", "hello",
		"--count", "42",
		"--limit", "7",
		"--enabled",
		"--timeout", "45s",
		"--retry", "5s",
		"--tags", "a",
		"--tags", "b",
		"--label", "shout",
	))
	assertCoercionCfg(t, &c)
}

func TestCoercion_EveryFieldKindFromConfigFile_WithBareOptions(t *testing.T) {
	path := writeConfig(t, `
str = "hello"
count = 42
limit = 7
enabled = true
timeout = "45s"
retry = "5s"
tags = ["a", "b"]
label = "shout"
`)

	var c coercionCfg
	require.NoError(t, runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}}, "--config", path))
	assertCoercionCfg(t, &c)
}

func TestCoercion_InvalidValueStillErrors_WithBareOptions(t *testing.T) {
	t.Setenv("TEST_COUNT", "not-a-number")

	var c coercionCfg
	err := runWithOptions(t, &c, Options{Prefixes: []string{"TEST"}})
	require.ErrorContains(t, err, "Count")
}

// --- generated docs ---

func TestDocsMarksDefaultsAndRequiredFields(t *testing.T) {
	type cfg struct {
		Host    string `toml:"host" usage:"the host" validate:"required" example:"'example.com'"`
		Retries int    `toml:"retries" usage:"how many times to retry"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{Retries: 3}, DefaultTOMLOptions("TEST")))

	doc, err := GenerateDocs(root)
	require.NoError(t, err)
	// A required field has no real default, so it's documented by example.
	assert.Contains(t, doc, "host = 'example.com' # Example")
	assert.Contains(t, doc, "retries = 3 # Default")
}

func TestDocsPutsNestedStructUsageInItsTableHeader(t *testing.T) {
	type chain struct {
		Host string `toml:"host" usage:"the host"`
	}
	type cfg struct {
		Chain chain `toml:"chain" usage:"which chain to talk to"`
	}

	toml, err := structToDocs(&cfg{}, "", DefaultTOMLOptions())
	require.NoError(t, err)

	// The struct's description sits in its table's header, above the fields.
	assert.Contains(t, toml, "# which chain to talk to\n[chain]\n")
}

func TestDocsFoldsSquashedStructUsageIntoEnclosingHeader(t *testing.T) {
	type shared struct {
		Host string `toml:"host" usage:"the host"`
	}
	type chain struct {
		Shared shared `toml:",inline" usage:"settings shared by every chain"`
		Name   string `toml:"name" usage:"chain name"`
	}
	type cfg struct {
		Chain chain `toml:"chain" usage:"which chain to talk to"`
	}

	toml, err := structToDocs(&cfg{}, "", DefaultTOMLOptions())
	require.NoError(t, err)

	// A squashed struct gets no table of its own, so its description joins the header of the
	// table it was flattened into, after that table's own description.
	assert.Contains(t, toml, "# which chain to talk to\n# settings shared by every chain\n[chain]\n")
	// Its fields still belong to that table, un-prefixed.
	assert.Contains(t, toml, "host = ")
	assert.NotContains(t, toml, "[chain.shared]")
}

func TestDocsKeepsTopLevelSquashedStructUsage(t *testing.T) {
	type shared struct {
		Host string `toml:"host" usage:"the host"`
	}
	type cfg struct {
		Shared shared `toml:",inline" usage:"settings shared by everything"`
	}

	toml, err := structToDocs(&cfg{}, "", DefaultTOMLOptions())
	require.NoError(t, err)

	// No enclosing table at the top level, so it survives as standalone prose: a comment
	// block ended by a blank line, which is what keeps it off the next field's description.
	assert.True(t, strings.HasPrefix(toml, "# settings shared by everything\n\n"), toml)

	// And it still renders, rather than tripping configdoc's description checks.
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))
	doc, err := GenerateDocs(root)
	require.NoError(t, err)
	assert.Contains(t, doc, "settings shared by everything")
}

func TestNamespaceRootsKeysFlagsAndEnv(t *testing.T) {
	type cfg struct {
		URL string `toml:"url"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "database"

	t.Run("flag", func(t *testing.T) {
		root := newRoot(t)
		var c cfg
		require.NoError(t, RegisterCommandFlags(root, &c, opts))

		root.SetArgs([]string{"--database.url", "postgres://x"})
		require.NoError(t, root.Execute())
		assert.Equal(t, "postgres://x", c.URL)
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TEST_DATABASE_URL", "postgres://env")

		root := newRoot(t)
		var c cfg
		require.NoError(t, RegisterCommandFlags(root, &c, opts))

		root.SetArgs(nil)
		require.NoError(t, root.Execute())
		assert.Equal(t, "postgres://env", c.URL)
	})

	t.Run("config file", func(t *testing.T) {
		path := writeConfig(t, "[database]\nurl = 'postgres://file'\n")

		root := newRoot(t)
		var c cfg
		require.NoError(t, RegisterCommandFlags(root, &c, opts))

		root.SetArgs([]string{"--config", path})
		require.NoError(t, root.Execute())
		assert.Equal(t, "postgres://file", c.URL)
	})
}

func TestNamespaceSeparatesSameNamedFields(t *testing.T) {
	type dbCfg struct {
		URL string `toml:"url"`
	}
	type evmCfg struct {
		URL string `toml:"url"`
	}

	dbOpts := DefaultTOMLOptions("TEST")
	dbOpts.Namespace = "database"
	evmOpts := DefaultTOMLOptions("TEST")
	evmOpts.Namespace = "evm"

	root := newRoot(t)
	var db dbCfg
	var evm evmCfg
	require.NoError(t, RegisterCommandFlags(root, &db, dbOpts))
	// Same field name, different namespace: no flag collision, and each keeps its own value.
	require.NoError(t, RegisterCommandFlags(root, &evm, evmOpts))

	root.SetArgs([]string{"--database.url", "postgres://x", "--evm.url", "https://y"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "postgres://x", db.URL)
	assert.Equal(t, "https://y", evm.URL)
}

func TestNamespaceGroupsDocsUnderOneTable(t *testing.T) {
	type cfg struct {
		URL string `toml:"url" usage:"database url"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.Namespace = "database"

	toml, err := structToDocs(&cfg{}, opts.Namespace, opts)
	require.NoError(t, err)
	assert.Contains(t, toml, "[database]")
	assert.Contains(t, toml, "url = ")
}

func TestFlagDocsNoExampleOmitsFromExampleButKeepsDocs(t *testing.T) {
	type cfg struct {
		URL    string `toml:"url" usage:"database url"`
		RealDB bool   `toml:"real-db" usage:"use a real database in fake mode" flagdocs:"noexample"`
	}

	opts := DefaultTOMLOptions("TEST")

	example, err := exampleDoc(&cfg{}, "", opts)
	require.NoError(t, err)
	assert.NotContains(t, example, "real-db", "the example must not carry it")
	assert.Contains(t, example, "url = ")

	docs, err := structToDocs(&cfg{}, "", opts)
	require.NoError(t, err)
	// Still documented, but marked so it is kept out of its table's code block too - every
	// example in the document then describes the same working configuration.
	assert.Contains(t, docs, "real-db = false "+configdoc.FieldDocsOnly)
}

func TestDocsOnlyFieldIsAbsentFromItsTableCodeBlock(t *testing.T) {
	type inner struct {
		URL    string `toml:"url" usage:"database url"`
		RealDB bool   `toml:"real-db" usage:"use a real database in fake mode" flagdocs:"noexample"`
	}
	type cfg struct {
		Database inner `toml:"database"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

	doc, err := GenerateDocs(root)
	require.NoError(t, err)

	// The table's own code block lists only what an example would contain...
	section := doc[strings.Index(doc, "## database"):]
	block := section[:strings.Index(section, "###")]
	assert.Contains(t, block, "url = ")
	assert.NotContains(t, block, "real-db")

	// ...while the field still gets its own entry further down.
	assert.Contains(t, doc, "### real-db")
}

func TestFlagDocsExampleOverridesTheValueShown(t *testing.T) {
	type cfg struct {
		Workers int `toml:"workers" usage:"how many workers" flagdocs:"example=8"`
	}

	opts := DefaultTOMLOptions("TEST")

	example, err := exampleDoc(&cfg{Workers: 1}, "", opts)
	require.NoError(t, err)
	assert.Contains(t, example, "workers = 8", "the example shows the override")

	docs, err := structToDocs(&cfg{Workers: 1}, "", opts)
	require.NoError(t, err)
	assert.Contains(t, docs, "workers = 1 # Default", "the docs still show the real default")
}

func TestExampleShowsOnlyTheFirstOfMutuallyExclusiveFields(t *testing.T) {
	type cfg struct {
		Listen []string `toml:"listen-addresses" usage:"listen here" validate:"required_without=Proxy,excluded_with=Proxy"`
		Proxy  string   `toml:"proxy-address" usage:"or proxy there" validate:"excluded_with=Listen"`
	}

	opts := DefaultTOMLOptions("TEST")

	example, err := exampleDoc(&cfg{}, "", opts)
	require.NoError(t, err)
	// Listing both would be a config that fails its own validation, so the example commits
	// to the first.
	assert.Contains(t, example, "listen-addresses = ")
	assert.NotContains(t, example, "proxy-address = ")

	// Both are still documented; only the example has to choose.
	docs, err := structToDocs(&cfg{}, "", opts)
	require.NoError(t, err)
	assert.Contains(t, docs, "listen-addresses = ")
	assert.Contains(t, docs, "proxy-address = ")
}

func TestExampleShowsOnlyTheFirstOfMutuallyExclusiveSections(t *testing.T) {
	var c modesCfg

	example, err := exampleDoc(&c, "", DefaultTOMLOptions("TEST"))
	require.NoError(t, err)
	assert.Contains(t, example, "[b]")
	assert.NotContains(t, example, "[c]", "the whole excluded section is dropped, not just its fields")
}

func TestDocsTreatsExampleTagAsHavingNoDefault(t *testing.T) {
	type cfg struct {
		// Required by a rule this struct can't express, so it says so with an example.
		URL string `toml:"url" usage:"database url" example:"'postgres://localhost/db'"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

	doc, err := GenerateDocs(root)
	require.NoError(t, err)
	assert.Contains(t, doc, "url = 'postgres://localhost/db' # Example")
}

func TestDocsExplainsCrossFieldRulesUsingConfigKeys(t *testing.T) {
	type cfg struct {
		URL    string `toml:"database-url" usage:"database url"`
		RealDB bool   `toml:"real-db" usage:"use a real db" validate:"excluded_without=URL"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

	doc, err := GenerateDocs(root)
	require.NoError(t, err)
	// Named by config key ("database-url"), not Go field name ("URL").
	assert.Contains(t, doc, "must not be set unless database-url is set")
}

func TestDocsCoversSubcommands(t *testing.T) {
	type rootCfg struct {
		Host string `toml:"host" usage:"the host"`
	}
	type subCfg struct {
		Retries int `toml:"retries" usage:"how many times to retry"`
	}

	root := newSubcommand(t, &rootCfg{}, &subCfg{Retries: 3}, "sub")

	doc, err := GenerateDocs(root)
	require.NoError(t, err)
	assert.Contains(t, doc, "Global Configuration")
	assert.Contains(t, doc, "Command: app sub")
	assert.Contains(t, doc, "retries = 3 # Default", "subcommand fields are documented too")
}

// spyFormat is configdoc.TOML with the value syntax changed, to prove the document really is
// assembled and rendered through Options.Format rather than hard-coded TOML.
type spyFormat struct {
	configdoc.TOML
}

func (f spyFormat) Field(key, value, marker string) string {
	if marker == "" {
		return key + ": " + value
	}
	return key + ": " + value + " " + marker
}

// A format's two sides have to agree: whatever Field writes, ParseLine has to read back.
func (f spyFormat) ParseLine(line string) configdoc.Line {
	l := f.TOML.ParseLine(line)
	if l.Kind == configdoc.LineField {
		l.Text = strings.TrimSuffix(l.Text, ":")
	}
	return l
}

func TestDocsAssemblesWithFormatFromOptions(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" usage:"the host"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.Format = spyFormat{}

	toml, err := structToDocs(&cfg{Host: "example.com"}, "", opts)
	require.NoError(t, err)
	// The format's Field syntax is used, not TOML's "key = value".
	assert.Contains(t, toml, "host: 'example.com' # Default")
}

func TestDocsCommandUsesFormatFromOptions(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" usage:"the host"`
	}

	opts := DefaultTOMLOptions("TEST")
	opts.Format = spyFormat{}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, opts))

	dir := t.TempDir()
	t.Chdir(dir)

	// The format must reach the auto-added docs command too, not just GenerateDocs.
	root.SetArgs([]string{"docs"})
	require.NoError(t, root.Execute())

	written, err := os.ReadFile(filepath.Join(dir, docsOutputPath))
	require.NoError(t, err)
	assert.Contains(t, string(written), "host: ")
}

func TestBuiltinCommandsSkipValidation(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" validate:"required"`
	}

	// Reading the help that explains a required setting must not require that setting. docs
	// is included because it is the same kind of command - it describes the config rather
	// than consuming it - even though it opts out by its own route.
	for _, args := range [][]string{
		{"help"},
		{"help", "docs"},
		{"docs"},
		{"completion", "bash"},
		{"__complete", ""},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Chdir(t.TempDir()) // docs writes a file

			root := newRoot(t)
			require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

			root.SetArgs(args)
			require.NoError(t, root.Execute())
		})
	}
}

func TestIsBuiltinCommandCoversCobrasOwnCommands(t *testing.T) {
	// Callers chaining their own PersistentPreRunE need this to guard checks the library
	// can't see; a chained check that misses one of these rejects `help` on a machine with no
	// configuration, which is exactly what the skip exists to prevent.
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &struct{}{}, DefaultTOMLOptions("TEST")))
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()

	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		byName[c.Name()] = c
	}

	for _, name := range []string{"help", "completion", "docs"} {
		sub, ok := byName[name]
		require.True(t, ok, "expected a %q command", name)
		if name == "docs" {
			// docs is ours: it opts out via its own PersistentPreRunE, not this check.
			assert.False(t, IsBuiltinCommand(sub))
			continue
		}
		assert.True(t, IsBuiltinCommand(sub), name)
	}

	assert.False(t, IsBuiltinCommand(root), "the root command itself runs the program")
}

func TestNonBuiltinCommandStillValidates(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" validate:"required"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.ErrorContains(t, root.Execute(), "'required'")
}

func TestDocsCommandIsAddedOnce(t *testing.T) {
	type aCfg struct {
		A string `toml:"a"`
	}
	type bCfg struct {
		B string `toml:"b"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &aCfg{}, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterCommandFlags(root, &bCfg{}, DefaultTOMLOptions("TEST")))

	var docsCmds int
	for _, c := range root.Commands() {
		if c.Name() == "docs" {
			docsCmds++
		}
	}
	assert.Equal(t, 1, docsCmds)
}

func TestDocsCommandSkipsValidation(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" validate:"required"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))

	// Writing docs must not require a valid config.
	dir := t.TempDir()
	t.Chdir(dir)

	root.SetArgs([]string{"docs"})
	require.NoError(t, root.Execute())
	assert.FileExists(t, filepath.Join(dir, docsOutputPath))
}

func TestUnsetFieldKeepsCallerDefault(t *testing.T) {
	type cfg struct {
		Retries int    `toml:"retries"`
		Host    string `toml:"host"`
	}

	c := cfg{Retries: 3, Host: "default-host"}
	require.NoError(t, run(t, &c, "--host", "override"))
	assert.Equal(t, 3, c.Retries, "untouched field should keep its compiled-in default")
	assert.Equal(t, "override", c.Host)
}
