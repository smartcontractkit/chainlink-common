package batch

import (
	"context"
	"time"
)

/*
NOTE:This is a copy of the services/stop.go file to avoid a circular dependency on the services package.
*/

// A stopCh signals when some work should stop.
// Use StopChanR if you already have a read only <-chan.
type stopCh chan struct{}

// CtxWithTimeout cancels a [context.Context] when StopChan is closed.
func (s stopCh) CtxWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return s.CtxCancel(context.WithTimeout(context.Background(), timeout))
}

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s stopCh) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	go func() {
		select {
		case <-s:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
