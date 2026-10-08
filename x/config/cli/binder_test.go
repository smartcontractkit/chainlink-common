package cli

import (
	"encoding"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	nested "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/appconfig"
	simple "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/appconfig"
	"github.com/smartcontractkit/chainlink-common/x/config/commentparsing"
	"github.com/smartcontractkit/chainlink-common/x/config/markup"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func TestSubcommandFlagsArePersistent(t *testing.T) {
	root := newRoot(t)
	group := &cobra.Command{Use: "sub"}
	leaf := &cobra.Command{Use: "leaf", RunE: func(*cobra.Command, []string) error { return nil }}
	var s BasicConfig
	bind(t, group, &s, testOptions)
	group.AddCommand(leaf)
	root.AddCommand(group)

	root.SetArgs([]string{"sub", "leaf", "--value", "v"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "v", s.Value)
}

func TestFlagsCollide(t *testing.T) {
	type embedsTwoValues struct {
		BasicConfig
		RequiredField
	}
	require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &embedsTwoValues{}), "flag --value is already defined on app")
}

func TestKeysAConfigFileCantSeparateFailEveryCommand(t *testing.T) {
	type hasCaseOnlyTwins struct {
		Twin []BasicConfig
		TWIN []BasicConfig
	}
	type hasSection struct{ Value BasicConfig }
	type hasFoldedEnvVar struct {
		Folded string `toml:"value_value"`
	}

	t.Run("differ only in case in one struct", func(t *testing.T) {
		requireClashFailsSibling(t, &struct{}{}, &hasCaseOnlyTwins{}, "sub: Twin and TWIN differ only in case")
	})
	t.Run("differ only in case across commands", func(t *testing.T) {
		requireClashFailsSibling(t, &BasicConfig{}, &struct{ VALUE string }{}, "sub: Value and VALUE differ only in case")
	})
	t.Run("a value, then a section", func(t *testing.T) {
		requireClashFailsSibling(t, &BasicConfig{}, &hasSection{}, "sub: two fields hold Value")
	})
	t.Run("a section, then a value", func(t *testing.T) {
		requireClashFailsSibling(t, &hasSection{}, &BasicConfig{}, "sub: two fields hold Value")
	})
	t.Run("one env var", func(t *testing.T) {
		requireClashFailsSibling(t, &hasSection{}, &hasFoldedEnvVar{}, "sub: value.value and value-value are both TEST_VALUE_VALUE")
	})
}

func requireClashFailsSibling[P, S any](t *testing.T, parent *P, sub *S, want string) {
	t.Helper()

	root := newRoot(t)
	b := bind(t, root, parent, testOptions)
	subCmd := &cobra.Command{Use: "sub"}
	require.NoError(t, b.Register(subCmd, sub))
	root.AddCommand(subCmd, &cobra.Command{Use: "other", RunE: func(*cobra.Command, []string) error { return nil }})

	root.SetArgs([]string{"other"})
	require.ErrorContains(t, root.Execute(), want)
}

func TestRegisterRejectsAnExistingConfigFlag(t *testing.T) {
	for name, define := range map[string]func(*cobra.Command){
		"local":      func(c *cobra.Command) { c.Flags().String(ConfigFlagName, "", "") },
		"persistent": func(c *cobra.Command) { c.PersistentFlags().StringArray(ConfigFlagName, nil, "") },
	} {
		t.Run(name, func(t *testing.T) {
			root := newRoot(t)
			define(root)
			require.ErrorContains(t, newBinder(t, testOptions).Register(root, &BasicConfig{}), "flag --config is already defined on app")
		})
	}
}

func TestConfigFilesHoldOnlyTheCommandsKeys(t *testing.T) {
	type runsConfig struct{ Runs string }
	type siblingConfig struct{ Sibling string }

	for _, tc := range []struct {
		name, file string
		want       []string
	}{
		{"its own and the root's", "Value = 'v'\nRuns = 'r'\n", nil},
		{"a sibling's", "Sibling = 's'\n", []string{"unknown configuration key(s)", "Sibling"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newRoot(t)
			runs := &cobra.Command{Use: "runs", RunE: func(*cobra.Command, []string) error { return nil }}
			sibling := &cobra.Command{Use: "sibling"}
			root.AddCommand(runs, sibling)
			var shared BasicConfig
			var own runsConfig
			b := bind(t, root, &shared, testOptions)
			require.NoError(t, b.Register(runs, &own))
			require.NoError(t, b.Register(sibling, &siblingConfig{}))

			root.SetArgs([]string{"runs", "--config", writeConfig(t, tc.file)})
			err := root.Execute()
			if tc.want == nil {
				require.NoError(t, err)
				assert.Equal(t, BasicConfig{"v"}, shared)
				assert.Equal(t, runsConfig{"r"}, own)
			}

			for _, want := range tc.want {
				require.ErrorContains(t, err, want)
			}
		})
	}
}

