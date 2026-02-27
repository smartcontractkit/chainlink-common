package beholder_test

import (
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
	return lggr
}

func TestNewChipIngressBatchEmitter(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), newTestConfig())
		require.NoError(t, err)
		assert.NotNil(t, emitter)
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		emitter, err := beholder.NewChipIngressBatchEmitter(nil, newTestLogger(t), newTestConfig())
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}

func TestChipIngressBatchEmitter_Emit(t *testing.T) {
	t.Run("returns error when domain/entity missing", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), newTestConfig())
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		err = emitter.Emit(t.Context(), []byte("test"), "bad_key", "bad_value")
		assert.Error(t, err)
	})

	t.Run("enqueues and does not call PublishBatch immediately", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), newTestConfig())
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), cfg)
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), cfg)
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), cfg)
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), cfg)
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), newTestConfig())
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

		emitter, err := beholder.NewChipIngressBatchEmitter(clientMock, newTestLogger(t), cfg)
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
