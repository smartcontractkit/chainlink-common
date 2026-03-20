package capabilities

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

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
	UpdateDelivery(ctx context.Context, triggerId string, eventId string, lastSentAt time.Time, attempts int) error
	List(ctx context.Context) ([]PendingEvent, error)
	DeleteEvent(ctx context.Context, triggerId string, eventId string) error
	DeleteEventsForTrigger(ctx context.Context, triggerID string) error
}

type BaseTriggerMetrics interface {
	IncActiveTriggers()
	DecActiveTriggers()
	IncRetry(triggerID, eventID string)
	IncAck(triggerID, eventID string)
	ObserveTimeToAck(triggerID, eventID string, d time.Duration, attempts int)
	IncInboxMissing(triggerID string)
	IncInboxFull(triggerID string)
	EmitUndeliveredWarning(triggerID, eventID string)
	EmitUndeliveredCritical(triggerID, eventID string)
}

type undeliveredState struct {
	emittedWarning  bool
	emittedCritical bool
}

// BaseTriggerCapability keeps track of trigger registrations and handles resending events until
// they are ACKd. Events are persisted to be resilient to node restarts.
type BaseTriggerCapability[T proto.Message] struct {
	tRetransmit  time.Duration // time window for an event being ACKd before we retransmit
	store        EventStore
	newMsg       func() T // factory to allocate a new T for unmarshalling
	lggr         logger.Logger
	capabilityId string

	mu      sync.Mutex
	inboxes map[string]chan<- TriggerAndId[T]   // triggerID --> registered send channel
	pending map[string]map[string]*PendingEvent // triggerID --> eventID --> PendingEvent

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	metrics BaseTriggerMetrics
	// emit undelivered metrics after these thresholds
	undeliveredWarning     time.Duration
	undeliveredCritical    time.Duration
	undeliveredAlertStates map[string]map[string]*undeliveredState // triggerID -> eventID -> flags
}

func NewBaseTriggerCapability[T proto.Message](
	store EventStore,
	newMsg func() T,
	lggr logger.Logger,
	capabilityId string,
	tRetransmit time.Duration,
	undeliveredWarning time.Duration,
	undeliveredCritical time.Duration,
) *BaseTriggerCapability[T] {
	ctx, cancel := context.WithCancel(context.Background())
	metrics, err := NewBaseTriggerBeholderMetrics(capabilityId)
	if err != nil {
		lggr.Warnw("failed to initialize base trigger beholder metrics; continuing with metrics disabled", "err", err)
		metrics = &noopBaseTriggerMetrics{}
	}

	return &BaseTriggerCapability[T]{
		store:                  store,
		newMsg:                 newMsg,
		lggr:                   lggr,
		capabilityId:           capabilityId,
		tRetransmit:            tRetransmit,
		metrics:                metrics,
		undeliveredWarning:     undeliveredWarning,
		undeliveredCritical:    undeliveredCritical,
		undeliveredAlertStates: make(map[string]map[string]*undeliveredState),
		mu:                     sync.Mutex{},
		inboxes:                make(map[string]chan<- TriggerAndId[T]),
		pending:                make(map[string]map[string]*PendingEvent),
		ctx:                    ctx,
		cancel:                 cancel,
	}
}

func (b *BaseTriggerCapability[T]) retransmitEnabled() bool {
	return b.tRetransmit > 0
}

