//go:build race

package services

import (
	"context"
	"time"
)

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s StopRChan) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	go func() {
		select {
		case <-s:
			cancel()
		case <-ctx.Done():
		}
	}()
	return &syncCtx{
		deadline: ctx.Deadline,
		done:     ctx.Done,
		value:    ctx.Value,
		err:      ctx.Err,
	}, cancel
}

var _ context.Context = &syncCtx{}

// syncCtx is a context.Context implementation that is safe to format via %#v, which mockery uses.
type syncCtx struct {
	deadline func() (time.Time, bool)
	done     func() <-chan struct{}
	value    func(any) any
	err      func() error
}

func (s *syncCtx) Deadline() (time.Time, bool) { return s.deadline() }
func (s *syncCtx) Done() <-chan struct{}       { return s.done() }
func (s *syncCtx) Value(k any) any             { return s.value(k) }
func (c *syncCtx) Err() error                  { return c.err() }
