package loop_test

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestPluginRelayer(t *testing.T) {
	t.Parallel()

	stopCh := newStopCh(t)
	test.PluginTest(t, loop.PluginRelayerName, &loop.GRPCPluginRelayer{PluginServer: test.StaticPluginRelayer{}, BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}, test.RunPluginRelayer)
}

func TestPluginRelayerExec(t *testing.T) {
	t.Parallel()
	stopCh := newStopCh(t)

	pr := newPluginRelayerExec(t, stopCh)

	test.RunPluginRelayer(t, pr)
}

func FuzzPluginRelayer(f *testing.F) {
	test.RunFuzzPluginRelayer(f, fuzzTestWrapPluginRelayConstructor(f))
}

func FuzzRelayer(f *testing.F) {
	test.RunFuzzRelayer(f, fuzzTestWrapRelayConstructor(f))
}

func fuzzTestWrapPluginRelayConstructor(f *testing.F) func(*testing.T) loop.PluginRelayer {
	f.Helper()

	return func(t *testing.T) loop.PluginRelayer {
		t.Helper()

		stopCh := newStopCh(t)
		relayer := newPluginRelayerExec(t, stopCh)

		return relayer
	}
}

func fuzzTestWrapRelayConstructor(f *testing.F) func(*testing.T) loop.Relayer {
	f.Helper()

	return func(t *testing.T) loop.Relayer {
		t.Helper()

		stopCh := newStopCh(t)
		p := newPluginRelayerExec(t, stopCh)
		ctx := tests.Context(t)
		relayer, err := p.NewRelayer(ctx, test.ConfigTOML, test.StaticKeystore{})

		require.NoError(t, err)

		return relayer
	}
}

func newPluginRelayerExec(t *testing.T, stopCh <-chan struct{}) loop.PluginRelayer {
	relayer := loop.GRPCPluginRelayer{BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}
	cc := relayer.ClientConfig()
	cc.Cmd = NewHelperProcessCommand(loop.PluginRelayerName)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	client, err := c.Client()
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping())
	i, err := client.Dispense(loop.PluginRelayerName)
	require.NoError(t, err)
	return i.(loop.PluginRelayer)
}
