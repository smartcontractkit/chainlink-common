package ocr3

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type OCR3Capability struct {
	lggr logger.Logger
	loop.Plugin
	factory EncoderFactory
	reportingplugins.PluginProviderServer
}

type EncoderFactory func(config *values.Map) (types.Encoder, error)

func NewOCR3(lggr logger.Logger, factory EncoderFactory) *OCR3Capability {
	return &OCR3Capability{
		lggr:                 lggr,
		Plugin:               loop.Plugin{Logger: lggr},
		factory:              factory,
		PluginProviderServer: reportingplugins.PluginProviderServer{},
	}
}

func (o *OCR3Capability) NewReportingPluginFactory(ctx context.Context, cfg commontypes.ReportingPluginServiceConfig, provider commontypes.PluginProvider, pipelineRunner commontypes.PipelineRunnerService, telemetry commontypes.TelemetryClient, errorLog commontypes.ErrorLog, capabilityRegistry commontypes.CapabilitiesRegistry) (commontypes.OCR3ReportingPluginFactory, error) {
	conf := &config{Logger: o.lggr, EncoderFactory: o.factory}
	factory, err := newFactoryService(conf)
	if err != nil {
		return nil, err
	}

	err = capabilityRegistry.Add(ctx, factory.capability)
	if err != nil {
		return nil, err
	}

	o.SubService(factory)
	return factory, err
}
