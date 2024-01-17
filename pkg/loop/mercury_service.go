package loop

import (
	"context"
	"fmt"
	"os/exec"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
)

var _ ocrtypes.ReportingPluginFactory = (*MercuryV3Service)(nil)

// MercuryV3Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV3Service struct {
	internal.PluginService[*GRPCPluginMercury, types.ReportingPluginFactory]
}

// NewMercuryV3Service returns a new [*MercuryV3Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV3Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v3_types.DataSource) *MercuryV3Service {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, fmt.Errorf("expected PluginMercury but got %T", instance)
		}
		return plug.NewMercuryV3Factory(ctx, provider, dataSource)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV3")
	var ms MercuryV3Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMercuryName, &GRPCPluginMercury{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV3Service) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := m.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(config)
}

var _ ocrtypes.ReportingPluginFactory = (*MercuryV3Service)(nil)

/*
// MercuryV2Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV2Service struct {
	internal.PluginService[*GRPCPluginMedian, types.ReportingPluginFactory]
}

// NewMercuryV2Service returns a new [*MercuryV2Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV2Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v2_types.DataSource) *MercuryV2Service {
	newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, fmt.Errorf("expected PluginMedian but got %T", instance)
		}
		return plug.NewMercuryV2Factory(ctx, provider, dataSource)
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV3")
	var ms MercuryV2Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMedianName, &GRPCPluginMedian{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV2Service) NewReportingPlugin(config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
	if err := m.Wait(); err != nil {
		return nil, ocrtypes.ReportingPluginInfo{}, err
	}
	return m.Service.NewReportingPlugin(config)
}
*/
