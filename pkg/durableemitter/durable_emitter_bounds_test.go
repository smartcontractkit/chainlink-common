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

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

func (d *DurableEmitter) inFlightCount() int {
	d.inFlightMu.Lock()
	defer d.inFlightMu.Unlock()
	return len(d.inFlight)
}

// rejectingBatchEmitter refuses every message, like a full batch client buffer.
type rejectingBatchEmitter struct{ attempts atomic.Int64 }

func (b *rejectingBatchEmitter) QueueMessage(*chipingress.CloudEventPb, func(error)) error {
	b.attempts.Add(1)
	return batch.ErrMessageBufferFull
}
func (b *rejectingBatchEmitter) Start(context.Context) {}
func (b *rejectingBatchEmitter) Stop()                 {}

func boundsCfg() Config {
	cfg := DefaultConfig()
	cfg.DisablePruning = true
	cfg.RetransmitInterval = 5 * time.Millisecond
	cfg.RetransmitAfter = -time.Hour // every row is eligible at once
	cfg.RetransmitBatchSize = 100
	return cfg
}

func counterSum(t *testing.T, rm metricdata.ResourceMetrics, name string) int64 {
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
				total += dp.Value
			}
		}
	}
	return total
}

func TestDurableEmitter_RetransmitSkipsInFlightRows(t *testing.T) {
	store := NewMemDurableEventStore()
	ctx := context.Background()
	const numRows = 5
	for range numRows {
		_, err := store.Insert(ctx, makeCloudEventPayload(t, "t", "body"))
		require.NoError(t, err)
	}

	meter, reader := newTestMeter(t)
	cfg := boundsCfg()
	cfg.Metrics = &DurableEmitterMetricsConfig{PollInterval: time.Hour}
	be := &heldBatchEmitter{}
	em, err := NewDurableEmitter(store, be, true, cfg, logger.Test(t), meter)
	require.NoError(t, err)
	servicetest.Run(t, em)

	// Many ticks pass while the batch emitter never acks: each row is queued once.
	require.Eventually(t, func() bool {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))
		return counterSum(t, rm, "durable_emitter.retransmit.skipped_in_flight") >= 10*numRows
	}, 5*time.Second, 5*time.Millisecond)
	require.Len(t, be.queued(), numRows, "a row in flight must not be queued again")
	assert.Equal(t, numRows, em.inFlightCount())

	// Failed delivery releases the rows, so the next tick queues them again.
	for _, cb := range be.queued() {
		cb(errors.New("send failed"))
	}
	require.Eventually(t, func() bool { return len(be.queued()) == 2*numRows }, 5*time.Second, 5*time.Millisecond)
	time.Sleep(10 * cfg.RetransmitInterval)
	assert.Len(t, be.queued(), 2*numRows)

	// Successful delivery releases and deletes.
	for _, cb := range be.queued()[numRows:] {
		cb(nil)
	}
	require.Eventually(t, func() bool { return store.Len() == 0 && em.inFlightCount() == 0 }, 5*time.Second, 5*time.Millisecond)
	assert.Len(t, be.queued(), 2*numRows, "a delivered row is not queued again before its delete")
}

func TestDurableEmitter_EmitPathRowsNotRetransmittedWhileInFlight(t *testing.T) {
	store := NewMemDurableEventStore()
	cfg := boundsCfg()
	cfg.InsertBatchSize = 0
	be := &heldBatchEmitter{}
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	const numEvents = 3
	for range numEvents {
		require.NoError(t, em.Emit(t.Context(), []byte("body"), testEmitAttrs()...))
	}
	time.Sleep(20 * cfg.RetransmitInterval)
	assert.Len(t, be.queued(), numEvents, "retransmit must skip rows queued by Emit")
	assert.Equal(t, numEvents, em.inFlightCount())
}

func TestDurableEmitter_InFlightReleasedOnEnqueueError(t *testing.T) {
	store := NewMemDurableEventStore()
	cfg := boundsCfg()
	cfg.InsertBatchSize = 0
	be := &rejectingBatchEmitter{}
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("body"), testEmitAttrs()...))
	// Each tick retries the rejected row, so it was never left marked in flight.
	require.Eventually(t, func() bool { return be.attempts.Load() >= 5 }, 5*time.Second, 5*time.Millisecond)
	assert.Zero(t, em.inFlightCount())
	assert.Equal(t, 1, store.Len())
}

