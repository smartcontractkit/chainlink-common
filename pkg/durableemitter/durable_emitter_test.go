package durableemitter

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

// newTestMeter returns a metric.Meter backed by a ManualReader so tests can
// inspect recorded metrics. No global state is mutated — the meter is fully
// owned by the test.
func newTestMeter(t *testing.T) (metric.Meter, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
	return mp.Meter("durableemitter"), reader
}

// testBatchEmitter is a minimal BatchEmitter for unit tests.
// QueueMessage invokes the callback asynchronously (like batch.Client), using
// publishErr as the result. callCount tracks how many events were enqueued.
type testBatchEmitter struct {
	mu         sync.Mutex
	publishErr error
	callCount  atomic.Int64

	stopOnce sync.Once
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func newTestBatchEmitter() *testBatchEmitter {
	return &testBatchEmitter{stopCh: make(chan struct{})}
}

func (b *testBatchEmitter) QueueMessage(event *chipingress.CloudEventPb, cb func(error)) error {
	select {
	case <-b.stopCh:
		return errors.New("batch emitter stopped")
	default:
	}
	b.mu.Lock()
	err := b.publishErr
	b.mu.Unlock()

	b.callCount.Add(1)
	if cb != nil {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			cb(err)
		}()
	}
	return nil
}

func (b *testBatchEmitter) Start(_ context.Context) {}

func (b *testBatchEmitter) Stop() {
	b.stopOnce.Do(func() {
		close(b.stopCh)
		b.wg.Wait()
	})
}

func (b *testBatchEmitter) setPublishErr(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.publishErr = err
}

func testEmitAttrs() []any {
	return []any{"source", "test-source", "type", "test-type"}
}

func newTestDurableEmitter(t *testing.T, store DurableEventStore, be BatchEmitter, cfgOverride *Config) *DurableEmitter {
	t.Helper()
	cfg := DefaultConfig()
	if cfgOverride != nil {
		cfg = *cfgOverride
	}
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	return em
}

// stallBatchStore wraps MemDurableEventStore so tests can block InsertBatch until stall is unlocked.
type stallBatchStore struct {
	*MemDurableEventStore
	stall *sync.Mutex
}

func (s *stallBatchStore) InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error) {
	s.stall.Lock()
	defer s.stall.Unlock()
	return s.MemDurableEventStore.InsertBatch(ctx, payloads)
}

// markRecordingStore wraps MemDurableEventStore and records the size of every
// MarkDeliveredBatch call so tests can assert how marks were coalesced.
type markRecordingStore struct {
	*MemDurableEventStore
	mu        sync.Mutex
	callSizes []int
}

func (s *markRecordingStore) MarkDeliveredBatch(ctx context.Context, ids []int64) (int64, error) {
	s.mu.Lock()
	s.callSizes = append(s.callSizes, len(ids))
	s.mu.Unlock()
	return s.MemDurableEventStore.MarkDeliveredBatch(ctx, ids)
}

func (s *markRecordingStore) sizes() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int(nil), s.callSizes...)
}

