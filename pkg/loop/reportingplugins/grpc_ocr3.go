package reportingplugins

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func OCR3ReportingPluginHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_OCR3_PLUGIN_GENERIC_MAGIC_COOKIE",
		MagicCookieValue: "2ad983747cd86c2ab3e23170970020fa",
	}
}

type OCR3ProviderServer[T types.PluginProvider] interface {
	types.OCR3ReportingPluginServer[T]
	ConnToProvider(conn grpc.ClientConnInterface, broker internal.Broker, brokerConfig loop.BrokerConfig) T
}

type OCR3GRPCService[T types.PluginProvider] struct {
	plugin.NetRPCUnsupportedPlugin

	loop.BrokerConfig

	PluginServer OCR3ProviderServer[T]

	pluginClient *internal.OCR3ReportingPluginServiceClient
}

type ocr3serverAdapter func(
	context.Context,
	types.ReportingPluginServiceConfig,
	grpc.ClientConnInterface,
	types.PipelineRunnerService,
	types.TelemetryService,
	types.ErrorLog,
) (types.OCR3ReportingPluginFactory, error)

func (s ocr3serverAdapter) NewReportingPluginFactory(
	ctx context.Context,
	config types.ReportingPluginServiceConfig,
	conn grpc.ClientConnInterface,
	pr types.PipelineRunnerService,
	ts types.TelemetryService,
	errorLog types.ErrorLog,
) (types.OCR3ReportingPluginFactory, error) {
	return s(ctx, config, conn, pr, ts, errorLog)
}

func (g *OCR3GRPCService[T]) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	adapter := func(
		ctx context.Context,
		cfg types.ReportingPluginServiceConfig,
		conn grpc.ClientConnInterface,
		pr types.PipelineRunnerService,
		ts types.TelemetryService,
		el types.ErrorLog,
	) (types.OCR3ReportingPluginFactory, error) {
		provider := g.PluginServer.ConnToProvider(conn, broker, g.BrokerConfig)
		tc := internal.NewTelemetryClient(ts)
		return g.PluginServer.NewReportingPluginFactory(ctx, cfg, provider, pr, tc, el)
	}
	return internal.RegisterOCR3ReportingPluginServiceServer(server, broker, g.BrokerConfig, ocr3serverAdapter(adapter))
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginClient], updated with the new broker and conn.
func (g *OCR3GRPCService[T]) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if g.pluginClient == nil {
		g.pluginClient = internal.NewOCR3ReportingPluginServiceClient(broker, g.BrokerConfig, conn)
	} else {
		g.pluginClient.Refresh(broker, conn)
	}

	return types.OCR3ReportingPluginClient(g.pluginClient), nil
}

func (g *OCR3GRPCService[T]) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  ReportingPluginHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{PluginServiceName: g},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  g.BrokerConfig.DialOpts,
		Logger:           loop.HCLogLogger(g.BrokerConfig.Logger),
	}
}
