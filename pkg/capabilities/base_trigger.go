package capabilities

import (
	"context"
	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
	"sync"
	"time"
)

type PendingEvent struct {
	TriggerId  string
	WorkflowId string
	EventId    string
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

type OutboundSend func(ctx context.Context, ev TriggerEvent, workflowId string) error
type LostHook func(ctx context.Context, rec PendingEvent)

// TODO Implement BaseTriggerCapability - CRE-1523
type BaseTriggerCapability struct {
	/*
	 Keeps track of workflow registrations (similar to LLO streams trigger).
	 Handles retransmits based on T_retransmit and T_max.
	 Persists pending events in the DB to be resilient to node restarts.
	*/
	tRetransmit time.Duration // time window for an event being ACKd before we retransmit
	tMax        time.Duration // timeout before events are considered lost if not ACKd

	store EventStore
	send  OutboundSend
	lost  LostHook
	lggr  *log.Logger

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

func (b *BaseTriggerCapability) deliverEvent(
	ctx context.Context,
	ev TriggerEvent,
	workflowIds []string,
) error {
	/*
	 Base Trigger Capability can interact with the Don2Don layer (in the remote capability setting)
	 as well as directly with a consumer (in the local setting).
	*/
	now := time.Now()

	for _, workflowId := range workflowIds {
		rec := PendingEvent{
			TriggerId:  ev.TriggerId,
			WorkflowId: workflowId,
			EventId:    ev.EventId,
			Payload:    ev.Payload,
			FirstAt:    now,
		}

		if err := b.store.Insert(ctx, rec); err != nil {
			return err
		}

		b.mu.Lock()
		b.pending[key(ev.TriggerId, workflowId, ev.EventId)] = &rec
		b.mu.Unlock()

		_ = b.trySend(ctx, ev.TriggerId, workflowId, ev.EventId)
	}
	return nil // only when the event is successfully persisted and ready to be relaibly delivered
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
