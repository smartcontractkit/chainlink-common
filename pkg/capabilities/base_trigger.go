package capabilities

import (
	"context"
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
	IncRetry(triggerID, eventID string)
	IncAck(triggerID, eventID string, attempts int)
	ObserveTimeToAck(triggerID, eventID string, d time.Duration, attempts int)
	IncInboxMissing(triggerID string)
	IncInboxFull(triggerID string)
	EmitUndelivered(triggerID, eventID string, age time.Duration, attempts int)  // after T
	EmitUndelivered2(triggerID, eventID string, age time.Duration, attempts int) // after T2
}

type noopBaseTriggerMetrics struct{}

func (noopBaseTriggerMetrics) IncRetry(string, string)                             {}
func (noopBaseTriggerMetrics) IncAck(string, string, int)                          {}
func (noopBaseTriggerMetrics) ObserveTimeToAck(string, string, time.Duration, int) {}
func (noopBaseTriggerMetrics) IncInboxMissing(string)                              {}
func (noopBaseTriggerMetrics) IncInboxFull(string)                                 {}
func (noopBaseTriggerMetrics) EmitUndelivered(string, string, time.Duration, int)  {}
func (noopBaseTriggerMetrics) EmitUndelivered2(string, string, time.Duration, int) {}

type BaseTriggerOpts struct {
	Metrics BaseTriggerMetrics
	// Emit undelivered metric after this duration since FirstAt. If 0, disabled.
	UndeliveredAfter time.Duration
	// Emit undelivered_2 metric after this duration since FirstAt. If 0, disabled.
	UndeliveredAfter2 time.Duration
}

type undeliveredState struct {
	emitted1 bool
	emitted2 bool
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
	undeliveredAfter       time.Duration
	undeliveredAfter2      time.Duration
	undeliveredAlertStates map[string]map[string]*undeliveredState // triggerID -> eventID -> flags
}

func NewBaseTriggerCapability[T proto.Message](
	store EventStore,
	newMsg func() T,
	lggr logger.Logger,
	capabilityId string,
	tRetransmit time.Duration,
	opts *BaseTriggerOpts,
) *BaseTriggerCapability[T] {
	ctx, cancel := context.WithCancel(context.Background())

	var metrics BaseTriggerMetrics = noopBaseTriggerMetrics{}
	var undeliveredAfter, undeliveredAfter2 time.Duration
	if opts != nil && opts.Metrics != nil {
		metrics = opts.Metrics
		undeliveredAfter = opts.UndeliveredAfter
		undeliveredAfter2 = opts.UndeliveredAfter2
	}

	return &BaseTriggerCapability[T]{
		store:                  store,
		newMsg:                 newMsg,
		lggr:                   lggr,
		capabilityId:           capabilityId,
		tRetransmit:            tRetransmit,
		metrics:                metrics,
		undeliveredAfter:       undeliveredAfter,
		undeliveredAfter2:      undeliveredAfter2,
		undeliveredAlertStates: make(map[string]map[string]*undeliveredState),
		mu:                     sync.Mutex{},
		inboxes:                make(map[string]chan<- TriggerAndId[T]),
		pending:                make(map[string]map[string]*PendingEvent),
		ctx:                    ctx,
		cancel:                 cancel,
	}
}

func (b *BaseTriggerCapability[T]) Start(ctx context.Context) error {
	b.lggr.Info("starting base trigger")

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
	b.inboxes[triggerID] = sendCh
	b.mu.Unlock()
}

func (b *BaseTriggerCapability[T]) UnregisterTrigger(triggerID string) {
	b.mu.Lock()
	delete(b.inboxes, triggerID)
	delete(b.pending, triggerID)
	delete(b.undeliveredAlertStates, triggerID)
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

func (b *BaseTriggerCapability[T]) AckEvent(ctx context.Context, triggerId string, eventId string) error {
	b.lggr.Infof("Event ACK (triggerID: %s, eventID %s)", triggerId, eventId)

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
		b.metrics.IncAck(triggerId, eventId, attempts)
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

			if b.undeliveredAfter == 0 && b.undeliveredAfter2 == 0 {
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

			if b.undeliveredAfter > 0 && !state.emitted1 && age >= b.undeliveredAfter {
				b.metrics.EmitUndelivered(triggerID, eventID, age, rec.Attempts)
				state.emitted1 = true
			}

			if b.undeliveredAfter2 > 0 && !state.emitted2 && age >= b.undeliveredAfter2 {
				b.metrics.EmitUndelivered2(triggerID, eventID, age, rec.Attempts)
				state.emitted2 = true
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
	b.lggr.Infof("resending event (triggerID: %s, eventID: %s)", event.TriggerId, event.EventId)
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

	typeURL := rec.AnyTypeURL
	payloadCopy := append([]byte(nil), rec.Payload...)
	sendCh, inboxOk := b.inboxes[event.TriggerId]
	attempts := rec.Attempts
	lastSent := rec.LastSentAt
	b.mu.Unlock()

	b.metrics.IncRetry(event.TriggerId, event.EventId)
	if err := b.store.UpdateDelivery(b.ctx, event.TriggerId, event.EventId, lastSent, attempts); err != nil {
		b.lggr.Errorf("failed to persist delivery update for trigger=%s event=%s: %v", event.TriggerId, event.EventId, err)
	}

	if !inboxOk {
		b.metrics.IncInboxMissing(event.TriggerId)
		b.lggr.Errorf("no inbox registered for trigger %s", event.TriggerId)
		return
	}

	msg := b.newMsg()
	if err := proto.Unmarshal(payloadCopy, msg); err != nil {
		b.lggr.Errorf("failed to unmarshal payload to message type (typeURL=%s): %v", typeURL, err)
		return
	}

	wrapped := TriggerAndId[T]{
		Trigger: msg,
		Id:      event.EventId,
	}

	select {
	case sendCh <- wrapped:
		b.lggr.Infof("event dispatched: capability =%s trigger=%s event=%s attempt=%d",
			b.capabilityId, event.TriggerId, event.EventId, attempts)
	default:
		b.metrics.IncInboxFull(event.TriggerId)
		b.lggr.Warnf("inbox full for trigger %s", event.TriggerId)
	}
}
