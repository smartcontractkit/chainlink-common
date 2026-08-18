package config_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/config/flags"
)

const (
	first  = "12D3KooWQzePGqHw66cV1Qsm71eGZKiPEgALYYM3inPtFYibZ67e"
	second = "12D3KooWKh28EhBVfiiFh39w3zqtBxzYJhmGfBZNmoL4tRjMWSor"
)

func TestBootstrapperLocator_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("peer and one address", func(t *testing.T) {
		var locator config.BootstrapperLocator
		require.NoError(t, locator.UnmarshalText([]byte(first+"@127.0.0.1:6690")))
		assert.Equal(t, commontypes.BootstrapperLocator{
			PeerID: first,
			Addrs:  []string{"127.0.0.1:6690"},
		}, locator.ToBootstrapperLocator())
	})

	t.Run("several addresses", func(t *testing.T) {
		var locator config.BootstrapperLocator
		require.NoError(t, locator.UnmarshalText([]byte(first+"@127.0.0.1:6690/chain.link:443")))
		assert.Equal(t, []string{"127.0.0.1:6690", "chain.link:443"}, locator.ToBootstrapperLocator().Addrs)
	})

	t.Run("malformed", func(t *testing.T) {
		var locator config.BootstrapperLocator
		assert.Error(t, locator.UnmarshalText([]byte("not-a-bootstrapper")))
	})

	t.Run("round trips", func(t *testing.T) {
		const text = first + "@127.0.0.1:6690/chain.link:443"
		var locator config.BootstrapperLocator
		require.NoError(t, locator.UnmarshalText([]byte(text)))

		out, err := locator.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, text, string(out))
	})
}

// TestBootstrapperLocators_Flag is the point of the type: a binary declares the
// setting and is handed locators, rather than strings it has to parse.
func TestBootstrapperLocators_Flag(t *testing.T) {
	t.Parallel()

	cfg := struct {
		Bootstrappers config.BootstrapperLocators `usage:"the DON's bootstrap peers"`
	}{}

	cmd := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
	require.NoError(t, flags.RegisterCommandFlags(cmd, &cfg, flags.DefaultTOMLOptions("TEST")))

	cmd.SetArgs([]string{"--bootstrappers", first + "@127.0.0.1:6690," + second + "@127.0.0.1:6691"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, []commontypes.BootstrapperLocator{
		{PeerID: first, Addrs: []string{"127.0.0.1:6690"}},
		{PeerID: second, Addrs: []string{"127.0.0.1:6691"}},
	}, cfg.Bootstrappers.ToBootstrapperLocators())
}
