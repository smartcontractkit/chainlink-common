package utils

import (
	"context"
	"math"
	mrand "math/rand"
	"time"

	"github.com/jpillora/backoff"
)

// WithJitter adds +/- 10% to a duration
func WithJitter(d time.Duration) time.Duration {
	// #nosec
	if d == 0 {
		return 0
	}
	// ensure non-zero arg to Intn to avoid panic
	max := math.Max(float64(d.Abs())/5.0, 1.)
	// #nosec - non critical randomness
	jitter := mrand.Intn(int(max))
	jitter = jitter - (jitter / 2)
	return time.Duration(int(d) + jitter)
}

// ContextFromChan creates a context that finishes when the provided channel
// receives or is closed.
// When channel closes, the ctx.Err() will always be context.Canceled
// NOTE: Spins up a goroutine that exits on cancellation.
// REMEMBER TO CALL CANCEL OTHERWISE IT CAN LEAD TO MEMORY LEAKS
func ContextFromChan(chStop <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-chStop:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// ContextWithDeadlineFn returns a copy of the parent context with the deadline modified by deadlineFn.
// deadlineFn will only be called if the parent has a deadline.
// The new deadline must be sooner than the old to have an effect.
func ContextWithDeadlineFn(ctx context.Context, deadlineFn func(orig time.Time) time.Time) (context.Context, context.CancelFunc) {
	cancel := func() {}
	if d, ok := ctx.Deadline(); ok {
		if m := deadlineFn(d); m.Before(d) {
			ctx, cancel = context.WithDeadline(ctx, m)
		}
	}
	return ctx, cancel
}

// NewRedialBackoff is a standard backoff to use for redialling or reconnecting to
// unreachable network endpoints
func NewRedialBackoff() backoff.Backoff {
	return backoff.Backoff{
		Min:    1 * time.Second,
		Max:    15 * time.Second,
		Jitter: true,
	}

}

// UnwrapError returns a list of underlying errors if passed error implements joinedError or return the err in a single-element list otherwise.
//
//nolint:errorlint // error type checks will fail on wrapped errors. Disabled since we are not doing checks on error types.
func UnwrapError(err error) []error {
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return []error{err}
	}
	return joined.Unwrap()
}
