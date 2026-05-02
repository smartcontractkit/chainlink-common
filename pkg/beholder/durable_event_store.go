package beholder

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
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
