package batch

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/protobuf/proto"

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

	t.Run("WithEventClone", func(t *testing.T) {
		client, err := NewBatchClient(nil)
		require.NoError(t, err)
		assert.True(t, client.cloneEvent)

		client, err = NewBatchClient(nil, WithEventClone(false))
		require.NoError(t, err)
		assert.False(t, client.cloneEvent)
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

	t.Run("splits oversized batch by max gRPC request size", func(t *testing.T) {
		events := []*chipingress.CloudEventPb{
			largeTestEvent("test-id-1"),
			largeTestEvent("test-id-2"),
			largeTestEvent("test-id-3"),
			largeTestEvent("test-id-4"),
			largeTestEvent("test-id-5"),
		}
		maxRequestSize := proto.Size(&chipingress.CloudEventBatch{Events: events[:2]})
		require.LessOrEqual(t, proto.Size(&chipingress.CloudEventBatch{Events: events[:1]}), maxRequestSize)
		require.Greater(t, proto.Size(&chipingress.CloudEventBatch{Events: events[:3]}), maxRequestSize)

		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		callbackDone := make(chan error, len(events))
		var mu sync.Mutex
		var publishedIDs []string
		var publishedSizes []int

		mockClient.
			On("PublishBatch",
				mock.Anything,
				mock.MatchedBy(func(batch *chipingress.CloudEventBatch) bool {
					return len(batch.Events) > 0 && proto.Size(batch) <= maxRequestSize
				}),
			).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(args mock.Arguments) {
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				mu.Lock()
				for _, event := range batch.Events {
					publishedIDs = append(publishedIDs, event.Id)
				}
				publishedSizes = append(publishedSizes, proto.Size(batch))
				if len(publishedIDs) == len(events) {
					close(done)
				}
				mu.Unlock()
			}).
			Times(3)

		client, err := NewBatchClient(mockClient, WithMaxGRPCRequestSize(maxRequestSize))
		require.NoError(t, err)

		messages := make([]*messageWithCallback, 0, len(events))
		for _, event := range events {
			messages = append(messages, &messageWithCallback{
				event: event,
				callback: func(err error) {
					callbackDone <- err
				},
			})
		}

		client.sendBatch(t.Context(), messages)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for split batches to be sent")
		}
		for range events {
			select {
			case err := <-callbackDone:
				require.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for split batch callback")
			}
		}

		assert.Equal(t, []string{"test-id-1", "test-id-2", "test-id-3", "test-id-4", "test-id-5"}, publishedIDs)
		for _, size := range publishedSizes {
			assert.LessOrEqual(t, size, maxRequestSize)
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("doesn't publish a single event over max gRPC request size", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		callbackDone := make(chan error, 1)
		event := largeTestEvent("oversized-id")
		maxRequestSize := proto.Size(&chipingress.CloudEventBatch{Events: []*chipingress.CloudEventPb{event}}) - 1

		client, err := NewBatchClient(mockClient, WithMaxGRPCRequestSize(maxRequestSize))
		require.NoError(t, err)

		client.sendBatch(t.Context(), []*messageWithCallback{
			{
				event: event,
				callback: func(err error) {
					callbackDone <- err
				},
			},
		})

		select {
		case err := <-callbackDone:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exceeds max gRPC request size")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for oversized batch callback")
		}

		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
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
		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, nil).
			Maybe()

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

		// Stop the client — drains any buffered messages
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

	t.Run("clears seqnum counters on Stop", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		client, err := NewBatchClient(mockClient, WithBatchSize(10))
		require.NoError(t, err)

		_ = client.seqnumFor("domain-a", "entity-x")
		_ = client.seqnumFor("domain-b", "entity-y")
		assert.Equal(t, 2, countCounters(&client.counters))

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		client.Start(ctx)
		client.Stop()

		assert.Equal(t, 0, countCounters(&client.counters))
	})
}

