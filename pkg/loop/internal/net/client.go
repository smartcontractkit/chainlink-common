package net

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jpillora/backoff"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var _ ClientConnInterface = (*grpc.ClientConn)(nil)

type ClientConnInterface interface {
	grpc.ClientConnInterface
	GetState() connectivity.State
	Close() error
}

func ClientConnInterfaceFromGRPC(conn grpc.ClientConnInterface) ClientConnInterface {
	connCloser, ok := conn.(ClientConnInterface)
	if !ok {
		connCloser = &noopClientConnInterface{conn}
	}
	return connCloser
}

// noopClientConnInterface adapts ClientConnInterface to implement net.ClientConnInterface with no-ops.
type noopClientConnInterface struct {
	grpc.ClientConnInterface
}

func (c *noopClientConnInterface) GetState() connectivity.State {
	return connectivity.State(-1)
}

func (*noopClientConnInterface) Close() error { return nil }

var _ ClientConnInterface = (*AtomicClient)(nil)

// An AtomicClient implements [grpc.ClientConnInterface] and is backed by a swappable [*grpc.ClientConn].
type AtomicClient struct {
	cc atomic.Pointer[grpc.ClientConn]
}

func (a *AtomicClient) Close() error {
	if v := a.cc.Swap(nil); v != nil {
		return (*v).Close()
	}
	return nil
}

func (a *AtomicClient) GetState() connectivity.State {
	return a.cc.Load().GetState()
}

func (a *AtomicClient) Store(cc *grpc.ClientConn) { a.cc.Store(cc) }

func (a *AtomicClient) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return a.cc.Load().Invoke(ctx, method, args, reply, opts...)
}

func (a *AtomicClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return a.cc.Load().NewStream(ctx, desc, method, opts...)
}

var _ ClientConnInterface = (*clientConn)(nil)

// newClientFn returns a new client connection id to dial, and a set of Resource dependencies to close.
type newClientFn func(context.Context) (id uint32, deps Resources, err error)

// clientConn is a [grpc.ClientConnInterface] backed by a [*grpc.ClientConn] which can be recreated and swapped out
// via the provided [newClientFn].
// New instances should be created via BrokerExt.NewClientConn.
type clientConn struct {
	*BrokerExt
	newClient newClientFn
	name      string

	// refreshSem serializes connection refreshes. It is a weight-1 semaphore acting as a
	// context-aware mutex: Acquire honors ctx cancellation, so a caller never blocks past its
	// own deadline while another refresh is in progress.
	// It must be created with capacity 1; a nil value would block forever.
	refreshSem *semaphore.Weighted

	// mu guards cc and deps only. It is never held while dialing, so an in-progress refresh
	// cannot block concurrent callers that are reading a still-valid connection.
	mu   sync.RWMutex
	deps Resources
	cc   *grpc.ClientConn
}

func (c *clientConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.close()
}

func (c *clientConn) close() error {
	if c.cc != nil {
		err := c.cc.Close()
		c.CloseAll(c.deps...)
		c.cc = nil
		c.deps = nil
		return err
	}
	return nil
}

func (c *clientConn) GetState() connectivity.State {
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()
	if cc != nil {
		return cc.GetState()
	}
	// fall back to Shutdown to reflect underlying state
	return connectivity.Shutdown
}

func (c *clientConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()

	var refErr error
	if cc == nil {
		cc, refErr = c.refresh(ctx, nil)
	}
	for cc != nil {
		err := cc.Invoke(ctx, method, args, reply, opts...)
		if isErrTerminal(err) {
			if method == pb.Service_Close_FullMethodName {
				// don't reconnect just to call Close
				c.Logger.Warnw("clientConn: Invoke: terminal error", "method", method, "err", err)
				return err
			}
			c.Logger.Errorw("clientConn: Invoke: terminal error, refreshing connection", "method", method, "err", err)
			cc, refErr = c.refresh(ctx, cc)
			continue
		}
		return err
	}
	return refErr
}

func (c *clientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()

	var refErr error
	if cc == nil {
		cc, refErr = c.refresh(ctx, nil)
	}
	for cc != nil {
		s, err := cc.NewStream(ctx, desc, method, opts...)
		if isErrTerminal(err) {
			c.Logger.Errorw("clientConn: NewStream: terminal error, refreshing connection", "err", err)
			cc, refErr = c.refresh(ctx, cc)
			continue
		}
		return s, err
	}
	return nil, refErr
}

// refresh replaces c.cc with a new (different from orig) *grpc.ClientConn, and returns it as well.
// It will block until a new connection is successfully dialed, or return nil if the context expires.
//
// Dialing happens without holding c.mu, so a slow or stuck refresh
// never blocks concurrent callers that are reading a still-valid c.cc.
// Refreshes are serialized through a context-aware lock, so a caller aborts
// on its own deadline rather than waiting out another in-flight refresh.
func (c *clientConn) refresh(ctx context.Context, orig *grpc.ClientConn) (*grpc.ClientConn, error) {
	// Serialize refreshes without blocking past the caller's deadline: if another refresh holds
	// the semaphore (potentially under a long backoff), a caller with a deadline still returns
	// promptly because Acquire honors ctx cancellation.
	if err := c.refreshSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer c.refreshSem.Release(1)

	// Another refresh may have already replaced orig while we waited for the lock.
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()
	if cc != orig {
		return cc, nil
	}

	// Tear down the stale connection before dialing a replacement. Held only briefly.
	c.mu.Lock()
	if err := c.close(); err != nil {
		c.Logger.Errorw("Client close failed", "err", err)
	}
	c.mu.Unlock()

	// try dials a fresh connection, returning it and its deps on success. It does not touch
	// c.cc/c.deps; the caller installs them under c.mu only once dialing succeeds.
	try := func() (*grpc.ClientConn, Resources, error) {
		if d, ok := ctx.Deadline(); ok {
			c.Logger.Debugw("Client refresh", "deadline", d, "until", time.Until(d))
		}
		id, deps, err := c.newClient(ctx)
		if err != nil {
			c.Logger.Errorw("Client refresh attempt failed", "err", err)
			c.CloseAll(deps...)
			return nil, nil, err
		}

		lggr := logger.With(c.Logger, "id", id)
		lggr.Debug("Client dial")
		cc, err := c.Dial(id)
		if err != nil {
			if ctx.Err() != nil {
				lggr.Errorw("Client dial failed", "err", ErrConnDial{Name: c.name, ID: id, Err: err})
			}
			c.CloseAll(deps...)
			return nil, nil, err
		}
		return cc, deps, nil
	}

	b := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		cc, deps, err := try()
		if err == nil {
			c.mu.Lock()
			c.cc = cc
			c.deps = deps
			c.mu.Unlock()
			return cc, nil
		}
		if ctx.Err() != nil {
			err = fmt.Errorf("%w: last error: %w", context.Cause(ctx), err)
			c.Logger.Errorw("Client refresh failed: aborting refresh", "err", err)
			return nil, err
		}
		wait := b.Duration()
		c.Logger.Infow("Waiting to refresh", "wait", wait)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: last error: %w", context.Cause(ctx), err)
		case <-time.After(wait):
		}
	}
}
