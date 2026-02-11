package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

func RequireEqualKeys(t *testing.T, a, b interface {
	ID() string
	Raw() internal.Raw
}) {
	t.Helper()
	require.Equal(t, a.ID(), b.ID(), "ids be equal")
	require.Equal(t, a.Raw(), b.Raw(), "raw bytes must be equal")
	require.EqualExportedValues(t, a, b)
}
