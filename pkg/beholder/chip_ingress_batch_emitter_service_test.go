package beholder_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func newTestConfig() beholder.Config {
	return beholder.Config{
		ChipIngressBufferSize:         10,
		ChipIngressMaxBatchSize:       5,
		ChipIngressMaxConcurrentSends: 3,
		ChipIngressSendInterval:       50 * time.Millisecond,
		ChipIngressSendTimeout:        5 * time.Second,
		ChipIngressDrainTimeout:       5 * time.Second,
	}
}

func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()
	lggr, err := logger.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = lggr.Sync() })
	return lggr
}

func TestNewChipIngressBatchEmitterService(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		assert.NotNil(t, emitter)
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		emitter, err := beholder.NewChipIngressBatchEmitterService(nil, newTestConfig(), newTestLogger(t))
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}

func TestChipIngressBatchEmitterService_Emit(t *testing.T) {
	t.Run("returns error when domain/entity missing", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, newTestConfig(), newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer emitter.Close() //nolint:errcheck

		err = emitter.Emit(t.Context(), []byte("test"), "bad_key", "bad_value")
		assert.Error(t, err)
	})

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

		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
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

func TestChipIngressBatchEmitterService_CloudEventFormat(t *testing.T) {
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

	emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
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
}

func TestChipIngressBatchEmitterService_PublishBatchError(t *testing.T) {
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

	emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
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
}

func TestChipIngressBatchEmitterService_ContextCancellation(t *testing.T) {
	clientMock := mocks.NewClient(t)
	clientMock.
		On("PublishBatch", mock.Anything, mock.Anything).
		Return(nil, nil).
		Maybe()

	cfg := newTestConfig()
	cfg.ChipIngressBufferSize = 1
	cfg.ChipIngressSendInterval = 10 * time.Second

	emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
	require.NoError(t, err)
	require.NoError(t, emitter.Start(t.Context()))
	defer emitter.Close() //nolint:errcheck

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = emitter.Emit(ctx, []byte("should-fail"),
		beholder.AttrKeyDomain, "platform",
		beholder.AttrKeyEntity, "TestEvent",
	)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestChipIngressBatchEmitterService_DefaultConfig(t *testing.T) {
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

	emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, beholder.Config{}, newTestLogger(t))
	require.NoError(t, err)
	require.NoError(t, emitter.Start(t.Context()))

	err = emitter.Emit(t.Context(), []byte("body"),
		beholder.AttrKeyDomain, "platform",
		beholder.AttrKeyEntity, "TestEvent",
	)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedBatch != nil
	}, 3*time.Second, 50*time.Millisecond)

	require.NoError(t, emitter.Close())

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, receivedBatch.Events, 1)
}

func TestChipIngressBatchEmitterService_EmitAfterClose(t *testing.T) {
	clientMock := mocks.NewClient(t)
	clientMock.
		On("PublishBatch", mock.Anything, mock.Anything).
		Return(nil, nil).
		Maybe()

	emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, newTestConfig(), newTestLogger(t))
	require.NoError(t, err)
	require.NoError(t, emitter.Start(t.Context()))
	require.NoError(t, emitter.Close())

	err = emitter.Emit(t.Context(), []byte("body"),
		beholder.AttrKeyDomain, "platform",
		beholder.AttrKeyEntity, "TestEvent",
	)
	assert.Error(t, err)
}