func TestDurableEmitter_CloseCoalescedInsertShutdown(t *testing.T) {
	stall := new(sync.Mutex)
	stall.Lock()
	store := &stallBatchStore{
		MemDurableEventStore: NewMemDurableEventStore(),
		stall:                stall,
	}
	be := newTestBatchEmitter()
	cfg := DefaultConfig()
	cfg.InsertBatchSize = 1
	cfg.InsertBatchWorkers = 1
	cfg.DisablePruning = true

	em := newTestDurableEmitter(t, store, be, &cfg)
	ctx := t.Context()
	require.NoError(t, em.Start(ctx))

	emitErr := make(chan error, 1)
	go func() { emitErr <- em.Emit(ctx, []byte("during-close"), testEmitAttrs()...) }()

	require.Eventually(t, func() bool {
		return em.insertInFlight.Load() == 1
	}, time.Second, 5*time.Millisecond, "Emit should be in-flight waiting on InsertBatch")

	closeErr := make(chan error, 1)
	go func() { closeErr <- em.Close() }()

	select {
	case err := <-closeErr:
		require.NoError(t, err)
		t.Fatal("Close returned before coalesced insert finished; shutdown wait is broken")
	case <-time.After(150 * time.Millisecond):
	}

	stall.Unlock()

	select {
	case err := <-closeErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Close timed out after releasing InsertBatch")
	}

	require.NoError(t, <-emitErr, "Emit should complete after insert path drains")

	err := em.Emit(ctx, []byte("after-close"), testEmitAttrs()...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestDurableEmitter_MarkCoalescingBatchesIds(t *testing.T) {
	store := &markRecordingStore{MemDurableEventStore: NewMemDurableEventStore()}
	be := newTestBatchEmitter()

	cfg := DefaultConfig()
	cfg.DisablePruning = true
	cfg.InsertBatchSize = 0 // disable insert coalescing so deliveries arrive in a burst
	cfg.MarkBatchSize = 100
	cfg.MarkBatchWorkers = 1
	cfg.MarkBatchFlushInterval = 200 * time.Millisecond

	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	const n = 25
	for i := 0; i < n; i++ {
		require.NoError(t, em.Emit(ctx, []byte("coalesce-me"), testEmitAttrs()...))
	}

	// Every delivered event must end up marked (MemStore deletes on mark).
	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 3*time.Second, 10*time.Millisecond, "all delivered events should be marked")

	sizes := store.sizes()
	total, maxBatch := 0, 0
	for _, s := range sizes {
		total += s
		if s > maxBatch {
			maxBatch = s
		}
	}
	assert.Equal(t, n, total, "every delivered id must be marked exactly once")
	assert.Less(t, len(sizes), n, "marks must be coalesced into fewer UPDATEs than events")
	assert.Greater(t, maxBatch, 1, "at least one UPDATE must mark multiple ids")
}

func TestDurableEmitter_MarkCoalescingFlushesPendingOnClose(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()

	cfg := DefaultConfig()
	cfg.DisablePruning = true
	cfg.MarkBatchSize = 100
	cfg.MarkBatchWorkers = 1
	// Long linger so marks stay buffered until Close drains them.
	cfg.MarkBatchFlushInterval = time.Hour

	em := newTestDurableEmitter(t, store, be, &cfg)
	require.NoError(t, em.Start(t.Context()))

	const n = 5
	for i := 0; i < n; i++ {
		require.NoError(t, em.Emit(t.Context(), []byte("buffer-me"), testEmitAttrs()...))
	}

	require.Eventually(t, func() bool {
		return be.callCount.Load() == int64(n)
	}, 2*time.Second, 10*time.Millisecond, "all events should be handed to the batch emitter")

	// Marks are buffered in the coalescer (1h linger), so nothing is flushed yet
	// regardless of worker scheduling — rows are only removed on a flush.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, n, store.Len(), "marks should still be buffered before Close")

	require.NoError(t, em.Close())

	assert.Equal(t, 0, store.Len(), "Close must flush buffered marks before returning")
}

func TestDurableEmitter_MarkCoalescingDisabledMarksInline(t *testing.T) {
	store := &markRecordingStore{MemDurableEventStore: NewMemDurableEventStore()}
	be := newTestBatchEmitter()

	cfg := DefaultConfig()
	cfg.DisablePruning = true
	cfg.MarkBatchSize = 0 // disable coalescing

	em := newTestDurableEmitter(t, store, be, &cfg)
	require.Nil(t, em.markCh, "mark coalescer channel must be nil when disabled")
	servicetest.Run(t, em)
	ctx := t.Context()

	const n = 3
	for i := 0; i < n; i++ {
		require.NoError(t, em.Emit(ctx, []byte("inline"), testEmitAttrs()...))
	}
	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond)

	sizes := store.sizes()
	assert.Len(t, sizes, n, "one inline UPDATE per delivered event when coalescing is disabled")
	for _, s := range sizes {
		assert.Equal(t, 1, s, "inline marks must be single-id UPDATEs")
	}
}

