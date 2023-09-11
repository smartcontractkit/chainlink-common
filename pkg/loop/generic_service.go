package loop

import (
	"context"
	"fmt"
	"os/exec"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
)

var _ ocrtypes.ReportingPluginFactory = (*GenericService)(nil)

// GenericService is a [types.Service] that maintains an internal [types.PluginGeneric].
type GenericService struct {
	pluginService[*GRPCPluginGeneric, types.ReportingPluginFactory]
}

// NewGenericService returns a new [*GenericService].
// cmd must return a new exec.Cmd each time it is called.
func NewGenericService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, config []byte, provider grpc.ClientConnInterface, errorLog types.ErrorLog) *GenericService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.PluginGeneric)
		if !ok {
			return nil, fmt.Errorf("expected PluginGeneric but got %T", instance)
		}
		return plug.NewGenericServiceFactory(ctx, config, provider, errorLog)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "GenericService")
	var gs GenericService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	gs.init(PluginGenericName, &GRPCPluginGeneric{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &gs
}

func (g *GenericService) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	ctx, cancel := utils.ContextFromChan(g.pluginService.stopCh)
	defer cancel()
	if err := g.wait(ctx); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return g.service.NewReportingPlugin(config)
}
