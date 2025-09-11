package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	securemint "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// GRPCPluginSecureMint implements a go-plugin [plugin.GRPCPlugin] for [core.PluginSecureMint].
type GRPCPluginSecureMint struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer core.PluginSecureMint

	pluginClient *securemint.PluginSecureMintClient
}

var _ plugin.GRPCPlugin = (*GRPCPluginSecureMint)(nil)

// GRPCServer is called by the go-plugin framework. It registers the plugin server with the given broker and server.
func (p *GRPCPluginSecureMint) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return securemint.RegisterPluginSecureMintServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient is called by the go-plugin framework. It returns the pluginClient [types.PluginSecureMint], updated with refreshed broker and conn.
func (p *GRPCPluginSecureMint) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (any, error) {
	if p.pluginClient == nil {
		p.pluginClient = securemint.NewPluginSecureMintClient(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)

	return core.PluginSecureMint(p.pluginClient), nil
}

// ClientConfig is called by the loopp plugin framework to configure the plugin.
func (p *GRPCPluginSecureMint) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginSecureMintHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{core.PluginSecureMintName: p},
	}
	if p.pluginClient == nil {
		p.pluginClient = securemint.NewPluginSecureMintClient(p.BrokerConfig)
	}
	return ManagedGRPCClientConfig(c, p.BrokerConfig)
}

// PluginSecureMintHandshakeConfig is used for making a connection between the loopp plugin client and server.
func PluginSecureMintHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_SECURE_MINT_MAGIC_COOKIE",
		MagicCookieValue: "2cba6293d2ae66563d6838334f8b9c3b11c8d3388a1763835a7104d63f44b932",
	}
}
