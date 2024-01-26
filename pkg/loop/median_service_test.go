package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestMedianService(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	median := loop.NewMedianService(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMedianName)
	}, test.NewStaticMedianProvider(lggr), test.StaticDataSource(), test.StaticJuelsPerFeeCoinDataSource(), &test.StaticErrorLog{})
	hook := median.PluginService.XXXTestHook()
	servicetest.Run(t, median)

	t.Run("control", func(t *testing.T) {
		test.ReportingPluginFactory(t, median)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		test.ReportingPluginFactory(t, median)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		test.ReportingPluginFactory(t, median)
	})
}

func TestMedianService_recovery(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	var limit atomic.Int32
	median := loop.NewMedianService(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return HelperProcessCommand{
			Command: loop.PluginMedianName,
			Limit:   int(limit.Add(1)),
		}.New()
	}, test.NewStaticMedianProvider(lggr), test.StaticDataSource(), test.StaticJuelsPerFeeCoinDataSource(), &test.StaticErrorLog{})
	servicetest.Run(t, median)

	test.ReportingPluginFactory(t, median)
}
