package reportingplugins

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// PluginServiceName is the name for [types.PluginClient]/[NewGRPCService].
const PluginServiceName = "plugin-service"

func ReportingPluginHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_GENERIC_MAGIC_COOKIE",
		MagicCookieValue: "2ad981747cd86c4ab3e23170970020fd",
	}
}

type ProviderServer[T types.PluginProvider] interface {
	types.ReportingPluginServer[T]
	ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig loop.BrokerConfig) T
}

// GRPCService is the loopp interface for a plugin that can
// run an arbitrary product in the core node. By specifying
// `T`, server instances can request a specific provider type.
type GRPCService[T types.PluginProvider] struct {
	plugin.NetRPCUnsupportedPlugin

	loop.BrokerConfig

	PluginServer ProviderServer[T]

	pluginClient *internal.ReportingPluginServiceClient
}

type serverAdapter[T types.PluginProvider] struct {
	services.StateMachine
	BrokerConfig   loop.BrokerConfig
	ProviderServer ProviderServer[T]
	GRPCBroker     *plugin.GRPCBroker
}

func (s *serverAdapter[T]) Start(ctx context.Context) error { return s.ProviderServer.Start(ctx) }

func (s *serverAdapter[T]) Close() error { return s.ProviderServer.Close() }

func (s *serverAdapter[T]) HealthReport() map[string]error { return s.ProviderServer.HealthReport() }

func (s *serverAdapter[T]) Name() string { return s.ProviderServer.Name() }

func (s *serverAdapter[T]) NewReportingPluginFactory(
	ctx context.Context,
	config types.ReportingPluginServiceConfig,
	conn grpc.ClientConnInterface,
	pr types.PipelineRunnerService,
	ts types.TelemetryService,
	errorLog types.ErrorLog,
) (types.ReportingPluginFactory, error) {
	provider := s.ProviderServer.ConnToProvider(conn, s.GRPCBroker, s.BrokerConfig)
	tc := internal.NewTelemetryClient(ts)
	return s.ProviderServer.NewReportingPluginFactory(ctx, config, provider, pr, tc, errorLog)
}

func (g *GRPCService[T]) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	impl := &serverAdapter[T]{BrokerConfig: g.BrokerConfig, ProviderServer: g.PluginServer, GRPCBroker: broker}
	//TODO when to start
	return internal.RegisterReportingPluginServiceServer(server, broker, g.BrokerConfig, impl)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginClient], updated with the new broker and conn.
func (g *GRPCService[T]) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if g.pluginClient == nil {
		g.pluginClient = internal.NewReportingPluginServiceClient(broker, g.BrokerConfig, conn)
	} else {
		g.pluginClient.Refresh(broker, conn)
	}

	return types.ReportingPluginClient(g.pluginClient), nil
}

func (g *GRPCService[T]) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  ReportingPluginHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{PluginServiceName: g},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  g.BrokerConfig.DialOpts,
		Logger:           loop.HCLogLogger(g.BrokerConfig.Logger),
	}
}

// These implement `ConnToProvider` and return the conn wrapped as
// the specified provider type. They can be embedded into the server struct
// for ease of use.
type PluginProviderServer = internal.PluginProviderServer
type MedianProviderServer = internal.MedianProviderServer