func TestDurableEmitter_HooksBatchPublishPath(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	var pubCalls, markCalls atomic.Int32
	cfg := DefaultConfig()
	cfg.Hooks = &Hooks{
		OnBatchPublish:       func(time.Duration, int, error) { pubCalls.Add(1) },
		OnBatchMarkDelivered: func(time.Duration, int) { markCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(1), pubCalls.Load())
	require.Eventually(t, func() bool { return markCalls.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_HooksPublishFailureSkipsMarkHook(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("down"))
	var pubCalls, markCalls atomic.Int32
	cfg := DefaultConfig()
	cfg.Hooks = &Hooks{
		OnBatchPublish:       func(time.Duration, int, error) { pubCalls.Add(1) },
		OnBatchMarkDelivered: func(time.Duration, int) { markCalls.Add(1) },
	}
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("hello"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return pubCalls.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, int32(0), markCalls.Load())
}

func TestDurableEmitter_NonHostProcessSkipsRetransmitAndExpiry(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable"))

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 40 * time.Millisecond
	cfg.RetransmitAfter = 15 * time.Millisecond
	cfg.ExpiryInterval = 40 * time.Millisecond
	cfg.EventTTL = 25 * time.Millisecond

	em, err := NewDurableEmitter(store, be, nil, false, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("plugin-row"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		return be.callCount.Load() >= 1 && store.Len() == 1
	}, 2*time.Second, 5*time.Millisecond, "initial QueueMessage should fail and leave the row")

	// Several host-only ticks would have cleared or retried by now.
	time.Sleep(250 * time.Millisecond)

	assert.Equal(t, 1, store.Len(), "non-host must not run retransmit or expiry loops")
	assert.Equal(t, int64(1), be.callCount.Load(), "non-host must not schedule extra QueueMessage via retransmit")
}

func TestDurableEmitter_NonHostProcessStillDeliversViaBatchWorkers(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()

	em, err := NewDurableEmitter(store, be, nil, false, DefaultConfig(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("loop-plugin"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		return store.Len() == 0 && be.callCount.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond, "batch emitter must deliver even when retransmitEnabled is false")
}

func TestDurableEmitter_EmitPersistsAndPublishes(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	em := newTestDurableEmitter(t, store, be, nil)
	servicetest.Run(t, em)
	ctx := t.Context()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return be.callCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestDurableEmitter_EmitReturnSuccessEvenWhenPublishFails(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))

	em := newTestDurableEmitter(t, store, be, nil)
	servicetest.Run(t, em)
	ctx := t.Context()

	err := em.Emit(ctx, []byte("hello"), testEmitAttrs()...)
	require.NoError(t, err, "Emit must succeed once the DB insert succeeds")

	require.Eventually(t, func() bool {
		return be.callCount.Load() == 1
	}, 2*time.Second, 10*time.Millisecond)

	// Event must remain in the store for retransmit.
	assert.Equal(t, 1, store.Len())
}

func TestDurableEmitter_RetransmitLoopDeliversFailedEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	err := em.Emit(ctx, []byte("retry-me"), testEmitAttrs()...)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return be.callCount.Load() >= 1 && store.Len() == 1
	}, 2*time.Second, 5*time.Millisecond, "failed delivery should leave the row")

	be.setPublishErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should eventually deliver and delete the event")

	assert.GreaterOrEqual(t, be.callCount.Load(), int64(2))
}

func TestDurableEmitter_RetransmitSerialDistinctCloudEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("immediate fail"))

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond

	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("first"), testEmitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("second"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		return be.callCount.Load() >= 2 && store.Len() == 2
	}, 2*time.Second, 5*time.Millisecond, "both failed deliveries should leave two rows")

	be.setPublishErr(nil)

	require.Eventually(t, func() bool { return store.Len() == 0 }, 5*time.Second, 50*time.Millisecond)
	assert.GreaterOrEqual(t, be.callCount.Load(), int64(4))
}

func TestDurableEmitter_ExpiryLoopDeletesOldEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("always fail"))

	cfg := DefaultConfig()
	cfg.ExpiryInterval = 100 * time.Millisecond
	cfg.EventTTL = 50 * time.Millisecond
	cfg.RetransmitInterval = 10 * time.Minute // effectively disable retransmit

	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	err := em.Emit(ctx, []byte("will-expire"), testEmitAttrs()...)
	require.NoError(t, err)
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "expiry loop should purge the event")
}

func TestDurableEmitter_RetransmitDeliversManuallyInsertedRow(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()

	ev, err := chipingress.NewEvent("unknown-domain", "t", []byte("b"), nil)
	require.NoError(t, err)
	evPb, err := chipingress.EventToProto(ev)
	require.NoError(t, err)
	payload, err := proto.Marshal(evPb)
	require.NoError(t, err)

	_, err = store.Insert(context.Background(), payload)
	require.NoError(t, err)

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 50 * time.Millisecond
	cfg.RetransmitAfter = 30 * time.Millisecond

	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	require.Eventually(t, func() bool {
		return store.Len() == 0 && be.callCount.Load() >= 1
	}, 3*time.Second, 20*time.Millisecond, "pending row should be delivered via batch emitter")
}

func TestDurableEmitter_EmitRejectsInvalidAttributes(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	em := newTestDurableEmitter(t, store, be, nil)

	err := em.Emit(context.Background(), []byte("no-attrs"))
	require.Error(t, err)
	assert.Equal(t, 0, store.Len(), "nothing should be persisted when attributes are invalid")
}

