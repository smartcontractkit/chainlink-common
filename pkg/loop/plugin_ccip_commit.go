package loop

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// PluginCCIPCommitName is the name for [types.PluginCCIPCommit]/[NewGRPCPluginCCIPCommit].
const PluginCCIPCommitName = "ccipcommit"

func PluginCCIPCommitHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_CCIPCOMMIT_MAGIC_COOKIE",
		MagicCookieValue: "abca10d83d68186cdb2076b87465c0e1920b18bc56bf912acdde0b4dc0506dc3",
	}
}

type GRPCPluginCCIPCommit struct {
	plugin.NetRPCUnsupportedPlugin

	BrokerConfig

	PluginServer core.PluginCCIPCommit

	pluginClient *ccip.PluginCCIPCommitClient
}

func (p *GRPCPluginCCIPCommit) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return ccip.RegisterPluginCCIPCommitServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginCCIPCommit], updated with the new broker and conn.
func (p *GRPCPluginCCIPCommit) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = ccip.NewPluginCCIPCommitClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return core.PluginCCIPCommit(p.pluginClient), nil
}

func (p *GRPCPluginCCIPCommit) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: PluginCCIPCommitHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginCCIPCommitName: p},
	}
	return ManagedGRPCClientConfig(c, p.BrokerConfig)
}
