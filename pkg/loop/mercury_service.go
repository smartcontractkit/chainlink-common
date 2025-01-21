package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	mercury_v1_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v1"
	mercury_v2_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v2"
	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
	mercury_v4_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v4"
)

var _ ocr3types.MercuryPluginFactory = (*MercuryV4Service)(nil)

// MercuryV4Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV4Service struct {
	goplugin.PluginService[*GRPCPluginMercury, types.MercuryPluginFactory]
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV4Service)(nil)

// NewMercuryV4Service returns a new [*MercuryV4Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV4Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v4_types.DataSource) *MercuryV4Service {
	newService := func(ctx context.Context, instance any) (types.MercuryPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginMercury but got %T", instance)
		}
		factory, err := plug.NewMercuryV4Factory(ctx, provider, dataSource)
		if err != nil {
			return nil, nil, err

		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV3")
	var ms MercuryV4Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMercuryName, &GRPCPluginMercury{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV4Service) NewMercuryPlugin(ctx context.Context, config ocr3types.MercuryPluginConfig) (ocr3types.MercuryPlugin, ocr3types.MercuryPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}
	return m.Service.NewMercuryPlugin(ctx, config)
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV3Service)(nil)

// MercuryV3Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV3Service struct {
	goplugin.PluginService[*GRPCPluginMercury, types.MercuryPluginFactory]
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV3Service)(nil)

// NewMercuryV3Service returns a new [*MercuryV3Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV3Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v3_types.DataSource) *MercuryV3Service {
	newService := func(ctx context.Context, instance any) (types.MercuryPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginMercury but got %T", instance)
		}
		factory, err := plug.NewMercuryV3Factory(ctx, provider, dataSource)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV3")
	var ms MercuryV3Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMercuryName, &GRPCPluginMercury{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV3Service) NewMercuryPlugin(ctx context.Context, config ocr3types.MercuryPluginConfig) (ocr3types.MercuryPlugin, ocr3types.MercuryPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}
	return m.Service.NewMercuryPlugin(ctx, config)
}

// MercuryV1Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV1Service struct {
	goplugin.PluginService[*GRPCPluginMercury, types.MercuryPluginFactory]
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV1Service)(nil)

// NewMercuryV1Service returns a new [*MercuryV1Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV1Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v1_types.DataSource) *MercuryV1Service {
	newService := func(ctx context.Context, instance any) (types.MercuryPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginMercury but got %T", instance)
		}
		factory, err := plug.NewMercuryV1Factory(ctx, provider, dataSource)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV1")
	var ms MercuryV1Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMercuryName, &GRPCPluginMercury{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV1Service) NewMercuryPlugin(ctx context.Context, config ocr3types.MercuryPluginConfig) (ocr3types.MercuryPlugin, ocr3types.MercuryPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}
	return m.Service.NewMercuryPlugin(ctx, config)
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV1Service)(nil)

// MercuryV2Service is a [types.Service] that maintains an internal [types.PluginMedian].
type MercuryV2Service struct {
	goplugin.PluginService[*GRPCPluginMercury, types.MercuryPluginFactory]
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV2Service)(nil)

// NewMercuryV2Service returns a new [*MercuryV2Service].
// cmd must return a new exec.Cmd each time it is called.
func NewMercuryV2Service(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd, provider types.MercuryProvider, dataSource mercury_v2_types.DataSource) *MercuryV2Service {
	newService := func(ctx context.Context, instance any) (types.MercuryPluginFactory, services.HealthReporter, error) {
		plug, ok := instance.(types.PluginMercury)
		if !ok {
			return nil, nil, fmt.Errorf("expected PluginMercury but got %T", instance)
		}
		factory, err := plug.NewMercuryV2Factory(ctx, provider, dataSource)
		if err != nil {
			return nil, nil, err
		}
		return factory, plug, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "MercuryV2")
	var ms MercuryV2Service
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	ms.Init(PluginMercuryName, &GRPCPluginMercury{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &ms
}

func (m *MercuryV2Service) NewMercuryPlugin(ctx context.Context, config ocr3types.MercuryPluginConfig) (ocr3types.MercuryPlugin, ocr3types.MercuryPluginInfo, error) {
	if err := m.WaitCtx(ctx); err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}
	return m.Service.NewMercuryPlugin(ctx, config)
}

var _ ocr3types.MercuryPluginFactory = (*MercuryV2Service)(nil)
