package durableemitter

import (
	"context"
	"errors"
	"sync"
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
// EstimateBacklog errors while fail is set and returns the exact row count when
// exact is set. A non-nil insertHold blocks batch inserts until it is closed;
// failInserts makes them fail.
type estimatingStore struct {
	*MemDurableEventStore
	estimate    atomic.Int64
	fail        atomic.Bool
	exact       bool
	failInserts atomic.Bool
	insertHold  chan struct{}
}

var _ BacklogEstimator = (*estimatingStore)(nil)

func (s *estimatingStore) EstimateBacklog(context.Context) (int64, error) {
	if s.fail.Load() {
		return 0, errors.New("estimate timeout")
	}
	if s.exact {
		return int64(s.Len()), nil
	}
	return s.estimate.Load(), nil
}

func (s *estimatingStore) InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error) {
	if s.insertHold != nil {
		<-s.insertHold
	}
	if s.failInserts.Load() {
		return nil, errors.New("insert timeout")
	}
	return s.MemDurableEventStore.InsertBatch(ctx, payloads)
}

// gatedEstimatingStore holds EstimateBacklog open while gate is set, so a test
// can write rows while a refresh is in flight.
type gatedEstimatingStore struct {
	*MemDurableEventStore
	estimate atomic.Int64
	gate     atomic.Bool
	entered  chan struct{}
	release  chan struct{}
}

func (s *gatedEstimatingStore) EstimateBacklog(context.Context) (int64, error) {
	if s.gate.Load() {
		s.entered <- struct{}{}
		<-s.release
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

func TestDurableEmitter_BackpressureCountsOwnInsertsBetweenPolls(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable")) // rows are never deleted

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 2
	cfg.BacklogPollInterval = time.Hour
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("1"), testEmitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("2"), testEmitAttrs()...))
	require.ErrorIs(t, em.Emit(ctx, []byte("3"), testEmitAttrs()...), ErrBacklogged)
	assert.Equal(t, 2, store.Len())
}

func TestDurableEmitter_BackpressureOffByDefault(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(DefaultMaxQueueDepth)

	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), nil)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("x"), testEmitAttrs()...))
	assert.Nil(t, em.backlogEstimator, "the backlog is not polled unless back-pressure is enabled")
}

func TestDurableEmitter_BackpressureRefreshKeepsInsertsMadeDuringPoll(t *testing.T) {
	store := &gatedEstimatingStore{
		MemDurableEventStore: NewMemDurableEventStore(),
		entered:              make(chan struct{}),
		release:              make(chan struct{}),
	}
	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 10
	cfg.BacklogPollInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	store.gate.Store(true)
	done := make(chan struct{})
	go func() {
		defer close(done)
		em.refreshBacklog(ctx)
	}()
	<-store.entered
	for range 3 {
		require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	}
	// The estimate was taken before the inserts, so it does not include them.
	store.estimate.Store(8)
	store.release <- struct{}{}
	<-done

	assert.Equal(t, int64(11), em.estimatedBacklog(), "inserts made during the poll stay counted")
	require.ErrorIs(t, em.Emit(ctx, []byte("rejected"), testEmitAttrs()...), ErrBacklogged)
}

func TestDurableEmitter_BackpressureStalePollKeepsRecentInserts(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(90)
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable")) // rows are never deleted

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 100
	cfg.BacklogPollInterval = time.Hour
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	for range 10 {
		require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	}
	require.ErrorIs(t, em.Emit(ctx, []byte("over"), testEmitAttrs()...), ErrBacklogged)

	// Postgres statistics may not include the 10 inserts yet: a poll still
	// returning 90 must not drop them from the estimate.
	em.refreshBacklog(ctx)
	require.ErrorIs(t, em.Emit(ctx, []byte("over"), testEmitAttrs()...), ErrBacklogged)

	// One poll later they are assumed to be reflected in the statistics.
	em.refreshBacklog(ctx)
	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
}

func TestDurableEmitter_BackpressureCountsRemovalsWhilePollsFail(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 5
	cfg.BacklogPollInterval = 10 * time.Millisecond
	cfg.DeleteBatchFlushInterval = time.Millisecond
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	store.fail.Store(true) // seeded at 0, every later poll fails
	for range 3 {
		for range 5 {
			require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
		}
		require.Eventually(t, func() bool { return em.estimatedBacklog() == 0 }, 2*time.Second, 5*time.Millisecond,
			"delivered rows are removed from the estimate without a successful poll")
	}
	assert.Equal(t, 0, store.Len())
}

func TestDurableEmitter_BackpressureReservesBeforeInsert(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore(), insertHold: make(chan struct{})}
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable"))

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 5
	cfg.BacklogPollInterval = time.Hour
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(store.insertHold) }) }
	t.Cleanup(release)
	ctx := t.Context()

	// Inserts are held, so no emit completes before all of them are admitted or rejected.
	var accepted, rejected atomic.Int64
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			switch err := em.Emit(ctx, []byte("x"), testEmitAttrs()...); {
			case err == nil:
				accepted.Add(1)
			case errors.Is(err, ErrBacklogged):
				rejected.Add(1)
			}
		})
	}
	require.Eventually(t, func() bool { return rejected.Load() == 15 }, 2*time.Second, 5*time.Millisecond,
		"emits beyond MaxQueueDepth are rejected while the admitted inserts are still in flight")
	release()
	wg.Wait()
	assert.Equal(t, int64(5), accepted.Load())
	assert.Equal(t, 5, store.Len())
}

