package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// PluginMedianName is the name for [types.PluginMedian]/[NewGRPCPluginMedian].
const PluginMercuryName = "mercury"

func PluginMercuryHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey: "CL_PLUGIN_MERCURY_MAGIC_COOKIE",
		// TODO how to generate this?
		MagicCookieValue: "fffa697e19748cd695dd1690c09745ee7cc03717179958e8eadd5a7ca4646000",
	}
}

type GRPCPluginMercury struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer types.PluginMercury

	pluginClient *internal.MercuryAdapterClient
}

func (p *GRPCPluginMercury) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return internal.RegisterMercuryAdapterServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginMercury], updated with the new broker and conn.
func (p *GRPCPluginMercury) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = internal.NewMercuryAdapterClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return types.PluginMercury(p.pluginClient), nil
}

func (p *GRPCPluginMercury) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  PluginMercuryHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{PluginMercuryName: p},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  p.DialOpts,
		Logger:           HCLogLogger(p.Logger),
	}
}
