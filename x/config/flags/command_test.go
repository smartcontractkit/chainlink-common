package flags

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
// cover the reaching-down form end to end, from a flag and from an env var.

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

// --- namespaces ---

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

func TestBuiltinCommandsSkipValidation(t *testing.T) {
	type cfg struct {
		Host string `toml:"host" validate:"required"`
	}

	// Reading the help that explains a required setting must not require that setting.
	for _, args := range [][]string{
		{"help"},
		{"completion", "bash"},
		{"__complete", ""},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := newRoot(t)
			require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))
			// Cobra only generates help/completion for a command that has children.
			root.AddCommand(&cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }})

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
	root.AddCommand(&cobra.Command{Use: "sub"})
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()

	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		byName[c.Name()] = c
	}

	for _, name := range []string{"help", "completion"} {
		sub, ok := byName[name]
		require.True(t, ok, "expected a %q command", name)
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
