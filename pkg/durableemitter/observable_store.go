package durableemitter

import (
	"context"
	"errors"
	"time"
)

// DurableEvent represents a persisted event awaiting delivery to Chip.
type DurableEvent struct {
	ID        int64
	Payload   []byte // serialized CloudEventPb proto
	CreatedAt time.Time
}

// DurableQueueStats is a point-in-time snapshot of the pending queue for metrics.
type DurableQueueStats struct {
	// Depth is the number of undelivered (delivered_at IS NULL) rows — the
	// delivery backlog.
	Depth int64
	// TotalRows is the number of rows physically present in the table, including
	// delivered-but-not-yet-purged rows. This is the authoritative "queue depth"
	// (actual table count): it is read directly from the DB so it stays correct
	// regardless of how many writers share the table or which in-memory delta
	// updates were lost to failed/partial DB operations.
	TotalRows        int64
	PayloadBytes     int64
	OldestPendingAge time.Duration // 0 if the queue is empty
	// NearTTLCount is the number of rows within nearExpiryLead of EventTTL (still
	// pending, not yet removed by expiry). Serves as a DLQ-pressure proxy; there is
	// no separate dead-letter table in the default design.
	NearTTLCount int64
}

// DurableQueueObserver is optionally implemented by DurableEventStore implementations
// so DurableEmitter can export queue depth and age gauges when metrics are enabled.
type DurableQueueObserver interface {
	// ObserveDurableQueue returns live queue statistics. eventTTL and nearExpiryLead
	// match Config (nearExpiryLead should be << eventTTL).
	ObserveDurableQueue(ctx context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error)
}

// BatchInserter is optionally implemented by DurableEventStore implementations
// to support multi-row inserts for higher throughput. When the store implements
// this interface and InsertBatchSize > 0, DurableEmitter coalesces Emit() calls
// into batched INSERTs, dramatically reducing per-event transaction overhead.
type BatchInserter interface {
	InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error)
}

// DurableEventStore abstracts the persistence layer for durable chip events.
// Implementations must be safe for concurrent use.
type DurableEventStore interface {
	// Insert persists a serialized event and returns its assigned ID.
	Insert(ctx context.Context, payload []byte) (int64, error)
	// Delete physically removes a row (corrupt payloads, policy drops, tests).
	Delete(ctx context.Context, id int64) error
	// BatchDelete records successful delivery of multiple events to Chip by
	// deleting them in a single operation (delete-on-delivery)
	BatchDelete(ctx context.Context, ids []int64) (int64, error)
	// ListPending returns events created before the given cutoff, ordered by
	// creation time ascending, up to limit rows. Under delete-on-delivery every
	// row still present is undelivered, so this is the pending backlog.
	ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error)
	// DeleteExpired removes any events older than ttl and returns the count
	// deleted — a time-based garbage collector that also reclaims rows which
	// failed to delete on delivery (e.g. a DB error in the delivery callback), so
	// nothing lingers past EventTTL.
	DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
}

// metricsInstrumentedStore wraps DurableEventStore to record store operation metrics.
type metricsInstrumentedStore struct {
	inner DurableEventStore
	m     *durableEmitterMetrics
}

var _ DurableEventStore = (*metricsInstrumentedStore)(nil)
var _ DurableQueueObserver = (*metricsInstrumentedStore)(nil)

func newMetricsInstrumentedStore(inner DurableEventStore, m *durableEmitterMetrics) DurableEventStore {
	if m == nil {
		return inner
	}
	return &metricsInstrumentedStore{inner: inner, m: m}
}

func (s *metricsInstrumentedStore) Insert(ctx context.Context, payload []byte) (int64, error) {
	t0 := time.Now()
	id, err := s.inner.Insert(ctx, payload)
	s.m.recordStoreOp(ctx, "insert", time.Since(t0), err)
	return id, err
}

func (s *metricsInstrumentedStore) Delete(ctx context.Context, id int64) error {
	t0 := time.Now()
	err := s.inner.Delete(ctx, id)
	s.m.recordStoreOp(ctx, "delete", time.Since(t0), err)
	return err
}

func (s *metricsInstrumentedStore) BatchDelete(ctx context.Context, ids []int64) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.BatchDelete(ctx, ids)
	s.m.recordStoreOp(ctx, "batch_delete", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error) {
	t0 := time.Now()
	evs, err := s.inner.ListPending(ctx, createdBefore, limit)
	s.m.recordStoreOp(ctx, "list_pending", time.Since(t0), err)
	return evs, err
}

func (s *metricsInstrumentedStore) DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.DeleteExpired(ctx, ttl)
	s.m.recordStoreOp(ctx, "delete_expired", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) ObserveDurableQueue(ctx context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error) {
	o, ok := s.inner.(DurableQueueObserver)
	if !ok {
		return DurableQueueStats{}, errors.New("inner DurableEventStore does not implement DurableQueueObserver")
	}
	return o.ObserveDurableQueue(ctx, eventTTL, nearExpiryLead)
}

func (s *metricsInstrumentedStore) InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error) {
	bi, ok := s.inner.(BatchInserter)
	if !ok {
		return nil, errors.New("inner DurableEventStore does not implement BatchInserter")
	}
	t0 := time.Now()
	ids, err := bi.InsertBatch(ctx, payloads)
	s.m.recordStoreOp(ctx, "insert_batch", time.Since(t0), err)
	return ids, err
}