func TestDurableEmitter_MultipleEvents(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	em := newTestDurableEmitter(t, store, be, nil)
	servicetest.Run(t, em)
	ctx := t.Context()

	const n = 50
	for i := 0; i < n; i++ {
		err := em.Emit(ctx, []byte("event"), testEmitAttrs()...)
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		return be.callCount.Load() == int64(n)
	}, 5*time.Second, 10*time.Millisecond)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 10*time.Millisecond, "all events should be delivered and deleted")
}

func TestNewDurableEmitter_ValidationErrors(t *testing.T) {
	log := logger.Test(t)
	cfg := DefaultConfig()
	be := newTestBatchEmitter()

	_, err := NewDurableEmitter(nil, be, nil, true, cfg, log, nil)
	assert.ErrorContains(t, err, "store")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), nil, nil, true, cfg, log, nil)
	assert.ErrorContains(t, err, "batch emitter")

	_, err = NewDurableEmitter(NewMemDurableEventStore(), be, nil, true, cfg, nil, nil)
	assert.ErrorContains(t, err, "logger")

	cfgWithMetrics := cfg
	cfgWithMetrics.Metrics = &DurableEmitterMetricsConfig{}
	_, err = NewDurableEmitter(NewMemDurableEventStore(), be, nil, true, cfgWithMetrics, log, nil)
	assert.ErrorContains(t, err, "meter")
}

func TestDurableEmitter_HealthReport(t *testing.T) {
	em := newTestDurableEmitter(t, NewMemDurableEventStore(), newTestBatchEmitter(), nil)
	servicetest.Run(t, em)

	report := em.HealthReport()
	require.Contains(t, report, "DurableEmitter")
	require.NoError(t, report["DurableEmitter"], "service should be healthy after Start")
	require.NoError(t, em.Ready())
}

func TestDurableEmitter_MetricsRegistersEmitSuccess(t *testing.T) {
	meter, reader := newTestMeter(t)

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	cfg := DefaultConfig()
	cfg.RetransmitInterval = time.Hour
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 25 * time.Millisecond}

	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("m"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 0 }, 2*time.Second, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "durable_emitter.emit.success" {
				found = true
			}
		}
	}
	assert.True(t, found, "expected durable_emitter.emit.success in exported metrics")
}

func counterSumByPhase(t *testing.T, rm metricdata.ResourceMetrics, name, phase string) int64 {
	t.Helper()
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok, "expected Sum[int64] for %s", name)
			for _, dp := range sum.DataPoints {
				var gotPhase string
				for _, kv := range dp.Attributes.ToSlice() {
					if kv.Key == "phase" {
						gotPhase = kv.Value.AsString()
					}
				}
				if gotPhase == phase {
					total += dp.Value
				}
			}
		}
	}
	return total
}

func TestDurableEmitter_MetricsPublishBatchEventPhase(t *testing.T) {
	meter, reader := newTestMeter(t)

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("connection refused"))

	cfg := DefaultConfig()
	cfg.InsertBatchSize = 0
	cfg.MarkBatchSize = 0
	cfg.RetransmitInterval = 100 * time.Millisecond
	cfg.RetransmitAfter = 50 * time.Millisecond
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 25 * time.Millisecond}

	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("phase-test"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			return false
		}
		return counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.failure", "batch") >= 1
	}, 2*time.Second, 10*time.Millisecond, "initial batch attempt should record failure with phase=batch")

	be.setPublishErr(nil)

	require.Eventually(t, func() bool {
		var rm metricdata.ResourceMetrics
		if err := reader.Collect(ctx, &rm); err != nil {
			return false
		}
		return counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.success", "retransmit") >= 1
	}, 5*time.Second, 50*time.Millisecond, "retransmit should record success with phase=retransmit")

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit should deliver")

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))
	assert.GreaterOrEqual(t, counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.failure", "batch"), int64(1))
	assert.Equal(t, int64(0), counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.failure", "retransmit"))
	assert.GreaterOrEqual(t, counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.success", "retransmit"), int64(1))
	assert.Equal(t, int64(0), counterSumByPhase(t, rm, "durable_emitter.publish.batch.events.success", "batch"))
}

// mockChipServer implements ChipIngressServer with controllable behaviour.
type mockChipServer struct {
	pb.UnimplementedChipIngressServer

	mu            sync.Mutex
	publishErr    error
	batchErr      error
	received      []*cepb.CloudEvent
	batchReceived [][]*cepb.CloudEvent
	publishCount  atomic.Int64
	batchCount    atomic.Int64
	publishDelay  time.Duration
}

func (s *mockChipServer) Publish(_ context.Context, in *cepb.CloudEvent) (*pb.PublishResponse, error) {
	if s.publishDelay > 0 {
		time.Sleep(s.publishDelay)
	}
	s.publishCount.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.publishErr != nil {
		return nil, s.publishErr
	}
	s.received = append(s.received, in)
	return &pb.PublishResponse{}, nil
}

func (s *mockChipServer) PublishBatch(_ context.Context, in *pb.CloudEventBatch) (*pb.PublishResponse, error) {
	s.batchCount.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.batchErr != nil {
		return nil, s.batchErr
	}
	s.batchReceived = append(s.batchReceived, in.Events)
	s.received = append(s.received, in.Events...)
	return &pb.PublishResponse{}, nil
}

func (s *mockChipServer) Ping(context.Context, *pb.EmptyRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "pong"}, nil
}

func (s *mockChipServer) setPublishErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publishErr = err
}