func (b *BaseTriggerCapability[T]) Start(ctx context.Context) error {
	b.lggr.Info("starting base trigger")

	if !b.retransmitEnabled() {
		b.lggr.Warn("retransmits disabled (tRetransmit <= 0), events will be delivered once without persistence or ACK tracking")
		return nil
	}

	recs, err := b.store.List(ctx)
	if err != nil {
		b.lggr.Errorf("failed to load persisted trigger events")
		return err
	}

	// Initialize in-memory persistence
	b.mu.Lock()
	for i := range recs {
		r := &recs[i]
		if _, ok := b.pending[r.TriggerId]; !ok {
			b.pending[r.TriggerId] = map[string]*PendingEvent{}
		}
		b.pending[r.TriggerId][r.EventId] = r
	}
	b.mu.Unlock()

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

func (b *BaseTriggerCapability[T]) RegisterTrigger(triggerID string, sendCh chan<- TriggerAndId[T]) {
	b.mu.Lock()
	_, existed := b.inboxes[triggerID]
	b.inboxes[triggerID] = sendCh
	b.mu.Unlock()

	if !existed {
		b.metrics.IncActiveTriggers()
	}
}

func (b *BaseTriggerCapability[T]) UnregisterTrigger(triggerID string) {
	b.mu.Lock()
	_, existed := b.inboxes[triggerID]
	delete(b.inboxes, triggerID)
	delete(b.pending, triggerID)
	delete(b.undeliveredAlertStates, triggerID)
	b.mu.Unlock()

	if existed {
		b.metrics.DecActiveTriggers()
	}

	if err := b.store.DeleteEventsForTrigger(b.ctx, triggerID); err != nil {
		b.lggr.Errorf("Failed to delete events for trigger (TriggerID=%s): %v", triggerID, err)
	}
}

func (b *BaseTriggerCapability[T]) DeliverEvent(
	ctx context.Context,
	te TriggerEvent,
	triggerID string,
) error {
	if !b.retransmitEnabled() {
		return b.sendToInbox(triggerID, te.ID, te.Payload.GetValue())
	}

	rec := PendingEvent{
		TriggerId:  triggerID,
		EventId:    te.ID,
		AnyTypeURL: te.Payload.GetTypeUrl(),
		Payload:    te.Payload.GetValue(),
		FirstAt:    time.Now(),
	}

	if err := b.store.Insert(ctx, rec); err != nil {
		return err
	}

	b.mu.Lock()
	if b.pending[triggerID] == nil {
		b.pending[triggerID] = map[string]*PendingEvent{}
	}
	b.pending[triggerID][te.ID] = &rec
	b.mu.Unlock()

	b.trySend(rec)
	return nil
}

// sendToInbox unmarshals the payload and delivers it to the registered inbox channel.
func (b *BaseTriggerCapability[T]) sendToInbox(triggerID, eventID string, payload []byte) error {
	b.mu.Lock()
	sendCh, ok := b.inboxes[triggerID]
	b.mu.Unlock()

	if !ok {
		b.metrics.IncInboxMissing(triggerID)
		return fmt.Errorf("no inbox registered for trigger %s", triggerID)
	}

	msg := b.newMsg()
	if err := proto.Unmarshal(payload, msg); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	wrapped := TriggerAndId[T]{Trigger: msg, Id: eventID}
	if !safeSend(sendCh, wrapped) {
		b.metrics.IncInboxFull(triggerID)
		return fmt.Errorf("inbox full or closed for trigger %s", triggerID)
	}

	b.lggr.Infof("event dispatched: capability=%s trigger=%s event=%s",
		b.capabilityId, triggerID, eventID)
	return nil
}

func (b *BaseTriggerCapability[T]) AckEvent(ctx context.Context, triggerId string, eventId string) error {
	b.lggr.Infow("Event ACK", "triggerID", triggerId, "eventID", eventId)
	if !b.retransmitEnabled() {
		return nil
	}

	var (
		attempts int
		firstAt  time.Time
		found    bool
	)

	b.mu.Lock()
	if eventsForTrigger, ok := b.pending[triggerId]; ok && eventsForTrigger != nil {
		if rec, recOk := eventsForTrigger[eventId]; recOk && rec != nil {
			attempts = rec.Attempts
			firstAt = rec.FirstAt
			found = true
		}

		delete(eventsForTrigger, eventId)
		if len(eventsForTrigger) == 0 {
			delete(b.pending, triggerId)
		}
	}

	if m, ok := b.undeliveredAlertStates[triggerId]; ok {
		delete(m, eventId)
		if len(m) == 0 {
			delete(b.undeliveredAlertStates, triggerId)
		}
	}
	b.mu.Unlock()

	if found {
		b.metrics.IncAck(triggerId, eventId)
		b.metrics.ObserveTimeToAck(triggerId, eventId, time.Since(firstAt), attempts)
	}

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
	for triggerID, pendingForTrigger := range b.pending {
		for eventID, rec := range pendingForTrigger {
			if rec.LastSentAt.IsZero() || now.Sub(rec.LastSentAt) >= b.tRetransmit {
				toResend = append(toResend, PendingEvent{
					TriggerId: rec.TriggerId,
					EventId:   rec.EventId,
				})
			}

			if b.undeliveredWarning == 0 && b.undeliveredCritical == 0 {
				continue
			}
			age := now.Sub(rec.FirstAt)

			if b.undeliveredAlertStates[triggerID] == nil {
				b.undeliveredAlertStates[triggerID] = make(map[string]*undeliveredState)
			}

			state := b.undeliveredAlertStates[triggerID][eventID]
			if state == nil {
				state = &undeliveredState{}
				b.undeliveredAlertStates[triggerID][eventID] = state
			}

			if b.undeliveredWarning > 0 && !state.emittedWarning && age >= b.undeliveredWarning {
				b.metrics.EmitUndeliveredWarning(triggerID, eventID)
				state.emittedWarning = true
			}

			if b.undeliveredCritical > 0 && !state.emittedCritical && age >= b.undeliveredCritical {
				b.metrics.EmitUndeliveredCritical(triggerID, eventID)
				state.emittedCritical = true
			}
		}
	}
	b.mu.Unlock()

	for _, event := range toResend {
		b.trySend(event)
	}
}

// trySend attempts a delivery for the given event.
// It updates Attempts and LastSentAt on every attempt locally. Success is determined
// later by an AckEvent call.
func (b *BaseTriggerCapability[T]) trySend(event PendingEvent) {
	b.mu.Lock()
	eventsForTrigger, ok := b.pending[event.TriggerId]
	if !ok || eventsForTrigger == nil {
		b.mu.Unlock()
		return
	}

	rec, ok := eventsForTrigger[event.EventId]
	if !ok || rec == nil {
		b.mu.Unlock()
		return
	}

	rec.Attempts++
	rec.LastSentAt = time.Now()

	payloadCopy := append([]byte(nil), rec.Payload...)
	attempts := rec.Attempts
	lastSent := rec.LastSentAt
	b.mu.Unlock()

	b.metrics.IncRetry(event.TriggerId, event.EventId)
	if err := b.store.UpdateDelivery(b.ctx, event.TriggerId, event.EventId, lastSent, attempts); err != nil {
		b.lggr.Errorf("failed to persist delivery update for trigger=%s event=%s: %v", event.TriggerId, event.EventId, err)
	}

	if err := b.sendToInbox(event.TriggerId, event.EventId, payloadCopy); err != nil {
		b.lggr.Errorf("trySend failed: %v", err)
		return
	}
}

func safeSend[T any](ch chan<- T, val T) (sent bool) {
	defer func() {
		if recover() != nil {
			sent = false
		}
	}()

	select {
	case ch <- val:
		return true
	default:
		return false
	}
}
