package capabilities

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MemEventStore struct {
	mu   sync.Mutex
	recs map[string]map[string]PendingEvent // triggerID -> eventID -> event
}

func NewMemEventStore() *MemEventStore {
	return &MemEventStore{
		recs: make(map[string]map[string]PendingEvent),
	}
}

func (m *MemEventStore) Insert(ctx context.Context, r PendingEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eventsForTrigger := m.recs[r.TriggerId]
	if eventsForTrigger == nil {
		eventsForTrigger = make(map[string]PendingEvent)
		m.recs[r.TriggerId] = eventsForTrigger
	}
	eventsForTrigger[r.EventId] = r
	return nil
}

func (m *MemEventStore) UpdateDelivery(ctx context.Context, triggerId string, eventId string, lastSentAt time.Time, attempts int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eventsForTrigger := m.recs[triggerId]
	if eventsForTrigger == nil {
		return fmt.Errorf("event not found trigger=%s event=%s", triggerId, eventId)
	}

	rec, ok := eventsForTrigger[eventId]
	if !ok {
		return fmt.Errorf("event not found trigger=%s event=%s", triggerId, eventId)
	}

	rec.Attempts = attempts
	rec.LastSentAt = lastSentAt
	eventsForTrigger[eventId] = rec
	return nil
}

func (m *MemEventStore) DeleteEvent(ctx context.Context, triggerId, eventId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eventsForTrigger := m.recs[triggerId]
	if eventsForTrigger == nil {
		return nil
	}
	delete(eventsForTrigger, eventId)
	if len(eventsForTrigger) == 0 {
		delete(m.recs, triggerId)
	}
	return nil
}

func (m *MemEventStore) DeleteEventsForTrigger(ctx context.Context, triggerId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.recs, triggerId)
	return nil
}

func (m *MemEventStore) List(ctx context.Context) ([]PendingEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]PendingEvent, 0)
	for _, eventsForTrigger := range m.recs {
		for _, r := range eventsForTrigger {
			out = append(out, r)
		}
	}
	return out, nil
}
