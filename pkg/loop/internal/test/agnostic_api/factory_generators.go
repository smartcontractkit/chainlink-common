package agnosticapi_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	median_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/median"
	pluginprovider_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/plugin_provider"
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	resources_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var MedianGeneratorImpl = medianFactoryGenerator{
	medianGeneratorConfig: medianGeneratorConfig{
		medianProvider: median_test.MedianProviderImpl,
		pipeline:       resources_test.PipelineRunnerImpl,
		telemetry:      resources_test.TelemetryImpl,
	},
}

const MedianID = "ocr2-reporting-plugin-with-median-provider"

type medianGeneratorConfig struct {
	medianProvider median_test.MedianProviderTester
	pipeline       resources_test.PipelineRunnerEvaluator
	telemetry      resources_test.TelemetryEvaluator
}

type medianFactoryGenerator struct {
	medianGeneratorConfig
}

var _ reportingplugins.ProviderServer[types.MedianProvider] = medianFactoryGenerator{}

func (s medianFactoryGenerator) ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig internal.BrokerConfig) types.MedianProvider {
	return s.medianProvider
}

func (s medianFactoryGenerator) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.MedianProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog) (types.ReportingPluginFactory, error) {
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

	return reportingplugin_test.FactoryImpl, nil
}

var AgnosticPluginGeneratorImpl = AgnosticPluginGenerator{
	provider:       pluginprovider_test.AgnosticPluginProviderImpl,
	pipelineRunner: resources_test.PipelineRunnerImpl,
	telemetry:      resources_test.TelemetryImpl,
}

var _ reportingplugins.ProviderServer[types.PluginProvider] = AgnosticPluginGenerator{}

type AgnosticPluginGenerator struct {
	provider       pluginprovider_test.PluginProviderTester
	pipelineRunner resources_test.PipelineRunnerEvaluator
	telemetry      resources_test.TelemetryEvaluator
}

func (s AgnosticPluginGenerator) ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig internal.BrokerConfig) types.PluginProvider {
	return s.provider
}

func (s AgnosticPluginGenerator) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.PluginProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog) (types.ReportingPluginFactory, error) {
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

	return reportingplugin_test.FactoryImpl, nil
}
