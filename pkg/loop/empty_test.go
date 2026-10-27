package loop_test

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestPluginEmpty(t *testing.T) {
	t.Parallel()

	srv := servicetest.Run(t, services.Config{Name: "EmptyServer"}.NewService(logger.Test(t)))
	stopCh := newStopCh(t)
	test.PluginTest(t, loop.PluginEmptyName,
		&loop.EmptyLoop{
			PluginServer: srv,
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh}},
		func(t *testing.T, s services.Service) {
			assertEmptyHealthy(t, s)
		})
}

func TestRunningEmptyPluginOutOfProcess(t *testing.T) {
	t.Parallel()
	stopCh := newStopCh(t)

	s := newOutOfProcessEmptyService(t, stopCh)

	assertEmptyHealthy(t, s)
}

func assertEmptyHealthy(t *testing.T, s services.Service) {
	t.Helper()
	require.NoError(t, s.Ready())
	hr := s.HealthReport()
	servicetest.AssertHealthReportNames(t, hr, "EmptyClient", "EmptyClient.EmptyServer")
	for name, err := range hr {
		assert.NoError(t, err, name)
	}
}

func newOutOfProcessEmptyService(t *testing.T, stopCh <-chan struct{}) services.Service {
	el := loop.EmptyLoop{BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}
	cc := el.ClientConfig()
	cc.SkipHostEnv = true
	cc.Cmd = NewHelperProcessCommand(loop.PluginEmptyName, false, 0)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	client, err := c.Client()
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping())
	i, err := client.Dispense(loop.PluginEmptyName)
	require.NoError(t, err)
	return i.(services.Service)
}
