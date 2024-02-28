package reportingplugins_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	resources_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources"
	coreapi_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/core_api"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type HelperProcessCommand test.HelperProcessCommand

func (h *HelperProcessCommand) New() *exec.Cmd {
	h.CommandLocation = "../internal/test/cmd/main.go"
	return (test.HelperProcessCommand)(*h).New()
}

func NewHelperProcessCommand(command string) *exec.Cmd {
	h := HelperProcessCommand{
		Command: command,
	}
	return h.New()
}

func TestLOOPPService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Plugin string
	}{
		// A generic plugin with a median provider
		{Plugin: coreapi_test.MedianID},
		// A generic plugin with a plugin provider
		{Plugin: reportingplugins.PluginServiceName},
	}
	for _, ts := range tests {
		looppSvc := reportingplugins.NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
			return NewHelperProcessCommand(ts.Plugin)
		},
			types.ReportingPluginServiceConfig{},
			resources_test.MockConn{},
			resources_test.PipelineRunnerImpl,
			resources_test.TelemetryImpl,
			&resources_test.ErrorLogImpl)
		hook := looppSvc.XXXTestHook()
		servicetest.Run(t, looppSvc)

		t.Run("control", func(t *testing.T) {
			reportingplugin_test.Factory(t, looppSvc)
		})

		t.Run("Kill", func(t *testing.T) {
			hook.Kill()

			// wait for relaunch
			time.Sleep(2 * internal.KeepAliveTickDuration)

			reportingplugin_test.Factory(t, looppSvc)
		})

		t.Run("Reset", func(t *testing.T) {
			hook.Reset()

			// wait for relaunch
			time.Sleep(2 * internal.KeepAliveTickDuration)

			reportingplugin_test.Factory(t, looppSvc)
		})
	}
}

func TestLOOPPService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	looppSvc := reportingplugins.NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: coreapi_test.MedianID,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	},
		types.ReportingPluginServiceConfig{},
		resources_test.MockConn{},
		resources_test.PipelineRunnerImpl,
		resources_test.TelemetryImpl,
		&resources_test.ErrorLogImpl)
	servicetest.Run(t, looppSvc)

	reportingplugin_test.Factory(t, looppSvc)
}
