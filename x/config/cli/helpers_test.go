package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

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

func newBinder(t *testing.T, opts Options) *Binder {
	t.Helper()

	b, err := New(opts)
	require.NoError(t, err)
	return b
}

func bind[T any](t *testing.T, cmd *cobra.Command, target *T, opts Options) *Binder {
	t.Helper()

	b := newBinder(t, opts)
	require.NoError(t, b.Register(cmd, target))
	return b
}

func flagsOf[T any](t *testing.T, target *T, opts Options) *pflag.FlagSet {
	t.Helper()
	root := newRoot(t)
	bind(t, root, target, opts)
	return root.PersistentFlags()
}

func run[T any](t *testing.T, target *T, opts Options, args ...string) error {
	t.Helper()

	root := newRoot(t)
	bind(t, root, target, opts)
	root.SetArgs(args)
	return root.Execute()
}

func supplyConfig(t *testing.T, envVar, env, file string, args ...string) []string {
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

type readsItselfAsText struct{ Value string }

func (s *readsItselfAsText) UnmarshalText(data []byte) error { s.Value = string(data); return nil }

type writesItselfAsText struct{ Value string }

func (writesItselfAsText) MarshalText() ([]byte, error) { return []byte("x"), nil }