func TestConfigFileErrors(t *testing.T) {
	type hasMapAndSection struct {
		Int     int
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := run(t, &hasMapAndSection{}, testOptions, tc.args...)
			for _, want := range tc.want {
				require.ErrorContains(t, err, want)
			}
		})
	}
}

func TestConfigFilesAreLayeredInOrder(t *testing.T) {
	opts := testOptions
	opts.BaseConfig = []byte("LogLevel = 'base'\n[Server]\nHost = 'base'\n")
	first := writeConfig(t, "[Server]\nHost = 'first'\n")
	second := writeConfig(t, "[Server]\nHost = 'second'\n")

	var c nested.Config
	require.NoError(t, run(t, &c, opts, "--config", first, "--config", second))
	assert.Equal(t, "base", c.LogLevel)
	assert.Equal(t, "second", c.Server.Host)
}

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

		// tomlmarkup's decoder skips these alone, but a key no field reads is an error here, a list element's too.
		for file, key := range map[string]string{"Ignored = 's'\n": "line 1: Ignored", "'-' = 's'\n": "line 1: -", "[[List]]\nIgnored = 's'\n": "line 2: List.Ignored"} {
			require.ErrorContains(t, run(t, &hasIgnoredAndListsIgnored{}, testOptions, "--config", writeConfig(t, file)), "unknown configuration key(s): "+key, file)
		}
	})

	// pflag can't parse the --- its flag would need.
	t.Run("a field named -", func(t *testing.T) {
		type hasDashName struct {
			Dash string `toml:"-,"` //nolint:revive // the key "-" is what's under test
		}
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &hasDashName{}), `Dash is named "-"`)
	})
}

func TestDefaultConfigPath(t *testing.T) {
	t.Run("config.<extension> in the working directory when unset", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config."+testOptions.Markup.Extension()), []byte("Value = 'v'\n"), 0o600))
		t.Chdir(dir)

		var c BasicConfig
		require.NoError(t, run(t, &c, testOptions))
		assert.Equal(t, "v", c.Value)
	})

	opts := testOptions
	opts.DefaultConfigPath = writeConfig(t, "Value = 'default-path'\n")

	for _, tc := range []struct{ name, file, want string }{
		{"read with no --config", "", "default-path"},
		{"--config overrides it", "Value = 'named'", "named"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c BasicConfig
			require.NoError(t, run(t, &c, opts, supplyConfig(t, "", "", tc.file)...))
			assert.Equal(t, tc.want, c.Value)
		})
	}
}

func TestMultipleTargetsBothReportTheirOwnErrors(t *testing.T) {
	type hasOther struct {
		Other string `validate:"required"`
	}

	root := newRoot(t)
	sub := &cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(sub)
	require.NoError(t, bind(t, root, &RequiredField{}, testOptions).Register(sub, &hasOther{}))
	root.SetArgs([]string{"sub"})
	err := root.Execute()
	require.ErrorContains(t, err, "invalid configuration: Value failed")
	assert.ErrorContains(t, err, "invalid configuration: Other failed")
}

func TestBuiltinCommandsSkipValidation(t *testing.T) {
	root := newRoot(t)
	bind(t, root, &RequiredField{}, testOptions)
	// Cobra only adds help to a command with children.
	root.AddCommand(&cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }})
	root.SetArgs([]string{"help"})
	require.NoError(t, root.Execute())
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

func TestKeyNames(t *testing.T) {
	type hasUntaggedCamelCase struct {
		AcronymID uint32
		TwoWords  config.Duration
	}
	type hasLanguageAndOtherTags struct {
		LanguageTagged string `toml:"renamed"`
		OtherTagged    string `mapstructure:"other"`
	}
	type embedsBasic struct{ BasicConfig }
	type embedsUnexported struct {
		unexportedBasicConfig
		Own int
	}
	type embedsAnEmbedder struct{ embedsUnexported }
	type embedsBasicPointer struct {
		*BasicConfig
		Own string
	}
	type embedsNamedPointer struct {
		*BasicConfig `toml:"aws,omitempty"`
	}

	t.Run("untagged fields are kebab-cased", func(t *testing.T) {
		decodesTo(t, hasUntaggedCamelCase{137, *config.MustNewDuration(7 * time.Second)}, "--acronym-id", "137", "--two-words", "7s")
	})
	t.Run("only the language's tag names a key", func(t *testing.T) {
		decodesTo(t, hasLanguageAndOtherTags{"l", "o"}, "--renamed", "l", "--other-tagged", "o")
	})
	t.Run("an embedded struct flattens", func(t *testing.T) {
		decodesTo(t, embedsBasic{BasicConfig{"v"}}, "--value", "v")
	})
	t.Run("unexported embeds flatten all the way up", func(t *testing.T) {
		decodesTo(t, embedsAnEmbedder{embedsUnexported{unexportedBasicConfig{"v"}, 1}}, "--value", "v", "--own", "1")
	})
	t.Run("an embedded pointer flattens and is allocated", func(t *testing.T) {
		decodesTo(t, embedsBasicPointer{&BasicConfig{"v"}, "x"}, "--value", "v", "--own", "x")
	})
	// A named embedded pointer is how a polymorphic section is written, so the name must win.
	t.Run("a named embedded pointer keeps its section", func(t *testing.T) {
		decodesTo(t, embedsNamedPointer{&BasicConfig{"p"}}, "--aws.value", "p")
	})
}

