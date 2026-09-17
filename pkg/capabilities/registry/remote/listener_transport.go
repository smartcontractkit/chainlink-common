package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// ErrNoLegacyHandle is returned when a peer named a capability only by the numeric
// handle that predates target strings. Resolving one needs the go-plugin broker that
// issued it, so a transport without a broker cannot reach that capability.
var ErrNoLegacyHandle = errors.New("peer named the capability by a legacy numeric handle, which requires a go-plugin broker")

// ListenerTransport implements [Transport] over ordinary gRPC connections: each
// published capability is served on its own listener, and a [Locator] names the
// address a peer dials to reach it.
//
// It issues no legacy handle, so a peer that names a capability only by the numeric
// handle predating targets cannot be reached - that peer is asking for a go-plugin
// broker, which this transport does not have.
type ListenerTransport struct {
	// Listen returns a listener for a newly published capability. Required.
	Listen func() (net.Listener, error)

	// Target returns the address a peer dials to reach a capability served on lis.
	// Defaults to lis.Addr().String(), which is reachable only from the same host;
	// override it to advertise a routable address.
	Target func(lis net.Listener) string

	// Connect opens a connection to target. Defaults to [grpc.NewClient] with DialOpts.
	Connect func(target string) (*grpc.ClientConn, error)

	// DialOpts are passed to the default Connect.
	DialOpts []grpc.DialOption

	// NewServer optionally overrides the *grpc.Server constructor.
	NewServer func() *grpc.Server

	// Logger reports errors that surface after Publish has returned.
	Logger logger.Logger
}

var _ Transport = (*ListenerTransport)(nil)

func (t *ListenerTransport) Publish(name string, register func(*grpc.Server)) (Locator, io.Closer, error) {
	if t.Listen == nil {
		return Locator{}, nil, errors.New("ListenerTransport.Listen is not set")
	}
	lis, err := t.Listen()
	if err != nil {
		return Locator{}, nil, fmt.Errorf("failed to listen for %s: %w", name, err)
	}

	server := grpc.NewServer()
	if t.NewServer != nil {
		server = t.NewServer()
	}
	register(server)

	var wg sync.WaitGroup
	wg.Go(func() {
		// Serve returns ErrServerStopped on a deliberate Stop, which is not a fault.
		if err := server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.logger().Errorw("Failed to serve capability", "name", name, "err", err)
		}
	})

	return Locator{Target: t.target(lis)}, closeFunc(func() error {
		server.Stop()
		wg.Wait()
		return nil
	}), nil
}

func (t *ListenerTransport) Dial(name string, resolve func(context.Context) (Locator, error)) ClientConn {
	return &lazyConn{name: name, resolve: resolve, connect: t.connect}
}

func (t *ListenerTransport) target(lis net.Listener) string {
	if t.Target != nil {
		return t.Target(lis)
	}
	return lis.Addr().String()
}

func (t *ListenerTransport) connect(target string) (*grpc.ClientConn, error) {
	if t.Connect != nil {
		return t.Connect(target)
	}
	return grpc.NewClient(target, t.DialOpts...)
}

func (t *ListenerTransport) logger() logger.Logger {
	if t.Logger != nil {
		return t.Logger
	}
	return logger.Nop()
}

// lazyConn resolves its target on first use, so the RPC that discovers a
// capability's address is not made until someone actually calls the capability.
// That matches what callers of [Dialer.Dial] already expect from the broker-backed
// transport, and lets a capability be resolved after the registry client is built.
type lazyConn struct {
	name    string
	resolve func(context.Context) (Locator, error)
	connect func(target string) (*grpc.ClientConn, error)

	mu sync.Mutex
	cc *grpc.ClientConn
}

var _ ClientConn = (*lazyConn)(nil)

func (c *lazyConn) conn(ctx context.Context) (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cc != nil {
		return c.cc, nil
	}
	loc, err := c.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if loc.Target == "" {
		return nil, fmt.Errorf("cannot reach %s: %w", c.name, ErrNoLegacyHandle)
	}
	cc, err := c.connect(loc.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s at %q: %w", c.name, loc.Target, err)
	}
	c.cc = cc
	return cc, nil
}

func (c *lazyConn) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	cc, err := c.conn(ctx)
	if err != nil {
		return err
	}
	return cc.Invoke(ctx, method, args, reply, opts...)
}

func (c *lazyConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	cc, err := c.conn(ctx)
	if err != nil {
		return nil, err
	}
	return cc.NewStream(ctx, desc, method, opts...)
}

func (c *lazyConn) GetState() connectivity.State {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cc == nil {
		return connectivity.Idle
	}
	return c.cc.GetState()
}

func (c *lazyConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cc == nil {
		return nil
	}
	cc := c.cc
	c.cc = nil
	return cc.Close()
}

// closeFunc adapts a func() error to io.Closer.
type closeFunc func() error

func (f closeFunc) Close() error { return f() }
