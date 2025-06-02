package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ ocrtypes.ReportingPluginFactory = (*SecureMintService)(nil)

// TODO(gg): this should use the secure mint plugin from the repo, I think?

// SecureMintService is a [types.Service] that maintains an internal [types.PluginSecureMint].
type SecureMintService struct {
	goplugin.PluginService[*GRPCPluginSecureMint, types.ReportingPluginFactory]
}

// NewSecureMintService returns a new [*SecureMintService].
// cmd must return a new exec.Cmd each time it is called.
func NewSecureMintService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.SecureMintProvider, contractAddress string, dataSource, juelsPerFeeCoin, gasPriceSubunits median.DataSource, errorLog core.ErrorLog, deviationFuncDefinition map[string]any) *SecureMintService {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(core.PluginSecureMint)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginSecureMint but got %T", instance)
		}
		factory, err := plug.NewSecureMintFactory(ctx, provider, contractAddress, dataSource, juelsPerFeeCoin, gasPriceSubunits, errorLog, deviationFuncDefinition)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "SecureMintService")
	var sms SecureMintService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	sms.Init(PluginSecureMintName, &GRPCPluginSecureMint{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &sms
}

func (m *SecureMintService) NewReportingPlugin(ctx context.Context, config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(ctx, config)
}
