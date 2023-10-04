package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	relayertest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestRelayerService(t *testing.T) {
	t.Parallel()
	capRegistry := mocks.NewCapabilitiesRegistry(t)
	relayer := loop.NewRelayerService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginRelayerName, false, 0)
	}, test.ConfigTOML, keystoretest.Keystore, capRegistry)
	hook := relayer.XXXTestHook()
	servicetest.Run(t, relayer)

	t.Run("control", func(t *testing.T) {
		relayertest.Run(t, relayer)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		relayertest.Run(t, relayer)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		relayertest.Run(t, relayer)
	})
}

func TestRelayerService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	relayer := loop.NewRelayerService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginRelayerName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, test.ConfigTOML, keystoretest.Keystore, nil)
	servicetest.Run(t, relayer)

	relayertest.Run(t, relayer)
}

func TestRelayerService_HealthReport(t *testing.T) {
	lggr := logger.Named(logger.Test(t), "Foo")
	capRegistry := mocks.NewCapabilitiesRegistry(t)
	s := loop.NewRelayerService(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return test.HelperProcessCommand{Command: loop.PluginRelayerName}.New()
	}, test.ConfigTOML, keystoretest.Keystore, capRegistry)

	servicetest.AssertHealthReportNames(t, s.HealthReport(),
		"Foo.RelayerService")

	require.NoError(t, s.Start(tests.Context(t)))
	t.Cleanup(func() { require.NoError(t, s.Close()) })

	require.Eventually(t, func() bool { return s.Ready() == nil }, tests.WaitTimeout(t)/2, time.Second, s.Ready())

	servicetest.AssertHealthReportNames(t, s.HealthReport(),
		"Foo.RelayerService",
		"Foo.RelayerService.PluginRelayerClient.ChainRelayerClient",
		"staticRelayer")
}
