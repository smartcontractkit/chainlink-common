package goplugin

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

type PluginClient struct {
	net.AtomicBroker
	net.AtomicClient
	*net.BrokerExt
}

// NewPluginClient creates a *PluginClient. Refresh must be called to initialize the net.Broker and *grpc.ClientConn.
func NewPluginClient(brokerCfg net.BrokerConfig) *PluginClient {
	var pc PluginClient
	pc.BrokerExt = &net.BrokerExt{Broker: &pc.AtomicBroker, BrokerConfig: brokerCfg}
	return &pc
}

func (p *PluginClient) Refresh(broker net.Broker, conn *grpc.ClientConn) {
	p.AtomicBroker.Store(broker)
	p.AtomicClient.Store(conn)
	p.Logger.Debugw("Refreshed PluginClient connection", "state", conn.GetState())
}

// GRPCClientConn is implemented by clients to expose their connection for efficient proxying.
type GRPCClientConn interface {
	// ClientConn returns the underlying client connection.
	ClientConn() grpc.ClientConnInterface
}
