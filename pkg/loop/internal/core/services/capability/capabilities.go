package capability

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	registryremote "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

// ExecutableCapability is a capability that can be executed.
type ExecutableCapability interface {
	capabilities.Executable
	capabilities.BaseCapability
}

// The clients and servers below are adapters over the shared capability wire
// implementation in pkg/capabilities/registry/remote. The only thing they add
// is the go-plugin broker: clients are built over broker connections so they
// can be re-dialed when the serving process restarts, and servers are handed
// the broker for the same reason, though the capability services themselves
// never use it.

func NewTriggerCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) capabilities.TriggerCapability {
	return registryremote.Wrap(brokerExt.Logger, conn, capabilities.CapabilityTypeTrigger).(capabilities.TriggerCapability)
}

func NewExecutableCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) ExecutableCapability {
	return registryremote.Wrap(brokerExt.Logger, conn, capabilities.CapabilityTypeAction).(ExecutableCapability)
}

func NewCombinedCapabilityClient(brokerExt *net.BrokerExt, conn net.ClientConnInterface) ExecutableCapability {
	return registryremote.Wrap(brokerExt.Logger, conn, capabilities.CapabilityTypeCombined).(ExecutableCapability)
}

func RegisterExecutableCapabilityServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl ExecutableCapability) error {
	bext := &net.BrokerExt{BrokerConfig: brokerCfg, Broker: broker}
	return registryremote.RegisterCapability(bext.Logger, server, impl, capabilities.CapabilityTypeAction)
}

func RegisterTriggerCapabilityServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl capabilities.TriggerCapability) error {
	bext := &net.BrokerExt{BrokerConfig: brokerCfg, Broker: broker}
	return registryremote.RegisterCapability(bext.Logger, server, impl, capabilities.CapabilityTypeTrigger)
}
