package internal

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/network"
)

type PluginClient struct {
	network.AtomicBroker
	network.AtomicClient
	*network.BrokerExt
}

func NewPluginClient(broker network.Broker, brokerCfg network.BrokerConfig, conn *grpc.ClientConn) *PluginClient {
	var pc PluginClient
	pc.BrokerExt = &network.BrokerExt{Broker: &pc.AtomicBroker, BrokerConfig: brokerCfg}
	pc.Refresh(broker, conn)
	return &pc
}

func (p *PluginClient) Refresh(broker network.Broker, conn *grpc.ClientConn) {
	p.AtomicBroker.Store(broker)
	p.AtomicClient.Store(conn)
	p.Logger.Debugw("Refreshed PluginClient connection", "state", conn.GetState())
}

// GRPCClientConn is implemented by clients to expose their connection for efficient proxying.
type GRPCClientConn interface {
	// ClientConn returns the underlying client connection.
	ClientConn() grpc.ClientConnInterface
}
