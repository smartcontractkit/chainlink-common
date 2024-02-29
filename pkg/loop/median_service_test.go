package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	median_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/median"
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	resources_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestMedianService(t *testing.T) {
	t.Parallel()

	median := loop.NewMedianService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMedianName, false)
	}, median_test.MedianProvider, median_test.DataSource, median_test.JuelsPerFeeCoinDataSource, &resources_test.ErrorLogImpl)
	hook := median.PluginService.XXXTestHook()
	servicetest.Run(t, median)

	t.Run("control", func(t *testing.T) {
		reportingplugin_test.Factory(t, median)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		reportingplugin_test.Factory(t, median)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		reportingplugin_test.Factory(t, median)
	})
}

func TestMedianService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	median := loop.NewMedianService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMedianName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, median_test.MedianProvider, median_test.DataSource, median_test.JuelsPerFeeCoinDataSource, &resources_test.ErrorLogImpl)
	servicetest.Run(t, median)

	reportingplugin_test.Factory(t, median)
}
