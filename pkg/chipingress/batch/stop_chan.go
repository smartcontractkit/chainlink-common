package batch

import (
	"context"
	"time"
)

// A stopCh signals when some work should stop.
// Use StopChanR if you already have a read only <-chan.
type stopCh chan struct{}

// NewCtx returns a background [context.Context] that is cancelled when StopChan is closed.
func (s stopCh) NewCtx() (context.Context, context.CancelFunc) {
	return stopRchan((<-chan struct{})(s)).NewCtx()
}

// Ctx cancels a [context.Context] when StopChan is closed.
func (s stopCh) Ctx(ctx context.Context) (context.Context, context.CancelFunc) {
	return stopRchan((<-chan struct{})(s)).Ctx(ctx)
}

// CtxWithTimeout cancels a [context.Context] when StopChan is closed.
func (s stopCh) CtxWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return s.CtxCancel(context.WithTimeout(context.Background(), timeout))
}

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s stopCh) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	return stopRchan((<-chan struct{})(s)).CtxCancel(ctx, cancel)
}

// A stopRchan signals when some work should stop.
// This is a receive-only version of StopChan, for casting an existing <-chan.
type stopRchan <-chan struct{}

// NewCtx returns a background [context.Context] that is cancelled when StopChan is closed.
func (s stopRchan) NewCtx() (context.Context, context.CancelFunc) {
	return s.Ctx(context.Background())
}

// Ctx cancels a [context.Context] when StopChan is closed.
func (s stopRchan) Ctx(ctx context.Context) (context.Context, context.CancelFunc) {
	return s.CtxCancel(context.WithCancel(ctx))
}

// CtxWithTimeout cancels a [context.Context] when StopChan is closed.
func (s stopRchan) CtxWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return s.CtxCancel(context.WithTimeout(context.Background(), timeout))
}

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s stopRchan) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	go func() {
		select {
		case <-s:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
