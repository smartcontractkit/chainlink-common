package capabilities

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

const (
	// backoffMultiplierCap limits the exponential backoff multiplier so retry
	// intervals don't grow unbounded. With a 30s base interval, this caps at
	// 30s * 10 = 5 minutes between retries.
	backoffMultiplierCap = 10

	// jitterFraction is the ±percentage applied to backoff intervals to
	// desynchronize retries across DON nodes and prevent P2P burst traffic.
	jitterFraction = 0.25

	// defaultMaxSendsPerTick limits how many events are sent per scanPending
	// cycle (initial deliveries AND retransmissions). This prevents a large
	// backlog from flooding P2P.
	defaultMaxSendsPerTick = 20

	// maxLoopTick caps how long the retransmit loop sleeps between scans.
	// With DeliverEvent queueing events instead of sending immediately,
	// a short tick is needed so first delivery latency stays low (~1s).
	maxLoopTick = time.Second
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
	// IncAckError counts ACK paths that return an error (e.g. store delete failure). reason is a stable identifier for dashboards.
	IncAckError(reason string)
	// IncAckMemoryOutcome records how an ACK related to the in-memory pending map: hit, miss_no_trigger_bucket, miss_no_event, miss_nil_record.
	IncAckMemoryOutcome(outcome string)
	// AddPendingEvents adjusts the live gauge of events awaiting ACK. Positive on insert, negative on ACK/unregister.
	AddPendingEvents(delta int64)
	// IncStuckEvent increments the live gauge of events stuck past the critical undelivered threshold.
	// Keyed by (capability_id, trigger_id, event_id) so you can see exactly which events are stuck.
	IncStuckEvent(triggerID, eventID string)
	// DecStuckEvent decrements the stuck-event gauge when a previously-critical event is ACKed or unregistered.
	DecStuckEvent(triggerID, eventID string)
	// IncStoppedResending increments the counter of events where the node exhausted max retries and stopped resending.
	IncStoppedResending(triggerID, eventID string, attempts int)
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
	// settings provides live CRE globals (BaseTriggerRetransmitEnabled, BaseTriggerRetryInterval).
	// When nil, tRetransmit > 0 enables persistence/retry with fixed spacing.
	settings settings.Getter

	mu      sync.Mutex
	inboxes map[string]chan<- TriggerAndId[T]   // triggerID --> registered send channel
	pending map[string]map[string]*PendingEvent // triggerID --> eventID --> PendingEvent

	// preAcked remembers ACKs that arrived before DeliverEvent was called on this node.
	// In a multi-node capability DON, other nodes may deliver the event first, causing
	// the workflow engine to ACK before the slower node has even persisted the event.
	// Without this cache, the late node would persist and retransmit forever since no
	// new ACK will arrive (other nodes already stopped retransmitting, so the subscriber
	// can't reach aggregation quorum).
	preAcked map[string]map[string]time.Time // triggerID -> eventID -> ackedAt

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
	settings settings.Getter,
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
		settings:               settings,
		metrics:                metrics,
		undeliveredWarning:     undeliveredWarning,
		undeliveredCritical:    undeliveredCritical,
		undeliveredAlertStates: make(map[string]map[string]*undeliveredState),
		mu:                     sync.Mutex{},
		inboxes:                make(map[string]chan<- TriggerAndId[T]),
		pending:                make(map[string]map[string]*PendingEvent),
		preAcked:               make(map[string]map[string]time.Time),
		ctx:                    ctx,
		cancel:                 cancel,
	}
}

// retransmitAllowed is true when events should be persisted and eligible for resend / ACK tracking.
func (b *BaseTriggerCapability[T]) retransmitAllowed(ctx context.Context) bool {
	if b.settings == nil {
		return b.tRetransmit > 0
	}
	enabled, err := cresettings.Default.BaseTriggerRetransmitEnabled.GetOrDefault(ctx, b.settings)
	if err != nil {
		b.lggr.Warnw("CRE settings read failed for BaseTriggerRetransmitEnabled; treating retransmit as disabled", "err", err)
		return false
	}
	return enabled
}

