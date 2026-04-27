package net

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jpillora/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var _ ClientConnInterface = (*grpc.ClientConn)(nil)

type ClientConnInterface interface {
	grpc.ClientConnInterface
	GetState() connectivity.State
}

var _ ClientConnInterface = (*AtomicClient)(nil)

// An AtomicClient implements [grpc.ClientConnInterface] and is backed by a swappable [*grpc.ClientConn].
type AtomicClient struct {
	cc atomic.Pointer[grpc.ClientConn]
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

	mu   sync.RWMutex
	deps Resources
	cc   *grpc.ClientConn

	refreshing  bool          // indicates whether a refresh operation is currently in progress.
	refreshDone chan struct{} // closed when the current refresh attempt finishes so waiters can re-check state.
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
		if method == pb.Service_Close_FullMethodName {
			// If the underlying plugin is already gone, treat Close as a no-op rather than
			// attempting to recreate the client just to close it again.
			return nil
		}
		cc, refErr = c.refresh(ctx, nil)
	}
	for cc != nil {
		err := cc.Invoke(ctx, method, args, reply, opts...)
		if isErrTerminal(err) {
			if method == pb.Service_Close_FullMethodName {
				// don't reconnect just to call Close
				c.Logger.Warnw("clientConn: Invoke: terminal error", "method", method, "err", err)
				return nil
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

// refresh replaces c.cc with a new (different from orig) *grpc.ClientConn and returns the currently
// published replacement.
//
// Only one goroutine is allowed to perform reconnect work for a given broken generation. Other
// callers wait for that reconnect attempt to finish rather than holding the lock across the entire
// dial/backoff loop themselves. This keeps recovery from turning one dead plugin generation into a
// wider lock convoy across all callers trying to use the client.
//
// The broken connection is unpublished before redialing, and the replacement is published only after
// both the broker-side resources and the new gRPC connection are ready, so observers either see the
// previous complete generation or the next complete generation, never a partial one.
func (c *clientConn) refresh(ctx context.Context, orig *grpc.ClientConn) (*grpc.ClientConn, error) {
	// refreshDone is the completion signal for the specific in-flight refresh attempt we either own
	// or wait on. Waiters snapshot it under lock, then block on it without holding the mutex.
	var refreshDone chan struct{}
	for {
		c.mu.Lock()
		if c.cc != orig {
			// Another goroutine already installed a newer generation while we were racing to
			// refresh this one.
			cc := c.cc
			c.mu.Unlock()
			return cc, nil
		}
		if c.refreshing {
			// Only one goroutine performs the reconnect work. Everyone else waits without
			// holding the state lock, which avoids node-wide stalls during backoff.
			//
			// refreshDone belongs to the goroutine that already won the refresh race. When that
			// goroutine closes the channel, it means "this refresh attempt is finished; re-check
			// the currently published connection and decide again from there."
			refreshDone = c.refreshDone
			c.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%w", context.Cause(ctx))
			case <-refreshDone:
				// Do not assume success here. The refresh goroutine may have published a new
				// connection, or it may have failed and cleared the in-flight marker. Loop back
				// under lock and inspect the current state again.
			}
			continue
		}

		c.refreshing = true
		// This goroutine is now the single owner of the current refresh attempt. Anyone else
		// that encounters the broken generation will wait on this channel until we finish.
		refreshDone = make(chan struct{})
		c.refreshDone = refreshDone

		// Mark this generation unavailable before redialing so new callers either wait for the
		// replacement or observe a fully published successor.

		// Tear down the broken connection before dialing a replacement generation.
		if c.cc != nil {
			if err := c.cc.Close(); err != nil {
				c.Logger.Errorw("Client close failed", "err", err)
			}
			c.CloseAll(c.deps...)
			c.cc = nil
			c.deps = nil
		}
		c.mu.Unlock()
		break
	}

	defer func() {
		c.mu.Lock()
		c.refreshing = false
		// Closing refreshDone wakes every waiter that snapped this attempt's completion signal.
		// They will re-enter the loop, observe either the newly published connection or the lack
		// of one, and then either use it or start the next refresh attempt.
		close(refreshDone)
		c.refreshDone = nil
		c.mu.Unlock()
	}()

	try := func() error {
		// Each attempt acquires a fresh broker client ID/resources and dials the matching gRPC
		// endpoint. Failed attempts must release those resources before backing off.
		if d, ok := ctx.Deadline(); ok {
			c.Logger.Debugw("Client refresh", "deadline", d, "until", time.Until(d))
		}
		id, deps, err := c.newClient(ctx)
		if err != nil {
			c.Logger.Errorw("Client refresh attempt failed", "err", err)
			c.CloseAll(deps...)
			return err
		}

		lggr := logger.With(c.Logger, "id", id)
		lggr.Debug("Client dial")
		cc, err := c.Dial(id)
		if err != nil {
			if ctx.Err() != nil {
				lggr.Errorw("Client dial failed", "err", ErrConnDial{Name: c.name, ID: id, Err: err})
			}
			c.CloseAll(deps...)
			return err
		}

		// Publish the new connection only after both the broker-side dependencies and the
		// gRPC client are ready, so waiters see a fully initialized generation.
		c.mu.Lock()
		c.deps = deps
		c.cc = cc
		c.mu.Unlock()
		return nil
	}

	b := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for err := try(); err != nil; err = try() {
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

	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()
	return cc, nil
}
