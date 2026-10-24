package goplugin

import "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"

// Aliases for the internal net types that appear in this package's public API, so that
// consumers outside of pkg/loop can name them.
type (
	AtomicBroker        = net.AtomicBroker
	AtomicClient        = net.AtomicClient
	Broker              = net.Broker
	BrokerConfig        = net.BrokerConfig
	BrokerExt           = net.BrokerExt
	ClientConnInterface = net.ClientConnInterface
	GRPCOpts            = net.GRPCOpts
	Resource            = net.Resource
	Resources           = net.Resources
)

var ClientConnInterfaceFromGRPC = net.ClientConnInterfaceFromGRPC
