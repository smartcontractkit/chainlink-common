package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestEmptyService(t *testing.T) {
	var launches atomic.Int32
	svc := loop.NewEmptyService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		launches.Add(1)
		return NewHelperProcessCommand(loop.PluginEmptyName, false, 0)
	})
	hook := svc.XXXTestHook()
	servicetest.Run(t, svc)

	require.NoError(t, svc.WaitCtx(t.Context()))
	require.NoError(t, svc.Ready())
	servicetest.AssertHealthReportNames(t, svc.HealthReport(), "EmptyService", "EmptyService.EmptyClient")
	assert.Equal(t, int32(1), launches.Load())

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		require.Eventually(t, func() bool { return launches.Load() == 2 }, 2*goplugin.KeepAliveTickDuration, 100*time.Millisecond,
			"the service must launch the process again")
		servicetest.AssertHealthReportNames(t, svc.HealthReport(), "EmptyService", "EmptyService.EmptyClient")
	})
}
