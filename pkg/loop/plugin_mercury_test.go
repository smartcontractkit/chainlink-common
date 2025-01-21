package loop_test

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	mercurytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/test"
	relayertest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestPluginMercury(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	stopCh := newStopCh(t)
	test.PluginTest(t, loop.PluginMercuryName, &loop.GRPCPluginMercury{PluginServer: mercurytest.FactoryServer(lggr), BrokerConfig: loop.BrokerConfig{Logger: lggr, StopCh: stopCh}}, mercurytest.PluginMercury)

	t.Run("proxy", func(t *testing.T) {
		test.PluginTest(t, loop.PluginRelayerName,
			&loop.GRPCPluginRelayer{
				PluginServer: relayertest.NewPluginRelayer(lggr, false),
				BrokerConfig: loop.BrokerConfig{Logger: lggr, StopCh: stopCh}},
			func(t *testing.T, pr loop.PluginRelayer) {
				p := newMercuryProvider(t, pr)
				pm := mercurytest.PluginMercuryTest{MercuryProvider: p}
				test.PluginTest(t, loop.PluginMercuryName,
					&loop.GRPCPluginMercury{PluginServer: mercurytest.FactoryServer(lggr),
						BrokerConfig: loop.BrokerConfig{Logger: lggr, StopCh: stopCh}},
					pm.TestPluginMercury)
			})
	})
}

func TestPluginMercuryExec(t *testing.T) {
	t.Parallel()
	stopCh := newStopCh(t)
	mercury := loop.GRPCPluginMercury{BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}
	cc := mercury.ClientConfig()
	cc.Cmd = NewHelperProcessCommand(loop.PluginMercuryName, true, 0)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	client, err := c.Client()
	require.NoError(t, err)
	defer client.Close()
	require.NoError(t, client.Ping())

	i, err := client.Dispense(loop.PluginMercuryName)
	require.NoError(t, err)
	require.NotNil(t, i)
	mercurytest.PluginMercury(t, i.(types.PluginMercury))

	t.Run("proxy", func(t *testing.T) {
		pr := newPluginRelayerExec(t, true, stopCh)
		p := newMercuryProvider(t, pr)
		pm := mercurytest.PluginMercuryTest{MercuryProvider: p}
		pm.TestPluginMercury(t, i.(types.PluginMercury))
	})
}

func newMercuryProvider(t *testing.T, pr loop.PluginRelayer) types.MercuryProvider {
	ctx := tests.Context(t)
	r, err := pr.NewRelayer(ctx, test.ConfigTOML, keystoretest.Keystore, nil)
	require.NoError(t, err)
	servicetest.Run(t, r)
	p, err := r.NewPluginProvider(ctx, mercurytest.RelayArgs, mercurytest.PluginArgs)
	mp, ok := p.(types.MercuryProvider)
	require.True(t, ok)
	require.NoError(t, err)
	servicetest.Run(t, mp)
	return mp
}