func decodesTo[T any](t *testing.T, want T, args ...string) {
	t.Helper()

	var got T
	require.NoError(t, run(t, &got, testOptions, args...))
	assert.Equal(t, want, got)
}

type opaqueWithPointerKeys struct{ byAddress map[*int]string }

func (o *opaqueWithPointerKeys) UnmarshalText([]byte) error { o.byAddress = nil; return nil }

func TestWhatGetsAFlag(t *testing.T) {
	type nestsItself struct {
		Value string
		Child *nestsItself
	}
	type hasEveryShape struct {
		String        string
		Bytes         []byte
		ListOfStrings []string
		ListOfPtrs    []*int
		MapOfPtrs     map[string]*int
		ListOfMaps    []map[string]string
		ListOfStructs []nestsItself
		MapOfStructs  map[string]nestsItself
		MapOfBytes    map[string][]byte
		TextReader    readsItselfAsText
		Opaque        opaqueWithPointerKeys
		Interface     encoding.TextUnmarshaler
		Interfaces    []encoding.TextUnmarshaler
		InterfaceMap  map[string]encoding.TextUnmarshaler
		TextWriter    writesItselfAsText
		Nested        nestsItself
		Complex       complex128
		unexported    string //nolint:unused // nothing can set it, so nothing binds it
	}

	flags := flagsOf(t, &hasEveryShape{}, testOptions)
	for _, name := range []string{"string", "bytes", "list-of-strings", "map-of-bytes", "text-reader", "opaque", "text-writer.value", "nested.value"} {
		assert.NotNil(t, flags.Lookup(name), name)
	}

	// Without a single text form a flag's default could not be read back, so these are file-only.
	for _, name := range []string{"interface", "interfaces", "interface-map", "list-of-ptrs", "map-of-ptrs", "list-of-maps", "list-of-structs", "list-of-structs.value", "map-of-structs",
		"complex", "unexported", "text-reader.value", "nested.child.value"} {
		assert.Nil(t, flags.Lookup(name), name)
	}
}

func TestRegistrationRejects(t *testing.T) {
	var nilPointer *struct{}
	host := "example.com"
	type embedsUnexportedPointer struct {
		*unexportedBasicConfig
	}
	type hasPointerKeys struct {
		Keys map[*int]string
	}
	type nestsAPointerKey struct {
		Lists []map[string]hasPointerKeys
	}
	type hasAPointerInsideKeys struct {
		Keys map[struct{ P [1]*int }]string
	}
	type hasAnInterfaceInsideKeys struct {
		Keys map[struct{ V any }]string
	}
	type embedsAndHoldsIDs struct {
		BasicConfig
		IDs map[uint32]string
	}
	type holdsARecursiveUintMap struct {
		Tree recursiveUintMap
	}
	type nestsAnEmbedder struct {
		Inner embedsAndHoldsIDs
	}
	type listsEmbeddersOfIDs struct {
		List []nestsAnEmbedder
	}

	t.Run("nil pointer", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), nilPointer), "target pointer cannot be nil")
	})
	t.Run("not a struct", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &host), "target must be a struct or pointer to struct")
	})
	// Reflect can't allocate it, so decoding would panic.
	t.Run("an unexported embedded pointer", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &embedsUnexportedPointer{}), "embedded *unexportedBasicConfig is unexported")
	})
	t.Run("a pointer map key", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &hasPointerKeys{}), "cli.hasPointerKeys: Keys: map[*int]string has keys that hold a pointer, channel or interface")
	})
	t.Run("a pointer map key nested in a list, map and struct", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &nestsAPointerKey{}), "cli.nestsAPointerKey: Lists: map[*int]string has keys that hold a pointer, channel or interface")
	})
	t.Run("a pointer inside a struct or array key", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &hasAPointerInsideKeys{}), "cli.hasAPointerInsideKeys: Keys: map[struct { P [1]*int }]string has keys that hold a pointer, channel or interface")
	})
	t.Run("an interface inside a key", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &hasAnInterfaceInsideKeys{}), "cli.hasAnInterfaceInsideKeys: Keys: map[struct { V interface {} }]string has keys that hold a pointer, channel or interface")
	})
	t.Run("a recursive type with non-string keys", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &holdsARecursiveUintMap{}),
			"cli.holdsARecursiveUintMap: Tree: cli.recursiveUintMap holds itself and a map with non-string keys")
	})
	t.Run("a struct nested in a list that embeds a field and holds non-string keys", func(t *testing.T) {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &listsEmbeddersOfIDs{}),
			"cli.listsEmbeddersOfIDs: List: cli.embedsAndHoldsIDs holds a map with non-string keys and embeds a field")
	})

	_, err := New(Options{})
	require.ErrorIs(t, err, markup.Err)
}

