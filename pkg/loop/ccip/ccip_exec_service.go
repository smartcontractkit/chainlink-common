package ccip

import (
	"context"
	"fmt"
	"os/exec"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ ocrtypes.ReportingPluginFactory = (*ExecFactoryService)(nil)

// ExecFactoryService is a [types.Service] that maintains an internal [types.PluginMedian].
type ExecFactoryService struct {
	internal.PluginService[*ExecLoop, types.ReportingPluginFactory]
}

// NewExecService returns a new [*ExecFactoryService].
// cmd must return a new exec.Cmd each time it is called.
func NewExecService(lggr logger.Logger, grpcOpts loop.GRPCOpts, cmd func() *exec.Cmd, provider types.CCIPExecProvider, config types.CCIPExecFactoryGeneratorConfig) *ExecFactoryService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.CCIPExecFactoryGenerator)
		if !ok {
			return nil, fmt.Errorf("expected PluginMedian but got %T", instance)
		}
		return plug.NewExecFactory(ctx, provider, config)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MedianService")
	var efs ExecFactoryService
	broker := loop.BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	efs.Init(ExecName, &ExecLoop{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &efs
}

func (m *ExecFactoryService) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := m.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(config)
}
