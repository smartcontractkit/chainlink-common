package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	cciptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/ccip/test"
	testreportingplugin "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestExecService(t *testing.T) {
	t.Parallel()

	exec := loop.NewExecutionService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.ExecName, false)
	}, cciptest.ExecProvider, cciptest.ExecConfig)
	hook := exec.PluginService.XXXTestHook()
	servicetest.Run(t, exec)

	t.Run("control", func(t *testing.T) {
		testreportingplugin.RunFactory(t, exec)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		testreportingplugin.RunFactory(t, exec)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		testreportingplugin.RunFactory(t, exec)
	})
}

func TestExecService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	exec := loop.NewExecutionService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.ExecName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, cciptest.ExecProvider, cciptest.ExecConfig)
	servicetest.Run(t, exec)

	testreportingplugin.RunFactory(t, exec)
}