func (s *mockChipServer) setBatchErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batchErr = err
}

func (s *mockChipServer) receivedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.received)
}

func (s *mockChipServer) batchCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.batchReceived)
}

// startMockServer starts a gRPC server on a random port and returns the
// server, address, and a cleanup function.
func startMockServer(t *testing.T, srv *mockChipServer) (*grpc.Server, string) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := grpc.NewServer()
	pb.RegisterChipIngressServer(gs, srv)

	go func() {
		if err := gs.Serve(lis); err != nil {
			// Ignore errors from server being stopped during cleanup.
		}
	}()

	t.Cleanup(func() { gs.GracefulStop() })
	return gs, lis.Addr().String()
}

func newChipClient(t *testing.T, addr string) chipingress.Client {
	t.Helper()
	c, err := chipingress.NewClient(addr, chipingress.WithInsecureConnection())
	require.NoError(t, err)
	return c
}

// newIntegrationBatchEmitter creates a batch.Client suitable for integration tests.
// Batch size 1 and a short interval ensure one event per RPC, matching the
// assertion style of the integration tests.
func newIntegrationBatchEmitter(t *testing.T, addr string) *batch.Client {
	t.Helper()
	chipClient := newChipClient(t, addr)
	bc, err := batch.NewBatchClient(chipClient,
		batch.WithBatchSize(1),
		batch.WithBatchInterval(10*time.Millisecond),
		batch.WithMaxPublishTimeout(2*time.Second),
		batch.WithShutdownTimeout(5*time.Second),
	)
	require.NoError(t, err)
	return bc
}

func emitAttrs() []any {
	return []any{"source", "test-domain", "type", "test-entity"}
}

func fastCfg() Config {
	return Config{
		RetransmitInterval:  100 * time.Millisecond,
		RetransmitAfter:     50 * time.Millisecond,
		RetransmitBatchSize: 50,
		ExpiryInterval:      200 * time.Millisecond,
		EventTTL:            500 * time.Millisecond,
		PublishTimeout:      2 * time.Second,
	}
}

// testFallbackClient is a minimal chipingress.Client for fallback testing.
type testFallbackClient struct {
	chipingress.NoopClient

	mu           sync.Mutex
	publishErr   error
	publishCount atomic.Int64
	closed       atomic.Bool
}

func (c *testFallbackClient) Publish(_ context.Context, _ *chipingress.CloudEventPb, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	c.mu.Lock()
	err := c.publishErr
	c.mu.Unlock()
	c.publishCount.Add(1)
	return &chipingress.PublishResponse{}, err
}

func (c *testFallbackClient) Close() error {
	c.closed.Store(true)
	return nil
}

func (c *testFallbackClient) setPublishErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishErr = err
}

// TestDurableEmitter_FallbackDeliversOnBatchFailure verifies that when the
// batch emitter fails, the fallback client delivers the event directly and
// marks it delivered (removing it from the DB).
func TestDurableEmitter_FallbackDeliversOnBatchFailure(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("batch down"))
	fallback := &testFallbackClient{}

	em, err := NewDurableEmitter(store, be, fallback, true, DefaultConfig(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("needs-fallback"), testEmitAttrs()...))

	// Batch emitter fires callback with error → fallback client should deliver.
	require.Eventually(t, func() bool {
		return fallback.publishCount.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond, "fallback client should receive one Publish call")

	// Event should be marked delivered and removed from the store.
	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 2*time.Second, 10*time.Millisecond, "event should be marked delivered after fallback")
}

