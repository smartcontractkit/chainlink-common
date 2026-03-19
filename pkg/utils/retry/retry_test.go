package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// exampleFunc is a function type used for testing retry strategies
type exampleFunc func(ctx context.Context) (string, error)
type testCase struct {
	name     string
	fn       exampleFunc
	expected string
	errMsg   []string
	timeout  time.Duration
	strategy *Strategy[string]
}

func TestWithRetry(t *testing.T) {
	lggr := logger.Test(t)

	tests := []testCase{
		{
			name: "successful function",
			fn: func(ctx context.Context) (string, error) {
				return "success", nil
			},
			expected: "success",
			timeout:  100 * time.Millisecond,
		},
		{
			name: "always failing function",
			fn: func(ctx context.Context) (string, error) {
				return "", errors.New("permanent error")
			},
			errMsg:  []string{"context done while executing function", "permanent error"},
			timeout: 100 * time.Millisecond,
		},
		{
			name: "eventually successful function",
			fn: func() exampleFunc {
				attempts := 0
				return func(ctx context.Context) (string, error) {
					attempts++
					if attempts < 3 {
						return "", errors.New("temporary error")
					}
					return "eventual success", nil
				}
			}(),
			expected: "eventual success",
			timeout:  500 * time.Millisecond,
		},
		{
			name: "eventually successful function (fail - exceeding context timeout)",
			fn: func() exampleFunc {
				attempts := 0
				return func(ctx context.Context) (string, error) {
					attempts++
					if attempts < 3 {
						return "", errors.New("temporary error")
					}
					return "eventual success", nil
				}
			}(),
			errMsg:  []string{"context done while executing function", "temporary error"},
			timeout: 100 * time.Millisecond,
		},
		{
			name: "eventually (in time) successful function",
			fn: func() exampleFunc {
				// Start timer, successful after 400ms
				timeout := 400 * time.Millisecond
				start := time.Now()
				return func(ctx context.Context) (string, error) {
					if time.Since(start) < timeout {
						return "", errors.New("temporary error")
					}
					return "eventual success", nil
				}
			}(),
			expected: "eventual success",
			timeout:  1 * time.Second,
		},
		{
			name: "eventually (in time) successful function (fail - exceeding context timeout)",
			fn: func() exampleFunc {
				// Start timer, successful after 4s
				timeout := 4 * time.Second
				start := time.Now()
				return func(ctx context.Context) (string, error) {
					if time.Since(start) < timeout {
						return "", errors.New("temporary error")
					}
					return "eventual success", nil
				}
			}(),
			errMsg:  []string{"context done while executing function", "temporary error"},
			timeout: 1 * time.Second,
		},
		{
			name: "obeys limit of 1 max retry",
			strategy: &Strategy[string]{
				MaxRetries: 1,
			},
			fn: func() exampleFunc {
				calls := make(chan struct{}, 1)
				called := 0
				return func(ctx context.Context) (string, error) {
					called++
					select {
					case calls <- struct{}{}:
						return "", errors.New("temporary error")
					default:
						if called != 2 {
							return "", errors.New("expected to be called once to fail and once to succeed")
						}
						return "eventual success", nil
					}
				}
			}(),
			expected: "eventual success",
			timeout:  1 * time.Second,
		},
		{
			name: "fails until max retries is met",
			fn: func(ctx context.Context) (string, error) {
				return "", errors.New("temporary error")
			},
			strategy: &Strategy[string]{
				MaxRetries: 1,
			},
			errMsg:  []string{"max retry attempts reached", "numRetries=2", "temporary error"},
			timeout: 1 * time.Second,
		},
		{
			name: "context timeout surfaces the last callback error, not earlier ones",
			fn: func() exampleFunc {
				attempt := 0
				return func(ctx context.Context) (string, error) {
					attempt++
					return "", fmt.Errorf("error on attempt %d", attempt)
				}
			}(),
			errMsg:  []string{"context done while executing function", "error on attempt 2"},
			timeout: 300 * time.Millisecond,
		},
		{
			name: "max retries surfaces the last callback error, not earlier ones",
			fn: func() exampleFunc {
				attempt := 0
				return func(ctx context.Context) (string, error) {
					attempt++
					return "", fmt.Errorf("error on attempt %d", attempt)
				}
			}(),
			strategy: &Strategy[string]{
				MaxRetries: 3,
			},
			errMsg:  []string{"max retry attempts reached", "error on attempt 4"},
			timeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, tt.timeout)
			defer cancel()

			var (
				result string
				err    error
			)
			if tt.strategy == nil {
				result, err = Do(ctx, lggr, tt.fn)
			} else {
				result, err = tt.strategy.Do(ctx, lggr, tt.fn)
			}

			if len(tt.errMsg) > 0 {
				require.Error(t, err)
				for _, msg := range tt.errMsg {
					require.Contains(t, err.Error(), msg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
