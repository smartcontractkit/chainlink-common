package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

// PluginGenericName is the name for [types.PluginGeneric]/[NewGRPCPluginGeneric].
const PluginGenericName = "generic"

func PluginGenericHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_GENERIC_MAGIC_COOKIE",
		MagicCookieValue: "2ad981747cd86c4ab3e23170970020fd",
	}
}

type GRPCPluginGeneric struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer types.PluginGeneric

	pluginClient *internal.PluginGenericClient
}

func (p *GRPCPluginGeneric) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return internal.RegisterPluginGenericServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginMedian], updated with the new broker and conn.
func (p *GRPCPluginGeneric) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = internal.NewPluginGenericClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return types.PluginGeneric(p.pluginClient), nil
}

func (p *GRPCPluginGeneric) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  PluginGenericHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{PluginGenericName: p},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  p.DialOpts,
		Logger:           HCLogLogger(p.Logger),
	}
}