func TestDurableEmitter_BackpressureInflightSurvivesPolls(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore(), insertHold: make(chan struct{})}
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable"))

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 5
	cfg.BacklogPollInterval = time.Hour
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(store.insertHold) }) }
	t.Cleanup(release)
	ctx := t.Context()

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() { assert.NoError(t, em.Emit(ctx, []byte("held"), testEmitAttrs()...)) })
	}
	require.Eventually(t, func() bool { return em.backlogInflight.Load() == 5 }, 2*time.Second, 5*time.Millisecond)

	// The held inserts have not committed, so polls cannot see them.
	em.refreshBacklog(ctx)
	em.refreshBacklog(ctx)
	require.ErrorIs(t, em.Emit(ctx, []byte("over"), testEmitAttrs()...), ErrBacklogged,
		"in-flight inserts stay counted across polls")

	release()
	wg.Wait()
	assert.Equal(t, int64(5), em.estimatedBacklog())
}

func TestDurableEmitter_BackpressureCountsFailedInsertsAsPossiblyPersisted(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.failInserts.Store(true)

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 1
	cfg.BacklogPollInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.Error(t, em.Emit(ctx, []byte("timed out"), testEmitAttrs()...))
	require.ErrorIs(t, em.Emit(ctx, []byte("over"), testEmitAttrs()...), ErrBacklogged,
		"an insert that reached the store may have committed")

	// Two polls later it is assumed to be reflected by the store estimate (0 here).
	store.failInserts.Store(false)
	em.refreshBacklog(ctx)
	em.refreshBacklog(ctx)
	require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
}

func TestDurableEmitter_BackpressureIgnoresRemovalsStartedBeforePoll(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	store.estimate.Store(10)

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.BacklogPollInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	before := em.removalGen()
	em.refreshBacklog(ctx) // may already reflect a removal that started before it
	em.recordRemovals(before, nil, 3)
	assert.Equal(t, int64(10), em.estimatedBacklog())

	em.recordRemovals(em.removalGen(), nil, 3)
	assert.Equal(t, int64(7), em.estimatedBacklog())
}

func TestDurableEmitter_BackpressureRefreshDoesNotRecountDeliveredInserts(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore(), exact: true}

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 5
	cfg.BacklogPollInterval = time.Hour
	cfg.DeleteBatchFlushInterval = time.Millisecond
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	for range 5 {
		require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	}
	require.Eventually(t, func() bool { return store.Len() == 0 && em.estimatedBacklog() == 0 },
		2*time.Second, 5*time.Millisecond)

	// The inserts move to the previous generation, but they are already deleted.
	em.refreshBacklog(ctx)
	assert.Equal(t, int64(0), em.estimatedBacklog())
	for range 5 {
		require.NoError(t, em.Emit(ctx, []byte("x"), testEmitAttrs()...))
	}
}

func TestDurableEmitter_BackpressureNeverCountsARemovalTwice(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore()}
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip unavailable")) // rows are never deleted by delivery

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.BacklogPollInterval = time.Hour
	cfg.RetransmitInterval = time.Hour
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	require.NoError(t, em.Emit(ctx, []byte("a"), testEmitAttrs()...))
	require.NoError(t, em.Emit(ctx, []byte("b"), testEmitAttrs()...))
	own := em.backlogSnap.Load().cur
	started := em.removalGen()
	require.Equal(t, int64(2), em.estimatedBacklog())

	// Retransmit deletes row a, then the original delivery's delete finds it gone.
	em.recordRemovals(started, []*backlogGen{nil}, 1)
	em.recordRemovals(started, []*backlogGen{own}, 0)
	assert.Equal(t, int64(1), em.estimatedBacklog())

	// One of two rows removed, without knowing which: neither group is credited.
	em.recordRemovals(started, []*backlogGen{own, nil}, 1)
	assert.Equal(t, int64(1), em.estimatedBacklog())

	em.recordRemovals(started, []*backlogGen{own}, 1)
	assert.Equal(t, int64(0), em.estimatedBacklog())
}

// Exercises inserts, deliveries, deletes, retransmit and polls concurrently for
// the race detector, then checks the estimate converges to the real row count.
func TestDurableEmitter_BackpressureConcurrentAccounting(t *testing.T) {
	store := &estimatingStore{MemDurableEventStore: NewMemDurableEventStore(), exact: true}

	cfg := DefaultConfig()
	cfg.BackpressureEnabled = true
	cfg.MaxQueueDepth = 50
	cfg.BacklogPollInterval = time.Millisecond
	cfg.DeleteBatchFlushInterval = time.Millisecond
	cfg.RetransmitInterval = 5 * time.Millisecond
	cfg.RetransmitAfter = time.Millisecond
	cfg.InsertBatchFlushInterval = time.Millisecond
	em := newTestDurableEmitter(t, store, newTestBatchEmitter(), &cfg)
	servicetest.Run(t, em)
	ctx := t.Context()

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 50 {
				err := em.Emit(ctx, []byte("x"), testEmitAttrs()...)
				if err != nil && !errors.Is(err, ErrBacklogged) {
					assert.NoError(t, err)
				}
			}
		})
	}
	wg.Wait()

	require.Eventually(t, func() bool { return store.Len() == 0 }, 5*time.Second, 5*time.Millisecond)
	assert.Zero(t, em.backlogInflight.Load())
	// Two polls after the last write, only the store estimate remains.
	require.Eventually(t, func() bool { return em.estimatedBacklog() == 0 }, 5*time.Second, 5*time.Millisecond)
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