func TestPointers(t *testing.T) {
	type hasRequiredAndOptional struct {
		Required int32 `validate:"required"`
		Optional int32
	}
	type hasPointers struct {
		*BasicConfig
		Section *hasRequiredAndOptional
		String  *string
	}

	var unset hasPointers
	require.NoError(t, run(t, &unset, testOptions))
	assert.Equal(t, hasPointers{}, unset)

	var empty hasPointers
	require.NoError(t, run(t, &empty, testOptions, "--string", ""))
	require.NotNil(t, empty.String)

	var partial hasPointers
	require.ErrorContains(t, run(t, &partial, testOptions, "--section.optional", "9"), "Section.Required failed")
	require.NotNil(t, partial.Section)
}

func TestNumericFlagsAreSized(t *testing.T) {
	type hasEveryNumber struct {
		Int     int
		Int8    int8
		Int16   int16
		Int32   int32
		Int64   int64
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Float32 float32
		Float64 float64
	}

	for _, tc := range []struct{ flag, limit, past string }{
		{"int", "9223372036854775807", "9223372036854775808"},
		{"int-8", "-128", "-129"},
		{"int-16", "32767", "32768"},
		{"int-32", "-2147483648", "-2147483649"},
		{"int-64", "9223372036854775807", "9223372036854775808"},
		{"uint", "18446744073709551615", "18446744073709551616"},
		{"uint-8", "255", "256"},
		{"uint-16", "65535", "65536"},
		{"uint-32", "4294967295", "4294967296"},
		{"uint-64", "18446744073709551615", "18446744073709551616"},
		{"float-32", "3.4e38", "3.5e38"},
		{"float-64", "1.7e308", "1.8e308"},
	} {
		t.Run(tc.flag, func(t *testing.T) {
			require.NoError(t, run(t, &hasEveryNumber{}, testOptions, "--"+tc.flag, tc.limit))
			require.ErrorContains(t, run(t, &hasEveryNumber{}, testOptions, "--"+tc.flag, tc.past), "out of range")
		})
	}
}

func TestExplicitDefaultValueOverridesANonDefault(t *testing.T) {
	flag := []string{"--bool=false", "--uint", "0", "--string", ""}
	for _, tc := range everySource(flag, "", "Bool = false\nUint = 0\nString = ''") {
		t.Run(tc.name, func(t *testing.T) {
			c := hasEveryKind{Bool: true, Uint: 5, String: "set"}
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, hasEveryKind{}, c)
		})
	}
}

func TestPrecedence(t *testing.T) {
	type hasTwoStrings struct {
		Value     string
		Untouched string
	}

	for _, tc := range []struct {
		name, env, file string
		args            []string
		want            string
	}{
		{"file", "", "Value = 'file'", nil, "file"},
		{"file and env", "env", "Value = 'file'", nil, "env"},
		{"file, env and flag", "env", "Value = 'file'", []string{"--value", "flag"}, "flag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_VALUE", "")
			c := hasTwoStrings{Value: "default", Untouched: "untouched"}
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "TEST_VALUE", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, hasTwoStrings{tc.want, "untouched"}, c)
		})
	}
}

func TestEnvPrefixesTriedInOrder(t *testing.T) {
	t.Setenv("FIRST_VALUE", "first")
	t.Setenv("SECOND_VALUE", "second")

	var c BasicConfig
	require.NoError(t, run(t, &c, Options{Markup: tomlmarkup.New(), Prefixes: []string{"FIRST", "SECOND"}}))
	assert.Equal(t, "first", c.Value)
}

func TestValidationErrorsUseConfigKeys(t *testing.T) {
	// Renamed keys show a config file key comes from the tag, not the Go name.
	type nestsRenamedRequired struct {
		GoName RequiredField `toml:"Renamed"`
	}
	type hasExclusivePointers struct {
		Excluded *RequiredField `validate:"excluded_with=Excluder"`
		Excluder *RequiredField `toml:"RenamedExcluder"`
	}

	t.Run("a leaf names how to set it", func(t *testing.T) {
		require.ErrorContains(t, run(t, &nestsRenamedRequired{}, testOptions), "invalid configuration: Renamed.Value failed on the 'required' tag; "+
			"set it with --renamed.value, TEST_RENAMED_VALUE, Renamed.Value in a config file")
	})
	t.Run("a section and its rule's sibling", func(t *testing.T) {
		require.ErrorContains(t, run(t, &hasExclusivePointers{}, testOptions, "--excluded.value", "v", "--renamed-excluder.value", "v"),
			"invalid configuration: Excluded failed on the 'excluded_with=RenamedExcluder' tag")
	})
}

