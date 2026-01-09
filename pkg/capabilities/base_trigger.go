package capabilities

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type PendingEvent struct {
	TriggerId  string
	WorkflowId string
	EventId    string
	AnyTypeURL string // Payload type
	Payload    []byte
	FirstAt    time.Time
	LastSentAt time.Time
	Attempts   int
}

type EventStore interface {
	Insert(ctx context.Context, rec PendingEvent) error
	Delete(ctx context.Context, triggerId, workflowId, eventId string) error
	List(ctx context.Context) ([]PendingEvent, error)
}

type OutboundSend func(ctx context.Context, te TriggerEvent, workflowId string) error
type LostHook func(ctx context.Context, rec PendingEvent)

// key builds the composite lookup key used in b.pending
func key(triggerId, workflowId, eventId string) string {
	return triggerId + "|" + workflowId + "|" + eventId
}

type BaseTriggerCapability struct {
	/*
	 Keeps track of workflow registrations (similar to LLO streams trigger).
	 Handles retransmits based on T_retransmit and T_max.
	 Persists pending events in the DB to be resilient to node restarts.
	*/
	// TODO: We will want these to be configurable per chain
	tRetransmit time.Duration // time window for an event being ACKd before we retransmit
	tMax        time.Duration // timeout before events are considered lost if not ACKd

	store EventStore
	send  OutboundSend
	lost  LostHook
	lggr  logger.Logger

	mu      sync.Mutex
	pending map[string]*PendingEvent // key(triggerID|workflowID|eventID)

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (b *BaseTriggerCapability) Start(ctx context.Context) error {
	b.ctx, b.cancel = context.WithCancel(ctx)

	recs, err := b.store.List(ctx)
	if err != nil {
		return err
	}

	// Initialize in-memory persistence
	b.pending = make(map[string]*PendingEvent)
	for i := range recs {
		r := recs[i]
		b.pending[key(r.TriggerId, r.WorkflowId, r.EventId)] = &r
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.retransmitLoop()
	}()

	for _, r := range recs {
		_ = b.trySend(ctx, r.TriggerId, r.WorkflowId, r.EventId)
	}
	return nil
}

func (b *BaseTriggerCapability) DeliverEvent(
	ctx context.Context,
	te TriggerEvent,
	workflowIds []string,
) error {
	/*
	 Base Trigger Capability can interact with the Don2Don layer (in the remote capability setting)
	 as well as directly with a consumer (in the local setting).
	*/
	now := time.Now()

	for _, workflowId := range workflowIds {
		rec := PendingEvent{
			TriggerId:  te.TriggerType,
			WorkflowId: workflowId,
			EventId:    te.ID,
			AnyTypeURL: te.Payload.GetTypeUrl(),
			Payload:    te.Payload.GetValue(),
			FirstAt:    now,
		}

		if err := b.store.Insert(ctx, rec); err != nil {
			return err
		}

		b.mu.Lock()
		b.pending[key(te.TriggerType, workflowId, te.ID)] = &rec
		b.mu.Unlock()

		_ = b.trySend(ctx, te.TriggerType, workflowId, te.ID)
	}
	return nil // only when the event is successfully persisted and ready to be reliably delivered
}

func (b *BaseTriggerCapability) AckEvent(
	ctx context.Context,
	triggerId, workflowId, eventId string,
) error {
	k := key(triggerId, workflowId, eventId) // NOTE: WorkflowID we want to start ;P

	b.mu.Lock()
	delete(b.pending, k)
	b.mu.Unlock()

	return b.store.Delete(ctx, triggerId, workflowId, eventId)
}

func (b *BaseTriggerCapability) retransmitLoop() {
	ticker := time.NewTicker(b.tRetransmit / 2)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.scanPending()
		}
	}
}

func (b *BaseTriggerCapability) scanPending() {
	now := time.Now()

	for _, rec := range b.pending {
		if now.Sub(rec.FirstAt) >= b.tMax {
			_ = b.AckEvent(b.ctx, rec.TriggerId, rec.WorkflowId, rec.EventId)
			b.lost(b.ctx, *rec)
			continue
		}
		if rec.LastSentAt.IsZero() || now.Sub(rec.LastSentAt) >= b.tRetransmit {
			_ = b.trySend(b.ctx, rec.TriggerId, rec.WorkflowId, rec.EventId)
		}
	}
}

// trySend attempts a delivery for the given (triggerId, workflowId, eventId).
// It updates Attempts and LastSentAt on every attempt. Success is determined
// by a later AckEvent; this method does NOT remove the record from memory/DB.
func (b *BaseTriggerCapability) trySend(ctx context.Context, triggerId, workflowId, eventId string) error {
	k := key(triggerId, workflowId, eventId)

	b.mu.Lock()
	rec, ok := b.pending[k]
	if !ok || rec == nil {
		b.mu.Unlock()
		return nil
	}
	rec.Attempts++
	rec.LastSentAt = time.Now()

	anyPayload := &anypb.Any{
		TypeUrl: rec.AnyTypeURL,
		Value:   append([]byte(nil), rec.Payload...),
	}

	te := TriggerEvent{
		TriggerType: triggerId,
		ID:          eventId,
		Payload:     anyPayload,
	}
	b.mu.Unlock()

	if err := b.send(ctx, te, workflowId); err != nil {
		if b.lggr != nil {
			b.lggr.Errorf("trySend failed: trigger=%s workflow=%s event=%s attempt=%d err=%v",
				triggerId, workflowId, eventId, rec.Attempts, err)
		}
		return err
	}
	if b.lggr != nil {
		b.lggr.Debugf("trySend dispatched: trigger=%s workflow=%s event=%s attempt=%d",
			triggerId, workflowId, eventId, rec.Attempts)
	}
	return nil
}
