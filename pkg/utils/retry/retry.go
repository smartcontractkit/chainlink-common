package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jpillora/backoff"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// CtxKeyTracingID is the context key for tracing ID
type ctxKey string

const CtxKeyTracingID ctxKey = "tracingID"

// Exponential backoff (default) is used to handle retries with increasing wait times in case of errors
var BackoffStrategyDefault = backoff.Backoff{
	Min:    100 * time.Millisecond,
	Max:    3 * time.Second,
	Factor: 2,
}

// WithRetryStrategy applies a retry strategy to a given function.
func WithRetryStrategy[R any](ctx context.Context, lggr logger.Logger, strategy backoff.Backoff, fn func(ctx context.Context) (R, error)) (R, error) {
	// Generate a new tracing ID if not present, used to track retries
	tracingID, ok := ctx.Value(CtxKeyTracingID).(string)
	if !ok {
		tracingID = uuid.New().String()
		// Add the generated tracing ID to the context (as it was not already present)
		ctx = context.WithValue(ctx, CtxKeyTracingID, tracingID)
	}

	// Track the number of retries
	numRetries := 0
	for {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		wait := strategy.Duration()
		message := fmt.Sprintf("Failed to execute function, retrying in %s ...", wait)
		lggr.Warnw(message, "wait", wait, "numRetries", numRetries, "tracingID", tracingID, "err", err)

		select {
		case <-ctx.Done():
			return result, fmt.Errorf("context done while executing function {tracingID=%s, numRetries=%d}: %w", tracingID, numRetries, ctx.Err())
		case <-time.After(wait):
			numRetries++
			// Continue with the next retry
		}
	}
}

// WithRetry applies a default retry strategy to a given function.
func WithRetry[R any](ctx context.Context, lggr logger.Logger, fn func(ctx context.Context) (R, error)) (R, error) {
	return WithRetryStrategy(ctx, lggr, BackoffStrategyDefault, fn)
}