func TestDurableEmitter_InFlightClearedOnStop(t *testing.T) {
	store := NewMemDurableEventStore()
	cfg := boundsCfg()
	cfg.InsertBatchSize = 0
	be := &heldBatchEmitter{} // Stop returns without running callbacks
	em := newTestDurableEmitter(t, store, be, &cfg)
	require.NoError(t, em.Start(t.Context()))
	require.NoError(t, em.Emit(t.Context(), []byte("body"), testEmitAttrs()...))
	require.Equal(t, 1, em.inFlightCount())
	require.NoError(t, em.Close())
	assert.Zero(t, em.inFlightCount())
}

func TestDurableEmitter_InFlightConcurrentCallbacks(t *testing.T) {
	store := NewMemDurableEventStore()
	cfg := boundsCfg()
	be := newTestBatchEmitter()
	em := newTestDurableEmitter(t, store, be, &cfg)
	servicetest.Run(t, em)

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_ = em.Emit(t.Context(), []byte("body"), testEmitAttrs()...)
		})
	}
	wg.Wait()
	require.Eventually(t, func() bool { return store.Len() == 0 && em.inFlightCount() == 0 }, 5*time.Second, 5*time.Millisecond)
}

func TestMemListPendingByteBudget(t *testing.T) {
	store := NewMemDurableEventStore()
	ctx := context.Background()
	sizes := []int{10, 10, 10, 50, 10}
	payloads := make([][]byte, len(sizes))
	for i, n := range sizes {
		payloads[i] = make([]byte, n)
	}
	ids, err := store.InsertBatch(ctx, payloads)
	require.NoError(t, err)
	before := time.Now().Add(time.Hour)

	page, err := store.ListPending(ctx, before, time.Time{}, 0, 10, 25)
	require.NoError(t, err)
	assert.Equal(t, ids[:2], eventIDs(page), "stops before the row that crosses the budget")

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 10, 30)
	require.NoError(t, err)
	assert.Equal(t, ids[:3], eventIDs(page), "a row landing exactly on the budget is kept")

	last := page[len(page)-1]
	page, err = store.ListPending(ctx, before, last.CreatedAt, last.ID, 10, 25)
	require.NoError(t, err)
	assert.Equal(t, ids[3:4], eventIDs(page), "a single row over the budget is still returned")

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, ids, eventIDs(page), "zero budget is unbounded")

	page, err = store.ListPending(ctx, before, time.Time{}, 0, 2, 1000)
	require.NoError(t, err)
	assert.Equal(t, ids[:2], eventIDs(page), "row limit still applies")
}

func TestDurableEmitter_RetransmitCursorAdvancesOnByteCutPage(t *testing.T) {
	store := NewMemDurableEventStore()
	ctx := context.Background()
	const numRows = 3
	var ids []int64
	maxPayload := 0
	for range numRows {
		p := makeCloudEventPayload(t, "t", "body")
		maxPayload = max(maxPayload, len(p))
		id, err := store.Insert(ctx, p)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	cfg := boundsCfg()
	cfg.RetransmitBatchSize = 10
	cfg.RetransmitMaxPageBytes = maxPayload + 1 // one row per page
	be := &heldBatchEmitter{}
	em := newTestDurableEmitter(t, store, be, &cfg)

	for i := range numRows {
		em.retransmitPending()
		assert.Equal(t, ids[i], em.retransmitCursorID, "byte-cut page %d must advance the cursor", i)
		assert.Len(t, be.queued(), i+1)
	}
	// Past the end: the empty page wraps to the oldest row, which is in flight.
	em.retransmitPending()
	assert.Equal(t, ids[0], em.retransmitCursorID)
	assert.Len(t, be.queued(), numRows)
}

func TestDurableEmitter_RetransmitMaxPageBytesDefault(t *testing.T) {
	em := newTestDurableEmitter(t, NewMemDurableEventStore(), newTestBatchEmitter(), &Config{})
	assert.Equal(t, DefaultRetransmitMaxPageBytes, em.retransmitMaxPageBytes())
	assert.Equal(t, DefaultRetransmitMaxPageBytes, DefaultConfig().RetransmitMaxPageBytes)
}

func eventIDs(evs []DurableEvent) []int64 {
	out := make([]int64, len(evs))
	for i, e := range evs {
		out[i] = e.ID
	}
	return out
}
