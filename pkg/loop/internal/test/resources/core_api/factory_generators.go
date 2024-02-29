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
	test_types "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var MedianProviderServerImpl = medianFactoryGenerator{
	medianGeneratorConfig: medianGeneratorConfig{
		medianProvider: median_test.MedianProvider,
		pipeline:       resources_test.PipelineRunnerImpl,
		telemetry:      resources_test.Telemetry,
	},
}

const MedianID = "ocr2-reporting-plugin-with-median-provider"

type medianGeneratorConfig struct {
	medianProvider test_types.MedianProviderTester
	pipeline       test_types.Evaluator[types.PipelineRunnerService]
	telemetry      test_types.Evaluator[types.TelemetryClient]
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

var AgnosticProviderServerImpl = agnosticPluginGenerator{
	provider:       pluginprovider_test.AgnosticPluginProviderImpl,
	pipelineRunner: resources_test.PipelineRunnerImpl,
	telemetry:      resources_test.Telemetry,
}

var _ reportingplugins.ProviderServer[types.PluginProvider] = agnosticPluginGenerator{}

type agnosticPluginGenerator struct {
	provider       pluginprovider_test.PluginProviderTester
	pipelineRunner test_types.Evaluator[types.PipelineRunnerService]
	telemetry      test_types.TelemetryEvaluator
}

func (s agnosticPluginGenerator) ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig internal.BrokerConfig) types.PluginProvider {
	return s.provider
}

func (s agnosticPluginGenerator) NewReportingPluginFactory(ctx context.Context, config types.ReportingPluginServiceConfig, provider types.PluginProvider, pipelineRunner types.PipelineRunnerService, telemetry types.TelemetryClient, errorLog types.ErrorLog) (types.ReportingPluginFactory, error) {
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
