package remote

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func NewGRPCRegistry(ctx context.Context, lggr logger.Logger, url string) (registry.CapabilitiesRegistry, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create registry proxy client for %s: %w", url, err)
	}

	transport, err := NewListenerTransport(ListenerTransportParams{
		Lggr: lggr,
		Listen: func() (net.Listener, error) {
			return net.Listen("tcp", "localhost:0")
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create listener transport: %w", err)
	}

	return NewCapabilitiesRegistryClient(lggr, conn, transport), nil
}
