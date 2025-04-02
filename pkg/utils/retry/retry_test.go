package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/stretchr/testify/require"
)

// exampleFunc is a function type used for testing retry strategies
type exampleFunc func(ctx context.Context) (string, error)

func TestWithRetry(t *testing.T) {
	lggr := logger.Test(t)

	tests := []struct {
		name     string
		fn       exampleFunc
		expected string
		errMsg   string
		timeout  time.Duration
	}{
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
			errMsg:  "context done while executing function",
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
			errMsg:  "context done while executing function",
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
			errMsg:  "context done while executing function",
			timeout: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, tt.timeout)
			defer cancel()

			result, err := WithRetry(ctx, lggr, tt.fn)
			if tt.errMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
