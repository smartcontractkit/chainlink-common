package standalone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffsetPort(t *testing.T) {
	for _, tc := range []struct {
		name     string
		address  string
		delta    int
		expected string
	}{
		{"an address without a host keeps its shape", ":50051", 2, ":50053"},
		{"instance 0 is the configured address", ":50051", 0, ":50051"},
		{"the host part is left alone", "127.0.0.1:1234", 1, "127.0.0.1:1235"},
		{"an ipv6 host stays bracketed", "[::1]:1234", 1, "[::1]:1235"},
		// Every instance asking for an ephemeral port already gets a distinct one, so offsetting
		// would only turn "any port" into a specific one nothing asked for.
		{"an ephemeral port is left as it is", "127.0.0.1:0", 3, "127.0.0.1:0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			offset, err := OffsetPort(tc.address, tc.delta)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, offset)
		})
	}

	t.Run("an address without a port is an error", func(t *testing.T) {
		_, err := OffsetPort("localhost", 1)
		require.ErrorContains(t, err, "failed to parse address")
	})

	t.Run("a non-numeric port is an error", func(t *testing.T) {
		_, err := OffsetPort("localhost:https", 1)
		require.ErrorContains(t, err, "failed to parse port")
	})
}
