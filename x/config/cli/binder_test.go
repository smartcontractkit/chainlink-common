package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	simple "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings"
)

func TestSubcommandFlagsArePersistent(t *testing.T) {
	root := newRoot(t)
	group := &cobra.Command{Use: "sub"}
	leaf := &cobra.Command{Use: "leaf", RunE: func(*cobra.Command, []string) error { return nil }}
	var s BasicConfig
	require.NoError(t, newBinder(t, root, testOptions).Register(group, "sub", &s))
	group.AddCommand(leaf)
	root.AddCommand(group)

	root.SetArgs([]string{"sub", "leaf", "--sub.value", "v"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "v", s.Value)
}

func TestFlagsCollideInOneNamespace(t *testing.T) {
	b := newBinder(t, newRoot(t), testOptions)
	sub := &cobra.Command{Use: "sub"}
	require.NoError(t, b.Register(sub, "a", &BasicConfig{}))
	require.ErrorContains(t, b.Register(sub, "a", &BasicConfig{}), "flag --a.value is already defined on sub")
}

// Either would let one config file key reach two fields, so registration rejects them.
func TestRegisterRejectsKeysAConfigFileCantSeparate(t *testing.T) {
	type hasCaseOnlyTwins struct {
		Twin []BasicConfig
		TWIN []BasicConfig
	}

	// Each registers on its own command, so pflag has no collision to report first.
	for _, tc := range []struct {
		name     string
		register func(b *Binder, cmd *cobra.Command) error
		err      string
	}{
		{"differ only in case across structs", func(b *Binder, cmd *cobra.Command) error {
			return errors.Join(b.Register(b.root, "", &BasicConfig{}), b.Register(cmd, "", &struct{ VALUE string }{}))
		}, "Value and VALUE differ only in case"},
		{"differ only in case in one struct", func(b *Binder, cmd *cobra.Command) error {
			return b.Register(cmd, "", &hasCaseOnlyTwins{})
		}, "Twin and TWIN differ only in case"},
		{"a value, then a table", func(b *Binder, cmd *cobra.Command) error {
			return errors.Join(b.Register(b.root, "", &BasicConfig{}), b.Register(cmd, "Value", &BasicConfig{}))
		}, "Value is a value in one struct and a table in another"},
		{"a table, then a value", func(b *Binder, cmd *cobra.Command) error {
			return errors.Join(b.Register(b.root, "Value", &BasicConfig{}), b.Register(cmd, "", &BasicConfig{}))
		}, "Value is a value in one struct and a table in another"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newRoot(t)
			sub := &cobra.Command{Use: "sub"}
			root.AddCommand(sub)
			require.ErrorContains(t, tc.register(newBinder(t, root, testOptions), sub), tc.err)
		})
	}

	t.Run("the same key from two commands", func(t *testing.T) {
		root := newRoot(t)
		b := newBinder(t, root, testOptions)
		one, two := &cobra.Command{Use: "one"}, &cobra.Command{Use: "two"}
		root.AddCommand(one, two)
		require.NoError(t, b.Register(one, "database", &BasicConfig{}))
		require.NoError(t, b.Register(two, "database", &BasicConfig{}))
	})
}

func TestNewRejectsAnExistingConfigFlag(t *testing.T) {
	for name, define := range map[string]func(*cobra.Command){
		"local":      func(c *cobra.Command) { c.Flags().String(ConfigFlagName, "", "") },
		"persistent": func(c *cobra.Command) { c.PersistentFlags().StringArray(ConfigFlagName, nil, "") },
	} {
		t.Run(name, func(t *testing.T) {
			root := newRoot(t)
			define(root)
			_, err := New(root, testOptions)
			require.ErrorContains(t, err, "flag --config is already defined on app")
		})
	}
}

func TestEachCommandDecodesOnlyWhenItRuns(t *testing.T) {
	root := newRoot(t)
	b := newBinder(t, root, testOptions)
	var foo, bar BasicConfig
	fooCmd := &cobra.Command{Use: "foo", RunE: func(*cobra.Command, []string) error { return nil }}
	barCmd := &cobra.Command{Use: "bar", RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, b.Register(fooCmd, "foo", &foo))
	require.NoError(t, b.Register(barCmd, "bar", &bar))
	root.AddCommand(fooCmd, barCmd)

	t.Setenv("TEST_BAR_VALUE", "theirs")
	root.SetArgs([]string{"foo", "--foo.value", "mine"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "mine", foo.Value)
	assert.Empty(t, bar.Value)
}

func TestConfigFileIsDiscoveredInTheWorkingDirectory(t *testing.T) {
	for _, tc := range []struct{ name, file, want string }{
		{"present", "Value = 'from-cwd'\n", "from-cwd"},
		{"absent", "", "default"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.file != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config."+testOptions.Markup.Extension()), []byte(tc.file), 0o600))
			}
			t.Chdir(dir)

			c := BasicConfig{Value: "default"}
			require.NoError(t, run(t, &c, testOptions))
			assert.Equal(t, tc.want, c.Value)
		})
	}
}

