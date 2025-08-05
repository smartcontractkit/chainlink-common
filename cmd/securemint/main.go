package main

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
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
	_ = NewRelayerExternalAdapter(provider, s.Logger)

	// Create contract reader implementation using Relayer
	_ = NewRelayerContractReader(provider, s.Logger)

	// Create report marshaler implementation
	_ = NewChainlinkReportMarshaler(s.Logger)

	// Create the external plugin factory using the imported por package
	// TODO(gg): Import and use the external por package when available
	// porFactory := &por.PorReportingPluginFactory{
	//     Logger:          s.Logger,
	//     ExternalAdapter: externalAdapter,
	//     ContractReader:  contractReader,
	//     ReportMarshaler: reportMarshaler,
	// }

	// Wrap the external factory in our LOOPP interface
	factory := NewSecureMintFactory(config, s.Logger)

	return factory, nil
}
