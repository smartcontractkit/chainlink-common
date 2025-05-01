package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	looptypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// PluginRelayerName is the name for [types.PluginRelayer]/[NewGRPCPluginRelayer].
const PluginRelayerName = "relayer"

type PluginRelayer = looptypes.PluginRelayer

func PluginRelayerHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_RELAYER_MAGIC_COOKIE",
		MagicCookieValue: "dae753d4542311b33cf041b930db0150647e806175c2818a0c88a9ab745e45aa",
	}
}

// Deprecated
type Keystore = core.Keystore

type Relayer = looptypes.Relayer

type BrokerConfig = net.BrokerConfig

var _ plugin.GRPCPlugin = (*GRPCPluginRelayer)(nil)

// GRPCPluginRelayer implements [plugin.GRPCPlugin] for [types.PluginRelayer].
type GRPCPluginRelayer struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer PluginRelayer

	pluginClient *relayer.PluginRelayerClient
}

func (p *GRPCPluginRelayer) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return relayer.RegisterPluginRelayerServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginRelayer], updated with the new broker and conn.

func (p *GRPCPluginRelayer) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = relayer.NewPluginRelayerClient(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)
	return PluginRelayer(p.pluginClient), nil
}

func (p *GRPCPluginRelayer) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginRelayerHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginRelayerName: p},
	}
	if p.pluginClient == nil {
		p.pluginClient = relayer.NewPluginRelayerClient(p.BrokerConfig)
	}
	return ManagedGRPCClientConfig(c, p.pluginClient.BrokerConfig)
}
