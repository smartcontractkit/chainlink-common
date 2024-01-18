package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	mercury_v3_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v3/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestMercuryV3Service(t *testing.T) {
	t.Parallel()

	mercuryV3 := loop.NewMercuryV3Service(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMercuryName)
	}, test.StaticMercuryProvider{}, mercury_v3_test.StaticDataSource{})
	hook := mercuryV3.PluginService.XXXTestHook()
	servicetest.Run(t, mercuryV3)

	t.Run("control", func(t *testing.T) {
		test.ReportingPluginFactory(t, mercuryV3)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		test.ReportingPluginFactory(t, mercuryV3)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * internal.KeepAliveTickDuration)

		test.ReportingPluginFactory(t, mercuryV3)
	})
}

func TestMercuryV3Service_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	mercury := loop.NewMercuryV3Service(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMercuryName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, test.StaticMercuryProvider{}, mercury_v3_test.StaticDataSource{})
	servicetest.Run(t, mercury)

	test.ReportingPluginFactory(t, mercury)
}
