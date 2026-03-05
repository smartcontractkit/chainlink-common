package beholder_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func newTestConfig() beholder.Config {
	return beholder.Config{
		ChipIngressBufferSize:   10,
		ChipIngressMaxBatchSize: 5,
		ChipIngressSendInterval: 50 * time.Millisecond,
		ChipIngressSendTimeout:  5 * time.Second,
	}
}

func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()
	lggr, err := logger.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = lggr.Sync() })
	return lggr
}

func TestNewChipIngressBatchEmitter(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		assert.NotNil(t, emitter)
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		emitter, err := beholder.NewChipIngressBatchEmitter(nil, newTestConfig(), newTestLogger(t))
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}

func TestChipIngressBatchEmitter_Emit(t *testing.T) {
	t.Run("returns error when domain/entity missing", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		err = emitter.Emit(t.Context(), []byte("test"), "bad_key", "bad_value")
		assert.Error(t, err)
	})

	t.Run("enqueues and does not call PublishBatch immediately", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "test-domain",
			beholder.AttrKeyEntity, "test-entity",
		)
		require.NoError(t, err)

		// PublishBatch should NOT have been called yet (event is just buffered)
		clientMock.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
	})
}

func TestChipIngressBatchEmitter_BatchAssembly(t *testing.T) {
	t.Run("events are batched and sent via PublishBatch", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		var receivedBatches []*chipingress.CloudEventBatch
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				receivedBatches = append(receivedBatches, batch)
			}).
			Return(nil, nil)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		// Wait for flush to occur
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(receivedBatches) > 0
		}, 2*time.Second, 10*time.Millisecond)

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()

		totalEvents := 0
		for _, batch := range receivedBatches {
			totalEvents += len(batch.Events)
		}
		assert.Equal(t, 3, totalEvents)
	})
}

func TestChipIngressBatchEmitter_MaxBatchSize(t *testing.T) {
	t.Run("batch is capped at maxBatchSize", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		var batchSizes []int
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				batchSizes = append(batchSizes, len(batch.Events))
			}).
			Return(nil, nil)

		cfg := newTestConfig()
		cfg.ChipIngressBufferSize = 20
		cfg.ChipIngressMaxBatchSize = 3
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 7; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		// Wait for all events to be flushed
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			total := 0
			for _, s := range batchSizes {
				total += s
			}
			return total >= 7
		}, 2*time.Second, 10*time.Millisecond)

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		for _, size := range batchSizes {
			assert.LessOrEqual(t, size, 3, "batch size should not exceed maxBatchSize")
		}
	})
}

func TestChipIngressBatchEmitter_PerDomainEntityIsolation(t *testing.T) {
	t.Run("separate workers for different domain/entity pairs", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		domainEntitySeen := make(map[string]int)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				for _, event := range batch.Events {
					key := event.Source + "/" + event.Type
					domainEntitySeen[key] += 1
				}
			}).
			Return(nil, nil)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		// Emit events for two different domain/entity pairs
		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("workflow"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "WorkflowEvent",
			)
			require.NoError(t, err)
		}
		for i := 0; i < 2; i++ {
			err = emitter.Emit(t.Context(), []byte("bridge"),
				beholder.AttrKeyDomain, "data-feeds",
				beholder.AttrKeyEntity, "BridgeStatus",
			)
			require.NoError(t, err)
		}

		// Wait for both to be flushed
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return domainEntitySeen["platform/WorkflowEvent"] >= 3 &&
				domainEntitySeen["data-feeds/BridgeStatus"] >= 2
		}, 2*time.Second, 10*time.Millisecond)

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 3, domainEntitySeen["platform/WorkflowEvent"])
		assert.Equal(t, 2, domainEntitySeen["data-feeds/BridgeStatus"])
	})
}