func TestChipIngressBatchEmitterService_EmitWithCallback(t *testing.T) {
	t.Run("callback receives nil on success", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		done := make(chan error, 1)
		err = emitter.EmitWithCallback(t.Context(), []byte("body"), func(sendErr error) {
			done <- sendErr
		},
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		require.NoError(t, err)

		select {
		case sendErr := <-done:
			assert.NoError(t, sendErr)
		case <-time.After(3 * time.Second):
			t.Fatal("callback was not invoked within timeout")
		}

		require.NoError(t, emitter.Close())
	})

	t.Run("callback receives error on PublishBatch failure", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		done := make(chan error, 1)
		err = emitter.EmitWithCallback(t.Context(), []byte("body"), func(sendErr error) {
			done <- sendErr
		},
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		require.NoError(t, err)

		select {
		case sendErr := <-done:
			assert.Error(t, sendErr)
		case <-time.After(3 * time.Second):
			t.Fatal("callback was not invoked within timeout")
		}

		require.NoError(t, emitter.Close())
	})

	t.Run("callback receives error when buffer is full", func(t *testing.T) {
		clientMock := mocks.NewClient(t)

		sendBlocked := make(chan struct{})
		firstCallSignal := make(chan struct{}, 1)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Run(func(_ mock.Arguments) {
				select {
				case firstCallSignal <- struct{}{}:
				default:
				}
				<-sendBlocked
			}).
			Return(nil, nil).
			Maybe()

		cfg := newTestConfig()
		cfg.ChipIngressBufferSize = 2
		cfg.ChipIngressMaxBatchSize = 1
		cfg.ChipIngressMaxConcurrentSends = 1
		cfg.ChipIngressSendInterval = 50 * time.Millisecond
		cfg.ChipIngressDrainTimeout = 200 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))
		defer close(sendBlocked)
		defer emitter.Close() //nolint:errcheck

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		require.NoError(t, err)

		<-firstCallSignal
		time.Sleep(100 * time.Millisecond)

		for i := 0; i < 10; i++ {
			_ = emitter.Emit(t.Context(), []byte("filler"),
				beholder.AttrKeyDomain, "platform",
				beholder.AttrKeyEntity, "TestEvent",
			)
		}

		dropped := make(chan error, 1)
		err = emitter.EmitWithCallback(t.Context(), []byte("overflow"), func(sendErr error) {
			dropped <- sendErr
		},
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		assert.NoError(t, err)

		select {
		case dropErr := <-dropped:
			assert.Error(t, dropErr)
		case <-time.After(time.Second):
			t.Fatal("callback was not invoked for dropped event")
		}
	})

	t.Run("nil callback behaves like Emit", func(t *testing.T) {
		clientMock := mocks.NewClient(t)
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Maybe()

		cfg := newTestConfig()
		cfg.ChipIngressSendInterval = 50 * time.Millisecond

		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		err = emitter.EmitWithCallback(t.Context(), []byte("body"), nil,
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "TestEvent",
		)
		assert.NoError(t, err)

		require.NoError(t, emitter.Close())
	})
}

func TestChipIngressBatchEmitterService_Metrics(t *testing.T) {
	t.Run("records events_sent on successful publish", func(t *testing.T) {
		reader, restore := useEmitterTestMeterProvider(t)
		defer restore()

		clientMock := mocks.NewClient(t)
		done := make(chan struct{})
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, nil).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		cfg := newTestConfig()
		cfg.ChipIngressMaxBatchSize = 1
		cfg.ChipIngressSendInterval = time.Second
		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "MetricEvent",
		)
		require.NoError(t, err)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for publish")
		}
		require.NoError(t, emitter.Close())

		rm := collectEmitterMetrics(t, reader)
		metricdatatest.AssertEqual(t, metricdata.ScopeMetrics{
			Scope: instrumentation.Scope{Name: "beholder/chip_ingress_batch_emitter"},
			Metrics: []metricdata.Metrics{
				{
					Name:        "chip_ingress.events_sent",
					Description: "Total events successfully sent via PublishBatch",
					Unit:        "{event}",
					Data: metricdata.Sum[int64]{
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("domain", "platform"),
									attribute.String("entity", "MetricEvent"),
								),
								Value: 1,
							},
						},
					},
				},
			},
		}, mustEmitterScopeMetrics(t, rm, "beholder/chip_ingress_batch_emitter"), metricdatatest.IgnoreTimestamp())

		metric := mustEmitterMetric(t, rm, "chip_ingress.events_sent")
		sum, ok := metric.Data.(metricdata.Sum[int64])
		require.True(t, ok)
		dp := mustEmitterInt64SumPoint(t, sum, "domain", "platform", "entity", "MetricEvent")
		assert.GreaterOrEqual(t, dp.Value, int64(1))
	})

	t.Run("records events_dropped on publish error", func(t *testing.T) {
		reader, restore := useEmitterTestMeterProvider(t)
		defer restore()

		clientMock := mocks.NewClient(t)
		done := make(chan struct{})
		clientMock.
			On("PublishBatch", mock.Anything, mock.Anything).
			Return(nil, assert.AnError).
			Run(func(_ mock.Arguments) { close(done) }).
			Once()

		cfg := newTestConfig()
		cfg.ChipIngressMaxBatchSize = 1
		cfg.ChipIngressSendInterval = time.Second
		emitter, err := beholder.NewChipIngressBatchEmitterService(clientMock, cfg, newTestLogger(t))
		require.NoError(t, err)
		require.NoError(t, emitter.Start(t.Context()))

		err = emitter.Emit(t.Context(), []byte("body"),
			beholder.AttrKeyDomain, "platform",
			beholder.AttrKeyEntity, "MetricDropEvent",
		)
		require.NoError(t, err)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for publish")
		}
		require.NoError(t, emitter.Close())

		rm := collectEmitterMetrics(t, reader)
		metricdatatest.AssertEqual(t, metricdata.ScopeMetrics{
			Scope: instrumentation.Scope{Name: "beholder/chip_ingress_batch_emitter"},
			Metrics: []metricdata.Metrics{
				{
					Name:        "chip_ingress.events_dropped",
					Description: "Total events dropped (buffer full or send failure)",
					Unit:        "{event}",
					Data: metricdata.Sum[int64]{
						Temporality: metricdata.CumulativeTemporality,
						IsMonotonic: true,
						DataPoints: []metricdata.DataPoint[int64]{
							{
								Attributes: attribute.NewSet(
									attribute.String("domain", "platform"),
									attribute.String("entity", "MetricDropEvent"),
								),
								Value: 1,
							},
						},
					},
				},
			},
		}, mustEmitterScopeMetrics(t, rm, "beholder/chip_ingress_batch_emitter"), metricdatatest.IgnoreTimestamp())

		metric := mustEmitterMetric(t, rm, "chip_ingress.events_dropped")
		sum, ok := metric.Data.(metricdata.Sum[int64])
		require.True(t, ok)
		dp := mustEmitterInt64SumPoint(t, sum, "domain", "platform", "entity", "MetricDropEvent")
		assert.GreaterOrEqual(t, dp.Value, int64(1))
	})
}

