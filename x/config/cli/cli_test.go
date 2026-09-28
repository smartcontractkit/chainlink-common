package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/x/config/markup"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

var testOptions = Options{Markup: tomlmarkup.New(), Prefixes: []string{"TEST"}}

func newRoot(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "app", RunE: func(*cobra.Command, []string) error { return nil }}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// Without args cobra would parse the test binary's own os.Args.
	cmd.SetArgs([]string{})
	return cmd
}

func newBinder(t *testing.T, root *cobra.Command, opts Options) *Binder {
	t.Helper()

	b, err := New(root, opts)
	require.NoError(t, err)
	return b
}

// bind takes any, so a test can register targets the typed Register would not compile.
func bind(t *testing.T, root *cobra.Command, target any, opts Options) *Binder {
	t.Helper()

	b := newBinder(t, root, opts)
	require.NoError(t, b.register(root, "", target))
	return b
}

func flagsOf(t *testing.T, target any, opts Options) *pflag.FlagSet {
	t.Helper()
	return bind(t, newRoot(t), target, opts).root.PersistentFlags()
}

func runIn(t *testing.T, namespace string, target any, args ...string) error {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, newBinder(t, root, testOptions).register(root, namespace, target))
	root.SetArgs(args)
	return root.Execute()
}

func run(t *testing.T, target any, opts Options, args ...string) error {
	t.Helper()

	root := newRoot(t)
	bind(t, root, target, opts)
	root.SetArgs(args)
	return root.Execute()
}

// supply sets envVar to env and adds file as a --config file, each only when not empty.
func supply(t *testing.T, envVar, env, file string, args ...string) []string {
	t.Helper()

	if env != "" {
		t.Setenv(envVar, env)
	}
	if file != "" {
		args = append(args, "--config", writeConfig(t, file))
	}
	return args
}

type sourceCase struct {
	name, env, file string
	args            []string
}

// everySource is one case per non-empty source, so each test states one value three ways.
func everySource(flag []string, env, file string) []sourceCase {
	var cases []sourceCase
	if flag != nil {
		cases = append(cases, sourceCase{name: "flag", args: flag})
	}
	if env != "" {
		cases = append(cases, sourceCase{name: "env", env: env})
	}
	if file != "" {
		cases = append(cases, sourceCase{name: "config file", file: file})
	}
	return cases
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config."+testOptions.Markup.Extension())
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

type BasicConfig struct {
	Value string
}

type RequiredField struct {
	Value string `validate:"required"`
}

type unexportedBasicConfig struct {
	Value string
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

	for _, tc := range []struct {
		name         string
		target, want any
		args         []string
	}{
		{"untagged fields are kebab-cased", &hasUntaggedCamelCase{}, &hasUntaggedCamelCase{137, *config.MustNewDuration(7 * time.Second)},
			[]string{"--acronym-id", "137", "--two-words", "7s"}},
		{"only the language's tag names a key", &hasLanguageAndOtherTags{}, &hasLanguageAndOtherTags{"l", "o"}, []string{"--renamed", "l", "--other-tagged", "o"}},
		{"an embedded struct flattens", &embedsBasic{}, &embedsBasic{BasicConfig{"v"}}, []string{"--value", "v"}},
		{"unexported embeds flatten all the way up", &embedsAnEmbedder{}, &embedsAnEmbedder{embedsUnexported{unexportedBasicConfig{"v"}, 1}},
			[]string{"--value", "v", "--own", "1"}},
		{"an embedded pointer flattens and is allocated", &embedsBasicPointer{}, &embedsBasicPointer{&BasicConfig{"v"}, "x"},
			[]string{"--value", "v", "--own", "x"}},
		// A named embedded pointer is how a polymorphic section is written, so the name must win.
		{"a named embedded pointer keeps its table", &embedsNamedPointer{}, &embedsNamedPointer{&BasicConfig{"p"}},
			[]string{"--aws.value", "p"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, run(t, tc.target, testOptions, tc.args...))
			assert.Equal(t, tc.want, tc.target)
		})
	}
}

type readsItselfAsText struct{ Value string }

func (s *readsItselfAsText) UnmarshalText(data []byte) error { s.Value = string(data); return nil }

type writesItselfAsText struct{ Value string }

func (writesItselfAsText) MarshalText() ([]byte, error) { return []byte("x"), nil }

func TestWhatGetsAFlag(t *testing.T) {
	type nestsItself struct {
		Value string
		Child *nestsItself
	}
	type hasEveryShape struct {
		String        string
		Bytes         []byte
		ListOfMaps    []map[string]string
		ListOfStructs []nestsItself
		MapOfStructs  map[string]nestsItself
		MapOfBytes    map[string][]byte
		TextReader    readsItselfAsText
		TextWriter    writesItselfAsText
		Nested        nestsItself
		Complex       complex128
		unexported    string //nolint:unused // nothing can set it, so nothing binds it
	}

	flags := flagsOf(t, &hasEveryShape{}, testOptions)
	for _, name := range []string{"string", "bytes", "map-of-bytes", "text-reader", "text-writer.value", "nested.value"} {
		assert.NotNil(t, flags.Lookup(name), name)
	}
	// Without a single text form a flag's default could not be read back, so these are file-only.
	for _, name := range []string{"list-of-maps", "list-of-structs", "list-of-structs.value", "map-of-structs",
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

	for _, tc := range []struct {
		name   string
		target any
		want   string
	}{
		{"nil pointer", nilPointer, "target pointer cannot be nil"},
		{"not a struct", &host, "target must be a struct or pointer to struct"},
		// Reflect can't allocate it, so decoding would panic.
		{"an unexported embedded pointer", &embedsUnexportedPointer{}, "embedded *unexportedBasicConfig is unexported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newBinder(t, newRoot(t), testOptions)
			require.ErrorContains(t, b.register(b.root, "", tc.target), tc.want)
		})
	}

	_, err := New(newRoot(t), Options{})
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