func TestChipIngressBatchEmitter_BufferFull(t *testing.T) {
	t.Run("events are dropped when buffer is full", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		// Block PublishBatch so the buffer fills up
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		cfg := newTestConfig()
		cfg.ChipIngressBufferSize = 3
		cfg.ChipIngressSendInterval = 10 * time.Second // very long interval to prevent flushing

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		// Fill the buffer (3 events)
		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		// This should not error (it drops silently), but the event won't be delivered
		err = emitter.Emit(t.Context(), []byte("dropped"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		assert.NoError(t, err)
	})
}

func TestChipIngressBatchEmitter_Lifecycle(t *testing.T) {
	t.Run("start and close cleanly", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)

		require.NoError(t, emitter.Start(t.Context()))

		// Emit a few events to create workers
		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		// Close should not hang or error
		require.NoError(t, emitter.Close())
	})
}

func TestChipIngressBatchEmitter_CloudEventFormat(t *testing.T) {
	t.Run("CloudEvents have correct source, type, and data", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		var receivedBatch *chipingress.CloudEventBatch
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				receivedBatch = args.Get(1).(*chipingress.CloudEventBatch)
			}).
			Return(nil, nil)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		err = emitter.Emit(t.Context(), []byte("test-payload"),
			beholder.AttrKeyDomain, "my-domain",
			beholder.AttrKeyEntity, "my-entity",
		)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return receivedBatch != nil
		}, 2*time.Second, 10*time.Millisecond)

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		require.Len(t, receivedBatch.Events, 1)

		event := receivedBatch.Events[0]
		assert.Equal(t, "my-domain", event.Source)
		assert.Equal(t, "my-entity", event.Type)
		assert.NotEmpty(t, event.Id)
	})
}

func TestChipIngressBatchEmitter_PublishBatchError(t *testing.T) {
	t.Run("PublishBatch error is handled gracefully", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		callCount := 0
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(_ mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				callCount++
			}).
			Return(nil, assert.AnError)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return callCount > 0
		}, 2*time.Second, 10*time.Millisecond)

		require.NoError(t, emitter.Close())
	})
}

func TestChipIngressBatchEmitter_ContextCancellation(t *testing.T) {
	t.Run("Emit returns context error when context is cancelled", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		cfg := newTestConfig()
		cfg.ChipIngressBufferSize = 1
		cfg.ChipIngressSendInterval = 10 * time.Second

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		// Fill the buffer so the next Emit will block on channel send
		err = emitter.Emit(t.Context(), []byte("fill"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err = emitter.Emit(ctx, []byte("should-fail"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestChipIngressBatchEmitter_DefaultConfig(t *testing.T) {
	t.Run("zero config uses sane defaults", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		var receivedBatch *chipingress.CloudEventBatch
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				receivedBatch = args.Get(1).(*chipingress.CloudEventBatch)
			}).
			Return(nil, nil)

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, beholder.Config{}, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		require.NoError(t, err)

		// Default send interval is 500ms; wait for flush
		assert.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return receivedBatch != nil
		}, 3*time.Second, 50*time.Millisecond)

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		require.Len(t, receivedBatch.Events, 1)
	})
}

func TestChipIngressBatchEmitter_GracefulDrain(t *testing.T) {
	t.Run("flushes buffered events on close", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		totalEventsSent := 0
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				totalEventsSent += len(batch.Events)
			}).
			Return(nil, nil)

		cfg := beholder.Config{
			ChipIngressBufferSize:   20,
			ChipIngressMaxBatchSize: 10,
			ChipIngressSendInterval: 1 * time.Hour, // very long interval — events won't flush via tick
			ChipIngressSendTimeout:  5 * time.Second,
			ChipIngressDrainTimeout: 5 * time.Second,
		}

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 5; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		// Events are buffered but no tick has fired. Close should drain them.
		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 5, totalEventsSent, "all buffered events should be drained on close")
	})
}

func TestChipIngressBatchEmitter_DrainMultipleDomains(t *testing.T) {
	t.Run("drains events from all workers on close", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		domainEntitySent := make(map[string]int)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				for _, event := range batch.Events {
					key := event.Source + "/" + event.Type
					domainEntitySent[key] += 1
				}
			}).
			Return(nil, nil)

		cfg := beholder.Config{
			ChipIngressBufferSize:   20,
			ChipIngressMaxBatchSize: 10,
			ChipIngressSendInterval: 1 * time.Hour,
			ChipIngressSendTimeout:  5 * time.Second,
			ChipIngressDrainTimeout: 5 * time.Second,
		}

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 3; i++ {
			err = emitter.Emit(t.Context(), []byte("workflow"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "WorkflowEvent",
			)
			require.NoError(t, err)
		}
		for i := 0; i < 2; i++ {
			err = emitter.Emit(t.Context(), []byte("bridge"),
				beholder.AttrKeyDomain, "data-feeds",
				beholder.AttrKeyEntity, "BridgeStatus",
			)
			require.NoError(t, err)
		}

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 3, domainEntitySent["platform/WorkflowEvent"])
		assert.Equal(t, 2, domainEntitySent["data-feeds/BridgeStatus"])
	})
}

