package ocr3_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	median_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/median"
	pluginprovider_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/plugin_provider"
	resources_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources"
	test_types "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var MedianGeneratorImpl = MedianGenerator{
	medianGeneratorConfig: medianGeneratorConfig{
		medianProvider: median_test.MedianProvider,
		pipeline:       resources_test.PipelineRunnerImpl,
		telemetry:      resources_test.Telemetry,
	},
}

const OCR3ReportingPluginWithMedianProviderName = "ocr3-reporting-plugin-with-median-provider"

type medianGeneratorConfig struct {
	medianProvider test_types.MedianProviderTester
	pipeline       test_types.Evaluator[types.PipelineRunnerService]
	telemetry      test_types.Evaluator[types.TelemetryClient]
}

type MedianGenerator struct {
	medianGeneratorConfig
}

func (s MedianGenerator) ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig internal.BrokerConfig) types.MedianProvider {
	return s.medianProvider
}

func (s MedianGenerator) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.MedianProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog, capRegistry types.CapabilitiesRegistry) (types.OCR3ReportingPluginFactory, error) {
	err := s.medianProvider.Evaluate(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate median provider: %w", err)
	}

	err = s.pipeline.Evaluate(ctx, pipelineRunner)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate pipeline runner: %w", err)
	}

	err = s.telemetry.Evaluate(ctx, telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate telemetry: %w", err)
	}

	return Factory, nil
}

var AgnosticPluginGeneratorImpl = AgnosticPluginGenerator{
	provider:       pluginprovider_test.AgnosticPluginProviderImpl,
	pipelineRunner: resources_test.PipelineRunnerImpl,
	telemetry:      resources_test.Telemetry,
}

type AgnosticPluginGenerator struct {
	provider       pluginprovider_test.PluginProviderTester
	pipelineRunner test_types.Evaluator[types.PipelineRunnerService]
	telemetry      test_types.Evaluator[types.TelemetryClient]
}

func (s AgnosticPluginGenerator) ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig internal.BrokerConfig) types.PluginProvider {
	return s.provider
}

func (s AgnosticPluginGenerator) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.PluginProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog, capRegistry types.CapabilitiesRegistry) (types.OCR3ReportingPluginFactory, error) {
	err := s.provider.Evaluate(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate agnostic provider: %w", err)
	}

	err = s.pipelineRunner.Evaluate(ctx, pipelineRunner)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate pipeline runner: %w", err)
	}

	err = s.telemetry.Evaluate(ctx, telemetry)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate telemetry: %w", err)
	}

	return Factory, nil
}
