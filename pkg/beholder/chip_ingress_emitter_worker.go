package beholder

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// emitterPayload holds a single event to be batched and sent via chip ingress.
type emitterPayload struct {
	body       []byte
	attributes map[string]any
	domain     string
	entity     string
}

// chipIngressEmitterWorker buffers events for a single (domain, entity) pair
// and flushes them via PublishBatch on a periodic interval.
type chipIngressEmitterWorker struct {
	client       chipingress.Client
	ch           chan emitterPayload
	domain       string
	entity       string
	maxBatchSize uint
	sendTimeout  time.Duration
	lggr         logger.Logger
	dropCount    atomic.Uint32
}

func newChipIngressEmitterWorker(
	client chipingress.Client,
	ch chan emitterPayload,
	domain string,
	entity string,
	maxBatchSize uint,
	sendTimeout time.Duration,
	lggr logger.Logger,
) *chipIngressEmitterWorker {
	return &chipIngressEmitterWorker{
		client:       client,
		ch:           ch,
		domain:       domain,
		entity:       entity,
		maxBatchSize: maxBatchSize,
		sendTimeout:  sendTimeout,
		lggr:         logger.Named(lggr, "ChipIngressEmitterWorker"),
	}
}

// Send drains the channel and sends a batch. Called periodically by GoTick.
func (w *chipIngressEmitterWorker) Send(ctx context.Context) {
	if len(w.ch) == 0 {
		return
	}

	batch := w.buildBatch()
	if batch == nil || len(batch.Events) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, w.sendTimeout)
	defer cancel()

	_, err := w.client.PublishBatch(ctx, batch)
	if err != nil {
		w.lggr.Warnf("could not send batch via chip ingress: %v", err)
		return
	}
}

// buildBatch drains the channel up to maxBatchSize and converts payloads to a CloudEventBatch.
func (w *chipIngressEmitterWorker) buildBatch() *chipingress.CloudEventBatch {
	var events []chipingress.CloudEvent

	for len(w.ch) > 0 && len(events) < int(w.maxBatchSize) { // #nosec G115
		payload := <-w.ch
		event, err := w.payloadToEvent(payload)
		if err != nil {
			w.lggr.Warnf("failed to build CloudEvent, dropping: %v", err)
			continue
		}
		events = append(events, event)
	}

	if len(events) == 0 {
		return nil
	}

	batch, err := chipingress.EventsToBatch(events)
	if err != nil {
		w.lggr.Warnf("failed to convert events to batch: %v", err)
		return nil
	}

	return batch
}

func (w *chipIngressEmitterWorker) payloadToEvent(payload emitterPayload) (chipingress.CloudEvent, error) {
	event, err := chipingress.NewEvent(payload.domain, payload.entity, payload.body, payload.attributes)
	if err != nil {
		return chipingress.CloudEvent{}, fmt.Errorf("failed to create CloudEvent: %w", err)
	}
	return event, nil
}

// logBufferFullWithExpBackoff logs at 1, 2, 4, 8, 16, 32, 64, 100, 200, 300, ...
// to avoid flooding logs when the buffer is persistently full.
func (w *chipIngressEmitterWorker) logBufferFullWithExpBackoff(payload emitterPayload) {
	count := w.dropCount.Add(1)
	if count > 0 && (count%100 == 0 || count&(count-1) == 0) {
		w.lggr.Warnf("chip ingress emitter buffer full, dropping event (domain=%s, entity=%s, droppedCount=%d)",
			payload.domain, payload.entity, count)
	}
}
