// Package plugintest runs a go-plugin plugin in-process for tests via go-plugin's reattach
// mechanism: no subprocess, and therefore no binary checksum verification.
package plugintest

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"
)

// PluginTest serves p in-process under name, runs testFn against it, and asserts a clean shutdown.
func PluginTest[TB testing.TB, I any](tb TB, name string, p plugin.Plugin, testFn func(TB, I)) {
	ctx, cancel := context.WithCancel(tb.Context())
	defer cancel()

	ch := make(chan *plugin.ReattachConfig, 1)
	closeCh := make(chan struct{})
	go plugin.Serve(&plugin.ServeConfig{
		Test: &plugin.ServeTestConfig{
			Context:          ctx,
			ReattachConfigCh: ch,
			CloseCh:          closeCh,
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Plugins:    map[string]plugin.Plugin{name: p},
	})

	// We should get a config
	var config *plugin.ReattachConfig
	select {
	case config = <-ch:
	case <-time.After(5 * time.Second):
		tb.Fatal("should've received reattach")
	}
	require.NotNil(tb, config)

	c := plugin.NewClient(&plugin.ClientConfig{
		Reattach:    config,
		Plugins:     map[string]plugin.Plugin{name: p},
		SkipHostEnv: true,
	})
	tb.Cleanup(c.Kill)
	clientProtocol, err := c.Client()
	require.NoError(tb, err)
	defer clientProtocol.Close()
	i, err := clientProtocol.Dispense(name)
	require.NoError(tb, err)

	testFn(tb, i.(I))

	// stop plugin
	cancel()
	select {
	case <-closeCh:
	case <-time.After(5 * time.Second):
		tb.Fatal("should've stopped")
	}
	require.Error(tb, clientProtocol.Ping())
}