func countCounters(counters *sync.Map) int {
	n := 0
	counters.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

func largeTestEvent(id string) *chipingress.CloudEventPb {
	return &chipingress.CloudEventPb{
		Id:          id,
		Source:      "test-source",
		Type:        "test.event.type",
		SpecVersion: "1.0",
		Data: &cepb.CloudEvent_BinaryData{
			BinaryData: []byte("0123456789abcdefghijklmnopqrstuvwxyz"),
		},
	}
}

func TestSeqnum(t *testing.T) {
	t.Run("dropped messages consume seqnum and create detectable gaps", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(1))
		require.NoError(t, err)

		first := &chipingress.CloudEventPb{Id: "id-1", Source: "domain-a", Type: "entity-x"}
		second := &chipingress.CloudEventPb{Id: "id-2", Source: "domain-a", Type: "entity-x"}
		third := &chipingress.CloudEventPb{Id: "id-3", Source: "domain-a", Type: "entity-x"}

		err = client.QueueMessage(first, nil)
		require.NoError(t, err)

		err = client.QueueMessage(second, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "message buffer is full")

		msg := <-client.messageBuffer
		require.NotNil(t, msg.event.Attributes["seqnum"])
		assert.Equal(t, "1", msg.event.Attributes["seqnum"].GetCeString())

		err = client.QueueMessage(third, nil)
		require.NoError(t, err)

		msg = <-client.messageBuffer
		require.NotNil(t, msg.event.Attributes["seqnum"])
		assert.Equal(t, "3", msg.event.Attributes["seqnum"].GetCeString())
	})

	t.Run("reusing event pointer preserves queued seqnum snapshots", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(2))
		require.NoError(t, err)

		event := &chipingress.CloudEventPb{Id: "id-1", Source: "domain-a", Type: "entity-x"}

		err = client.QueueMessage(event, nil)
		require.NoError(t, err)
		err = client.QueueMessage(event, nil)
		require.NoError(t, err)

		first := <-client.messageBuffer
		second := <-client.messageBuffer

		require.NotNil(t, first.event.Attributes["seqnum"])
		require.NotNil(t, second.event.Attributes["seqnum"])
		assert.Equal(t, "1", first.event.Attributes["seqnum"].GetCeString())
		assert.Equal(t, "2", second.event.Attributes["seqnum"].GetCeString())
	})

	t.Run("reusing event pointer can overwrite queued seqnum when clone disabled", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(2), WithEventClone(false))
		require.NoError(t, err)

		event := &chipingress.CloudEventPb{Id: "id-1", Source: "domain-a", Type: "entity-x"}

		err = client.QueueMessage(event, nil)
		require.NoError(t, err)
		err = client.QueueMessage(event, nil)
		require.NoError(t, err)

		first := <-client.messageBuffer
		second := <-client.messageBuffer

		require.NotNil(t, first.event.Attributes["seqnum"])
		require.NotNil(t, second.event.Attributes["seqnum"])
		assert.Equal(t, "2", first.event.Attributes["seqnum"].GetCeString())
		assert.Equal(t, "2", second.event.Attributes["seqnum"].GetCeString())
	})

	t.Run("stamps sequential seqnum for same source+type", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(10))
		require.NoError(t, err)

		events := []*chipingress.CloudEventPb{
			{Id: "id-1", Source: "domain-a", Type: "entity-x"},
			{Id: "id-2", Source: "domain-a", Type: "entity-x"},
			{Id: "id-3", Source: "domain-a", Type: "entity-x"},
		}

		for _, e := range events {
			err := client.QueueMessage(e, nil)
			require.NoError(t, err)
		}

		// Drain buffer and verify seqnums
		for i, expected := range []string{"1", "2", "3"} {
			msg := <-client.messageBuffer
			require.NotNil(t, msg.event.Attributes, "event %d should have attributes", i)
			seqAttr, ok := msg.event.Attributes["seqnum"]
			require.True(t, ok, "event %d should have seqnum attribute", i)
			assert.Equal(t, expected, seqAttr.GetCeString(), "event %d seqnum mismatch", i)
		}
	})

	t.Run("independent counters per source+type pair", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(10))
		require.NoError(t, err)

		// Queue events with different source+type combinations
		events := []*chipingress.CloudEventPb{
			{Id: "a1", Source: "domain-a", Type: "entity-x"},
			{Id: "b1", Source: "domain-b", Type: "entity-y"},
			{Id: "a2", Source: "domain-a", Type: "entity-x"},
			{Id: "b2", Source: "domain-b", Type: "entity-y"},
			{Id: "c1", Source: "domain-a", Type: "entity-z"}, // same domain, different type
		}

		for _, e := range events {
			err := client.QueueMessage(e, nil)
			require.NoError(t, err)
		}

		// Expected seqnums by event ID
		expected := map[string]string{
			"a1": "1", // first for domain-a/entity-x
			"b1": "1", // first for domain-b/entity-y
			"a2": "2", // second for domain-a/entity-x
			"b2": "2", // second for domain-b/entity-y
			"c1": "1", // first for domain-a/entity-z (new type)
		}

		for range events {
			msg := <-client.messageBuffer
			seqAttr := msg.event.Attributes["seqnum"]
			require.NotNil(t, seqAttr)
			assert.Equal(t, expected[msg.event.Id], seqAttr.GetCeString(),
				"seqnum mismatch for event %s", msg.event.Id)
		}
	})

	t.Run("source and type values containing separator do not collide", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(10))
		require.NoError(t, err)

		events := []*chipingress.CloudEventPb{
			{Id: "sep-1", Source: "a\x00b", Type: "c"},
			{Id: "sep-2", Source: "a", Type: "b\x00c"},
		}

		for _, e := range events {
			err := client.QueueMessage(e, nil)
			require.NoError(t, err)
		}

		expected := map[string]string{
			"sep-1": "1",
			"sep-2": "1",
		}

		for range events {
			msg := <-client.messageBuffer
			seqAttr := msg.event.Attributes["seqnum"]
			require.NotNil(t, seqAttr)
			assert.Equal(t, expected[msg.event.Id], seqAttr.GetCeString(),
				"seqnum mismatch for event %s", msg.event.Id)
		}
	})

	t.Run("concurrent access produces unique seqnums", func(t *testing.T) {
		client, err := NewBatchClient(nil, WithMessageBuffer(1000))
		require.NoError(t, err)

		const numGoroutines = 50
		const eventsPerGoroutine = 10
		totalEvents := numGoroutines * eventsPerGoroutine

		queueErrors := make(chan error, totalEvents)
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				defer wg.Done()
				for i := 0; i < eventsPerGoroutine; i++ {
					event := &chipingress.CloudEventPb{
						Id:     strconv.Itoa(goroutineID*eventsPerGoroutine + i),
						Source: "concurrent-domain",
						Type:   "concurrent-type",
					}
					if err := client.QueueMessage(event, nil); err != nil {
						queueErrors <- err
					}
				}
			}(g)
		}

		wg.Wait()
		close(queueErrors)

		var collectedQueueErrors []error
		for err := range queueErrors {
			collectedQueueErrors = append(collectedQueueErrors, err)
		}
		require.Empty(t, collectedQueueErrors, "expected all concurrent QueueMessage calls to succeed")

		// Collect all seqnums
		seqnums := make([]uint64, 0, totalEvents)
		for i := 0; i < totalEvents; i++ {
			msg := <-client.messageBuffer
			seqAttr := msg.event.Attributes["seqnum"]
			require.NotNil(t, seqAttr)
			seq, err := strconv.ParseUint(seqAttr.GetCeString(), 10, 64)
			require.NoError(t, err)
			seqnums = append(seqnums, seq)
		}

		// Sort and verify all unique and in range [1, totalEvents]
		sort.Slice(seqnums, func(i, j int) bool { return seqnums[i] < seqnums[j] })

		expectedSeq := uint64(1)
		for i, seq := range seqnums {
			assert.Equal(t, expectedSeq, seq, "seqnum at index %d should be %d", i, expectedSeq)
			expectedSeq++
		}
	})
}