func TestChipIngressBatchEmitter_DrainPublishBatchFailure(t *testing.T) {
	t.Run("drain continues attempting batches after failure", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		var mu sync.Mutex
		callCount := 0
		totalEventsSent := 0
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				callCount++
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				if callCount == 1 {
					return // first call fails
				}
				totalEventsSent += len(batch.Events)
			}).
			Return(nil, assert.AnError).Once()

		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				mu.Lock()
				defer mu.Unlock()
				callCount++
				batch := args.Get(1).(*chipingress.CloudEventBatch)
				totalEventsSent += len(batch.Events)
			}).
			Return(nil, nil)

		cfg := beholder.Config{
			ChipIngressBufferSize:   20,
			ChipIngressMaxBatchSize: 3,
			ChipIngressSendInterval: 1 * time.Hour,
			ChipIngressSendTimeout:  5 * time.Second,
			ChipIngressDrainTimeout: 5 * time.Second,
		}

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		// Emit 6 events with maxBatchSize=3 => 2 batches during drain
		for i := 0; i < 6; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		require.NoError(t, emitter.Close())

		mu.Lock()
		defer mu.Unlock()
		assert.GreaterOrEqual(t, callCount, 2, "drain should have attempted at least 2 batches")
		assert.Equal(t, 3, totalEventsSent, "second batch should have succeeded despite first batch failure")
	})
}

func TestChipIngressBatchEmitter_DrainTimeout(t *testing.T) {
	t.Run("close returns promptly when drain timeout expires", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				ctx := args.Get(0).(context.Context)
				<-ctx.Done() // simulate a slow server that only returns on context cancellation
			}).
			Return(nil, context.DeadlineExceeded).
			Maybe()

		cfg := beholder.Config{
			ChipIngressBufferSize:   20,
			ChipIngressMaxBatchSize: 10,
			ChipIngressSendInterval: 1 * time.Hour,
			ChipIngressSendTimeout:  10 * time.Second,
			ChipIngressDrainTimeout: 200 * time.Millisecond,
		}

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		for i := 0; i < 5; i++ {
			err = emitter.Emit(t.Context(), []byte("body"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
			require.NoError(t, err)
		}

		closeDone := make(chan error, 1)
		go func() {
			closeDone <- emitter.Close()
		}()

		select {
		case err := <-closeDone:
			assert.NoError(t, err, "close should not error")
		case <-time.After(5 * time.Second):
			t.Fatal("Close() did not return within 5s; drain timeout is not working")
		}
	})
}

func TestChipIngressBatchEmitter_MaxWorkersCap(t *testing.T) {
	t.Run("drops events when max workers reached", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		cfg := newTestConfig()
		cfg.ChipIngressMaxWorkers = 2
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		// Create 2 workers (at the cap)
		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "domain-1",
			beholder.AttrKeyEntity, "entity-1",
		)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "domain-2",
			beholder.AttrKeyEntity, "entity-2",
		)
		require.NoError(t, err)

		// 3rd unique pair should be silently dropped (no error, no worker created)
		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "domain-3",
			beholder.AttrKeyEntity, "entity-3",
		)
		assert.NoError(t, err, "Emit should not error when max workers is reached")

		// Existing workers should still accept events
		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "domain-1",
			beholder.AttrKeyEntity, "entity-1",
		)
		assert.NoError(t, err, "Emit to existing worker should still work")
	})
}

func TestChipIngressBatchEmitter_EmitAfterClose(t *testing.T) {
	t.Run("Emit after Close returns error", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		require.NoError(t, emitter.Close())

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		assert.Error(t, err, "Emit after Close should return an error")
	})
}
