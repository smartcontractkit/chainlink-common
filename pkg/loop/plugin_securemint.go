package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// PluginSecureMintName is the name for [types.PluginSecureMint]/[NewGRPCPluginSecureMint].
const PluginSecureMintName = "secure_mint"

func PluginSecureMintHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_SECURE_MINT_MAGIC_COOKIE",
		MagicCookieValue: "b12a697e19748cd695dd1690c09745ee7cc03717179958e8eadd5a7ca4646720",
	}
}

type GRPCPluginSecureMint struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer core.PluginSecureMint

	pluginClient *securemint.PluginSecureMintClient
}

func (p *GRPCPluginSecureMint) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return median.RegisterPluginMedianServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginMedian], updated with the new broker and conn.
func (p *GRPCPluginSecureMint) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = median.NewPlugin(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)

	return core.PluginMedian(p.pluginClient), nil
}

func (p *GRPCPluginSecureMint) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginSecureMintHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginSecureMintName: p},
	}
	if p.pluginClient == nil {
		p.pluginClient = median.NewPluginSecureMintClient(p.BrokerConfig)
	}
	return ManagedGRPCClientConfig(c, p.pluginClient.BrokerConfig)
}