// retryInterval returns spacing between resend attempts. When settings is set, reads live CRE config.
func (b *BaseTriggerCapability[T]) retryInterval(ctx context.Context) time.Duration {
	if b.settings == nil {
		return b.tRetransmit
	}
	interval, err := cresettings.Default.BaseTriggerRetryInterval.GetOrDefault(ctx, b.settings)
	if err != nil {
		b.lggr.Warnw("CRE settings read failed for BaseTriggerRetryInterval; using schema default", "err", err)
		return cresettings.Default.BaseTriggerRetryInterval.DefaultValue
	}
	return interval
}

// maxRetries returns the configured maximum number of send attempts. 0 means unlimited.
func (b *BaseTriggerCapability[T]) maxRetries(ctx context.Context) int {
	defaultMaxRetries := 20
	if b.settings == nil {
		return defaultMaxRetries
	}
	v, err := cresettings.Default.BaseTriggerMaxRetries.GetOrDefault(ctx, b.settings)
	if err != nil {
		b.lggr.Warnw("CRE settings read failed for BaseTriggerMaxRetries; using default (20)", "err", err)
		return defaultMaxRetries
	}
	return v
}

// maxSendsPerTick returns how many events a single scanPending cycle may send.
// Tunable via CRE settings so operators can balance P2P load without code deploys.
func (b *BaseTriggerCapability[T]) maxSendsPerTick(ctx context.Context) int {
	if b.settings == nil {
		return defaultMaxSendsPerTick
	}
	v, err := cresettings.Default.BaseTriggerMaxSendsPerTick.GetOrDefault(ctx, b.settings)
	if err != nil {
		b.lggr.Warnw("CRE settings read failed for BaseTriggerMaxSendsPerTick; using default", "err", err)
		return defaultMaxSendsPerTick
	}
	if v <= 0 {
		return defaultMaxSendsPerTick
	}
	return v
}

// pruneAge returns how long a pending row must exist before the pruning loop may remove it.
func (b *BaseTriggerCapability[T]) pruneAge(ctx context.Context) time.Duration {
	if b.settings == nil {
		return 0
	}
	v, err := cresettings.Default.BaseTriggerPruneAge.GetOrDefault(ctx, b.settings)
	if err != nil {
		b.lggr.Warnw("CRE settings read failed for BaseTriggerPruneAge; treating as disabled", "err", err)
		return 0
	}
	return v
}

