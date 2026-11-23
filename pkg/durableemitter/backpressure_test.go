package durableemitter

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

// estimatingStore adds a settable BacklogEstimator to MemDurableEventStore.
// EstimateBacklog errors while fail is set.
type estimatingStore struct {
	*MemDurableEventStore
	estimate atomic.Int64
	fail     atomic.Bool
}

var _ BacklogEstimator = (*estimatingStore)(nil)

func (s *estimatingStore) EstimateBacklog(context.Context) (int64, error) {
	if s.fail.Load() {
		return 0, errors.New("estimate timeout")
	}
	return s.estimate.Load(), nil
}

// faultyStore fails BatchDelete while failDeletes is set and counts store reads.
type faultyStore struct {
	*MemDurableEventStore
	failDeletes  atomic.Bool
	listCalls    atomic.Int64
	observeCalls atomic.Int64
}

func (s *faultyStore) BatchDelete(ctx context.Context, ids []int64) (int64, error) {
	if s.failDeletes.Load() {
		return 0, errors.New("delete timeout")
	}
	return s.MemDurableEventStore.BatchDelete(ctx, ids)
}

func (s *faultyStore) ListPending(ctx context.Context, createdBefore, afterCreatedAt time.Time, afterID int64, limit int) ([]DurableEvent, error) {
	s.listCalls.Add(1)
	return s.MemDurableEventStore.ListPending(ctx, createdBefore, afterCreatedAt, afterID, limit)
}

func (s *faultyStore) ObserveDurableQueue(ctx context.Context, eventTTL time.Duration) (DurableQueueStats, error) {
	s.observeCalls.Add(1)
	return s.MemDurableEventStore.ObserveDurableQueue(ctx, eventTTL)
}

func TestDurableEmitter_BackpressureRejectsWhileBacklogged(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(100)

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 100
	cfg.BacklogPollInterval = 10 * time.Millisecond
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	// The estimate is seeded at start, so the first emit is already rejected.
	err := em.Emit(ctx, []byte("rejected"), testEmitAttrs()...)
	require.ErrorIs(t, err, ErrBacklogged)
	assert.Equal(t, 0, store.Len(), "rejected events must not be persisted")

	store.estimate.Store(99)
	require.Eventually(t, func() bool {
		return em.Emit(ctx, []byte("accepted"), testEmitAttrs()...) == nil
	}, 2*time.Second, 10*time.Millisecond, "emits resume once the polled estimate drops below MaxQueueDepth")
}

func TestDurableEmitter_BackpressureOffByDefault(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(DefaultMaxQueueDepth)

	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), nil)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("x"), testEmitAttrs()...))
	assert.Nil(t, em.backlogEstimator, "the backlog is not polled unless back-pressure is enabled")
}

func TestDurableEmitter_BackpressureKeepsLastEstimateWhenPollFails(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(100)

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 100
	cfg.BacklogPollInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	store.fail.Store(true)
	store.estimate.Store(0)
	em.refreshBacklog(ctx)
	require.ErrorIs(t, em.Emit(ctx, []byte("rejected"), testEmitAttrs()...), ErrBacklogged)

	store.fail.Store(false)
	em.refreshBacklog(ctx)
	require.NoError(t, em.Emit(ctx, []byte("accepted"), testEmitAttrs()...))
}

func TestDurableEmitter_BackpressureDisabledWithoutEstimator(t *testing.T) {
	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable"))

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 1
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	for range 3 {
		require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	}
	assert.Equal(t, 3, store.Len())
}

func TestDurableEmitter_RetransmitPausesWhileDeletesFail(t *testing.T) {
	store := &faultyStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.failDeletes.Store(true)

	cfg := DefaultConfig()
	cfg.RetransmitInterval = 10 * time.Millisecond
	cfg.RetransmitAfter = time.Millisecond
	cfg.DeleteFailureRetransmitBackoff = 2 * time.Second
	cfg.DeleteBatchFlushInterval = time.Millisecond
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("stuck"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return em.deleteFailedAt.Load() != 0 }, 2*time.Second, 5*time.Millisecond)

	listed := store.listCalls.Load()
	time.Sleep(150 * time.Millisecond) // ~15 ticks, all inside the backoff window
	assert.Equal(t, listed, store.listCalls.Load(), "retransmit must not list pending rows while deletes fail")

	// A later successful delete must not cut the pause short.
	store.failDeletes.Store(false)
	require.NoError(t, em.Emit(ctx, []byte("fresh"), testEmitAttrs()...))
	require.Eventually(t, func() bool { return store.Len() == 1 }, 2*time.Second, 5*time.Millisecond)
	assert.Equal(t, listed, store.listCalls.Load(), "retransmit stays paused for the full window after the last failure")

	require.Eventually(t, func() bool { return store.Len() == 0 }, 10*time.Second, 10*time.Millisecond,
		"retransmit resumes after the backoff and the stuck row is deleted")
}

func TestDurableEmitter_PluginMetricsLabelledAndSkipQueueScan(t *testing.T) {
	meter, reader := newTestMeter(t)
	store := &faultyStore{MemDurableEventStore: NewMemDurableEventStore()}

	cfg := DefaultConfig()
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 10 * time.Millisecond, Instance: "cron"}
	em, err := NewDurableEmitter(store, newTestBatchEmitter(), false, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("m"), testEmitAttrs()...))
	time.Sleep(80 * time.Millisecond) // several metrics polls

	assert.Zero(t, store.observeCalls.Load(), "only the host process scans the shared queue")

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))
	var sawEmit bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			assert.NotEqual(t, "durable_emitter.queue.depth", m.Name, "plugins must not record queue depth")
			if m.Name != "durable_emitter.emit.success" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok)
			for _, dp := range sum.DataPoints {
				sawEmit = true
				assert.True(t, hasMetricStringAttr(dp.Attributes, "emitter_instance", "cron"))
			}
		}
	}
	assert.True(t, sawEmit, "expected durable_emitter.emit.success data points")
}

func TestDurableEmitter_HostObservesQueueWithDefaultInstance(t *testing.T) {
	meter, reader := newTestMeter(t)
	store := &faultyStore{MemDurableEventStore: NewMemDurableEventStore()}

	cfg := DefaultConfig()
	cfg.RetransmitInterval = time.Hour
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: 10 * time.Millisecond}
	em, err := NewDurableEmitter(store, newTestBatchEmitter(), true, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.Eventually(t, func() bool { return store.observeCalls.Load() > 0 }, 2*time.Second, 10*time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))
	var sawDepth bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "durable_emitter.queue.depth" {
				continue
			}
			g, ok := m.Data.(metricdata.Gauge[int64])
			require.True(t, ok)
			for _, dp := range g.DataPoints {
				sawDepth = true
				assert.True(t, hasMetricStringAttr(dp.Attributes, "emitter_instance", "host"))
			}
		}
	}
	assert.True(t, sawDepth, "expected durable_emitter.queue.depth data points from the host")
}
