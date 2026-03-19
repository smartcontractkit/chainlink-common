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

// DurableEventStore abstracts the persistence layer for durable chip events.
// Implementations must be safe for concurrent use.
type DurableEventStore interface {
	// Insert persists a serialized event and returns its assigned ID.
	Insert(ctx context.Context, payload []byte) (int64, error)
	// Delete removes a successfully delivered event.
	Delete(ctx context.Context, id int64) error
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

var _ DurableEventStore = (*MemDurableEventStore)(nil)

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