// TestDurableEmitter_FallbackFailureEventRemainsForRetransmit verifies that
// when both the batch emitter and the fallback client fail, the event stays
// in the DB for the retransmit loop to pick up.
func TestDurableEmitter_FallbackFailureEventRemainsForRetransmit(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("batch down"))
	fallback := &testFallbackClient{}
	fallback.setPublishErr(errors.New("fallback down too"))

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 10 * time.Minute // disable retransmit for this test
	em, err := NewDurableEmitter(store, be, fallback, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("both-fail"), testEmitAttrs()...))

	require.Eventually(t, func() bool {
		return fallback.publishCount.Load() >= 1
	}, 2*time.Second, 10*time.Millisecond, "fallback should have been attempted")

	// Event must still be in the store for the retransmit loop.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, store.Len(), "event must remain in DB when fallback also fails")
}

// TestDurableEmitter_FallbackClientClosedOnStop verifies that the fallback
// client's Close() method is called when DurableEmitter shuts down.
func TestDurableEmitter_FallbackClientClosedOnStop(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	fallback := &testFallbackClient{}

	em, err := NewDurableEmitter(store, be, fallback, true, DefaultConfig(), logger.Test(t), nil)
	require.NoError(t, err)
	require.NoError(t, em.Start(t.Context()))

	require.NoError(t, em.Close())
	assert.True(t, fallback.closed.Load(), "fallback client should be closed after Stop")
}

// ---------- Test cases ----------

func TestIntegration_HappyPath(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, be, nil, true, fastCfg(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("billing-record-1"), emitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("billing-record-2"), emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.receivedCount() == 2
	}, 3*time.Second, 10*time.Millisecond, "server should receive both events")

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 3*time.Second, 10*time.Millisecond, "store should be empty after delivery")
}

func TestIntegration_ServerUnavailable_RetransmitRecovers(t *testing.T) {
	srv := &mockChipServer{}
	srv.setBatchErr(status.Error(codes.Unavailable, "chip down"))
	_, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, be, nil, true, fastCfg(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("will-retry"), emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.batchCount.Load() >= 1 && store.Len() == 1
	}, 2*time.Second, 10*time.Millisecond, "failed PublishBatch should leave the row pending")

	srv.setBatchErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "retransmit loop should deliver after recovery")

	assert.GreaterOrEqual(t, srv.batchCount.Load(), int64(2),
		"one failed PublishBatch then at least one successful PublishBatch (retransmit)")
	assert.Equal(t, int64(0), srv.publishCount.Load(), "unary Publish should not be used for durable path")
}

func TestIntegration_ServerDown_EventsSurvive(t *testing.T) {
	srv := &mockChipServer{}
	gs, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.PublishTimeout = 500 * time.Millisecond
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, em.Start(ctx))

	// Stop the gRPC server entirely.
	gs.Stop()
	time.Sleep(100 * time.Millisecond)

	// Emit while server is down — Emit() itself must succeed (DB insert works).
	require.NoError(t, em.Emit(ctx, []byte("server-is-down"), emitAttrs()...))
	assert.Equal(t, 1, store.Len(), "event should be persisted even with server down")

	// Stop the emitter to simulate a "node shutdown".
	require.NoError(t, em.Close())

	// Bring up a new server on the same address.
	srv2 := &mockChipServer{}
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	gs2 := grpc.NewServer()
	pb.RegisterChipIngressServer(gs2, srv2)
	go func() { _ = gs2.Serve(lis) }()
	t.Cleanup(func() { gs2.GracefulStop() })

	// Create a new batch emitter re-using the same store (simulating node restart with Postgres).
	be2 := newIntegrationBatchEmitter(t, addr)
	em2, err := NewDurableEmitter(store, be2, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em2)

	require.Eventually(t, func() bool {
		return srv2.receivedCount() == 1
	}, 5*time.Second, 50*time.Millisecond, "new emitter should retransmit the surviving event")

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond, "store should be empty after retransmit")
}

func TestIntegration_HighThroughput(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.RetransmitBatchSize = 200
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	const n = 500
	for i := 0; i < n; i++ {
		require.NoError(t, em.Emit(ctx, []byte("event"), emitAttrs()...))
	}

	require.Eventually(t, func() bool {
		return srv.receivedCount() >= n
	}, 10*time.Second, 50*time.Millisecond, "all %d events should be received", n)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 10*time.Second, 50*time.Millisecond, "store should drain completely")
}

