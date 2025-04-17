package retry

import (
	"context"
	"errors"
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

type Strategy[R any] struct {
	Backoff *backoff.Backoff

	// MaxRetries is the number of attempts to call a function that has errored.
	// A default 0 value retries indefinitely until the context is canceled.
	// Every function is called at least once.
	MaxRetries uint
}

// Do executes a func according to the strategy.
func (s *Strategy[R]) Do(ctx context.Context, lggr logger.Logger, fn func(ctx context.Context) (R, error)) (R, error) {
	if s.Backoff == nil {
		s.Backoff = BackoffStrategyDefault.Copy()
	}
	s.Backoff.Reset()

	// Generate a new tracing ID if not present, used to track retries
	retryID := ctx.Value(ctxKeyID)
	if retryID == nil {
		retryID = uuid.New().String()
		// Add the generated tracing ID to the context (as it was not already present)
		ctx = context.WithValue(ctx, ctxKeyID, retryID)
	}

	// Track the number of retries
	for numRetries := int(s.Backoff.Attempt()); ; numRetries++ {
		if s.MaxRetries > 0 {
			if numRetries > int(s.MaxRetries) {
				var empty R
				return empty, fmt.Errorf("max retry attempts reached")
			}
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		// Handle permanent errors without retrying.
		var permanent *PermanentError
		if errors.As(err, &permanent) {
			return result, permanent.Unwrap()
		}

		wait := s.Backoff.Duration()
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

// Do applies a default retry strategy to a given function.
func Do[R any](ctx context.Context, lggr logger.Logger, fn func(ctx context.Context) (R, error)) (R, error) {
	return new(Strategy[R]).Do(ctx, lggr, fn)
}
