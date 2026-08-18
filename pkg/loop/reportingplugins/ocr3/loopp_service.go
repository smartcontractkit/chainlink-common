package ocr3

import (
	"context"
	"fmt"
	"os/exec"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type LOOPPService struct {
	goplugin.PluginService[*GRPCService[types.PluginProvider], core.OCR3ReportingPluginFactory]
}

func NewLOOPPService(
	lggr logger.Logger,
	grpcOpts loop.GRPCOpts,
	cmd func() *exec.Cmd,
	config core.ReportingPluginServiceConfig,
	providerConn grpc.ClientConnInterface,
	pipelineRunner core.PipelineRunnerService,
	telemetryService core.TelemetryService,
	errorLog core.ErrorLog,
	capRegistry core.CapabilitiesRegistry,
	keyValueStore core.KeyValueStore,
	relayerSet core.RelayerSet,
) *LOOPPService {
	newService := func(ctx context.Context, instance any) (core.OCR3ReportingPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(core.OCR3ReportingPluginClient)
		if !ok {
			return nil, nil, fmt.Errorf("expected OCR3ReportingPluginClient but got %T", instance)
		}
		factory, err := plug.NewReportingPluginFactory(ctx, config, providerConn, pipelineRunner, telemetryService, errorLog, capRegistry, keyValueStore, relayerSet)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}

	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "OCR3GenericService")
	var ps LOOPPService
	broker := net.BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ps.Init(reportingplugins.PluginServiceName, &GRPCService[types.PluginProvider]{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ps
}

func (g *LOOPPService) NewReportingPlugin(ctx context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	if err := g.WaitCtx(ctx); err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	return g.Service.NewReportingPlugin(ctx, config)
}

func NewLOOPPServiceValidation(
	lggr logger.Logger,
	grpcOpts loop.GRPCOpts,
	cmd func() *exec.Cmd,
) *reportingplugins.LOOPPServiceValidation {
	newService := func(ctx context.Context, instance any) (core.ValidationService, services.HealthReporter, error) {
		plug, ok := instance.(core.OCR3ReportingPluginClient)
		if !ok {
			return nil, nil, fmt.Errorf("expected ValidationServiceClient but got %T", instance)
		}
		factory, err := plug.NewValidationService(ctx)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "GenericService")
	var ps reportingplugins.LOOPPServiceValidation
	broker := net.BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ps.Init(PluginServiceName, &reportingplugins.GRPCService[types.PluginProvider]{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ps
}
