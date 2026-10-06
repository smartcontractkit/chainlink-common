package beholder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// TestChipIngressRetryConfig_RetryPolicy covers the configuration-layer entry point used by
// hosts (node TOML, LOOP env): Enabled=false must leave retries off (nil policy), Enabled=true
// with zero overrides must resolve to chipingress.DefaultRetryPolicy rather than a zero-valued
// policy that gRPC would silently discard, and explicit overrides (status codes by name) must
// apply.
func TestChipIngressRetryConfig_RetryPolicy(t *testing.T) {
	t.Run("disabled returns nil policy", func(t *testing.T) {
		p, err := ChipIngressRetryConfig{Enabled: false, MaxAttempts: 5}.RetryPolicy()
		require.NoError(t, err)
		assert.Nil(t, p, "retries must stay off when not enabled, regardless of other fields")
	})

	t.Run("enabled with zero overrides resolves to defaults", func(t *testing.T) {
		p, err := ChipIngressRetryConfig{Enabled: true}.RetryPolicy()
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, chipingress.DefaultRetryPolicy(), *p)
	})

	t.Run("overrides apply", func(t *testing.T) {
		p, err := ChipIngressRetryConfig{
			Enabled:              true,
			MaxAttempts:          5,
			InitialBackoff:       250 * time.Millisecond,
			MaxBackoff:           5 * time.Second,
			BackoffMultiplier:    1.5,
			RetryableStatusCodes: []string{"Unavailable", "Internal"},
		}.RetryPolicy()
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, 5, p.MaxAttempts)
		assert.Equal(t, 250*time.Millisecond, p.InitialBackoff)
		assert.Equal(t, 5*time.Second, p.MaxBackoff)
		assert.InDelta(t, 1.5, p.BackoffMultiplier, 0.0001)
		assert.Equal(t, []codes.Code{codes.Unavailable, codes.Internal}, p.RetryableStatusCodes)
	})

	t.Run("unknown status code name errors", func(t *testing.T) {
		_, err := ChipIngressRetryConfig{
			Enabled:              true,
			RetryableStatusCodes: []string{"Unavailable", "Bogus"},
		}.RetryPolicy()
		require.Error(t, err)
		assert.ErrorContains(t, err, "Bogus")
	})
}
