package ocr3

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	commontypes "github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type OCR3Capability struct {
	lggr logger.Logger
	loop.Plugin
	factory EncoderFactory
}

type EncoderFactory func(config *values.Map) (types.Encoder, error)

func NewOCR3(lggr logger.Logger, factory EncoderFactory) *OCR3Capability {
	return &OCR3Capability{
		lggr:    lggr,
		Plugin:  loop.Plugin{Logger: lggr},
		factory: factory,
	}
}

func (o *OCR3Capability) NewReportingPluginFactory(ctx context.Context, cfg commontypes.ReportingPluginServiceConfig, provider commontypes.PluginProvider, pipelineRunner commontypes.PipelineRunnerService, telemetry commontypes.TelemetryClient, errorLog commontypes.ErrorLog) (commontypes.OCR3ReportingPluginFactory, error) {
	conf := &config{Logger: o.lggr, EncoderFactory: o.factory}
	factory, err := newFactoryService(conf)
	// TODO capabilityRegistry.Add(factory.capability)
	o.SubService(factory)
	return factory, err
}