func TestValidationErrorsInsideAListOrMapUseConfigKeys(t *testing.T) {
	type hasTaggedRequired struct {
		GoName string `toml:"renamed" validate:"required"`
	}
	type hasSiblingRule struct {
		Rule    string `validate:"required_without=Sibling"`
		Sibling string `toml:"renamed"`
	}

	type embedsRequired struct {
		RequiredField
	}
	type hasValidatedCollections struct {
		Structs []hasTaggedRequired `validate:"dive"`
		Strings []string            `validate:"dive,required"`
		Map     map[string]string   `validate:"dive,required"`
		Rules   []hasSiblingRule    `validate:"dive"`
		Embeds  []embedsRequired    `validate:"dive"`
	}

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"a list element's field, by its tag", []string{"--config", writeConfig(t, "[[Structs]]\nrenamed = 'a'\n[[Structs]]\nrenamed = ''\n")},
			"Structs.1.renamed failed on the 'required' tag"},
		{"a list element", []string{"--strings", "a,"}, "Strings.1 failed on the 'required' tag"},
		{"a map entry", []string{"--map", "k="}, "Map.k failed on the 'required' tag"},
		{"a rule's sibling in the same element", []string{"--config", writeConfig(t, "[[Rules]]\n")},
			"Rules.0.Rule failed on the 'required_without=Rules.0.renamed' tag"},
		{"a flattened embed adds no segment", []string{"--config", writeConfig(t, "[[Embeds]]\n")},
			"Embeds.0.Value failed on the 'required' tag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := run(t, &hasValidatedCollections{}, testOptions, tc.args...)
			require.ErrorContains(t, err, tc.want)
			// Nothing falls back to the validator's own text, which names Go fields.
			assert.NotContains(t, err.Error(), "Key: ")
		})
	}
}

func TestRepeatedListFlagAppends(t *testing.T) {
	type hasStrings struct{ Strings []string }

	var c hasStrings
	require.NoError(t, run(t, &c, testOptions, "--strings", "a", "--strings", "b"))
	assert.Equal(t, []string{"a", "b"}, c.Strings)
}

func TestListText(t *testing.T) {
	type hasStrings struct{ Strings []string }

	c := hasStrings{Strings: []string{"default"}}
	require.NoError(t, run(t, &c, testOptions, "--strings", ""))
	assert.Empty(t, c.Strings)

	require.ErrorContains(t, run(t, &hasStrings{}, testOptions, "--strings", `"a`), `extraneous or missing " in quoted-field`)
	require.ErrorContains(t, run(t, &hasStrings{}, testOptions, "--strings", "a\nb"), `"a\nb" must be one line`)
}

// A byte slice is text (JSON, PEM), so neither a flag nor an env var splits it on its commas.
func TestByteSliceIsText(t *testing.T) {
	type hasByteSlice struct {
		Bytes []byte
	}

	assert.Equal(t, `{"a":1}`, flagsOf(t, &hasByteSlice{Bytes: []byte(`{"a":1}`)}, testOptions).Lookup("bytes").DefValue)
	for _, tc := range everySource([]string{"--bytes", `{"a":1,"b":2}`}, `{"a":1,"b":2}`, "") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasByteSlice
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "TEST_BYTES", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, []byte(`{"a":1,"b":2}`), c.Bytes)
		})
	}
}

func TestSelfDecodingSliceOwnsItsForm(t *testing.T) {
	type hasSelfSplittingList struct {
		List selfSplittingList
	}
	t.Setenv("TEST_LIST", "80+443")

	var c hasSelfSplittingList
	require.NoError(t, run(t, &c, testOptions))
	assert.Equal(t, selfSplittingList{80, 443}, c.List)
}

type selfSplittingList []int

func (p *selfSplittingList) UnmarshalText(b []byte) error {
	*p = nil
	for part := range strings.SplitSeq(string(b), "+") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return err
		}

		*p = append(*p, n)
	}

	return nil
}

type hasStringMap struct {
	Map map[string]string
}

func TestStringMapFromEverySource(t *testing.T) {
	cases := everySource([]string{"--map", "a=1,b=2"}, "a=1,b=2", "[Map]\na = '1'\nb = '2'")
	cases = append(cases, sourceCase{name: "repeated flag", args: []string{"--map", "a=1", "--map", "b=2"}})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c hasStringMap
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "TEST_MAP", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, map[string]string{"a": "1", "b": "2"}, c.Map)
		})
	}
}

