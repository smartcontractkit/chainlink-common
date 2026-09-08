package grafana

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParallelFor(t *testing.T) {
	t.Run("processes all items with bounded concurrency", func(t *testing.T) {
		const (
			items = 50
			limit = 5
		)

		var mu sync.Mutex
		inFlight := 0
		maxInFlight := 0
		processed := 0

		err := parallelFor(make([]int, items), limit, func(int) error {
			mu.Lock()
			inFlight++
			if inFlight > maxInFlight {
				maxInFlight = inFlight
			}
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			inFlight--
			processed++
			mu.Unlock()
			return nil
		})

		require.NoError(t, err)
		require.Equal(t, items, processed)
		require.LessOrEqual(t, maxInFlight, limit)
		require.Greater(t, maxInFlight, 1)
	})

	t.Run("returns the first error", func(t *testing.T) {
		errBoom := errors.New("boom")
		err := parallelFor([]int{1, 2, 3}, 3, func(i int) error {
			if i == 2 {
				return errBoom
			}
			return nil
		})
		require.ErrorIs(t, err, errBoom)
	})

	t.Run("limit below 1 falls back to serial", func(t *testing.T) {
		processed := 0
		err := parallelFor([]int{1, 2, 3}, 0, func(int) error {
			processed++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 3, processed)
	})
}