func TestBatchClient_Metrics(t *testing.T) {
	t.Run("records success path metrics", func(t *testing.T) {
		reader, restore := useTestMeterProvider(t)
		defer restore()

		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(
			mockClient,
			WithBatchSize(1),
			WithBatchInterval(time.Second),
			WithMessageBuffer(10),
			WithMaxGRPCRequestSize(2048),
		)
		require.NoError(t, err)
		client.Start(t.Context())

		err = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "metric-success",
			Source: "platform",
			Type:   "MetricSuccess",
		}, nil)
		require.NoError(t, err)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for PublishBatch")
		}

		client.Stop()
		rm := collectResourceMetrics(t, reader)

		reqTotal := mustMetric(t, rm, "chip_ingress.batch.send_requests_total")
		reqSum, ok := reqTotal.Data.(metricdata.Sum[int64])
		require.True(t, ok)
		successPoint := mustInt64SumPointWithAttr(t, reqSum, "status", "success")
		assert.GreaterOrEqual(t, successPoint.Value, int64(1))

		msgSize := mustMetric(t, rm, "chip_ingress.batch.request_size_messages")
		msgSizeHist, ok := msgSize.Data.(metricdata.Histogram[int64])
		require.True(t, ok)
		msgSizePoint := mustInt64HistogramPointWithIntAttr(t, msgSizeHist, "max_batch_size", 1)
		assert.GreaterOrEqual(t, msgSizePoint.Count, uint64(1))

		reqSize := mustMetric(t, rm, "chip_ingress.batch.request_size_bytes")
		reqSizeHist, ok := reqSize.Data.(metricdata.Histogram[int64])
		require.True(t, ok)
		reqSizePoint := mustInt64HistogramPointWithIntAttr(t, reqSizeHist, "max_grpc_request_size_bytes", 2048)
		assert.GreaterOrEqual(t, reqSizePoint.Count, uint64(1))

		latency := mustMetric(t, rm, "chip_ingress.batch.request_latency_ms")
		latencyHist, ok := latency.Data.(metricdata.Histogram[float64])
		require.True(t, ok)
		latencyPoint := mustFloat64HistogramPointWithAttr(t, latencyHist, "status", "success")
		assert.GreaterOrEqual(t, latencyPoint.Count, uint64(1))

		config := mustMetric(t, rm, "chip_ingress.batch.config.info")
		configGauge, ok := config.Data.(metricdata.Gauge[int64])
		require.True(t, ok)
		require.NotEmpty(t, configGauge.DataPoints)
		assert.Equal(t, int64(1), configGauge.DataPoints[0].Value)
		assert.True(t, hasIntAttr(configGauge.DataPoints[0].Attributes, "max_batch_size", 1))
		assert.True(t, hasIntAttr(configGauge.DataPoints[0].Attributes, "max_grpc_request_size_bytes", 2048))
	})

	t.Run("records failure counters and latency", func(t *testing.T) {
		reader, restore := useTestMeterProvider(t)
		defer restore()

		mockClient := mocks.NewClient(t)
		done := make(chan struct{})
		mockClient.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, assert.AnError).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		client, err := NewBatchClient(mockClient, WithBatchSize(1), WithMessageBuffer(10))
		require.NoError(t, err)
		client.Start(t.Context())

		err = client.QueueMessage(&chipingress.CloudEventPb{
			Id:     "metric-failure",
			Source: "platform",
			Type:   "MetricFailure",
		}, nil)
		require.NoError(t, err)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for PublishBatch")
		}

		client.Stop()
		rm := collectResourceMetrics(t, reader)

		reqTotal := mustMetric(t, rm, "chip_ingress.batch.send_requests_total")
		reqSum, ok := reqTotal.Data.(metricdata.Sum[int64])
		require.True(t, ok)
		failureReq := mustInt64SumPointWithAttr(t, reqSum, "status", "failure")
		assert.GreaterOrEqual(t, failureReq.Value, int64(1))

		failures := mustMetric(t, rm, "chip_ingress.batch.send_failures_total")
		failuresSum, ok := failures.Data.(metricdata.Sum[int64])
		require.True(t, ok)
		require.NotEmpty(t, failuresSum.DataPoints)
		assert.GreaterOrEqual(t, failuresSum.DataPoints[0].Value, int64(1))

		latency := mustMetric(t, rm, "chip_ingress.batch.request_latency_ms")
		latencyHist, ok := latency.Data.(metricdata.Histogram[float64])
		require.True(t, ok)
		failureLatency := mustFloat64HistogramPointWithAttr(t, latencyHist, "status", "failure")
		assert.GreaterOrEqual(t, failureLatency.Count, uint64(1))
	})
}

