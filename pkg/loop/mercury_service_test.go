package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	mercurytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/test"
	mercuryv1test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/v1/test"
	mercuryv2test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/v2/test"
	mercuryv3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/v3/test"
	mercuryv4test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury/v4/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func TestMercuryV4Service(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	mercuryV4 := loop.NewMercuryV4Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMercuryName, true, 0)
	}, mercurytest.MercuryProvider(lggr), mercuryv4test.DataSource)
	hook := mercuryV4.PluginService.XXXTestHook()
	servicetest.Run(t, mercuryV4)

	t.Run("control", func(t *testing.T) {
		mercurytest.MercuryPluginFactory(t, mercuryV4)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV4)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV4)
	})
}

func TestMercuryV4Service_recovery(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	var limit atomic.Int32
	mercury := loop.NewMercuryV4Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMercuryName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, mercurytest.MercuryProvider(lggr), mercuryv4test.DataSource)
	servicetest.Run(t, mercury)

	mercurytest.MercuryPluginFactory(t, mercury)
}

func TestMercuryV3Service(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	mercuryV3 := loop.NewMercuryV3Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMercuryName, true, 0)
	}, mercurytest.MercuryProvider(lggr), mercuryv3test.DataSource)
	hook := mercuryV3.PluginService.XXXTestHook()
	servicetest.Run(t, mercuryV3)

	t.Run("control", func(t *testing.T) {
		mercurytest.MercuryPluginFactory(t, mercuryV3)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV3)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV3)
	})
}

func TestMercuryV3Service_recovery(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	var limit atomic.Int32
	mercury := loop.NewMercuryV3Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMercuryName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, mercurytest.MercuryProvider(lggr), mercuryv3test.DataSource)
	servicetest.Run(t, mercury)

	mercurytest.MercuryPluginFactory(t, mercury)
}

func TestMercuryV1Service(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	mercuryV1 := loop.NewMercuryV1Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMercuryName, true, 0)
	}, mercurytest.MercuryProvider(lggr), mercuryv1test.DataSource)
	hook := mercuryV1.PluginService.XXXTestHook()
	servicetest.Run(t, mercuryV1)

	t.Run("control", func(t *testing.T) {
		mercurytest.MercuryPluginFactory(t, mercuryV1)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV1)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV1)
	})
}

func TestMercuryV1Service_recovery(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	var limit atomic.Int32
	mercury := loop.NewMercuryV1Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMercuryName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, mercurytest.MercuryProvider(lggr), mercuryv1test.DataSource)
	servicetest.Run(t, mercury)

	mercurytest.MercuryPluginFactory(t, mercury)
}

func TestMercuryV2Service(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	mercuryV2 := loop.NewMercuryV2Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.PluginMercuryName, true, 0)
	}, mercurytest.MercuryProvider(lggr), mercuryv2test.DataSource)
	hook := mercuryV2.PluginService.XXXTestHook()
	servicetest.Run(t, mercuryV2)

	t.Run("control", func(t *testing.T) {
		mercurytest.MercuryPluginFactory(t, mercuryV2)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV2)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		mercurytest.MercuryPluginFactory(t, mercuryV2)
	})
}

func TestMercuryV2Service_recovery(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	var limit atomic.Int32
	mercury := loop.NewMercuryV2Service(lggr, loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.PluginMercuryName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, mercurytest.MercuryProvider(lggr), mercuryv2test.DataSource)
	servicetest.Run(t, mercury)

	mercurytest.MercuryPluginFactory(t, mercury)
}
