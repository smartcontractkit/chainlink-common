package reportingplugins

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ ocrtypes.ReportingPluginFactory = (*LOOPPService)(nil)

// LOOPPService is a [types.Service] that maintains an internal [types.PluginClient].
type LOOPPService struct {
	internal.PluginService[*GRPCService[types.PluginProvider], types.ReportingPluginFactory]
}

// NewLOOPPService returns a new [*PluginService].
// cmd must return a new exec.Cmd each time it is called.
// We use a `conn` here rather than a provider so that we can enforce proxy providers being passed in.
func NewLOOPPService(
	lggr logger.Logger,
	grpcOpts loop.GRPCOpts,
	cmd func() *exec.Cmd,
	config types.ReportingPluginServiceConfig,
	providerConn grpc.ClientConnInterface,
	pipelineRunner types.PipelineRunnerService,
	telemetryService types.TelemetryService,
	errorLog types.ErrorLog,
) *LOOPPService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.ReportingPluginClient)
		if !ok {
			return nil, fmt.Errorf("expected GenericPluginClient but got %T", instance)
		}
		return plug.NewReportingPluginFactory(ctx, config, providerConn, pipelineRunner, telemetryService, errorLog)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "GenericService")
	var ps LOOPPService
	broker := internal.BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ps.Init(PluginServiceName, &GRPCService[types.PluginProvider]{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ps
}

func (g *LOOPPService) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := g.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return g.Service.NewReportingPlugin(config)
}

type OCR3LOOPPService struct {
	internal.PluginService[*GRPCService[types.PluginProvider], types.OCR3ReportingPluginFactory]
}

func NewOCR3LOOPPService(
	lggr logger.Logger,
	grpcOpts loop.GRPCOpts,
	cmd func() *exec.Cmd,
	config types.ReportingPluginServiceConfig,
	providerConn grpc.ClientConnInterface,
	pipelineRunner types.PipelineRunnerService,
	telemetryService types.TelemetryService,
	errorLog types.ErrorLog,
) *OCR3LOOPPService {
	newService := func(ctx context.Context, instance any) (types.OCR3ReportingPluginFactory, error) {
		plug, ok := instance.(types.OCR3ReportingPluginClient)
		if !ok {
			return nil, fmt.Errorf("expected GenericPluginClient but got %T", instance)
		}
		return plug.NewReportingPluginFactory(ctx, config, providerConn, pipelineRunner, telemetryService, errorLog)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "OCR3GenericService")
	var ps OCR3LOOPPService
	broker := internal.BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ps.Init(PluginServiceName, &GRPCService[types.PluginProvider]{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ps
}

func (g *OCR3LOOPPService) NewReportingPlugin(config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[any], ocr3types.ReportingPluginInfo, error) {
	if err := g.Wait(); err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	return g.Service.NewReportingPlugin(config)
}
