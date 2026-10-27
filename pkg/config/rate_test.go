package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRate_UnmarshalText(t *testing.T) {
	var r Rate
	require.NoError(t, r.UnmarshalText([]byte("2.5rps:3")))
	require.Equal(t, Rate{Limit: 2.5, Burst: 3}, r)

	text, err := r.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "2.5rps:3", string(text))

	require.EqualError(t, r.UnmarshalText([]byte("1rps")), "invalid rate limit: 1rps: must contain two parts")
	require.Equal(t, Rate{Limit: 2.5, Burst: 3}, r)
}
