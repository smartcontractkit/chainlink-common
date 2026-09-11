package flags

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
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

	root.SetArgs([]string{"--remote.host", "example.com", "--remote.mine", "x"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "example.com", c.Host)
	assert.Equal(t, "x", c.Mine)
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
