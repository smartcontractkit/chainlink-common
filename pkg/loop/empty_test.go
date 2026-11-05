package loop_test

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestEmptyService(t *testing.T) {
	t.Parallel()

	lggr, obsLogs := logger.TestObserved(t, zapcore.DebugLevel)
	s := loop.NewEmptyService(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginEmptyName, false, 0)
	})
	hook := s.XXXTestHook()
	require.NoError(t, s.Start(t.Context()))

	assertEmptyHealthy(t, s)

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()
		assertPluginClosedGracefully(t, obsLogs, 1)

		assertEmptyHealthy(t, s)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()
		assertPluginClosedGracefully(t, obsLogs, 2)

		assertEmptyHealthy(t, s)
	})

	t.Run("Close", func(t *testing.T) {
		require.NoError(t, s.Close())
		assertPluginClosedGracefully(t, obsLogs, 3)
	})
}

func assertEmptyHealthy(t *testing.T, s services.Service) {
	t.Helper()
	require.Eventually(t, func() bool { return s.Ready() == nil }, tests.WaitTimeout(t)/2, 100*time.Millisecond)
	hr := s.HealthReport()
	servicetest.AssertHealthReportNames(t, hr, "EmptyService", "EmptyService.EmptyClient", "EmptyService.EmptyClient.EmptyServer")
	for name, err := range hr {
		assert.NoError(t, err, name)
	}
}

func assertPluginClosedGracefully(t *testing.T, obsLogs *observer.ObservedLogs, closes int) {
	t.Helper()
	serverClosed := func() int {
		return obsLogs.FilterLoggerName("EmptyService.EmptyServer").FilterMessage("Closed").Len()
	}
	require.Eventually(t, func() bool { return serverClosed() == closes }, tests.WaitTimeout(t)/2, 100*time.Millisecond,
		"plugin service closed %d times, want %d", serverClosed(), closes)
	assert.Zero(t, obsLogs.FilterMessage("plugin failed to exit gracefully").Len())
}
