package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

func TestAllEqual(t *testing.T) {
	t.Parallel()

	require.False(t, utils.AllEqual(1, 2, 3, 4, 5))
	require.True(t, utils.AllEqual(1, 1, 1, 1, 1))
	require.False(t, utils.AllEqual(1, 1, 1, 2, 1, 1, 1))
}