func TestMapsMergeAcrossSources(t *testing.T) {
	t.Setenv("TEST_MAP", "e=env,a=env")
	c := hasStringMap{Map: map[string]string{"d": "default", "a": "default"}}
	require.NoError(t, run(t, &c, testOptions,
		"--config", writeConfig(t, "[Map]\na = 'first'\nf = 'first'\n"),
		"--config", writeConfig(t, "[Map]\na = 'second'\ns = 'second'\n"),
		"--map", "a=flag"))
	assert.Equal(t, map[string]string{"d": "default", "f": "first", "s": "second", "e": "env", "a": "flag"}, c.Map)
}

func TestMapEntryText(t *testing.T) {
	var c hasStringMap
	require.NoError(t, run(t, &c, testOptions, "--map", `"a=1,b=2",c=3`))
	assert.Equal(t, map[string]string{"a": "1,b=2", "c": "3"}, c.Map)

	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions, "--map", "a=1,,b=2"), `"" must be formatted as key=value`)

	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions, "--map", `"a=1`), `extraneous or missing " in quoted-field`)

	t.Setenv("TEST_MAP", "a")
	require.ErrorContains(t, run(t, &hasStringMap{}, testOptions), `invalid value "a" for TEST_MAP: "a" must be formatted as key=value`)
}

type caseless string

func (c *caseless) UnmarshalText(b []byte) error {
	*c = caseless(strings.ToLower(string(b)))
	return nil
}

func TestMapKeysNeedNotBeStrings(t *testing.T) {
	type hasNonStringKeys struct {
		UintKeys map[uint32]string
		TextKeys map[config.Duration]string
	}
	want := hasNonStringKeys{
		UintKeys: map[uint32]string{1: "a", 42: "b"},
		TextKeys: map[config.Duration]string{*config.MustNewDuration(45 * time.Second): "c"},
	}

	flag := []string{"--uint-keys", "1=a,0x2a=b", "--text-keys", "45s=c"}
	for _, tc := range everySource(flag, "", "[UintKeys]\n1 = 'a'\n0x2a = 'b'\n[TextKeys]\n45s = 'c'\n") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasNonStringKeys
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, want, c)
		})
	}

	for _, args := range []string{"16=a,0x10=b", "16=a --uint-keys 0x10=b"} {
		require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, append([]string{"--uint-keys"}, strings.Fields(args)...)...),
			`keys "16" and "0x10" are both 16`, args)
	}

	require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, "--config", writeConfig(t, "[UintKeys]\n16 = 'a'\n0x10 = 'b'\n")),
		`keys "0x10" and "16" are both 16`)
	require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, "--config", writeConfig(t, "[TextKeys]\n45s = 'a'\n0m45s = 'b'\n")),
		`keys "0m45s" and "45s" are both 45s`)
	require.ErrorContains(t, run(t, &struct{ FloatKeys map[float64]string }{}, testOptions, "--float-keys", "NaN=a"),
		`key "NaN" never equals itself`)
	require.ErrorContains(t, run(t, &struct{ Caseless map[caseless]string }{}, testOptions, "--config", writeConfig(t, "[Caseless]\na = '1'\nA = '2'\n")),
		`keys "A" and "a" are both a`)
	require.ErrorContains(t, run(t, &hasNonStringKeys{}, testOptions, "--config", writeConfig(t, "[UintKeys]\nx = 'a'\n")),
		`uint-keys: key "x": strconv.ParseUint: parsing "x": invalid syntax`)
}

func TestMapKeysNestedInAConfigFileOnlyField(t *testing.T) {
	type holdsIDs struct {
		IDs         map[uint32]string
		IgnoredPtrs map[*int]string `toml:"-"`
		ignore      string          //nolint:unused // the markup never sets it, so the rebuilt struct leaves it out
	}
	type hasNestedNonStringKeys struct {
		Lists    []map[uint32]string
		Maps     map[string]map[uint32]string
		Pointers []*map[uint32]string
		Arrays   [1]map[uint32]string
		Structs  []holdsIDs
	}

	var c hasNestedNonStringKeys
	file := "Lists = [{1 = 'a'}]\nPointers = [{3 = 'c'}]\nArrays = [{4 = 'd'}]\n[Maps.m]\n2 = 'b'\n[[Structs]]\n[Structs.IDs]\n0x10 = 'e'\n"
	require.NoError(t, run(t, &c, testOptions, "--config", writeConfig(t, file)))
	assert.Equal(t, hasNestedNonStringKeys{
		Lists:    []map[uint32]string{{1: "a"}},
		Maps:     map[string]map[uint32]string{"m": {2: "b"}},
		Pointers: []*map[uint32]string{{3: "c"}},
		Arrays:   [1]map[uint32]string{{4: "d"}},
		Structs:  []holdsIDs{{IDs: map[uint32]string{16: "e"}}},
	}, c)

	for file, want := range map[string]string{
		"Lists = [{x = 'a'}]\n":                 `lists: key "x": strconv.ParseUint`,
		"[Maps.m]\nx = 'b'\n":                   `maps: key "x": strconv.ParseUint`,
		"Pointers = [{x = 'c'}]\n":              `pointers: key "x": strconv.ParseUint`,
		"Arrays = [{x = 'd'}]\n":                `arrays: key "x": strconv.ParseUint`,
		"[[Structs]]\n[Structs.IDs]\nx = 'e'\n": `structs: key "x": strconv.ParseUint`,
	} {
		require.ErrorContains(t, run(t, &hasNestedNonStringKeys{}, testOptions, "--config", writeConfig(t, file)), want, file)
	}
}