func BenchmarkBatchClient_QueueMessage(b *testing.B) {
	client, err := NewBatchClient(
		&chipingress.NoopClient{},
		WithBatchSize(b.N+1),
		WithMessageBuffer(b.N+10),
		WithBatchInterval(time.Hour),
	)
	if err != nil {
		b.Fatal(err)
	}
	client.Start(context.Background())
	defer client.Stop()

	payload := &chipingress.CloudEventPb{
		Source: "bench",
		Type:   "bench.event",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload.Id = strconv.Itoa(i)
		if err := client.QueueMessage(payload, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func useTestMeterProvider(t *testing.T) (*sdkmetric.ManualReader, func()) {
	t.Helper()
	prev := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	return reader, func() {
		require.NoError(t, provider.Shutdown(t.Context()))
		otel.SetMeterProvider(prev)
	}
}

func collectResourceMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))
	return rm
}

func mustMetric(t *testing.T, rm metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func mustInt64SumPointWithAttr(t *testing.T, sum metricdata.Sum[int64], key, want string) metricdata.DataPoint[int64] {
	t.Helper()
	for _, dp := range sum.DataPoints {
		if hasStringAttr(dp.Attributes, key, want) {
			return dp
		}
	}
	t.Fatalf("sum datapoint with %s=%s not found", key, want)
	return metricdata.DataPoint[int64]{}
}

func mustInt64HistogramPointWithIntAttr(t *testing.T, hist metricdata.Histogram[int64], key string, want int) metricdata.HistogramDataPoint[int64] {
	t.Helper()
	for _, dp := range hist.DataPoints {
		if hasIntAttr(dp.Attributes, key, want) {
			return dp
		}
	}
	t.Fatalf("histogram datapoint with %s=%d not found", key, want)
	return metricdata.HistogramDataPoint[int64]{}
}

func mustFloat64HistogramPointWithAttr(t *testing.T, hist metricdata.Histogram[float64], key, want string) metricdata.HistogramDataPoint[float64] {
	t.Helper()
	for _, dp := range hist.DataPoints {
		if hasStringAttr(dp.Attributes, key, want) {
			return dp
		}
	}
	t.Fatalf("histogram datapoint with %s=%s not found", key, want)
	return metricdata.HistogramDataPoint[float64]{}
}

func hasStringAttr(set attribute.Set, key, want string) bool {
	for _, kv := range set.ToSlice() {
		if string(kv.Key) == key {
			return kv.Value.AsString() == want
		}
	}
	return false
}

func hasIntAttr(set attribute.Set, key string, want int) bool {
	for _, kv := range set.ToSlice() {
		if string(kv.Key) == key {
			return int(kv.Value.AsInt64()) == want
		}
	}
	return false
}
