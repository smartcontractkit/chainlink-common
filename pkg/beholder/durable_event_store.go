package beholder

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
	Depth            int64
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
	// match DurableEmitterConfig (nearExpiryLead should be << eventTTL).
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
	// MarkDelivered records successful delivery to Chip. The row must no longer
	// appear in ListPending. Postgres implementations typically set delivered_at;
	// a background PurgeDelivered removes rows later. MemDurableEventStore removes
	// the row immediately (same as Delete).
	MarkDelivered(ctx context.Context, id int64) error
	// MarkDeliveredBatch marks multiple events as delivered in a single operation.
	// Semantically equivalent to calling MarkDelivered for each id.
	MarkDeliveredBatch(ctx context.Context, ids []int64) (int64, error)
	// PurgeDelivered deletes up to batchLimit rows already marked delivered.
	// Implementations that remove rows in MarkDelivered may return 0, nil always.
	PurgeDelivered(ctx context.Context, batchLimit int) (deleted int64, err error)
	// ListPending returns events created before the given cutoff, ordered by
	// creation time ascending, up to limit rows.
	ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error)
	// DeleteExpired removes events older than ttl and returns the count deleted.
	DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
	// MarkFailedBatch records a delivery failure for a batch of events, annotating
	// each row with the provided errorMessage. Failed events remain pending so the
	// retransmit loop can attempt redelivery.
	MarkFailedBatch(ctx context.Context, errorMessage string, ids []int64) error
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

func (s *metricsInstrumentedStore) MarkDelivered(ctx context.Context, id int64) error {
	t0 := time.Now()
	err := s.inner.MarkDelivered(ctx, id)
	s.m.recordStoreOp(ctx, "mark_delivered", time.Since(t0), err)
	return err
}

func (s *metricsInstrumentedStore) MarkDeliveredBatch(ctx context.Context, ids []int64) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.MarkDeliveredBatch(ctx, ids)
	s.m.recordStoreOp(ctx, "mark_delivered_batch", time.Since(t0), err)
	return n, err
}

func (s *metricsInstrumentedStore) PurgeDelivered(ctx context.Context, batchLimit int) (int64, error) {
	t0 := time.Now()
	n, err := s.inner.PurgeDelivered(ctx, batchLimit)
	s.m.recordStoreOp(ctx, "purge_delivered", time.Since(t0), err)
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

func (s *metricsInstrumentedStore) MarkFailedBatch(ctx context.Context, errorMessage string, ids []int64) error {
	t0 := time.Now()
	err := s.inner.MarkFailedBatch(ctx, errorMessage, ids)
	s.m.recordStoreOp(ctx, "mark_failed_batch", time.Since(t0), err)
	return err
}