func TestIntegration_EventExpiry(t *testing.T) {
	// Server always rejects — events can never be delivered.
	srv := &mockChipServer{}
	srv.setBatchErr(status.Error(codes.Internal, "permanent failure"))
	_, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	cfg := fastCfg()
	cfg.EventTTL = 100 * time.Millisecond
	cfg.ExpiryInterval = 100 * time.Millisecond
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("will-expire"), emitAttrs()...))
	assert.Equal(t, 1, store.Len())

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond,
		"expiry loop should purge undeliverable events after TTL")
}

func TestIntegration_RetransmitEnqueuesBatchWorkers(t *testing.T) {
	srv := &mockChipServer{}
	srv.setBatchErr(status.Error(codes.Unavailable, "reject batch"))
	_, addr := startMockServer(t, srv)
	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, be, nil, true, fastCfg(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	for i := 0; i < 5; i++ {
		require.NoError(t, em.Emit(ctx, []byte("retry-me"), emitAttrs()...))
	}

	require.Eventually(t, func() bool {
		return srv.batchCount.Load() >= 5 && store.Len() == 5
	}, 3*time.Second, 10*time.Millisecond, "all five PublishBatch RPCs should have failed and left five rows")

	srv.setBatchErr(nil)

	require.Eventually(t, func() bool {
		return store.Len() == 0
	}, 5*time.Second, 50*time.Millisecond,
		"retransmit should deliver via batch emitter")

	assert.GreaterOrEqual(t, srv.batchCount.Load(), int64(10),
		"five failed PublishBatch plus five successful PublishBatch (retransmit)")
	assert.Equal(t, int64(0), srv.publishCount.Load(), "no unary Publish on durable path")
}

// TestIntegration_GRPCConnection verifies the emitter works over a real gRPC
// connection with proper proto serialization round-trip.
func TestIntegration_GRPCConnection(t *testing.T) {
	srv := &mockChipServer{}
	_, addr := startMockServer(t, srv)

	// Use a raw gRPC dial to prove we're going over the wire.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	// Ping to verify connectivity.
	grpcClient := pb.NewChipIngressClient(conn)
	pong, err := grpcClient.Ping(context.Background(), &pb.EmptyRequest{})
	require.NoError(t, err)
	assert.Equal(t, "pong", pong.Message)

	be := newIntegrationBatchEmitter(t, addr)
	store := NewMemDurableEventStore()

	em, err := NewDurableEmitter(store, be, nil, true, fastCfg(), logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	payload := []byte("proto-round-trip-test")
	require.NoError(t, em.Emit(ctx, payload, emitAttrs()...))

	require.Eventually(t, func() bool {
		return srv.receivedCount() == 1
	}, 3*time.Second, 10*time.Millisecond)

	// Verify the CloudEvent arrived with correct source/type.
	srv.mu.Lock()
	received := srv.received[0]
	srv.mu.Unlock()

	assert.Equal(t, "test-domain", received.Source)
	assert.Equal(t, "test-entity", received.Type)
}

func TestDurableEmitter_IdempotencyKeyDefaultsToEventHash(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("hold"))
	cfg := DefaultConfig()
	cfg.RetransmitInterval = 10 * time.Minute // no retransmit during test
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	body := []byte("deterministic-body")
	require.NoError(t, em.Emit(ctx, body, testEmitAttrs()...))
	require.Eventually(t, func() bool { return be.callCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)

	// Event stays in store (publish failed). Unmarshal stored proto and check attribute.
	store.mu.Lock()
	var stored []byte
	for _, e := range store.events {
		stored = append([]byte(nil), e.Payload...)
		break
	}
	store.mu.Unlock()
	require.NotNil(t, stored, "event should still be in store")

	var eventPb cepb.CloudEvent
	require.NoError(t, proto.Unmarshal(stored, &eventPb))

	attr, ok := eventPb.Attributes[chipingress.IdempotencyKeyAttr]
	require.True(t, ok, "idempotencykey attribute must be set")

	// The idempotency key is computed as a SHA256 hash of:
	// source domain (with 4-byte big-endian length) + entity type (with 4-byte big-endian length) + body (with 4-byte big-endian length)
	h := sha256.New()
	for _, s := range []string{"test-source", "test-type"} {
		h.Write(binary.BigEndian.AppendUint32(nil, uint32(len(s))))
		h.Write([]byte(s))
	}
	h.Write(binary.BigEndian.AppendUint32(nil, uint32(len(body))))
	h.Write(body)
	expectedKey := hex.EncodeToString(h.Sum(nil))
	require.Equal(t, expectedKey, attr.GetCeString())
}

func TestDurableEmitter_IdempotencyKeyCallerSupplied(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("hold"))
	cfg := DefaultConfig()
	cfg.RetransmitInterval = 10 * time.Minute
	em, err := NewDurableEmitter(store, be, nil, true, cfg, logger.Test(t), nil)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	callerKey := "my-idempotency-key-abc123"
	body := []byte("some-body")
	attrs := append(testEmitAttrs(), chipingress.IdempotencyKeyAttr, callerKey)
	require.NoError(t, em.Emit(ctx, body, attrs...))

	require.Eventually(t, func() bool { return be.callCount.Load() == 1 }, 2*time.Second, 10*time.Millisecond)

	store.mu.Lock()
	var stored []byte
	for _, e := range store.events {
		stored = append([]byte(nil), e.Payload...)
		break
	}
	store.mu.Unlock()
	require.NotNil(t, stored)

	var eventPb cepb.CloudEvent
	require.NoError(t, proto.Unmarshal(stored, &eventPb))

	attr, ok := eventPb.Attributes[chipingress.IdempotencyKeyAttr]
	require.True(t, ok, "idempotencykey attribute must be present")
	require.Equal(t, callerKey, attr.GetCeString(), "caller-supplied key must be preserved")
}

// MemDurableEventStore is an in-memory DurableEventStore for unit tests.
type MemDurableEventStore struct {
	mu     sync.Mutex
	events map[int64]*DurableEvent
	nextID atomic.Int64
}

var (
	_ DurableEventStore    = (*MemDurableEventStore)(nil)
	_ DurableQueueObserver = (*MemDurableEventStore)(nil)
	_ BatchInserter        = (*MemDurableEventStore)(nil)
)

func NewMemDurableEventStore() *MemDurableEventStore {
	return &MemDurableEventStore{
		events: make(map[int64]*DurableEvent),
	}
}

func (m *MemDurableEventStore) Insert(_ context.Context, payload []byte) (int64, error) {
	id := m.nextID.Add(1)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[id] = &DurableEvent{
		ID:        id,
		Payload:   append([]byte(nil), payload...), // defensive copy
		CreatedAt: time.Now(),
	}
	return id, nil
}

func (m *MemDurableEventStore) InsertBatch(_ context.Context, payloads [][]byte) ([]int64, error) {
	now := time.Now()
	ids := make([]int64, len(payloads))
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range payloads {
		id := m.nextID.Add(1)
		m.events[id] = &DurableEvent{
			ID:        id,
			Payload:   append([]byte(nil), p...),
			CreatedAt: now,
		}
		ids[i] = id
	}
	return ids, nil
}

func (m *MemDurableEventStore) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.events, id)
	return nil
}

