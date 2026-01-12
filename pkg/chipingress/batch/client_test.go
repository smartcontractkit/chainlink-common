package batch

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
)

func TestNewBatchClient(t *testing.T) {
	t.Run("NewBatchClient", func(t *testing.T) {
		client, err := NewBatchClient(nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithBatchSize(100))
		require.NoError(t, err)
		assert.Equal(t, 100, client.batchSize)
	})

	t.Run("WithMaxConcurrentSends", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMaxConcurrentSends(10))
		require.NoError(t, err)
		assert.Equal(t, 10, cap(client.maxConcurrentSends))
	})

	t.Run("WithBatchTimeout", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithBatchTimeout(100*time.Millisecond))
		require.NoError(t, err)
		assert.Equal(t, 100*time.Millisecond, client.batchInterval)
	})

	t.Run("WithCompressionType", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithCompressionType("gzip"))
		require.NoError(t, err)
		assert.Equal(t, "gzip", client.compressionType)
	})

	t.Run("WithMessageBuffer", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(1000))
		require.NoError(t, err)
		assert.Equal(t, 1000, cap(client.messageBuffer))
	})
}

func TestQueueMessage(t *testing.T) {
	t.Run("successfully queues a message", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(5))
		require.NoError(t, err)

		event := &chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}

		client.QueueMessage(event)

		assert.Len(t, client.messageBuffer, 1)

		received := <-client.messageBuffer
		assert.Equal(t, event.Id, received.Id)
		assert.Equal(t, event.Source, received.Source)
		assert.Equal(t, event.Type, received.Type)
	})

	t.Run("drops message if buffer is full", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(1))
		require.NoError(t, err)
		require.NotNil(t, client)

		event := &chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}

		client.QueueMessage(event)
		client.QueueMessage(event)

		assert.Len(t, client.messageBuffer, 1)

		received := <-client.messageBuffer
		assert.Equal(t, event.Id, received.Id)
		assert.Equal(t, event.Source, received.Source)
		assert.Equal(t, event.Type, received.Type)
	})

	t.Run("handles nil event", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(5))
		require.NoError(t, err)

		client.QueueMessage(nil)
		assert.Empty(t, client.messageBuffer)
	})
}

func TestSendBatch(t *testing.T) {
	t.Run("successfully sends a batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					// verify the batch contains the expected events
					return len(batch.Events) == 3 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2" &&
						batch.Events[2].Id == "test-id-3"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).Run(func(args mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithMessageBuffer(5))
		require.NoError(t, err)

		events := []*chipingress.CloudEventPb{
			{Id: "test-id-1", Source: "test-source", Type: "test.event.type"},
			{Id: "test-id-2", Source: "test-source", Type: "test.event.type"},
			{Id: "test-id-3", Source: "test-source", Type: "test.event.type"},
		}

		client.sendBatch(t.Context(), events)

		// wait for the internal goroutine to complete
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("doesn't publish empty batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)

		client, err := NewBatchClient(mockClient, WithMessageBuffer(5))
		require.NoError(t, err)

		client.sendBatch(t.Context(), []*chipingress.CloudEventPb{})

		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
	})

	t.Run("sends multiple batches successfully", func(t *testing.T) {
		mockClient := mocks.NewClient(t)

		done := make(chan struct{})
		callCount := 0

		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) {
				callCount++
				if callCount == 3 {
					close(done)
				}
			}).
			Times(3)

		client, err := NewBatchClient(mockClient, WithMessageBuffer(5))
		require.NoError(t, err)

		batch1 := []*chipingress.CloudEventPb{
			{Id: "batch1-id-1", Source: "test-source", Type: "test.event.type"},
		}
		batch2 := []*chipingress.CloudEventPb{
			{Id: "batch2-id-1", Source: "test-source", Type: "test.event.type"},
			{Id: "batch2-id-2", Source: "test-source", Type: "test.event.type"},
		}
		batch3 := []*chipingress.CloudEventPb{
			{Id: "batch3-id-1", Source: "test-source", Type: "test.event.type"},
		}

		client.sendBatch(context.Background(), batch1)
		client.sendBatch(context.Background(), batch2)
		client.sendBatch(context.Background(), batch3)

		// wait for the internal goroutines to complete
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for multiple batches to be sent")
		}

		mockClient.AssertExpectations(t)
	})
}

func TestStart(t *testing.T) {
	t.Run("batch size trigger", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 3 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2" &&
						batch.Events[2].Id == "test-id-3"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(3), WithBatchTimeout(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"})
		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"})
		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-3", Source: "test-source", Type: "test.event.type"})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("timeout trigger", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchTimeout(50*time.Millisecond))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"})
		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent after timeout")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("context cancellation flushes pending batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.MatchedBy(func(ctx context.Context) bool {
					return ctx != nil
				}),
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchTimeout(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())

		client.Start(ctx)

		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"})
		client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"})

		time.Sleep(10 * time.Millisecond)

		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for flush on context cancellation")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("stop flushes pending batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchTimeout(100*time.Millisecond), WithMessageBuffer(10))
		require.NoError(t, err)

		client.Start(t.Context())

		queued1 := client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"})
		queued2 := client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"})
		require.True(t, queued1)
		require.True(t, queued2)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch timeout to trigger")
		}

		client.Stop()

		mockClient.AssertExpectations(t)
	})

	t.Run("no flush when batch is empty", func(t *testing.T) {
		mockClient := mocks.NewClient(t)

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchTimeout(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		client.Start(ctx)

		time.Sleep(10 * time.Millisecond)

		cancel()

		time.Sleep(50 * time.Millisecond)

		mockClient.AssertNotCalled(t, "PublishBatch")
	})

	t.Run("multiple batches via size trigger", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callCount := 0

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) {
				callCount++
				if callCount == 3 {
					close(done)
				}
			}).
			Times(3)

		client, err := NewBatchClient(mockClient, WithBatchSize(2), WithBatchTimeout(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		for i := 1; i <= 6; i++ {
			client.QueueMessage(&chipingress.CloudEventPb{
				Id:     "test-id-" + strconv.Itoa(i),
				Source: "test-source",
				Type:   "test.event.type",
			})
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for multiple batches to be sent")
		}

		mockClient.AssertExpectations(t)
	})
}
