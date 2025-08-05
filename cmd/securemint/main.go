package main

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/por_mock_ocr3plugin/por"
	"github.com/smartcontractkit/libocr/commontypes"
)

func main() {
	lggr, _ := logger.New()

	// Create the plugin server implementation
	pluginServer := &SecureMintPluginServer{
		Logger: lggr,
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: loop.PluginSecureMintHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			loop.PluginSecureMintName: &loop.GRPCPluginSecureMint{
				PluginServer: pluginServer,
				BrokerConfig: loop.BrokerConfig{Logger: lggr},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type SecureMintPluginServer struct {
	logger.Logger
}

func (s *SecureMintPluginServer) Start(ctx context.Context) error {
	return nil
}

func (s *SecureMintPluginServer) Close() error {
	return nil
}

func (s *SecureMintPluginServer) Ready() error {
	return nil
}

func (s *SecureMintPluginServer) HealthReport() map[string]error {
	return nil
}

func (s *SecureMintPluginServer) Name() string {
	return "SecureMintPluginServer"
}

func (s *SecureMintPluginServer) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.SecureMintFactoryGenerator, error) {
	// Create external adapter implementation using Relayer
	externalAdapter := NewRelayerExternalAdapter(provider, s.Logger)
	
	// Create contract reader implementation using Relayer
	contractReader := NewRelayerContractReader(provider, s.Logger)
	
	// Create report marshaler implementation
	reportMarshaler := NewChainlinkReportMarshaler(s.Logger)
	
	// Create a logger adapter for the external plugin
	loggerAdapter := &LoggerAdapter{Logger: s.Logger}
	
	// Create the external plugin factory using the imported por package
	porFactory := &por.PorReportingPluginFactory{
		Logger:          loggerAdapter,
		ExternalAdapter: externalAdapter,
		ContractReader:  contractReader,
		ReportMarshaler: reportMarshaler,
	}
	
	// Wrap the external factory in our LOOPP interface
	factory := NewSecureMintFactory(config, s.Logger, porFactory)
	
	return factory, nil
}

// LoggerAdapter bridges between our logger and the external plugin's logger
type LoggerAdapter struct {
	Logger logger.Logger
}

// Trace implements commontypes.Logger
func (l *LoggerAdapter) Trace(msg string, fields commontypes.LogFields) {
	l.Logger.Debugw(msg, convertFields(fields)...)
}

// Debug implements commontypes.Logger
func (l *LoggerAdapter) Debug(msg string, fields commontypes.LogFields) {
	l.Logger.Debugw(msg, convertFields(fields)...)
}

// Info implements commontypes.Logger
func (l *LoggerAdapter) Info(msg string, fields commontypes.LogFields) {
	l.Logger.Infow(msg, convertFields(fields)...)
}

// Warn implements commontypes.Logger
func (l *LoggerAdapter) Warn(msg string, fields commontypes.LogFields) {
	l.Logger.Warnw(msg, convertFields(fields)...)
}

// Error implements commontypes.Logger
func (l *LoggerAdapter) Error(msg string, fields commontypes.LogFields) {
	l.Logger.Errorw(msg, convertFields(fields)...)
}

// Critical implements commontypes.Logger
func (l *LoggerAdapter) Critical(msg string, fields commontypes.LogFields) {
	l.Logger.Errorw(msg, convertFields(fields)...) // Use Error for Critical since our logger doesn't have Critical
}

// convertFields converts commontypes.LogFields to []any for our logger
func convertFields(fields commontypes.LogFields) []any {
	result := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		result = append(result, k, v)
	}
	return result
}