type nodeName string

type recursiveMap map[nodeName]*recursiveMap

type wholeMap map[string]string

func (m *wholeMap) UnmarshalText(b []byte) error { *m = wholeMap{"text": string(b)}; return nil }

func TestAMapReadWholeIsReplacedNotMerged(t *testing.T) {
	type hasWholeMap struct {
		Whole wholeMap
	}

	c := hasWholeMap{Whole: wholeMap{"default": ""}}
	require.NoError(t, run(t, &c, testOptions, "--config", writeConfig(t, "Whole = 'a'\n"), "--config", writeConfig(t, "Whole = 'b'\n")))
	assert.Equal(t, wholeMap{"text": "b"}, c.Whole)

	require.NoError(t, run(t, &c, testOptions, "--whole", "c"))
	assert.Equal(t, wholeMap{"text": "c"}, c.Whole)
}

type recursiveUintMap map[uint32]*recursiveUintMap

func TestRecursiveConfigFileOnlyType(t *testing.T) {
	type hasRecursiveMap struct {
		Tree recursiveMap
	}

	var c hasRecursiveMap
	require.NoError(t, run(t, &c, testOptions, "--config", writeConfig(t, "[Tree.a.b]\n")))
	assert.Equal(t, hasRecursiveMap{Tree: recursiveMap{"a": {"b": new(recursiveMap)}}}, c)
}

type namedIntWithAString int

func (namedIntWithAString) String() string { return "name" }

func TestHelpDefaultsCanBePassedBackIn(t *testing.T) {
	type hasListMapAndText struct {
		Named     namedIntWithAString
		NamedList []namedIntWithAString
		Wait      time.Duration
		Ratio     float32
		List      []string
		LoneEmpty []string
		Map       map[string]string
		Text      *config.URL
	}

	c := hasListMapAndText{Named: 2, NamedList: []namedIntWithAString{3}, Wait: 90 * time.Second, Ratio: 0.1, List: []string{"a", "b,c", "d "}, LoneEmpty: []string{""}, Map: map[string]string{"a": "1", "b": "2,3", "c": " 4"}, Text: config.MustParseURL("https://x/rpc")}
	flags := flagsOf(t, &c, testOptions)
	assert.Equal(t, "value,...", flags.Lookup("list").Value.Type())
	assert.Equal(t, "key=value,...", flags.Lookup("map").Value.Type())

	type hasTextWithoutAWriter struct {
		Reader  readsItselfAsText
		Readers []readsItselfAsText
		ByName  map[string]readsItselfAsText
	}

	reader := readsItselfAsText{Value: "x"}
	readers := flagsOf(t, &hasTextWithoutAWriter{reader, []readsItselfAsText{reader}, map[string]readsItselfAsText{"a": reader}}, testOptions)
	for name, want := range map[string]string{"reader": cannotDisplay, "readers": cannotDisplay, "by-name": "a=" + cannotDisplay} {
		assert.Equal(t, want, readers.Lookup(name).DefValue, name)
	}

	var back hasListMapAndText
	args := []string{"--named", flags.Lookup("named").DefValue, "--named-list", flags.Lookup("named-list").DefValue,
		"--wait", flags.Lookup("wait").DefValue, "--ratio", flags.Lookup("ratio").DefValue,
		"--list", flags.Lookup("list").DefValue, "--lone-empty", flags.Lookup("lone-empty").DefValue, "--map", flags.Lookup("map").DefValue,
		"--text", flags.Lookup("text").DefValue}
	require.NoError(t, run(t, &back, testOptions, args...))
	assert.Equal(t, c, back)
}

type intKindText time.Duration

func (d *intKindText) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	*d = intKindText(parsed)
	return err
}

func TestTextUnmarshalerLeafFromEverySource(t *testing.T) {
	type hasIntKindText struct {
		Value intKindText
	}

	for _, tc := range everySource([]string{"--value", "30m"}, "30m", "Value = '30m'") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasIntKindText
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "TEST_VALUE", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, 30*time.Minute, time.Duration(c.Value))
		})
	}

	require.ErrorContains(t, run(t, &hasIntKindText{}, testOptions, "--value", "not-a-duration"), `time: invalid duration "not-a-duration"`)
}

