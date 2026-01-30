package capabilities

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type PendingEvent struct {
	TriggerId  string
	EventId    string
	AnyTypeURL string // Payload type
	Payload    []byte
	FirstAt    time.Time
	LastSentAt time.Time
	Attempts   int
}

type EventStore interface {
	Insert(ctx context.Context, rec PendingEvent) error
	List(ctx context.Context) ([]PendingEvent, error) // TODO: batching
	DeleteEvent(ctx context.Context, triggerId string, eventId string) error
	DeleteEventsForTrigger(ctx context.Context, triggerID string) error
}

// Decode takes a persisted record (type URL + raw bytes) and produces a typed message for the inbox.
type Decode[T any] func(te TriggerEvent) (T, error)

type BaseTriggerCapability[T any] struct {
	/*
	 Keeps track of workflow registrations (similar to LLO streams trigger).
	 Handles retransmits based on T_retransmit and T_max.
	 Persists pending events in the DB to be resilient to node restarts.
	*/
	tRetransmit time.Duration // time window for an event being ACKd before we retransmit

	store  EventStore
	decode Decode[T]
	lggr   logger.Logger

	mu      sync.Mutex
	inboxes map[string]chan<- T                 // triggerID -> registered send channel
	pending map[string]map[string]*PendingEvent // triggerID --> eventID

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewBaseTriggerCapability[T any](
	store EventStore,
	decode Decode[T],
	lggr logger.Logger,
	tRetransmit time.Duration,
) *BaseTriggerCapability[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseTriggerCapability[T]{
		store:       store,
		decode:      decode,
		lggr:        lggr,
		tRetransmit: tRetransmit,
		mu:          sync.Mutex{},
		inboxes:     make(map[string]chan<- T),
		pending:     make(map[string]map[string]*PendingEvent),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (b *BaseTriggerCapability[T]) Start(ctx context.Context) error {
	b.lggr.Info("starting base trigger")
	b.ctx, b.cancel = context.WithCancel(ctx)

	recs, err := b.store.List(ctx)
	if err != nil {
		b.lggr.Errorf("failed to load persisted trigger events")
		return err
	}

	// Initialize in-memory persistence
	b.pending = make(map[string]map[string]*PendingEvent)
	for i := range recs {
		r := &recs[i]
		if _, ok := b.pending[r.TriggerId]; !ok {
			b.pending[r.TriggerId] = map[string]*PendingEvent{}
		}
		b.pending[r.TriggerId][r.EventId] = r
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.retransmitLoop()
	}()
	return nil
}

func (b *BaseTriggerCapability[T]) Stop() {
	b.cancel()
	b.wg.Wait()
}

func (b *BaseTriggerCapability[T]) RegisterTrigger(triggerID string, sendCh chan<- T) {
	b.mu.Lock()
	b.inboxes[triggerID] = sendCh
	b.mu.Unlock()
}

func (b *BaseTriggerCapability[T]) UnregisterTrigger(triggerID string) {
	b.mu.Lock()
	delete(b.inboxes, triggerID)
	delete(b.pending, triggerID)
	b.mu.Unlock()
	if err := b.store.DeleteEventsForTrigger(b.ctx, triggerID); err != nil {
		b.lggr.Errorf("Failed to delete events for trigger (TriggerID=%s): %v", triggerID, err)
	}
}

func (b *BaseTriggerCapability[T]) DeliverEvent(
	ctx context.Context,
	te TriggerEvent,
	triggerID string,
) error {
	rec := PendingEvent{
		TriggerId:  te.TriggerType,
		EventId:    te.ID,
		AnyTypeURL: te.Payload.GetTypeUrl(),
		Payload:    te.Payload.GetValue(),
		FirstAt:    time.Now(),
	}

	if err := b.store.Insert(ctx, rec); err != nil {
		return err
	}

	// TODO: Problem here?
	b.mu.Lock()
	if b.pending[triggerID] == nil {
		b.pending[triggerID] = map[string]*PendingEvent{}
	}
	b.pending[triggerID][te.ID] = &rec
	b.mu.Unlock()

	if err := b.trySend(ctx, rec); err != nil {
		b.lggr.Infof("failed to send event: %v", err)
	}
	return nil // Retry will occur later
}

func (b *BaseTriggerCapability[T]) AckEvent(ctx context.Context, triggerId string, eventId string) error {
	b.lggr.Infof("Event ACK (triggerID: %s, eventID %s)", triggerId, eventId)
	b.mu.Lock()
	if eventsForTrigger, ok := b.pending[triggerId]; ok && eventsForTrigger != nil {
		delete(eventsForTrigger, eventId)
		if len(eventsForTrigger) == 0 {
			delete(b.pending, triggerId)
		}
	}
	b.mu.Unlock()
	return b.store.DeleteEvent(ctx, triggerId, eventId)
}

func (b *BaseTriggerCapability[T]) retransmitLoop() {
	ticker := time.NewTicker(b.tRetransmit / 2)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.lggr.Debug("retransmitting unacknowledged events")
			b.scanPending()
		}
	}
}

func (b *BaseTriggerCapability[T]) scanPending() {
	now := time.Now()

	b.mu.Lock()
	toResend := make([]PendingEvent, 0, len(b.pending))
	for _, pendingForTrigger := range b.pending {
		for _, rec := range pendingForTrigger {
			if rec.LastSentAt.IsZero() || now.Sub(rec.LastSentAt) >= b.tRetransmit {
				toResend = append(toResend, PendingEvent{
					TriggerId: rec.TriggerId,
					EventId:   rec.EventId,
				})
			}
		}
	}
	b.mu.Unlock()

	for _, event := range toResend {
		err := b.trySend(b.ctx, event)
		if err != nil {
			b.lggr.Errorf("failed to resend event (triggerID: %s, eventID: %s): %v", event.TriggerId, event.EventId, err)
		}
	}
}

// trySend attempts a delivery for the given event.
// It updates Attempts and LastSentAt on every attempt locally. Success is determined
// later by an AckEvent call.
func (b *BaseTriggerCapability[T]) trySend(ctx context.Context, event PendingEvent) error {
	b.lggr.Infof("resending event (triggerID: %s, eventID: %s)", event.TriggerId, event.EventId)
	b.mu.Lock()
	eventsForTrigger, ok := b.pending[event.TriggerId]
	if !ok || eventsForTrigger == nil {
		b.mu.Unlock()
		return nil
	}

	rec, ok := eventsForTrigger[event.EventId]
	if !ok || rec == nil {
		b.mu.Unlock()
		return nil
	}

	rec.Attempts++
	rec.LastSentAt = time.Now()

	typeURL := rec.AnyTypeURL
	payloadCopy := append([]byte(nil), rec.Payload...)

	sendCh, ok := b.inboxes[event.TriggerId]
	b.mu.Unlock()
	if !ok {
		err := fmt.Errorf("no inbox registered for trigger %s", event.TriggerId)
		b.lggr.Errorf(err.Error())
		return err
	}

	te := TriggerEvent{
		TriggerType: event.TriggerId,
		ID:          event.EventId,
		Payload:     &anypb.Any{TypeUrl: typeURL, Value: payloadCopy},
	}

	msg, err := b.decode(te)
	if err != nil {
		b.lggr.Errorf("failed to decode payload into trigger message type: %v", err)
		return err
	}

	select {
	case sendCh <- msg:
		b.lggr.Infof("event dispatched: trigger=%s event=%s attempt=%d",
			event.TriggerId, event.EventId, rec.Attempts)
		return nil
	default:
		return fmt.Errorf("inbox full for trigger %s", event.TriggerId)
	}
}
