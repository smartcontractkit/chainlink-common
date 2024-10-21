package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	keystorepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/types/keystore"
)

// PluginKeystoreName is the name for keystore.Keystore
const PluginKeystoreName = "keystore"

func PluginKeystoreHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_KEYSTORE_MAGIC_COOKIE",
		MagicCookieValue: "fe81b132-0d3d-4c16-9f13-c2f7bfd3c361",
	}
}

type GRPCPluginKeystore struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer keystorepb.GRPCService

	pluginClient *keystorepb.Client
}

func (p *GRPCPluginKeystore) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return keystorepb.RegisterKeystoreServer(server, broker, p.BrokerConfig, p.PluginServer)
}

func (p *GRPCPluginKeystore) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = keystorepb.NewKeystoreClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return keystore.Keystore(p.pluginClient), nil
}

func (p *GRPCPluginKeystore) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginKeystoreHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginKeystoreName: p},
	}
	return ManagedGRPCClientConfig(c, p.BrokerConfig)
}
