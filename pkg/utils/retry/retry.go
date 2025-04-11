package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jpillora/backoff"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type ctxKey string

// ctxKeyID is the context key for tracing ID
const ctxKeyID ctxKey = "retryID"

func CtxWithID(ctx context.Context, retryID string) context.Context {
	return context.WithValue(ctx, ctxKeyID, retryID)
}

// Exponential backoff (default) is used to handle retries with increasing wait times in case of errors
var BackoffStrategyDefault = backoff.Backoff{
	Min:    100 * time.Millisecond,
	Max:    3 * time.Second,
	Factor: 2,
}

type option struct {
	// Max Retries set the max number of times to execute a function that has failed.  Each function is called at least once.
	// By default, an infinite number of retries is allowed, limited only by the context.
	MaxRetries uint
	Strategy   backoff.Backoff
}

// WithMaxRetries sets the max number of times to execute a function that has failed.  Each function is called at least once.
// By default, an infinite number of retries is allowed, limited only by the context.
func WithMaxRetries(n uint) func(*option) {
	return func(o *option) {
		o.MaxRetries = n
	}
}

// WithStrategy sets the backoff strategy to apply
func WithStrategy(b backoff.Backoff) func(*option) {
	return func(o *option) {
		o.Strategy = b
	}
}

type Option func(*option)

// Do applies a default retry strategy to a given function.
func Do[R any](ctx context.Context, lggr logger.Logger, fn func(ctx context.Context) (R, error), opts ...Option) (R, error) {
	// Apply options
	option := &option{
		Strategy: BackoffStrategyDefault,
	}
	for _, opt := range opts {
		opt(option)
	}

	// Generate a new tracing ID if not present, used to track retries
	retryID := ctx.Value(ctxKeyID)
	if retryID == nil {
		retryID = uuid.New().String()
		// Add the generated tracing ID to the context (as it was not already present)
		ctx = context.WithValue(ctx, ctxKeyID, retryID)
	}

	// Track the number of retries
	for numRetries := int(option.Strategy.Attempt()); ; numRetries++ {
		if option.MaxRetries > 0 {
			if numRetries > int(option.MaxRetries) {
				var empty R
				return empty, fmt.Errorf("max retry attempts reached")
			}
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		wait := option.Strategy.Duration()
		message := fmt.Sprintf("Failed to execute function, retrying in %s ...", wait)
		lggr.Warnw(message, "wait", wait, "numRetries", numRetries, "retryID", retryID, "err", err)

		select {
		case <-ctx.Done():
			return result, fmt.Errorf("context done while executing function {retryID=%s, numRetries=%d}: %w", retryID, numRetries, ctx.Err())
		case <-time.After(wait):
			// Continue with the next retry
		}
	}
}
