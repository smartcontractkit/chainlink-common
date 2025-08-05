package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// PluginSecureMintName is the name for [types.PluginSecureMint]/[NewGRPCPluginSecureMint].
const PluginSecureMintName = "securemint"

func PluginSecureMintHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_SECUREMINT_MAGIC_COOKIE",
		MagicCookieValue: "secure-mint-magic-cookie-value", // Generate unique value
	}
}

type GRPCPluginSecureMint struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer types.PluginSecureMint

	pluginClient *securemint.PluginSecureMintClient
}

func (p *GRPCPluginSecureMint) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return securemint.RegisterPluginSecureMintServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginSecureMint], updated with the new broker and conn.
func (p *GRPCPluginSecureMint) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = securemint.NewPluginSecureMintClient(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)

	return types.PluginSecureMint(p.pluginClient), nil
}

func (p *GRPCPluginSecureMint) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginSecureMintHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginSecureMintName: p},
	}

	return c
}
