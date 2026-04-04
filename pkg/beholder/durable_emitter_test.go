package beholder

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// withTestBeholderMeter swaps the global beholder client meter for t's lifetime (for metrics assertions).
func withTestBeholderMeter(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()
	prev := GetClient()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	c := NewNoopClient()
	c.MeterProvider = mp
	c.Meter = mp.Meter(defaultPackageName)
	SetClient(c)
	t.Cleanup(func() {
		SetClient(prev)
		_ = mp.Shutdown(context.Background())
	})
	return reader
}

// testChipClient is a minimal chipingress.Client for tests.
type testChipClient struct {
	chipingress.NoopClient

	mu            sync.Mutex
	publishErr    error
	publishCount  atomic.Int64
	publishedIDs  []string
}

func (c *testChipClient) Publish(_ context.Context, ev *chipingress.CloudEventPb, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.publishCount.Add(1)
	c.mu.Lock()
	if ev != nil {
		c.publishedIDs = append(c.publishedIDs, ev.Id)
	}
	err := c.publishErr
	c.mu.Unlock()
	return &chipingress.PublishResponse{}, err
}

func (c *testChipClient) PublishBatch(_ context.Context, _ *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	return &chipingress.PublishResponse{}, nil
}

func (c *testChipClient) setPublishErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishErr = err
}

func (c *testChipClient) getPublishedIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.publishedIDs))
	copy(out, c.publishedIDs)
	return out
}

func testEmitAttrs() []any {
	return []any{"source", "test-source", "type", "test-type"}
}

func newTestDurableEmitter(t *testing.T, store DurableEventStore, client chipingress.Client, cfgOverride *DurableEmitterConfig) *DurableEmitter {
	t.Helper()
	cfg := DefaultDurableEmitterConfig()
	if cfgOverride != nil {
		cfg = *cfgOverride
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)
	return em
}

func TestDurableEmitter_HooksImmediatePath(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(1), pubCalls.Load())
	assert.Equal(t, int32(1), delCalls.Load())
}

func TestDurableEmitter_HooksPublishFailureSkipsDeleteHook(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("down"))
	var pubCalls, delCalls atomic.Int32
	cfg := DefaultDurableEmitterConfig()
	cfg.Hooks = &DurableEmitterHooks{
		OnImmediatePublish: func(time.Duration, error) { pubCalls.Add(1) },
		OnImmediateDelete:  func(time.Duration, error) { delCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return pubCalls.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(0), delCalls.Load())
}

func TestDurableEmitter_EmitPersistsAndPublishes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err)

	// Immediate async publish should fire and delete the record.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_EmitReturnSuccessEvenWhenPublishFails(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))

	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err, "Emit must succeed once the DB insert succeeds")

	// Wait for the async publish attempt to complete.
	require.Eventually(t, func() bool {
		return client.publishCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	// Event must remain in the store for retransmit.
	assert.Equal(t, 1, store.Len())
}

func TestDurableEmitter_RetransmitLoopDeliversFailedEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("connection refused"))

	cfg := DefaultDurableEmitterConfig()
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("retry-me"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	client.setPublishErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should eventually deliver and delete the event")

	assert.GreaterOrEqual(t, client.publishCount.Load(), int64(2))
}

func TestDurableEmitter_RetransmitSerialDistinctCloudEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("immediate fail"))

	cfg := DefaultDurableEmitterConfig()
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("first"), testEmitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("second"), testEmitAttrs()...))

	client.setPublishErr(nil)

	require.Eventually(t, func() bool { return store.Len() == 0 }, 5*time.Second, 50*time.Millisecond)

	ids := client.getPublishedIDs()
	require.GreaterOrEqual(t, len(ids), 4, "two immediate fails then two retransmit publishes")
	a, b := ids[len(ids)-2], ids[len(ids)-1]
	assert.NotEmpty(t, a)
	assert.NotEmpty(t, b)
	assert.NotEqualf(t, a, b, "retransmit must publish two distinct CloudEvents, not one pointer reused for every row")
}

func TestDurableEmitter_ExpiryLoopDeletesOldEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	client.setPublishErr(errors.New("always fail"))

	cfg := DefaultDurableEmitterConfig()
	cfg.ExpiryInterval = 100 * time.Millisecond
	cfg.EventTTL = 50 * time.Millisecond
	cfg.RetransmitInterval = 10 * time.Minute // effectively disable retransmit

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	err := em.Emit(ctx, []byte("will-expire"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "expiry loop should purge the event")
}

func TestDurableEmitter_PersistSourceFilter_skipsStoreBestEffortPublish(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"only-this"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return client.publishCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 0, store.Len())
}

func TestDurableEmitter_PersistSourceFilter_persistsAllowedSource(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"test-source"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return client.publishCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_PersistSourceWildcardStarAllowsAll(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"*"}
	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_RetransmitDropsDisallowedSource(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}

	ev, err := chipingress.NewEvent("unknown-domain", "t", []byte("b"), nil)
	require.NoError(t, err)
	evPb, err := chipingress.EventToProto(ev)
	require.NoError(t, err)
	payload, err := proto.Marshal(evPb)
	require.NoError(t, err)

	_, err = store.Insert(context.Background(), payload)
	require.NoError(t, err)

	cfg := DefaultDurableEmitterConfig()
	cfg.PersistCloudEventSources = []string{"test-source"}
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.RetransmitAfter = 30 * time.Millisecond

	em := newTestDurableEmitter(t, store, client, &cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	require.Eventually(t, func() bool {
		return store.Len() == 0 && client.publishCount.Load() == 0
	}, 3*time.Second, 20*time.Millisecond, "disallowed row should be deleted without Publish")
}

func TestDurableEmitter_EmitRejectsInvalidAttributes(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	err := em.Emit(context.Background(), []byte("no-attrs"))
	require.Error(t, err)
	assert.Equal(t, 0, store.Len(), "nothing should be persisted when attributes are invalid")
}

func TestDurableEmitter_MultipleEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	client := &testChipClient{}
	em := newTestDurableEmitter(t, store, client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer em.Close()

	const n = 50
	for i := 0; i < n; i++ {
		err := em.Emit(ctx, []byte("event"), testEmitAttrs()...)
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		return client.publishCount.Load() == int64(n)
	}, 5*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 10*time.Millisecond, "all events should be delivered and deleted")
}

func TestNewDurableEmitter_ValidationErrors(t *testing.T) {
	log := logger.Test(t)
	cfg := DefaultDurableEmitterConfig()

	_, err := NewDurableEmitter(nil, &testChipClient{}, cfg, log)
	assert.ErrorContains(t, err, "store")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), nil, cfg, log)
	assert.ErrorContains(t, err, "client")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), &testChipClient{}, cfg, nil)
	assert.ErrorContains(t, err, "logger")
}

func TestDurableEmitter_MetricsRegistersEmitSuccess(t *testing.T) {
	reader := withTestBeholderMeter(t)

	store := NewMemDurableEventStore()
	client := &testChipClient{}
	cfg := DefaultDurableEmitterConfig()
	cfg.RetransmitInterval = time.Hour
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 25 * time.Millisecond}

	em, err := NewDurableEmitter(store, client, cfg, logger.Test(t))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Start(ctx)
	defer func() { _ = em.Close() }()

	require.NoError(t, em.Emit(ctx, []byte("m"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "beholder.durable_emitter.emit.success" {
				found = true
			}
		}
	}
	assert.True(t, found, "expected beholder.durable_emitter.emit.success in exported metrics")
}