func TestConfigFileErrors(t *testing.T) {
	type hasMapAndSection struct {
		Int     int
		Map     map[string]string
		Section BasicConfig
	}

	for _, tc := range []struct {
		name string
		args []string
		want []string
	}{
		{"a named file that is missing", []string{"--config", filepath.Join(t.TempDir(), "nope")}, []string{"failed to read config file"}},
		{"a malformed file", []string{"--config", writeConfig(t, "Value = \n")}, []string{"invalid config file", "line 1: unexpected character"}},
		{"unknown keys", []string{"--config", writeConfig(t, "Itn = 5\nVaule = 'x'\n[Section]\nVaule = 1\n")},
			[]string{"unknown configuration key(s)", "Itn", "Vaule", "Section.Vaule"}},
		{"a map's own keys are not config keys", []string{"--config", writeConfig(t, "[Map]\nanything = 'goes'\nAnything = 'else'\n")}, nil},
		{"a casing the decoder accepts", []string{"--config", writeConfig(t, "int = 5\n")}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := run(t, &hasMapAndSection{}, testOptions, tc.args...)
			if tc.want == nil {
				require.NoError(t, err)
			}
			for _, want := range tc.want {
				require.ErrorContains(t, err, want)
			}
		})
	}
}

func TestUnknownKeysSpanTheCommandTree(t *testing.T) {
	root := newRoot(t)
	var r BasicConfig
	b := bind(t, root, &r, testOptions)
	sub := &cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, b.Register(sub, "sub", &BasicConfig{}))
	root.AddCommand(sub)
	root.SetArgs([]string{"--config", writeConfig(t, "Value = 'r'\n[sub]\nValue = 's'\n")})
	require.NoError(t, root.Execute())
	assert.Equal(t, "r", r.Value)
}

func TestConfigFilesAreLayeredInOrder(t *testing.T) {
	type hasReplacedAndKept struct {
		Replaced string
		Kept     string
	}
	type hasEveryLayering struct {
		EveryLayer string
		FirstOnly  int
		BaseOnly   int
		List       []string
		Table      hasReplacedAndKept
		Recased    uint32
	}

	opts := testOptions
	opts.BaseConfig = []byte("EveryLayer = 'base'\nBaseOnly = 3\n")
	first := writeConfig(t, "EveryLayer = 'first'\nFirstOnly = 1\nList = ['a', 'b']\nrecased = 1\n[Table]\nReplaced = 'first'\nKept = 'first'\n")
	// A different casing must still replace the first's, not race it in a map.
	second := writeConfig(t, "EveryLayer = 'second'\nList = ['c']\nRecased = 2\n[Table]\nReplaced = 'second'\n")

	var c hasEveryLayering
	require.NoError(t, run(t, &c, opts, "--config", first, "--config", second))
	// Tables merge, but a list is one value, or a lower list could never be replaced.
	assert.Equal(t, hasEveryLayering{"second", 1, 3, []string{"c"}, hasReplacedAndKept{"second", "first"}, 2}, c)
}

// It is checked as a file is, when a command first decodes.
func TestBaseConfigErrors(t *testing.T) {
	for base, want := range map[string]string{
		"Value = \n":    "invalid Options.BaseConfig: line 1: unexpected character",
		"Vaule = 'v'\n": "invalid Options.BaseConfig: unknown configuration key(s): line 1: Vaule",
	} {
		opts := testOptions
		opts.BaseConfig = []byte(base)
		require.ErrorContains(t, run(t, &BasicConfig{}, opts), want, base)
	}
}

func TestIgnoredFieldsMatchTheMarkup(t *testing.T) {
	type hasIgnored struct {
		Ignored string `toml:"-"`
		Value   string
	}
	type hasIgnoredAndListsIgnored struct {
		hasIgnored
		List []hasIgnored
	}

	t.Run("a config file", func(t *testing.T) {
		file := "Value = 'v'\n[[List]]\nValue = 'l'\n"
		var want, got hasIgnoredAndListsIgnored
		require.NoError(t, toml.Unmarshal([]byte(file), &want))
		require.NoError(t, run(t, &got, testOptions, "--config", writeConfig(t, file)))
		assert.Equal(t, hasIgnoredAndListsIgnored{hasIgnored{Value: "v"}, []hasIgnored{{Value: "l"}}}, want)
		assert.Equal(t, want, got)

		// tomlmarkup's decoder skips these alone, but a key no field reads is an error here, a list
		// element's too.
		for file, key := range map[string]string{"Ignored = 's'\n": "line 1: Ignored", "'-' = 's'\n": "line 1: -", "[[List]]\nIgnored = 's'\n": "line 2: List.Ignored"} {
			require.ErrorContains(t, run(t, &hasIgnoredAndListsIgnored{}, testOptions, "--config", writeConfig(t, file)), "unknown configuration key(s): "+key, file)
		}
	})

	t.Run("flags and env vars", func(t *testing.T) {
		assert.Nil(t, flagsOf(t, &hasIgnoredAndListsIgnored{}, testOptions).Lookup("ignored"))
		t.Setenv("TEST_IGNORED", "s")
		var c hasIgnoredAndListsIgnored
		require.NoError(t, run(t, &c, testOptions))
		assert.Empty(t, c.Ignored)
	})

	// pflag can't parse the --- its flag would need.
	t.Run("a field named -", func(t *testing.T) {
		type hasDashName struct {
			Dash string `toml:"-,"` //nolint:revive // the key "-" is what's under test
		}
		b := newBinder(t, newRoot(t), testOptions)
		require.ErrorContains(t, b.Register(b.root, "", &hasDashName{}), `Dash is named "-"`)
	})
}

