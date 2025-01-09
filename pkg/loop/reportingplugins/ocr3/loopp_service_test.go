package ocr3

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	errorlogtest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog/test"
	keyvaluestoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keyvalue/test"
	pipelinetest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/pipeline/test"
	relayersettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/relayerset/test"
	ocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3/test"
	telemetrytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/telemetry/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	nettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type HelperProcessCommand test.HelperProcessCommand

func (h *HelperProcessCommand) New() *exec.Cmd {
	h.CommandLocation = "../../internal/test/cmd/main.go"
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
		Plugin       string
		ProviderType string
	}{
		// A generic plugin with a median provider
		{
			Plugin:       ocr3test.OCR3ReportingPluginWithMedianProviderName,
			ProviderType: loop.PluginMedianName,
		},
		// A generic plugin with a plugin provider
		{
			Plugin:       PluginServiceName,
			ProviderType: loop.PluginRelayerName,
		},
	}
	for _, ts := range tests {
		t.Run(ts.Plugin, func(t *testing.T) {
			t.Parallel()
			looppSvc := NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
				return NewHelperProcessCommand(ts.Plugin)
			},
				core.ReportingPluginServiceConfig{},
				nettest.MockConn{},
				pipelinetest.PipelineRunner,
				telemetrytest.Telemetry,
				errorlogtest.ErrorLog,
				core.CapabilitiesRegistry(nil),
				keyvaluestoretest.KeyValueStore{},
				relayersettest.RelayerSet{})
			hook := looppSvc.XXXTestHook()
			servicetest.Run(t, looppSvc)

			t.Run("control", func(t *testing.T) {
				ocr3test.OCR3ReportingPluginFactory(t, looppSvc)
			})

			t.Run("Kill", func(t *testing.T) {
				hook.Kill()

				// wait for relaunch
				time.Sleep(2 * goplugin.KeepAliveTickDuration)

				ocr3test.OCR3ReportingPluginFactory(t, looppSvc)
			})

			t.Run("Reset", func(t *testing.T) {
				hook.Reset()

				// wait for relaunch
				time.Sleep(2 * goplugin.KeepAliveTickDuration)

				ocr3test.OCR3ReportingPluginFactory(t, looppSvc)
			})
		})
	}
}

func TestLOOPPService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	looppSvc := NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: ocr3test.OCR3ReportingPluginWithMedianProviderName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	},
		core.ReportingPluginServiceConfig{},
		nettest.MockConn{},
		pipelinetest.PipelineRunner,
		telemetrytest.Telemetry,
		errorlogtest.ErrorLog,
		core.CapabilitiesRegistry(nil),
		keyvaluestoretest.KeyValueStore{},
		relayersettest.RelayerSet{})
	servicetest.Run(t, looppSvc)

	ocr3test.OCR3ReportingPluginFactory(t, looppSvc)
}
