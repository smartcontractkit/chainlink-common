package loop

import (
	"context"
	"fmt"
	"os/exec"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ ocrtypes.ReportingPluginFactory = (*SecureMintService)(nil)

// SecureMintService is a [types.Service] that maintains an internal [types.SecureMintFactoryGenerator].
type SecureMintService struct {
	goplugin.PluginService[*GRPCPluginSecureMint, types.ReportingPluginFactory]
}

// NewSecureMintService returns a new [*SecureMintService].
// cmd must return a new exec.Cmd each time it is called.
func NewSecureMintService(
	lggr logger.Logger,
	grpcOpts GRPCOpts,
	cmd func() *exec.Cmd,
	provider types.SecureMintProvider,
	config types.SecureMintConfig,
) *SecureMintService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(types.PluginSecureMint)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginSecureMint but got %T", instance)
		}
		factoryGenerator, err := plug.NewSecureMintFactory(ctx, provider, config)
		if err != nil {
			return nil, nil, err
		}
		factory, err := factoryGenerator.NewSecureMintFactory(ctx, provider, config)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "SecureMintService")
	var ss SecureMintService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ss.Init(PluginSecureMintName, &GRPCPluginSecureMint{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ss
}

func (s *SecureMintService) NewReportingPlugin(ctx context.Context, config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := s.WaitCtx(ctx); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return s.Service.NewReportingPlugin(ctx, config)
} 