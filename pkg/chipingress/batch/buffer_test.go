package batch

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

func TestMessageBatch(t *testing.T) {
	t.Run("newMessageBatch creates empty batch", func(t *testing.T) {
		batch := newBuffer(10)
		require.NotNil(t, batch)
		assert.Equal(t, 0, batch.Len())
	})

	t.Run("Add appends messages", func(t *testing.T) {
		batch := newBuffer(10)

		msg1 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-1"},
		}
		msg2 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-2"},
		}

		batch.Add(msg1)
		assert.Equal(t, 1, batch.Len())

		batch.Add(msg2)
		assert.Equal(t, 2, batch.Len())
	})

	t.Run("Clear returns copy and empties batch", func(t *testing.T) {
		batch := newBuffer(10)

		msg1 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-1"},
		}
		msg2 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-2"},
		}

		batch.Add(msg1)
		batch.Add(msg2)

		result := batch.Clear()
		require.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "test-1", result[0].event.Id)
		assert.Equal(t, "test-2", result[1].event.Id)

		assert.Equal(t, 0, batch.Len())
	})

	t.Run("Clear on empty batch returns empty slice", func(t *testing.T) {
		batch := newBuffer(10)
		result := batch.Clear()
		require.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("Values returns copy without clearing", func(t *testing.T) {
		batch := newBuffer(10)

		msg1 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-1"},
		}
		msg2 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-2"},
		}

		batch.Add(msg1)
		batch.Add(msg2)

		result := batch.Values()
		require.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "test-1", result[0].event.Id)
		assert.Equal(t, "test-2", result[1].event.Id)

		// Batch should still have messages after Values
		assert.Equal(t, 2, batch.Len())
	})

	t.Run("Values on empty batch returns empty slice", func(t *testing.T) {
		batch := newBuffer(10)
		result := batch.Values()
		require.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("Values returns slice copy with same pointers", func(t *testing.T) {
		batch := newBuffer(10)

		msg1 := &messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-1"},
		}

		batch.Add(msg1)
		result := batch.Values()
		require.Len(t, result, 1)

		assert.Equal(t, msg1, result[0])
		assert.Same(t, msg1.event, result[0].event)

		batch.Add(&messageWithCallback{
			event: &chipingress.CloudEventPb{Id: "test-2"},
		})
		assert.Len(t, result, 1)
		assert.Equal(t, 2, batch.Len())
	})

	t.Run("concurrent Add operations are safe", func(t *testing.T) {
		batch := newBuffer(100)
		var wg sync.WaitGroup
		numGoroutines := 10
		messagesPerGoroutine := 10

		for range numGoroutines {
			wg.Go(func() {
				for range messagesPerGoroutine {
					batch.Add(&messageWithCallback{
						event: &chipingress.CloudEventPb{Id: "test"},
					})
				}
			})
		}

		wg.Wait()

		assert.Equal(t, numGoroutines*messagesPerGoroutine, batch.Len())
	})

	t.Run("concurrent Add and Clear operations are safe", func(t *testing.T) {
		batch := newBuffer(100)
		var wg sync.WaitGroup
		numAdders := 5
		numClears := 3

		for range numAdders {
			wg.Go(func() {
				for range 20 {
					batch.Add(&messageWithCallback{
						event: &chipingress.CloudEventPb{Id: "test"},
					})
				}
			})
		}

		// Start clearers
		for range numClears {
			wg.Go(func() {
				for range 10 {
					batch.Clear()
				}
			})
		}

		wg.Wait()
		assert.GreaterOrEqual(t, batch.Len(), 0)
	})

	t.Run("concurrent Len operations are safe", func(t *testing.T) {
		batch := newBuffer(100)
		var wg sync.WaitGroup
		for range 10 {
			batch.Add(&messageWithCallback{
				event: &chipingress.CloudEventPb{Id: "test"},
			})
		}

		for range 100 {
			wg.Go(func() {
				length := batch.Len()
				assert.GreaterOrEqual(t, length, 0)
			})
		}

		wg.Wait()
	})
}
