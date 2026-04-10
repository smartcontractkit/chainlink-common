package beholder

import (
	"context"
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
	_ DurableEventStore     = (*MemDurableEventStore)(nil)
	_ DurableQueueObserver  = (*MemDurableEventStore)(nil)
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