func TestLeafNestedInASectionListAndMap(t *testing.T) {
	type hasDuration struct {
		Value config.Duration
	}
	type hasDurationEverywhere struct {
		Section hasDuration
		List    []config.Duration
		Map     map[string]config.Duration
	}
	want := hasDurationEverywhere{
		Section: hasDuration{*config.MustNewDuration(time.Minute)},
		List:    []config.Duration{*config.MustNewDuration(time.Second), *config.MustNewDuration(2 * time.Second)},
		Map:     map[string]config.Duration{"a": *config.MustNewDuration(3 * time.Second)},
	}

	flag := []string{"--section.value", "1m", "--list", "1s,2s", "--map", "a=3s"}
	for _, tc := range everySource(flag, "", "List = ['1s', '2s']\n[Section]\nValue = '1m'\n[Map]\na = '3s'") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasDurationEverywhere
			require.NoError(t, run(t, &c, testOptions, supplyConfig(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, want, c)
		})
	}
}

type documentedOverTwoLines struct{ Value string }

func (documentedOverTwoLines) DocComments() map[string]commentparsing.FieldDoc {
	return map[string]commentparsing.FieldDoc{"Value": {Comment: "Value is one line.\nIt continues here."}}
}

func TestHelpText(t *testing.T) {
	noPrefix := Options{Markup: tomlmarkup.New()}
	t.Run("a multi-line comment is one line", func(t *testing.T) {
		assert.Equal(t, "Value is one line. It continues here. [env VALUE]", flagsOf(t, &documentedOverTwoLines{}, noPrefix).Lookup("value").Usage)
	})
	t.Run("an embedded struct is documented by its own type", func(t *testing.T) {
		assert.Equal(t, "LogLevel is the minimum level to log. [env LOG_LEVEL]", flagsOf(t, &nested.Config{}, noPrefix).Lookup("log-level").Usage)
	})
	t.Run("env vars in the order tried", func(t *testing.T) {
		opts := Options{Markup: tomlmarkup.New(), Prefixes: []string{"CRE", "CL"}}
		assert.Equal(t, "Host is the host to dial. [env CRE_HOST, CL_HOST]", flagsOf(t, &simple.Config{}, opts).Lookup("host").Usage)
	})
	t.Run("an empty prefix adds none", func(t *testing.T) {
		opts := Options{Markup: tomlmarkup.New(), Prefixes: []string{"APP", ""}}
		assert.Equal(t, "Host is the host to dial. [env APP_HOST, HOST]", flagsOf(t, &simple.Config{}, opts).Lookup("host").Usage)
	})
	t.Run("required", func(t *testing.T) {
		assert.Equal(t, "(required) [env VALUE]", flagsOf(t, &RequiredField{}, noPrefix).Lookup("value").Usage)
	})
}

// pflag would print a string kind's raw default; its own MarshalText is what redacts it.
func TestSecretDefaultIsRedacted(t *testing.T) {
	type hasSetAndUnsetSecrets struct {
		Set   config.SecretString
		Unset config.SecretString
	}

	flags := flagsOf(t, &hasSetAndUnsetSecrets{Set: "hunter2"}, testOptions)
	assert.Equal(t, "xxxxx", flags.Lookup("set").DefValue)
	assert.Empty(t, flags.Lookup("unset").DefValue)

	c := hasSetAndUnsetSecrets{Set: "hunter2"}
	require.NoError(t, run(t, &c, testOptions))
	assert.Equal(t, config.SecretString("hunter2"), c.Set)
}

type hasEveryKind struct {
	Int      int
	Int8     int8
	Uint     uint
	Float32  float32
	Bool     bool
	String   string
	Duration time.Duration
	IntList  []int
	IntMap   map[string]int
}

func TestParserErrorsArePropagated(t *testing.T) {
	err := run(t, &hasEveryKind{}, testOptions, "--int=abc")
	require.ErrorContains(t, err, `invalid argument "abc" for "--int" flag: strconv.ParseInt: parsing "abc": invalid syntax`)

	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--int-list=1,x"), `element "x": strconv.ParseInt: parsing "x"`)
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions, "--int-map=a=x"), `value of "a": strconv.ParseInt: parsing "x"`)

	t.Setenv("TEST_INT", "abc")
	require.ErrorContains(t, run(t, &hasEveryKind{}, testOptions), `invalid value "abc" for TEST_INT: strconv.ParseInt: parsing "abc": invalid syntax`)
}

func TestASubcommandsOwnConfigFlagIsRejected(t *testing.T) {
	root := newRoot(t)
	bind(t, root, &BasicConfig{}, testOptions)
	sub := &cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }}
	sub.Flags().String(ConfigFlagName, "", "")
	root.AddCommand(sub)

	root.SetArgs([]string{"sub"})
	require.ErrorContains(t, root.Execute(), "--config on sub is not the flag Register added")
}