func BenchmarkChipIngressBatchEmitterService_Emit(b *testing.B) {
	cfg := beholder.Config{
		ChipIngressBufferSize:         uint(b.N + 10),
		ChipIngressMaxBatchSize:       uint(b.N + 1),
		ChipIngressMaxConcurrentSends: 1,
		ChipIngressSendInterval:       time.Hour,
		ChipIngressSendTimeout:        5 * time.Second,
		ChipIngressDrainTimeout:       5 * time.Second,
	}
	emitter, err := beholder.NewChipIngressBatchEmitterService(&chipingress.NoopClient{}, cfg, logger.Nop())
	if err != nil {
		b.Fatal(err)
	}
	if err := emitter.Start(context.Background()); err != nil {
		b.Fatal(err)
	}
	defer func() { _ = emitter.Close() }()

	payload := []byte("benchmark-payload")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := emitter.Emit(context.Background(), payload,
			beholder.AttrKeyDomain, "bench",
			beholder.AttrKeyEntity, "BenchmarkEvent",
		); err != nil {
			b.Fatal(err)
		}
	}
}

func useEmitterTestMeterProvider(t *testing.T) (*sdkmetric.ManualReader, func()) {
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

func collectEmitterMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))
	return rm
}

func mustEmitterMetric(t *testing.T, rm metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, metric := range sm.Metrics {
			if metric.Name == name {
				return metric
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func mustEmitterScopeMetrics(t *testing.T, rm metricdata.ResourceMetrics, name string) metricdata.ScopeMetrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		if sm.Scope.Name == name {
			return sm
		}
	}
	t.Fatalf("scope metrics %q not found", name)
	return metricdata.ScopeMetrics{}
}

func mustEmitterInt64SumPoint(t *testing.T, sum metricdata.Sum[int64], k1, v1, k2, v2 string) metricdata.DataPoint[int64] {
	t.Helper()
	for _, dp := range sum.DataPoints {
		if hasEmitterStringAttr(dp.Attributes, k1, v1) && hasEmitterStringAttr(dp.Attributes, k2, v2) {
			return dp
		}
	}
	t.Fatalf("sum datapoint not found for attrs %s=%s,%s=%s", k1, v1, k2, v2)
	return metricdata.DataPoint[int64]{}
}

func hasEmitterStringAttr(set attribute.Set, key, want string) bool {
	for _, kv := range set.ToSlice() {
		if string(kv.Key) == key {
			return kv.Value.AsString() == want
		}
	}
	return false
}
