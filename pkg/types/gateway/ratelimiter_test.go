package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter_PerSender(t *testing.T) {
	t.Parallel()

	config := RateLimiterConfig{
		GlobalRPS:      3.0,
		GlobalBurst:    3,
		PerSenderRPS:   1.0,
		PerSenderBurst: 2,
	}
	rl, err := NewRateLimiter(config)
	require.NoError(t, err)
	require.True(t, rl.Allow("user1"))
	require.True(t, rl.Allow("user2"))
	require.True(t, rl.Allow("user1"))
	require.False(t, rl.Allow("user1"))
	require.False(t, rl.Allow("user3"))
}
