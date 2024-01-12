package transport

import (
	"google.golang.org/grpc"
)

type PluginClient struct {
	atomicBroker
	atomicClient
	//*brokerExt
	BrokerConfig
}

func NewPluginClient(broker Broker, brokerCfg BrokerConfig, conn *grpc.ClientConn) *PluginClient {
	var pc PluginClient
	//pc.brokeBrokerConfig
	pc.BrokerConfig = brokerCfg
	pc.Refresh(broker, conn)
	return &pc
}

func (p *PluginClient) Refresh(broker Broker, conn *grpc.ClientConn) {
	p.atomicBroker.store(broker)
	p.atomicClient.store(conn)
	p.Logger.Debugw("Refreshed pluginClient connection", "state", conn.GetState())
}

// GRPCClientConn is implemented by clients to expose their connection for efficient proxying.
type GRPCClientConn interface {
	// ClientConn returns the underlying client connection.
	ClientConn() grpc.ClientConnInterface
}
