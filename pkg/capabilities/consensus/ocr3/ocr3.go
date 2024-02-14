package consensus

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type OCR3 struct {
	lggr logger.Logger
}

func NewOCR3(lggr logger.Logger) *OCR3 {
	return &OCR3{lggr: lggr}
}

func (o *OCR3) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.PluginProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog) (types.OCR3ReportingPluginFactory, error) {
	factory, err := newFactoryService(nil)
	// TODO capabilityRegistry.Add(factory.capability)
	return factory, err
}
