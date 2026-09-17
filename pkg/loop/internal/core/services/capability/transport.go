package capability

import (
	"context"
	"io"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// brokerTransport implements [remote.Transport] over a go-plugin broker: a
// [remote.Locator] names a connection served on the broker, and this is the only
// place that knows that is what it means.
type brokerTransport struct {
	brokerExt *net.BrokerExt
}

var _ remote.Transport = brokerTransport{}

func NewBrokerTransport(b *net.BrokerExt) remote.Transport {
	return brokerTransport{brokerExt: b}
}

func (t brokerTransport) Publish(name string, register func(*grpc.Server)) (remote.Locator, io.Closer, error) {
	id, res, err := t.brokerExt.ServeNew(name, register)
	if err != nil {
		return remote.Locator{}, nil, err
	}
	return remote.Locator{Target: net.BrokerTarget(id), LegacyHandle: id}, res, nil
}

func (t brokerTransport) Dial(name string, resolve func(context.Context) (remote.Locator, error)) remote.ClientConn {
	return t.brokerExt.NewClientConn(name, func(ctx context.Context) (uint32, net.Resources, error) {
		loc, err := resolve(ctx)
		if err != nil {
			return 0, nil, err
		}
		id, err := brokerConnID(loc)
		if err != nil {
			return 0, nil, err
		}
		return id, nil, nil
	})
}

func brokerConnID(loc remote.Locator) (uint32, error) {
	if loc.Target == "" {
		return loc.LegacyHandle, nil
	}
	return net.ParseBrokerTarget(loc.Target)
}

func RegisterCapabilitiesRegistryServer(s *grpc.Server, b *net.BrokerExt, i core.CapabilitiesRegistry) {
	remote.RegisterCapabilitiesRegistryServer(s, b.Logger, i, NewBrokerTransport(b.WithName("CapabilitiesRegistryServer")))
}

func NewCapabilitiesRegistryClient(cc grpc.ClientConnInterface, b *net.BrokerExt) core.CapabilitiesRegistry {
	return remote.NewCapabilitiesRegistryClient(b.Logger, cc, NewBrokerTransport(b.WithName("CapabilitiesRegistryClient")))
}
