package beholder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// DurableEmitterConfig configures the DurableEmitter behaviour.
type DurableEmitterConfig struct {
	// RetransmitInterval controls how often the retransmit loop ticks.
	RetransmitInterval time.Duration
	// RetransmitAfter is the minimum age of an event before the retransmit
	// loop considers it. This gives the immediate-publish path time to succeed.
	RetransmitAfter time.Duration
	// RetransmitBatchSize caps the number of events sent per retransmit cycle.
	RetransmitBatchSize int
	// ExpiryInterval controls how often the expiry loop ticks.
	ExpiryInterval time.Duration
	// EventTTL is the maximum age of an event before it is expired.
	EventTTL time.Duration
	// PublishTimeout is the per-RPC deadline for Publish / PublishBatch calls.
	PublishTimeout time.Duration
}

func DefaultDurableEmitterConfig() DurableEmitterConfig {
	return DurableEmitterConfig{
		RetransmitInterval:  5 * time.Second,
		RetransmitAfter:     10 * time.Second,
		RetransmitBatchSize: 100,
		ExpiryInterval:      1 * time.Minute,
		EventTTL:            24 * time.Hour,
		PublishTimeout:      5 * time.Second,
	}
}

// DurableEmitter implements Emitter with persistence-backed delivery guarantees.
//
// On Emit the event is serialized and written to a DurableEventStore. Once the
// insert succeeds Emit returns nil — the caller has a durable guarantee. An
// immediate async Publish is attempted; on success the record is deleted. If
// that fails a background retransmit loop will pick the event up and retry via
// PublishBatch.
//
// A separate expiry loop garbage-collects events older than EventTTL to bound
// table growth.
type DurableEmitter struct {
	store  DurableEventStore
	client chipingress.Client
	cfg    DurableEmitterConfig
	log    logger.Logger

	stopCh chan struct{}
	wg     sync.WaitGroup
}

var _ Emitter = (*DurableEmitter)(nil)

func NewDurableEmitter(
	store DurableEventStore,
	client chipingress.Client,
	cfg DurableEmitterConfig,
	log logger.Logger,
) (*DurableEmitter, error) {
	if store == nil {
		return nil, fmt.Errorf("durable event store is nil")
	}
	if client == nil {
		return nil, fmt.Errorf("chipingress client is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	return &DurableEmitter{
		store:  store,
		client: client,
		cfg:    cfg,
		log:    log,
		stopCh: make(chan struct{}),
	}, nil
}

// Start launches the retransmit and expiry background loops.
// Cancel the supplied context or call Close to stop them.
func (d *DurableEmitter) Start(ctx context.Context) {
	d.wg.Add(2)
	go d.retransmitLoop(ctx)
	go d.expiryLoop(ctx)
}

// Emit persists the event then attempts async delivery.
// Returns nil once the store insert succeeds.
func (d *DurableEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		return err
	}

	event, err := chipingress.NewEvent(sourceDomain, entityType, body, newAttributes(attrKVs...))
	if err != nil {
		return err
	}

	eventPb, err := chipingress.EventToProto(event)
	if err != nil {
		return fmt.Errorf("failed to convert event to proto: %w", err)
	}

	payload, err := proto.Marshal(eventPb)
	if err != nil {
		return fmt.Errorf("failed to marshal event proto: %w", err)
	}

	id, err := d.store.Insert(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to persist event: %w", err)
	}

	// Fire-and-forget immediate delivery attempt.
	go d.publishAndDelete(id, eventPb)

	return nil
}

// Close signals background loops to stop and waits for them to finish.
func (d *DurableEmitter) Close() error {
	close(d.stopCh)
	d.wg.Wait()
	return nil
}

// publishAndDelete attempts a single Publish and deletes the record on success.
func (d *DurableEmitter) publishAndDelete(id int64, eventPb *chipingress.CloudEventPb) {
	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.PublishTimeout)
	defer cancel()

	if _, err := d.client.Publish(ctx, eventPb); err != nil {
		d.log.Debugw("immediate publish failed, retransmit loop will retry",
			"id", id, "error", err)
		return
	}

	if err := d.store.Delete(context.Background(), id); err != nil {
		d.log.Errorw("failed to delete delivered event", "id", id, "error", err)
	}
}

func (d *DurableEmitter) retransmitLoop(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(d.cfg.RetransmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.retransmitPending(ctx)
		}
	}
}

func (d *DurableEmitter) retransmitPending(ctx context.Context) {
	cutoff := time.Now().Add(-d.cfg.RetransmitAfter)
	pending, err := d.store.ListPending(ctx, cutoff, d.cfg.RetransmitBatchSize)
	if err != nil {
		d.log.Errorw("failed to list pending events", "error", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	events := make([]*chipingress.CloudEventPb, 0, len(pending))
	ids := make([]int64, 0, len(pending))

	for _, pe := range pending {
		var eventPb chipingress.CloudEventPb
		if err := proto.Unmarshal(pe.Payload, &eventPb); err != nil {
			d.log.Errorw("corrupt pending event, deleting", "id", pe.ID, "error", err)
			_ = d.store.Delete(ctx, pe.ID)
			continue
		}
		events = append(events, &eventPb)
		ids = append(ids, pe.ID)
	}
	if len(events) == 0 {
		return
	}

	publishCtx, cancel := context.WithTimeout(ctx, d.cfg.PublishTimeout)
	defer cancel()

	if _, err := d.client.PublishBatch(publishCtx, &chipingress.CloudEventBatch{Events: events}); err != nil {
		d.log.Warnw("retransmit batch failed", "count", len(events), "error", err)
		return
	}

	for _, id := range ids {
		if err := d.store.Delete(ctx, id); err != nil {
			d.log.Errorw("failed to delete retransmitted event", "id", id, "error", err)
		}
	}
	d.log.Debugw("retransmitted events", "count", len(ids))
}

func (d *DurableEmitter) expiryLoop(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(d.cfg.ExpiryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			deleted, err := d.store.DeleteExpired(ctx, d.cfg.EventTTL)
			if err != nil {
				d.log.Errorw("failed to delete expired events", "error", err)
				continue
			}
			if deleted > 0 {
				d.log.Infow("purged expired events", "count", deleted)
			}
		}
	}
}
