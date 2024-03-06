package ccip

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ExecName is the name for [types.PluginMedian]/[NewGRPCPluginMedian].
const ExecName = "exec"

func PluginMedianHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_CCIP_EXEC_MAGIC_COOKIE",
		MagicCookieValue: "b12a697e19748cd695dd1690c09745ee7cc03717179958e8eadd5a7ca4699999",
	}
}

type ExecLoop struct {
	plugin.NetRPCUnsupportedPlugin

	loop.BrokerConfig

	PluginServer types.CCIPExecFactoryGenerator

	pluginClient *internal.PluginMedianClient
}

func (p *ExecLoop) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return internal.RegisterPluginMedianServer(server, broker, p.BrokerConfig, p.PluginServer)
}

// GRPCClient implements [plugin.GRPCPlugin] and returns the pluginClient [types.PluginMedian], updated with the new broker and conn.
func (p *ExecLoop) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.pluginClient == nil {
		p.pluginClient = internal.NewPluginMedianClient(broker, p.BrokerConfig, conn)
	} else {
		p.pluginClient.Refresh(broker, conn)
	}

	return types.PluginMedian(p.pluginClient), nil
}

func (p *ExecLoop) ClientConfig() *plugin.ClientConfig {
	return &plugin.ClientConfig{
		HandshakeConfig:  PluginMedianHandshakeConfig(),
		Plugins:          map[string]plugin.Plugin{ExecName: p},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions:  p.DialOpts,
		Logger:           loop.HCLogLogger(p.Logger),
	}
}