func TestDefaultConfigPath(t *testing.T) {
	opts := testOptions
	opts.DefaultConfigPath = writeConfig(t, "Value = 'default-path'\n")

	for _, tc := range []struct{ name, file, want string }{
		{"read with no --config", "", "default-path"},
		{"--config overrides it", "Value = 'named'", "named"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c BasicConfig
			require.NoError(t, run(t, &c, opts, supply(t, "", "", tc.file)...))
			assert.Equal(t, tc.want, c.Value)
		})
	}
}

func TestNamespace(t *testing.T) {
	for _, tc := range everySource([]string{"--database.value", "v"}, "v", "[database]\nValue = 'v'") {
		t.Run(tc.name, func(t *testing.T) {
			var c BasicConfig
			require.NoError(t, runIn(t, "database", &c, supply(t, "TEST_DATABASE_VALUE", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, "v", c.Value)
		})
	}

	t.Run("separates same-named fields", func(t *testing.T) {
		root := newRoot(t)
		b := newBinder(t, root, testOptions)
		var db, evm BasicConfig
		require.NoError(t, b.Register(root, "database", &db))
		require.NoError(t, b.Register(root, "evm", &evm))

		root.SetArgs([]string{"--database.value", "d", "--evm.value", "e"})
		require.NoError(t, root.Execute())
		assert.Equal(t, "d", db.Value)
		assert.Equal(t, "e", evm.Value)
	})
}

func TestMultipleTargetsBothReportTheirOwnErrors(t *testing.T) {
	root := newRoot(t)
	require.NoError(t, bind(t, root, &RequiredField{}, testOptions).Register(root, "other", &RequiredField{}))
	err := root.Execute()
	require.ErrorContains(t, err, "invalid configuration: Value failed")
	assert.ErrorContains(t, err, "invalid configuration: other.Value failed")
}

func TestBuiltinCommandsSkipValidation(t *testing.T) {
	// Cobra only adds help and completion to a command with children.
	for _, args := range [][]string{{"help"}, {"completion", "bash"}, {"__complete", ""}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := newRoot(t)
			bind(t, root, &RequiredField{}, testOptions)
			root.AddCommand(&cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }})
			root.SetArgs(args)
			require.NoError(t, root.Execute())
		})
	}
}

func TestPersistentHooks(t *testing.T) {
	t.Run("one assigned after Register sees the decoded config", func(t *testing.T) {
		root := newRoot(t)
		var c BasicConfig
		bind(t, root, &c, testOptions)
		var seen string
		root.PersistentPreRunE = func(*cobra.Command, []string) error { seen = c.Value; return nil }

		root.SetArgs([]string{"--value", "v"})
		require.NoError(t, root.Execute())
		assert.Equal(t, "v", seen)
	})

	t.Run("a subcommand's own sees the root decoded", func(t *testing.T) {
		root := newRoot(t)
		var c BasicConfig
		bind(t, root, &c, testOptions)
		var seen string
		root.AddCommand(&cobra.Command{
			Use:               "sub",
			PersistentPreRunE: func(*cobra.Command, []string) error { seen = c.Value; return nil },
			RunE:              func(*cobra.Command, []string) error { return nil },
		})

		root.SetArgs([]string{"sub", "--value", "v"})
		require.NoError(t, root.Execute())
		assert.Equal(t, "v", seen)
	})

	t.Run("the caller's still runs", func(t *testing.T) {
		root := newRoot(t)
		bind(t, root, &struct{}{}, testOptions)
		called := false
		root.PersistentPreRun = func(*cobra.Command, []string) { called = true }

		require.NoError(t, root.Execute())
		assert.True(t, called)
	})
}

func TestUndocumentedListsFlagsWithoutHelp(t *testing.T) {
	assert.Equal(t, []string{"value"}, bind(t, newRoot(t), &BasicConfig{}, testOptions).Undocumented())
	assert.Empty(t, bind(t, newRoot(t), &simple.Config{}, testOptions).Undocumented())
}
