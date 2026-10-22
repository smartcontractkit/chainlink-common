package cli

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func TestParseTextRejectsATypeWithNoTextForm(t *testing.T) {
	_, err := parseText(reflect.TypeFor[struct{}](), "x")
	require.ErrorContains(t, err, "has no text form")
}

func TestRegisterValue(t *testing.T) {
	type config struct{ Host string }

	b, err := New(Options{Markup: tomlmarkup.New()})
	require.NoError(t, err)
	root := &cobra.Command{Use: "app", RunE: func(*cobra.Command, []string) error { return nil }}

	cfg := config{}
	var target any = &cfg
	require.NoError(t, b.RegisterValue(root, "Server", target))
	require.Error(t, b.RegisterValue(root, "Other", cfg))
	require.Error(t, b.RegisterValue(root, "Other", (*config)(nil)))

	root.SetArgs([]string{"--server.host", "example.com"})
	require.NoError(t, root.Execute())
	require.Equal(t, "example.com", cfg.Host)
}
