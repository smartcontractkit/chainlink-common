package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

var _ ocr3types.ReportingPluginFactory[core.ChainSelector] = (*PluginSecureMintService)(nil)

// PluginSecureMintService is a [types.Service] that maintains an internal [types.PluginSecureMint].
type PluginSecureMintService struct {
	goplugin.PluginService[*GRPCPluginSecureMint, core.ReportingPluginFactory[core.ChainSelector]]
}

// NewPluginSecureMintService returns a new [*PluginSecureMintService].
// cmd must return a new exec.Cmd each time it is called.
// This is called both client-side and server-side. The implementation of the NewSecureMintFactory differs between the two.
func NewPluginSecureMintService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, externalAdapter core.ExternalAdapter) *PluginSecureMintService {
	newService := func(ctx context.Context, instance any) (core.ReportingPluginFactory[core.ChainSelector], services.HealthReporter, error) {
		lggr.Infof("creating new PluginSecureMintService for client or server: type %T: %+v", instance, instance)
		plug, ok := instance.(core.PluginSecureMint)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginSecureMint but got %T", instance)
		}
		factory, err := plug.NewSecureMintFactory(ctx, lggr, externalAdapter)
		if err != nil {
			return nil, nil, err
		}
		lggr.Infof("created factory of type %T: %+v", factory, factory)

		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "PluginSecureMintService")
	var ms PluginSecureMintService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(core.PluginSecureMintName, &GRPCPluginSecureMint{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *PluginSecureMintService) NewReportingPlugin(ctx context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[core.ChainSelector], ocr3types.ReportingPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(ctx, config)
}
