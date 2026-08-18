//go:build !race

package services

import "context"

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
	return ctx, cancel
}
