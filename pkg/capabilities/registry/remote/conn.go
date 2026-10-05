package remote

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
)

// ClientConn is the gRPC client connection a capability wrapper needs: the standard
// invoke/stream surface, plus the connection state the registry reads to decide
// whether a cached capability client is still usable.
//
// It is deliberately expressed only in gRPC terms. A plain *grpc.ClientConn satisfies
// it, as does a connection managed by a plugin host on the caller's behalf.
type ClientConn interface {
	grpc.ClientConnInterface
	registry.StateGetter
}

var _ ClientConn = (*grpc.ClientConn)(nil)
