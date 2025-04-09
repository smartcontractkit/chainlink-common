package net

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jpillora/backoff"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var _ grpc.ClientConnInterface = (*AtomicClient)(nil)

// An AtomicClient implements [grpc.ClientConnInterface] and is backed by a swappable [*grpc.ClientConn].
type AtomicClient struct {
	cc atomic.Pointer[grpc.ClientConn]
}

func (a *AtomicClient) Store(cc *grpc.ClientConn) { a.cc.Store(cc) }

func (a *AtomicClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return a.cc.Load().Invoke(ctx, method, args, reply, opts...)
}

func (a *AtomicClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return a.cc.Load().NewStream(ctx, desc, method, opts...)
}

var _ grpc.ClientConnInterface = (*clientConn)(nil)

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
}

func (c *clientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()

	if cc == nil {
		var err error
		cc, err = c.refresh(ctx, nil)
		if err != nil {
			return err
		}
	}
	for cc != nil {
		err := cc.Invoke(ctx, method, args, reply, opts...)
		if isErrTerminal(err) {
			if method == pb.Service_Close_FullMethodName {
				// don't reconnect just to call Close
				c.Logger.Warnw("clientConn: Invoke: terminal error", "method", method, "err", err)
				return err
			}
			c.Logger.Warnw("clientConn: Invoke: terminal error, refreshing connection", "method", method, "err", err)
			cc, err = c.refresh(ctx, cc)
			if err != nil {
				return err
			}

			continue
		}
		return err
	}
	return context.Cause(ctx)
}

func (c *clientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	c.mu.RLock()
	cc := c.cc
	c.mu.RUnlock()

	if cc == nil {
		var err error
		cc, err = c.refresh(ctx, nil)
		if err != nil {
			return nil, err
		}
	}
	for cc != nil {
		s, err := cc.NewStream(ctx, desc, method, opts...)
		if isErrTerminal(err) {
			c.Logger.Warnw("clientConn: NewStream: terminal error, refreshing connection", "err", err)
			cc, err = c.refresh(ctx, cc)
			if err != nil {
				return nil, err
			}
			continue
		}
		return s, err
	}
	return nil, context.Cause(ctx)
}

// AppError represents a custom application error that wraps another error.
type AppError struct {
	Err error // Wrapped error
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// Unwrap allows access to the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError instance.
func NewAppError(err error) *AppError {
	return &AppError{
		Err: err,
	}
}

// refresh replaces c.cc with a new (different from orig) *grpc.ClientConn, and returns it as well.
// It will block until a new connection is successfully dialed, or return nil if the context expires.
func (c *clientConn) refresh(ctx context.Context, orig *grpc.ClientConn) (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cc != orig {
		return c.cc, nil
	}
	if c.cc != nil {
		if err := c.cc.Close(); err != nil {
			c.Logger.Errorw("Client close failed", "err", err)
		}
		c.CloseAll(c.deps...)
	}

	try := func() (bool, error) {
		c.Logger.Debug("Client refresh")
		id, deps, err := c.newClient(ctx)
		if err != nil {
			c.Logger.Errorw("Client refresh attempt failed", "err", err)
			c.CloseAll(deps...)

			var appErr *AppError
			if errors.As(err, &appErr) {
				return false, err
			}

			return false, nil
		}
		c.deps = deps

		lggr := logger.With(c.Logger, "id", id)
		lggr.Debug("Client dial")
		c.cc, err = c.Dial(id)
		if err != nil {
			if ctx.Err() != nil {
				lggr.Errorw("Client dial failed", "err", ErrConnDial{Name: c.name, ID: id, Err: err})
			}
			c.CloseAll(c.deps...)
			return false, nil
		}
		return true, nil
	}

	b := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		success, err := try()
		if success {
			break
		}

		if err != nil {
			return nil, err
		}

		if ctx.Err() != nil {
			c.Logger.Errorw("Client refresh failed: aborting refresh due to context error", "err", ctx.Err())
			return nil, nil
		}
		wait := b.Duration()
		c.Logger.Infow("Waiting to refresh", "wait", wait)
		select {
		case <-ctx.Done():
			return nil, nil
		case <-time.After(wait):
		}
	}

	return c.cc, nil
}
