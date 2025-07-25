package dontime

import (
	"context"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	ocr3rp "github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins/ocr3"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ ocr3rp.ProviderServer[commontypes.PluginProvider] = (*Capability)(nil)

type Capability struct {
	loop.Plugin
	reportingplugins.PluginProviderServer
	config             Config
	capabilityRegistry core.CapabilitiesRegistry
}

type Config struct {
	RequestTimeout *time.Duration
	Logger         logger.Logger
	SendBufferSize int

	store *Store
	//capability *capability
	//clock      clockwork.Clock
}

const (
	defaultSendBufferSize = 10
)

// Needs New ReportingPluginFactor and ValidationService
func NewOCR3(config Config) *Capability {
	if config.RequestTimeout == nil {
		dre := DefaultRequestTimeout
		config.RequestTimeout = &dre
	}

	if config.SendBufferSize == 0 {
		config.SendBufferSize = defaultSendBufferSize
	}

	if config.store == nil {
		config.store = NewStore(DefaultRequestTimeout)
	}

	/*
		if config.capability == nil {
			ci := NewCapability(config.store, config.clock, *config.RequestTimeout, config.Logger,
				config.SendBufferSize)
			config.capability = ci
		}
	*/

	cp := &Capability{
		Plugin:               loop.Plugin{Logger: config.Logger},
		PluginProviderServer: reportingplugins.PluginProviderServer{},
		config:               config,
	}

	//cp.SubService(config.capability)
	return cp
}

func (o *Capability) NewReportingPluginFactory(ctx context.Context, cfg core.ReportingPluginServiceConfig,
	provider commontypes.PluginProvider, pipelineRunner core.PipelineRunnerService, telemetry core.TelemetryClient,
	errorLog core.ErrorLog, capabilityRegistry core.CapabilitiesRegistry, keyValueStore core.KeyValueStore,
	relayerSet core.RelayerSet) (core.OCR3ReportingPluginFactory, error) {
	f, err := NewFactory(o.config.store, o.config.Logger)
	if err != nil {
		return nil, err
	}

	/*
		err = capabilityRegistry.Add(ctx, o.config.capability)
		if err != nil {
			return nil, err
		}

		o.capabilityRegistry = capabilityRegistry
	*/

	return f, err
}

func (o *Capability) NewValidationService(ctx context.Context) (core.ValidationService, error) {
	s := &validationService{lggr: o.Logger}
	o.SubService(s)
	return s, nil
}

func (o *Capability) Close() error {
	o.Plugin.Close()

	/*
		if o.capabilityRegistry == nil {
			return nil
		}
			if err := o.capabilityRegistry.Remove(context.TODO(), o.config.capability.ID); err != nil {
				return err
			}
	*/

	return nil
}