func (m *MemDurableEventStore) MarkDelivered(ctx context.Context, id int64) error {
	return m.Delete(ctx, id)
}

func (m *MemDurableEventStore) MarkDeliveredBatch(_ context.Context, ids []int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, id := range ids {
		if _, ok := m.events[id]; ok {
			delete(m.events, id)
			n++
		}
	}
	return n, nil
}

func (m *MemDurableEventStore) PurgeDelivered(_ context.Context, _ int) (int64, error) {
	return 0, nil
}

func (m *MemDurableEventStore) ListPending(_ context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []DurableEvent
	for _, e := range m.events {
		if e.CreatedAt.Before(createdBefore) {
			result = append(result, *e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MemDurableEventStore) DeleteExpired(_ context.Context, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-ttl)
	var deleted int64
	for id, e := range m.events {
		if e.CreatedAt.Before(cutoff) {
			delete(m.events, id)
			deleted++
		}
	}
	return deleted, nil
}

// Len returns the number of events in the store (test helper).
func (m *MemDurableEventStore) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

// ObserveDurableQueue implements DurableQueueObserver.
func (m *MemDurableEventStore) ObserveDurableQueue(_ context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var st DurableQueueStats
	st.TotalRows = int64(len(m.events))
	if len(m.events) == 0 {
		return st, nil
	}
	var oldest time.Time
	first := true
	for _, e := range m.events {
		st.Depth++
		st.PayloadBytes += int64(len(e.Payload))
		if first || e.CreatedAt.Before(oldest) {
			oldest = e.CreatedAt
			first = false
		}
		age := now.Sub(e.CreatedAt)
		if eventTTL > 0 && nearExpiryLead > 0 && nearExpiryLead < eventTTL {
			threshold := eventTTL - nearExpiryLead
			if age >= threshold && age < eventTTL {
				st.NearTTLCount++
			}
		}
	}
	st.OldestPendingAge = now.Sub(oldest)
	return st, nil
}