// loopTickDuration is recomputed before each loop wait so BaseTriggerRetryInterval and enablement
// changes take effect without restarting. Derived from half the retry interval, clamped to
// [1ms, maxLoopTick] so that newly queued events are picked up promptly.
func (b *BaseTriggerCapability[T]) loopTickDuration() time.Duration {
	d := time.Second
	if b.settings != nil {
		if iv := b.retryInterval(b.ctx); iv > 0 {
			d = iv / 2
		}
	} else if b.tRetransmit > 0 {
		d = b.tRetransmit / 2
	}
	if d < time.Millisecond {
		d = time.Millisecond
	}
	if d > maxLoopTick {
		d = maxLoopTick
	}
	return d
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

	if n := int64(len(recs)); n > 0 {
		b.metrics.AddPendingEvents(n)
	}

	b.wg.Add(2)
	go func() {
		defer b.wg.Done()
		b.retransmitLoop()
	}()
	go func() {
		defer b.wg.Done()
		b.pruneLoop()
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
	pendingCount := int64(len(b.pending[triggerID]))

	var criticalEvents []string
	if m, ok := b.undeliveredAlertStates[triggerID]; ok {
		for eventID, s := range m {
			if s != nil && s.emittedCritical {
				criticalEvents = append(criticalEvents, eventID)
			}
		}
	}

	delete(b.inboxes, triggerID)
	delete(b.pending, triggerID)
	delete(b.preAcked, triggerID)
	delete(b.undeliveredAlertStates, triggerID)
	b.mu.Unlock()

	for _, eventID := range criticalEvents {
		b.metrics.DecStuckEvent(triggerID, eventID)
	}

	if existed {
		b.metrics.DecActiveTriggers()
	}
	if pendingCount > 0 {
		b.metrics.AddPendingEvents(-pendingCount)
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
	if !b.retransmitAllowed(ctx) {
		b.lggr.Infow("base trigger retransmit not active")
		return b.sendToInbox(triggerID, te.ID, te.Payload.GetValue())
	}

	// Check if this event was already ACKed or is already pending.
	// Pre-ACK: other capability DON nodes deliver the event faster,
	// causing the workflow engine to ACK before this node calls DeliverEvent.
	// Already pending: the EVM trigger re-delivers after finalization
	// while the event is still awaiting ACK.
	b.mu.Lock()
	if pa, ok := b.preAcked[triggerID]; ok {
		if _, wasAcked := pa[te.ID]; wasAcked {
			delete(pa, te.ID)
			if len(pa) == 0 {
				delete(b.preAcked, triggerID)
			}
			b.mu.Unlock()
			b.lggr.Infow("base trigger DeliverEvent skipped: event was already ACKed (pre-ACK)",
				"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID)
			b.metrics.IncAckMemoryOutcome("pre_ack_delivery_skipped")
			return nil
		}
	}
	if pending, ok := b.pending[triggerID]; ok {
		if _, exists := pending[te.ID]; exists {
			b.mu.Unlock()
			b.lggr.Debugw("base trigger DeliverEvent skipped: event already pending",
				"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID)
			return nil
		}
	}
	b.mu.Unlock()

	rec := PendingEvent{
		TriggerId:  triggerID,
		EventId:    te.ID,
		AnyTypeURL: te.Payload.GetTypeUrl(),
		Payload:    te.Payload.GetValue(),
		FirstAt:    time.Now(),
	}

	if err := b.store.Insert(ctx, rec); err != nil {
		if isDuplicateKeyError(err) {
			b.lggr.Debugw("base trigger DeliverEvent: event already in store (re-delivery after give-up), skipping",
				"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID)
			return nil
		}
		b.lggr.Errorw("base trigger failed to persist pending event",
			"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID, "err", err)
		return err
	}
	b.lggr.Infow("base trigger persisted pending event for ACK tracking",
		"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID)

	// Double-check preAcked under the same lock as adding to pending.
	// An ACK may have arrived during the store.Insert call above. Without
	// this second check, the event would be retransmitted forever because
	// the first preAcked check (before Insert) narrowly missed the ACK.
	b.mu.Lock()
	if pa, ok := b.preAcked[triggerID]; ok {
		if _, wasAcked := pa[te.ID]; wasAcked {
			delete(pa, te.ID)
			if len(pa) == 0 {
				delete(b.preAcked, triggerID)
			}
			b.mu.Unlock()
			b.lggr.Infow("base trigger DeliverEvent skipped after persist: event was ACKed during store write (pre-ACK double-check)",
				"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID)
			b.metrics.IncAckMemoryOutcome("pre_ack_delivery_skipped")
			if err := b.store.DeleteEvent(ctx, triggerID, te.ID); err != nil {
				b.lggr.Errorw("base trigger failed to delete pre-ACKed event from store",
					"capabilityID", b.capabilityId, "triggerID", triggerID, "eventID", te.ID, "err", err)
			}
			return nil
		}
	}
	if b.pending[triggerID] == nil {
		b.pending[triggerID] = map[string]*PendingEvent{}
	}
	b.pending[triggerID][te.ID] = &rec
	b.mu.Unlock()

	b.metrics.AddPendingEvents(1)
	// The event is now queued in pending. scanPending will pick it up on the
	// next tick (≤1s) and send it, subject to the per-tick cap. This prevents
	// a burst of 2,500+ P2P sends when many registrations match one on-chain
	// event. First delivery latency increases by at most maxLoopTick.
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

	var (
		attempts            int
		firstAt             time.Time
		found               bool
		hadTriggerBucket    bool
		hadEventKey         bool
		hadNilPendingRecord bool
	)

	b.mu.Lock()
	eventsForTrigger, ok := b.pending[triggerId]
	hadTriggerBucket = ok && eventsForTrigger != nil
	if hadTriggerBucket {
		rec, recOk := eventsForTrigger[eventId]
		hadEventKey = recOk
		switch {
		case recOk && rec != nil:
			attempts = rec.Attempts
			firstAt = rec.FirstAt
			found = true
		case recOk && rec == nil:
			hadNilPendingRecord = true
			b.metrics.IncAckMemoryOutcome("miss_nil_record")
		default:
			b.metrics.IncAckMemoryOutcome("miss_no_event")
		}

		delete(eventsForTrigger, eventId)
		if len(eventsForTrigger) == 0 {
			delete(b.pending, triggerId)
		}
	} else {
		b.metrics.IncAckMemoryOutcome("miss_no_trigger_bucket")
	}

	// Always record the ACK so that a later DeliverEvent for this event is
	// skipped. This covers two scenarios:
	//   1. (pre-ACK) Other DON nodes delivered the event first, so the ACK
	//      arrives at this node before DeliverEvent is called.
	//   2. (re-delivery) The upstream trigger re-delivers the same event
	//      (e.g. EVM trigger after block finalization prunes its sent-set),
	//      even though the event was already ACKed during a prior delivery.
	if b.preAcked[triggerId] == nil {
		b.preAcked[triggerId] = make(map[string]time.Time)
	}
	b.preAcked[triggerId][eventId] = time.Now()

	var wasCritical bool
	if m, ok := b.undeliveredAlertStates[triggerId]; ok {
		if s, exists := m[eventId]; exists && s != nil && s.emittedCritical {
			wasCritical = true
		}
		delete(m, eventId)
		if len(m) == 0 {
			delete(b.undeliveredAlertStates, triggerId)
		}
	}
	b.mu.Unlock()

	if wasCritical {
		b.metrics.DecStuckEvent(triggerId, eventId)
	}

	switch {
	case found:
		b.lggr.Infow("base trigger ACK matched in-memory pending event",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId,
			"attempts", attempts, "firstAt", firstAt)
		b.metrics.IncAckMemoryOutcome("hit")
		b.metrics.IncAck(triggerId, eventId)
		b.metrics.ObserveTimeToAck(triggerId, eventId, time.Since(firstAt), attempts)
		b.metrics.AddPendingEvents(-1)
	case hadNilPendingRecord:
		b.lggr.Warnw("base trigger ACK: pending map had nil record for event (treating as miss; reconciling store)",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId)
	case hadTriggerBucket && !hadEventKey:
		b.lggr.Infow("base trigger ACK: event id not in in-memory pending map for trigger (may exist only in store; reconciling)",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId)
	case !hadTriggerBucket:
		b.lggr.Infow("base trigger ACK: no in-memory pending bucket for trigger (not pending here; still deleting from store if row exists)",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId)
	}

	if err := b.store.DeleteEvent(ctx, triggerId, eventId); err != nil {
		b.lggr.Errorw("base trigger ACK failed to delete event from store",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId,
			"foundInMemory", found, "err", err)
		b.metrics.IncAckError("store_delete_failed")
		return err
	}
	if found {
		b.lggr.Debugw("base trigger ACK store delete succeeded",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId)
	} else {
		b.lggr.Infow("base trigger ACK store delete succeeded (memory miss path; store row removed if present)",
			"capabilityID", b.capabilityId, "triggerID", triggerId, "eventID", eventId)
	}
	return nil
}

func (b *BaseTriggerCapability[T]) retransmitLoop() {
	for {
		timer := time.NewTimer(b.loopTickDuration())
		select {
		case <-b.ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
			b.lggr.Debug("retransmitting unacknowledged events")
			b.scanPending()
		}
	}
}

type stoppedResendingEvent struct {
	triggerID   string
	eventID     string
	attempts    int
	wasCritical bool
}

// reachedMaxRetries returns true when the event has exhausted its allowed send attempts.
func reachedMaxRetries(attempts, maxRetries int) bool {
	return maxRetries > 0 && attempts >= maxRetries
}

// retryBackoff computes an exponential backoff interval: baseInterval * 2^(attempts-1),
// capped at baseInterval * backoffMultiplierCap. The first retry (attempts=1) uses
// isDuplicateKeyError returns true if err is a PostgreSQL unique constraint
// violation (SQLSTATE 23505). The store is accessed via gRPC, so we rely on
// the error message text rather than typed errors.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}

// baseInterval unchanged so well-behaved events see no regression.
func retryBackoff(baseInterval time.Duration, attempts int) time.Duration {
	if attempts <= 1 {
		return baseInterval
	}
	shift := uint(attempts - 1)
	multiplier := int64(1) << shift
	if multiplier > backoffMultiplierCap || multiplier <= 0 {
		multiplier = backoffMultiplierCap
	}
	return baseInterval * time.Duration(multiplier)
}

// addJitter applies ±jitterFraction random noise to d so that retries from
// multiple nodes don't collide on the same P2P window.
func addJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	// rand.Float64 is concurrency-safe and auto-seeded since Go 1.20.
	noise := time.Duration(float64(d) * jitterFraction * (2*rand.Float64() - 1))
	return d + noise
}

func (b *BaseTriggerCapability[T]) scanPending() {
	now := time.Now()
	ctx := b.ctx

	interval := b.retryInterval(ctx)
	if !b.retransmitAllowed(ctx) || interval <= 0 {
		return
	}

	maxRetries := b.maxRetries(ctx)
	warnThreshold := 1 * interval
	critThreshold := 3 * interval

	b.mu.Lock()

	b.expirePreAcked(now, maxRetries, interval)

	toResend := make([]PendingEvent, 0, len(b.pending))
	var toStop []stoppedResendingEvent
	for triggerID, pendingForTrigger := range b.pending {
		if _, hasInbox := b.inboxes[triggerID]; !hasInbox {
			continue // registration hasn't arrived yet — skip to avoid wasting retry attempts
		}
		for eventID, rec := range pendingForTrigger {
			if reachedMaxRetries(rec.Attempts, maxRetries) {
				toStop = append(toStop, b.collectStoppedResending(triggerID, eventID, rec, pendingForTrigger))
				continue
			}
			b.collectResendCandidate(rec, now, interval, warnThreshold, critThreshold, &toResend)
		}
	}
	b.mu.Unlock()

	for _, ev := range toStop {
		b.emitStoppedResending(ctx, ev, maxRetries)
	}

	// Sort deterministically so all DON nodes select the same events when
	// the per-tick cap applies. Without this, Go map iteration randomness
	// causes each node to pick a different subset, preventing the subscriber
	// from reaching F+1 aggregation quorum. Unsent events (Attempts==0) are
	// prioritized so fresh events reach quorum on their first tick.
	sort.Slice(toResend, func(i, j int) bool {
		iNew := toResend[i].Attempts == 0
		jNew := toResend[j].Attempts == 0
		if iNew != jNew {
			return iNew
		}
		if toResend[i].EventId != toResend[j].EventId {
			return toResend[i].EventId < toResend[j].EventId
		}
		return toResend[i].TriggerId < toResend[j].TriggerId
	})

	cap := b.maxSendsPerTick(ctx)
	if len(toResend) > cap {
		b.lggr.Warnw("base trigger capping sends per tick",
			"capabilityID", b.capabilityId,
			"eligible", len(toResend), "cap", cap)
		toResend = toResend[:cap]
	}

	for _, event := range toResend {
		b.trySend(event)
	}
}

// expirePreAcked removes old preAcked entries so the cache doesn't grow unbounded.
// Uses the full retry window as TTL — a DeliverEvent arriving later than that would
// have been stopped anyway. Must be called under b.mu.
func (b *BaseTriggerCapability[T]) expirePreAcked(now time.Time, maxRetries int, interval time.Duration) {
	preAckTTL := time.Duration(maxRetries) * interval
	if maxRetries == 0 || preAckTTL <= 0 {
		preAckTTL = 10 * time.Minute
	}
	for triggerID, events := range b.preAcked {
		for eventID, ackedAt := range events {
			if now.Sub(ackedAt) > preAckTTL {
				delete(events, eventID)
			}
		}
		if len(events) == 0 {
			delete(b.preAcked, triggerID)
		}
	}
}

// collectStoppedResending removes the event from pending and alert tracking, returning
// the metadata needed to emit metrics/logs outside the lock. Must be called under b.mu.
func (b *BaseTriggerCapability[T]) collectStoppedResending(
	triggerID, eventID string,
	rec *PendingEvent,
	pendingForTrigger map[string]*PendingEvent,
) stoppedResendingEvent {
	wasCritical := false
	if m, ok := b.undeliveredAlertStates[triggerID]; ok {
		if s, exists := m[eventID]; exists && s != nil && s.emittedCritical {
			wasCritical = true
		}
		delete(m, eventID)
		if len(m) == 0 {
			delete(b.undeliveredAlertStates, triggerID)
		}
	}
	delete(pendingForTrigger, eventID)
	if len(pendingForTrigger) == 0 {
		delete(b.pending, triggerID)
	}
	return stoppedResendingEvent{
		triggerID:   triggerID,
		eventID:     eventID,
		attempts:    rec.Attempts,
		wasCritical: wasCritical,
	}
}

// collectResendCandidate appends the event to toResend if the retry backoff has elapsed,
// and checks undelivered alert thresholds. Must be called under b.mu.
func (b *BaseTriggerCapability[T]) collectResendCandidate(
	rec *PendingEvent,
	now time.Time,
	interval, warnThreshold, critThreshold time.Duration,
	toResend *[]PendingEvent,
) {
	backoff := addJitter(retryBackoff(interval, rec.Attempts))
	if rec.LastSentAt.IsZero() || now.Sub(rec.LastSentAt) >= backoff {
		*toResend = append(*toResend, PendingEvent{
			TriggerId: rec.TriggerId,
			EventId:   rec.EventId,
			Attempts:  rec.Attempts,
		})
	}

	if warnThreshold == 0 && critThreshold == 0 {
		return
	}
	age := now.Sub(rec.FirstAt)

	triggerID, eventID := rec.TriggerId, rec.EventId
	if b.undeliveredAlertStates[triggerID] == nil {
		b.undeliveredAlertStates[triggerID] = make(map[string]*undeliveredState)
	}

	state := b.undeliveredAlertStates[triggerID][eventID]
	if state == nil {
		state = &undeliveredState{}
		b.undeliveredAlertStates[triggerID][eventID] = state
	}

	if warnThreshold > 0 && !state.emittedWarning && age >= warnThreshold {
		b.metrics.EmitUndeliveredWarning(triggerID, eventID)
		state.emittedWarning = true
	}

	if critThreshold > 0 && !state.emittedCritical && age >= critThreshold {
		b.metrics.EmitUndeliveredCritical(triggerID, eventID)
		b.metrics.IncStuckEvent(triggerID, eventID)
		state.emittedCritical = true
	}
}

// emitStoppedResending logs and emits metrics for an event that exhausted its retries.
// The event is NOT deleted from the store here — the prune loop handles cleanup,
// giving operators time to investigate unrecoverable payloads (e.g. HTTP triggers).
func (b *BaseTriggerCapability[T]) emitStoppedResending(ctx context.Context, ev stoppedResendingEvent, maxRetries int) {
	b.lggr.Errorw("base trigger stopped resending event after max retries",
		"capabilityID", b.capabilityId, "triggerID", ev.triggerID, "eventID", ev.eventID,
		"attempts", ev.attempts, "maxRetries", maxRetries,
		"reason", "max_retries_exhausted")
	b.metrics.IncStoppedResending(ev.triggerID, ev.eventID, ev.attempts)
	b.metrics.AddPendingEvents(-1)
	if ev.wasCritical {
		b.metrics.DecStuckEvent(ev.triggerID, ev.eventID)
	}
}

// pruneLoop periodically removes old rows from the event store that are no longer tracked in memory.
// This catches rows orphaned by crashes, missed deletes, or events that were given up on before
// this code was deployed. It uses store.List + age comparison + store.DeleteEvent,
// avoiding any new EventStore interface methods.
func (b *BaseTriggerCapability[T]) pruneLoop() {
	const minPruneInterval = time.Minute
	for {
		age := b.pruneAge(b.ctx)
		if age <= 0 {
			age = 24 * time.Hour
		}
		tick := age / 4
		if tick < minPruneInterval {
			tick = minPruneInterval
		}

		timer := time.NewTimer(tick)
		select {
		case <-b.ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
			b.pruneStaleEvents()
		}
	}
}

func (b *BaseTriggerCapability[T]) pruneStaleEvents() {
	age := b.pruneAge(b.ctx)
	if age <= 0 {
		return
	}
	cutoff := time.Now().Add(-age)

	recs, err := b.store.List(b.ctx)
	if err != nil {
		b.lggr.Errorw("prune: failed to list events from store", "capabilityID", b.capabilityId, "err", err)
		return
	}

	for _, rec := range recs {
		// Use the most recent activity timestamp so events still being
		// retried aren't pruned prematurely.
		lastActivity := rec.FirstAt
		if rec.LastSentAt.After(lastActivity) {
			lastActivity = rec.LastSentAt
		}
		if lastActivity.After(cutoff) {
			continue
		}

		b.mu.Lock()
		inMemory := false
		if evts, ok := b.pending[rec.TriggerId]; ok {
			_, inMemory = evts[rec.EventId]
		}
		b.mu.Unlock()

		if inMemory {
			// An event old enough to prune should not still be in memory —
			// scanPending should have stopped resending it or an ACK should
			// have removed it. Log a warning so we can investigate.
			b.lggr.Warnw("prune: stale event still tracked in memory (possible inconsistency; skipping)",
				"capabilityID", b.capabilityId, "triggerID", rec.TriggerId, "eventID", rec.EventId,
				"firstAt", rec.FirstAt, "lastSentAt", rec.LastSentAt, "attempts", rec.Attempts, "pruneAge", age)
			continue
		}

		b.lggr.Infow("prune: removing stale event from store",
			"capabilityID", b.capabilityId, "triggerID", rec.TriggerId, "eventID", rec.EventId,
			"firstAt", rec.FirstAt, "lastSentAt", rec.LastSentAt, "attempts", rec.Attempts, "pruneAge", age)
		if err := b.store.DeleteEvent(b.ctx, rec.TriggerId, rec.EventId); err != nil {
			b.lggr.Errorw("prune: failed to delete stale event",
				"capabilityID", b.capabilityId, "triggerID", rec.TriggerId, "eventID", rec.EventId, "err", err)
		}
	}
}

// trySend attempts a delivery for the given event.
// It updates Attempts and LastSentAt on every attempt locally. Success is determined
// later by an AckEvent call.
func (b *BaseTriggerCapability[T]) trySend(event PendingEvent) {
	if !b.retransmitAllowed(b.ctx) {
		return
	}
	if b.retryInterval(b.ctx) <= 0 {
		return
	}

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
