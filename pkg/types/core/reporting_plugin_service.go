package core

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ReportingPluginServiceConfig struct {
	ProviderType  string
	Command       string
	PluginName    string
	TelemetryType string
	PluginConfig  string
}

// ReportingPluginClient is the client interface to a plugin running
// as a generic job (job type = GenericPlugin) inside the core node.
type ReportingPluginClient interface {
	services.Service
	NewReportingPluginFactory(ctx context.Context, config ReportingPluginServiceConfig, grpcProvider grpc.ClientConnInterface, pipelineRunner PipelineRunnerService, telemetry TelemetryService, errorLog ErrorLog, keyValueStore KeyValueStore, relayerSet RelayerSet) (types.ReportingPluginFactory, error)
	NewValidationService(ctx context.Context) (ValidationService, error)
}

// ReportingPluginServer is the server interface to a plugin running
// as a generic job (job type = GenericPlugin) inside the core node,
// with the passthrough provider connection converted to the provider
// expected by the plugin.
type ReportingPluginServer[T types.PluginProvider] interface {
	services.Service
	NewReportingPluginFactory(ctx context.Context, config ReportingPluginServiceConfig, provider T, pipelineRunner PipelineRunnerService, telemetry TelemetryClient, errorLog ErrorLog, keyValueStore KeyValueStore, relayerSet RelayerSet) (types.ReportingPluginFactory, error)
	NewValidationService(ctx context.Context) (ValidationService, error)
}

type OCR3ReportingPluginClient interface {
	services.Service
	NewReportingPluginFactory(ctx context.Context, config ReportingPluginServiceConfig, grpcProvider grpc.ClientConnInterface, pipelineRunner PipelineRunnerService, telemetry TelemetryService, errorLog ErrorLog, capRegistry CapabilitiesRegistry, keyValueStore KeyValueStore, relayerSet RelayerSet) (OCR3ReportingPluginFactory, error)
	NewValidationService(ctx context.Context) (ValidationService, error)
}

type OCR3ReportingPluginServer[T types.PluginProvider] interface {
	services.Service
	NewReportingPluginFactory(ctx context.Context, config ReportingPluginServiceConfig, provider T, pipelineRunner PipelineRunnerService, telemetry TelemetryClient, errorLog ErrorLog, capRegistry CapabilitiesRegistry, keyValueStore KeyValueStore, relayerSet RelayerSet) (OCR3ReportingPluginFactory, error)
	NewValidationService(ctx context.Context) (ValidationService, error)
}

type OCR3ReportingPluginFactory interface {
	services.Service
	ocr3types.ReportingPluginFactory[[]byte]
}
