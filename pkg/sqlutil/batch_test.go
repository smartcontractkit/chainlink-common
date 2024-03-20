package sqlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatching(t *testing.T) {
	const batchSize uint = 1000
	var callCount uint
	var maxCalls uint = 100

	err := Batch(func(offset, limit uint) (count uint, err error) {
		require.Equal(t, callCount*batchSize, offset)
		callCount++
		if callCount == maxCalls {
			return 0, nil
		}
		return limit, nil
	}, batchSize)

	require.NoError(t, err)
}
