package capabilities

import (
	"context"
	"sync"
)

type MemEventStore struct {
	mu   sync.Mutex
	recs map[string]PendingEvent
}

func NewMemEventStore() *MemEventStore {
	return &MemEventStore{recs: make(map[string]PendingEvent)}
}

func (m *MemEventStore) Insert(ctx context.Context, r PendingEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs[key(r.TriggerId, r.EventId)] = r
	return nil
}

func (m *MemEventStore) DeleteEvent(ctx context.Context, triggerId, eventId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recs, key(triggerId, eventId))
	return nil
}

func (m *MemEventStore) DeleteEventsForTrigger(ctx context.Context, triggerId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, r := range m.recs {
		if r.TriggerId == triggerId {
			delete(m.recs, k)
		}
	}
	return nil
}

func (m *MemEventStore) List(ctx context.Context) ([]PendingEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]PendingEvent, 0, len(m.recs))
	for _, r := range m.recs {
		out = append(out, r)
	}
	return out, nil
}
