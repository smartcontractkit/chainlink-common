package loop_test

import (
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

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

var relayerServiceNames = []string{
	"RelayerService",
	"RelayerService.PluginRelayerClient",
	"RelayerService.PluginRelayerClient.staticPluginRelayer",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticCommitProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticConfigProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticExecProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticMercuryProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticMedianProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticPluginProvider",
	"RelayerService.PluginRelayerClient.staticPluginRelayer.staticRelayer.staticPluginProvider.staticPluginProvider",
}

func TestRelayerService(t *testing.T) {
	t.Parallel()
	capRegistry := mocks.NewCapabilitiesRegistry(t)
	relayer := loop.NewRelayerService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginRelayerName, false, 0)
	}, test.ConfigTOML, keystoretest.Keystore, keystoretest.Keystore, capRegistry)
	hook := relayer.XXXTestHook()
	servicetest.Run(t, relayer)

	t.Run("control", func(t *testing.T) {
		relayertest.Run(t, relayer)
		servicetest.AssertHealthReportNames(t, relayer.HealthReport(), relayerServiceNames...)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		relayertest.Run(t, relayer)
		servicetest.AssertHealthReportNames(t, relayer.HealthReport(), relayerServiceNames...)

	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		relayertest.Run(t, relayer)
		servicetest.AssertHealthReportNames(t, relayer.HealthReport(), relayerServiceNames...)

	})
}

func TestRelayerService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	relayer := loop.NewRelayerService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return HelperProcessCommand{
			Command: loop.PluginRelayerName,
			Limit:   int(limit.Add(1)),
		}.New()
	}, test.ConfigTOML, keystoretest.Keystore, keystoretest.Keystore, nil)
	servicetest.Run(t, relayer)

	relayertest.Run(t, relayer)

	if hp := relayer.HealthReport(); len(hp) == 2 {
		servicetest.AssertHealthReportNames(t, hp, relayerServiceNames[:2]...)
	} else {
		servicetest.AssertHealthReportNames(t, hp, relayerServiceNames...)
	}

}

func TestRelayerService_HealthReport(t *testing.T) {
	t.Parallel()

	lggr, obsLogs := logger.TestObserved(t, zapcore.DebugLevel)
	t.Cleanup(AssertLogsObserved(t, obsLogs, relayerServiceNames))
	capRegistry := mocks.NewCapabilitiesRegistry(t)
	s := loop.NewRelayerService(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return HelperProcessCommand{Command: loop.PluginRelayerName}.New()
	}, test.ConfigTOML, keystoretest.Keystore, keystoretest.Keystore, capRegistry)

	servicetest.AssertHealthReportNames(t, s.HealthReport(), relayerServiceNames[0])

	servicetest.Run(t, s)

	require.Eventually(t, func() bool { return s.Ready() == nil }, tests.WaitTimeout(t)/2, time.Second, s.Ready())

	servicetest.AssertHealthReportNames(t, s.HealthReport(), relayerServiceNames...)

}

// AssertLogsObserved records an error for each name which does not have any corresponding log lines.
func AssertLogsObserved(t *testing.T, obsLogs *observer.ObservedLogs, names []string) func() {
	return func() {
		t.Helper()

		obsNames := map[string]struct{}{}
		for _, l := range obsLogs.All() {
			obsNames[l.LoggerName] = struct{}{}
		}
		var failed bool
		for _, n := range names {
			if _, ok := obsNames[n]; !ok {
				t.Errorf("No logs observed for service: %s", n)
				failed = true
			}
		}
		if failed {
			keys := maps.Keys(obsNames)
			slices.Sort(keys)
			t.Logf("Loggers observed:\n%s\n", strings.Join(keys, "\n"))
		}
	}
}
