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

// CtxKeyRetryID is the context key for tracing ID
const CtxKeyRetryID ctxKey = "retryID"

// Exponential backoff (default) is used to handle retries with increasing wait times in case of errors
var BackoffStrategyDefault = backoff.Backoff{
	Min:    100 * time.Millisecond,
	Max:    3 * time.Second,
	Factor: 2,
}

// WithStrategy applies a retry strategy to a given function.
func WithStrategy[R any](ctx context.Context, lggr logger.Logger, strategy backoff.Backoff, fn func(ctx context.Context) (R, error)) (R, error) {
	// Generate a new tracing ID if not present, used to track retries
	retryID, ok := ctx.Value(CtxKeyRetryID).(string)
	if !ok {
		retryID = uuid.New().String()
		// Add the generated tracing ID to the context (as it was not already present)
		ctx = context.WithValue(ctx, CtxKeyRetryID, retryID)
	}

	// Track the number of retries
	numRetries := int(strategy.Attempt())
	for {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		wait := strategy.Duration()
		message := fmt.Sprintf("Failed to execute function, retrying in %s ...", wait)
		lggr.Warnw(message, "wait", wait, "numRetries", numRetries, "retryID", retryID, "err", err)

		select {
		case <-ctx.Done():
			return result, fmt.Errorf("context done while executing function {retryID=%s, numRetries=%d}: %w", retryID, numRetries, ctx.Err())
		case <-time.After(wait):
			numRetries++
			// Continue with the next retry
		}
	}
}

// With applies a default retry strategy to a given function.
func With[R any](ctx context.Context, lggr logger.Logger, fn func(ctx context.Context) (R, error)) (R, error) {
	return WithStrategy(ctx, lggr, BackoffStrategyDefault, fn)
}
