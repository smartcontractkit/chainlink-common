package remote

import (
	"context"
	"io"

	"google.golang.org/grpc"
)

// Locator identifies a capability's gRPC services as named by a peer.
//
// Target is a gRPC target string. LegacyHandle is the opaque numeric handle that
// predates it, retained so that a peer built before targets existed can still be
// understood. A [Transport] resolving a Locator should treat Target as
// authoritative when it is non-empty and fall back to LegacyHandle otherwise.
//
// This package does not interpret either field: the Transport that issued a
// Locator is the only thing that knows what its values mean.
type Locator struct {
	Target       string
	LegacyHandle uint32
}

// Publisher serves a capability's gRPC services and names them for a peer.
type Publisher interface {
	// Publish serves register on a new gRPC service and returns the Locator a peer
	// uses to reach it, along with a Closer that stops serving it.
	Publish(name string, register func(*grpc.Server)) (Locator, io.Closer, error)
}

// Dialer resolves a Locator to a gRPC client connection.
type Dialer interface {
	// Dial returns a connection to the service named by the Locator that resolve
	// returns. resolve is called lazily, and may be called again by implementations
	// that re-establish a broken connection, so callers must make it repeatable.
	Dial(name string, resolve func(context.Context) (Locator, error)) ClientConn
}

// Transport is what the registry wrappers need from their host: publishing local
// capabilities and dialling the ones a peer published.
//
// Implementations decide what a [Locator] means - a go-plugin broker connection, a
// host:port a peer can reach, or anything else that yields a gRPC connection.
type Transport interface {
	Publisher
	Dialer
}
