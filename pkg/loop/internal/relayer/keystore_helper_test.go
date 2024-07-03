package relayer

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

const PluginKeystoreName = "keystore"

type GRPCPluginKeystore struct {
	plugin.NetRPCUnsupportedPlugin

	net.BrokerConfig

	PluginServer core.Keystore

	pluginClient *keystoreClient
}

func (p *GRPCPluginKeystore) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	pb.RegisterKeystoreServer(server, &keystoreServer{impl: p.PluginServer})
	return nil
}

func (p *GRPCPluginKeystore) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	p.pluginClient = newKeystoreClient(conn)
	return core.Keystore(p.pluginClient), nil
}
