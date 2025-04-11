package backoff

import (
	"context"
	"time"
)

// retryOptions holds configuration settings for the retry mechanism.
type retryOptions struct {
	BackOff  BackOff // Strategy for calculating backoff periods.
	MaxTries uint    // Maximum number of retry attempts.
}

type RetryOption func(*retryOptions)

// WithBackOff configures a custom backoff strategy.
func WithBackOff(b BackOff) RetryOption {
	return func(args *retryOptions) {
		args.BackOff = b
	}
}

// WithMaxTries limits the number of all attempts.
func WithMaxTries(n uint) RetryOption {
	return func(args *retryOptions) {
		args.MaxTries = n
	}
}

// Retry attempts the operation until success or backoff completion.
// It ensures the operation is executed at least once.
//
// Returns the operation result or error if retries are exhausted or context is cancelled.
func Retry[T any](ctx context.Context, operation func() (T, error), opts ...RetryOption) (T, error) {
	// Initialize default retry options.
	args := &retryOptions{
		BackOff: NewExponentialBackOff(),
	}

	// Apply user-provided options to the default settings.
	for _, opt := range opts {
		opt(args)
	}

	args.BackOff.Reset()
	for numTries := uint(1); ; numTries++ {
		// Execute the operation.
		res, err := operation()
		if err == nil {
			return res, nil
		}

		// Stop retrying if maximum tries exceeded.
		if args.MaxTries > 0 && numTries >= args.MaxTries {
			return res, err
		}

		// Stop retrying if context is cancelled.
		if cerr := context.Cause(ctx); cerr != nil {
			return res, cerr
		}

		// Calculate next backoff duration.
		next := args.BackOff.NextBackOff()
		if next == Stop {
			return res, err
		}

		// Wait for the next backoff period or context cancellation.
		select {
		case <-time.After(next):
		case <-ctx.Done():
			return res, context.Cause(ctx)
		}
	}
}
