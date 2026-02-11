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

	t.Run("WithBatchInterval", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithBatchInterval(100*time.Millisecond))
		require.NoError(t, err)
		assert.Equal(t, 100*time.Millisecond, client.batchInterval)
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

		err = client.QueueMessage(event, nil)
		require.NoError(t, err)

		assert.Len(t, client.messageBuffer, 1)

		received := <-client.messageBuffer
		assert.Equal(t, event.Id, received.event.Id)
		assert.Equal(t, event.Source, received.event.Source)
		assert.Equal(t, event.Type, received.event.Type)
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

		_ = client.QueueMessage(event, nil)
		_ = client.QueueMessage(event, nil)

		assert.Len(t, client.messageBuffer, 1)

		received := <-client.messageBuffer
		assert.Equal(t, event.Id, received.event.Id)
		assert.Equal(t, event.Source, received.event.Source)
		assert.Equal(t, event.Type, received.event.Type)
	})

	t.Run("handles nil event", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(5))
		require.NoError(t, err)

		err = client.QueueMessage(nil, nil)
		require.NoError(t, err)
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
			Return(&chipingress.PublishResponse{}, nil).Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithMessageBuffer(5))
		require.NoError(t, err)

		messages := []*messageWithCallback{
			{event: &chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"}},
			{event: &chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"}},
			{event: &chipingress.CloudEventPb{Id: "test-id-3", Source: "test-source", Type: "test.event.type"}},
		}

		client.sendBatch(t.Context(), messages)

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

		client.sendBatch(t.Context(), []*messageWithCallback{})

		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
	})

	t.Run("sends multiple batches successfully", func(t *testing.T) {
		mockClient := mocks.NewClient(t)

		done := make(chan struct{})
		callCount := 0

		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) {
				callCount++
				if callCount == 3 {
					close(done)
				}
			}).
			Times(3)

		client, err := NewBatchClient(mockClient, WithMessageBuffer(5))
		require.NoError(t, err)

		batch1 := []*messageWithCallback{
			{event: &chipingress.CloudEventPb{Id: "batch1-id-1", Source: "test-source", Type: "test.event.type"}},
		}
		batch2 := []*messageWithCallback{
			{event: &chipingress.CloudEventPb{Id: "batch2-id-1", Source: "test-source", Type: "test.event.type"}},
			{event: &chipingress.CloudEventPb{Id: "batch2-id-2", Source: "test-source", Type: "test.event.type"}},
		}
		batch3 := []*messageWithCallback{
			{event: &chipingress.CloudEventPb{Id: "batch3-id-1", Source: "test-source", Type: "test.event.type"}},
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
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(3), WithBatchInterval(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		err = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"}, nil)
		require.NoError(t, err)
		err = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"}, nil)
		require.NoError(t, err)
		err = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-3", Source: "test-source", Type: "test.event.type"}, nil)
		require.NoError(t, err)

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
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(50*time.Millisecond))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"}, nil)
		_ = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"}, nil)

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
					// Regression guard: flush on cancellation must not use an already-canceled context.
					return ctx != nil && ctx.Err() == nil
				}),
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2 &&
						batch.Events[0].Id == "test-id-1" &&
						batch.Events[1].Id == "test-id-2"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"}, nil)
		_ = client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"}, nil)

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
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(100*time.Millisecond), WithMessageBuffer(10))
		require.NoError(t, err)

		client.Start(t.Context())

		queued1 := client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-1", Source: "test-source", Type: "test.event.type"}, nil)
		queued2 := client.QueueMessage(&chipingress.CloudEventPb{Id: "test-id-2", Source: "test-source", Type: "test.event.type"}, nil)
		require.NoError(t, queued1)
		require.NoError(t, queued2)

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

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(5*time.Second))
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
			Run(func(_ mock.Arguments) {
				callCount++
				if callCount == 3 {
					close(done)
				}
			}).
			Times(3)

		client, err := NewBatchClient(mockClient, WithBatchSize(2), WithBatchInterval(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		for i := 1; i <= 6; i++ {
			_ = client.QueueMessage(&chipingress.CloudEventPb{
				Id:     "test-id-" + strconv.Itoa(i),
				Source: "test-source",
				Type:   "test.event.type",
			}, nil)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for multiple batches to be sent")
		}

		mockClient.AssertExpectations(t)
	})
}

func TestCallbacks(t *testing.T) {
	t.Run("callback invoked on successful send", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, 1)

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 1 &&
						batch.Events[0].Id == "test-id-1"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(1))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callbackDone <- err
		})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		// wait for callback
		select {
		case err := <-callbackDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("callback receives error on failed send", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, 1)
		expectedErr := assert.AnError

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 1 &&
						batch.Events[0].Id == "test-id-1"
				}),
			).
			Return(&chipingress.PublishResponse{}, expectedErr).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(1))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callbackDone <- err
		})

		// wait for batch to be sent
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		// wait for callback to be invoked with error
		select {
		case err := <-callbackDone:
			assert.Equal(t, expectedErr, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("nil callback works without panic", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 1 &&
						batch.Events[0].Id == "test-id-1"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(1))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		// Queue message with nil callback - should not panic
		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, nil)

		// wait for batch to be sent
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("multiple messages with different callbacks", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callback1Done := make(chan error, 1)
		callback2Done := make(chan error, 1)
		callback3Done := make(chan error, 1)

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 3
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(3))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callback1Done <- err
		})

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-2",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callback2Done <- err
		})

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-3",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callback3Done <- err
		})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		select {
		case err := <-callback1Done:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback 1")
		}

		select {
		case err := <-callback2Done:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback 2")
		}

		select {
		case err := <-callback3Done:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback 3")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("callback invoked for timeout-triggered batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, 1)

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 1 &&
						batch.Events[0].Id == "test-id-1"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(50*time.Millisecond))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callbackDone <- err
		})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		// wait for callback
		select {
		case err := <-callbackDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("callback invoked for size-triggered batch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, 1)

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 2
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(2), WithBatchInterval(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, nil)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-2",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callbackDone <- err
		})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for batch to be sent")
		}

		select {
		case err := <-callbackDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("callbacks invoked on context cancellation", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, 1)

		mockClient.
			On("PublishBatch",
				mock.MatchedBy(func(ctx context.Context) bool {
					// Regression guard: flush on cancellation must not use an already-canceled context.
					return ctx != nil && ctx.Err() == nil
				}),
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) == 1 &&
						batch.Events[0].Id == "test-id-1"
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(10), WithBatchInterval(5*time.Second))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())

		client.Start(ctx)

		_ = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, func(err error) {
			callbackDone <- err
		})

		time.Sleep(10 * time.Millisecond)

		// cancel context to trigger flush
		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for flush on cancellation")
		}

		select {
		case err := <-callbackDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mockClient.AssertExpectations(t)
	})
}

func TestStop(t *testing.T) {
	t.Run("can call Stop multiple times without panic", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		client, err := NewBatchClient(mockClient, WithBatchSize(10))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)
		client.Stop()
		client.Stop()
		client.Stop()
	})

	t.Run("QueueMessage returns error after Stop", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		client, err := NewBatchClient(mockClient, WithBatchSize(10))
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)

		// Queue message before stop - should succeed
		err = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-1",
			Source: "test-source",
			Type:   "test.event.type",
		}, nil)
		require.NoError(t, err)

		// Stop the client
		client.Stop()

		// Queue message after stop - should fail
		err = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "test-id-2",
			Source: "test-source",
			Type:   "test.event.type",
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shutdown")
	})
}
