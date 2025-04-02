package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	pipelinetest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/pipeline/test"
	telemetrytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/telemetry/test"
	validationtest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/validation/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	mediantest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/median/test"
	ocr2test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr2/test"
	reportingplugintest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

func MedianProviderServer(lggr logger.Logger) medianFactoryServer {
	return newMedianFactoryServer(lggr, medianGeneratorConfig{
		medianProvider:    mediantest.MedianProvider(lggr),
		pipeline:          pipelinetest.PipelineRunner,
		telemetry:         telemetrytest.Telemetry,
		validationService: validationtest.ValidationService,
	})
}

const MedianID = "ocr2-reporting-plugin-with-median-provider"

type medianGeneratorConfig struct {
	medianProvider    testtypes.MedianProviderTester
	pipeline          testtypes.Evaluator[core.PipelineRunnerService]
	telemetry         testtypes.Evaluator[core.TelemetryClient]
	validationService testtypes.ValidationEvaluator
}

type medianFactoryServer struct {
	services.Service
	medianGeneratorConfig
	factory types.ReportingPluginFactory
}

func newMedianFactoryServer(lggr logger.Logger, cfg medianGeneratorConfig) medianFactoryServer {
	lggr = logger.Named(lggr, "medianFactoryServer")
	return medianFactoryServer{
		Service:               test.NewStaticService(lggr),
		medianGeneratorConfig: cfg,
		factory:               reportingplugintest.Factory(lggr),
	}
}

var _ reportingplugins.ProviderServer[types.MedianProvider] = medianFactoryServer{}

func (s medianFactoryServer) NewValidationService(ctx context.Context) (core.ValidationService, error) {
	return s.validationService, nil
}

func (s medianFactoryServer) ConnToProvider(conn grpc.ClientConnInterface, broker net.Broker, brokerConfig net.BrokerConfig) types.MedianProvider {
	return s.medianProvider
}

func (s medianFactoryServer) NewReportingPluginFactory(ctx context.Context, config core.ReportingPluginServiceConfig,
	provider types.MedianProvider, pipelineRunner core.PipelineRunnerService, telemetry core.TelemetryClient,
	errorLog core.ErrorLog, keyValueStore core.KeyValueStore, relayerSet core.RelayerSet) (types.ReportingPluginFactory, error) {
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

	return s.factory, nil
}

func AgnosticProviderServer(lggr logger.Logger) agnosticPluginFactoryServer {
	lggr = logger.Named(lggr, "agnosticPluginFactoryServer")
	return agnosticPluginFactoryServer{
		Service:           test.NewStaticService(lggr),
		provider:          ocr2test.AgnosticPluginProvider(lggr),
		pipelineRunner:    pipelinetest.PipelineRunner,
		telemetry:         telemetrytest.Telemetry,
		validationService: validationtest.ValidationService,
		factory:           reportingplugintest.Factory(lggr),
	}
}

var _ reportingplugins.ProviderServer[types.PluginProvider] = agnosticPluginFactoryServer{}

type agnosticPluginFactoryServer struct {
	services.Service
	provider          testtypes.PluginProviderTester
	pipelineRunner    testtypes.PipelineEvaluator
	telemetry         testtypes.TelemetryEvaluator
	validationService testtypes.ValidationEvaluator
	factory           types.ReportingPluginFactory
}

func (s agnosticPluginFactoryServer) NewValidationService(ctx context.Context) (core.ValidationService, error) {
	return s.validationService, nil
}

func (s agnosticPluginFactoryServer) ConnToProvider(conn grpc.ClientConnInterface, broker net.Broker, brokerConfig net.BrokerConfig) types.PluginProvider {
	return s.provider
}

func (s agnosticPluginFactoryServer) NewReportingPluginFactory(ctx context.Context, config core.ReportingPluginServiceConfig,
	provider types.PluginProvider, pipelineRunner core.PipelineRunnerService, telemetry core.TelemetryClient,
	errorLog core.ErrorLog, keyValueStore core.KeyValueStore, relayerSet core.RelayerSet) (types.ReportingPluginFactory, error) {
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

	return s.factory, nil
}
