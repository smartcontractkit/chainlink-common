package beholder

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

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
// Transport-level retries (UNAVAILABLE, RESOURCE_EXHAUSTED) are handled by the
// gRPC client's built-in retry policy; no application-level retry is needed.
type chipIngressEmitterWorker struct {
	client       chipingress.Client
	ch           chan emitterPayload
	domain       string
	entity       string
	maxBatchSize uint
	sendTimeout  time.Duration
	lggr         logger.Logger
	dropCount    atomic.Uint32

	metrics     batchEmitterMetrics
	metricAttrs otelmetric.MeasurementOption
}

func newChipIngressEmitterWorker(
	client chipingress.Client,
	ch chan emitterPayload,
	domain string,
	entity string,
	maxBatchSize uint,
	sendTimeout time.Duration,
	metrics batchEmitterMetrics,
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
		metrics:      metrics,
		metricAttrs: otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("domain", domain),
			attribute.String("entity", entity),
		)),
	}
}

// Send drains the channel and sends a batch.
// Called periodically by the tick loop.
func (w *chipIngressEmitterWorker) Send(ctx context.Context) {
	if len(w.ch) == 0 {
		return
	}

	batch := w.buildBatch()
	if batch == nil || len(batch.Events) == 0 {
		return
	}

	w.publishOnce(ctx, batch)
}

func (w *chipIngressEmitterWorker) publishOnce(ctx context.Context, batch *chipingress.CloudEventBatch) {
	sendCtx, cancel := context.WithTimeout(ctx, w.sendTimeout)
	defer cancel()

	batchSize := int64(len(batch.Events))
	_, err := w.client.PublishBatch(sendCtx, batch)
	if err != nil {
		w.lggr.Warnf("could not send batch via chip ingress (domain=%s, entity=%s): %v",
			w.domain, w.entity, err)
		w.metrics.batchFailures.Add(context.Background(), 1, w.metricAttrs)
		w.metrics.eventsDropped.Add(context.Background(), batchSize, w.metricAttrs)
		return
	}
	w.metrics.eventsSent.Add(context.Background(), batchSize, w.metricAttrs)
}

// buildBatch drains the channel up to maxBatchSize and converts payloads to a CloudEventBatch.
func (w *chipIngressEmitterWorker) buildBatch() *chipingress.CloudEventBatch {
	var events []chipingress.CloudEvent

	max := int(w.maxBatchSize) // #nosec G115
drain:
	for len(events) < max {
		select {
		case payload := <-w.ch:
			event, err := w.payloadToEvent(payload)
			if err != nil {
				w.lggr.Warnf("failed to build CloudEvent, dropping: %v", err)
				w.metrics.eventsDropped.Add(context.Background(), 1, w.metricAttrs)
				continue
			}
			events = append(events, event)
		default:
			break drain
		}
	}

	if len(events) == 0 {
		return nil
	}

	batch, err := chipingress.EventsToBatch(events)
	if err != nil {
		w.lggr.Warnf("failed to convert events to batch: %v", err)
		w.metrics.eventsDropped.Add(context.Background(), int64(len(events)), w.metricAttrs)
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

// drain flushes all remaining buffered events before shutdown.
// Uses a fresh context with the given timeout (independent of the cancelled parent).
// Continues attempting subsequent batches even if one fails.
func (w *chipIngressEmitterWorker) drain(timeout time.Duration) {
	if len(w.ch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	w.lggr.Infof("draining %d buffered events (domain=%s, entity=%s)", len(w.ch), w.domain, w.entity)

	for len(w.ch) > 0 {
		if ctx.Err() != nil {
			remaining := len(w.ch)
			if remaining > 0 {
				w.lggr.Warnf("drain timeout exceeded, dropping %d remaining events (domain=%s, entity=%s)",
					remaining, w.domain, w.entity)
				w.metrics.eventsDropped.Add(context.Background(), int64(remaining), w.metricAttrs)
			}
			return
		}

		batch := w.buildBatch()
		if batch == nil || len(batch.Events) == 0 {
			break
		}

		batchSize := int64(len(batch.Events))
		sendCtx, sendCancel := context.WithTimeout(ctx, w.sendTimeout)
		_, err := w.client.PublishBatch(sendCtx, batch)
		sendCancel()

		if err != nil {
			w.lggr.Warnf("drain PublishBatch failed, dropping %d events (domain=%s, entity=%s): %v",
				len(batch.Events), w.domain, w.entity, err)
			w.metrics.eventsDropped.Add(context.Background(), batchSize, w.metricAttrs)
			continue
		}

		w.metrics.eventsDrained.Add(context.Background(), batchSize, w.metricAttrs)
	}
}

// logBufferFullWithExpBackoff logs at 1, 2, 4, 8, 16, 32, 64, 100, 200, 300, ...
// to avoid flooding logs when the buffer is persistently full.
// dropCount is intentionally racy with Emit's Store(0) — this only affects log frequency, not correctness.
func (w *chipIngressEmitterWorker) logBufferFullWithExpBackoff(payload emitterPayload) {
	w.metrics.eventsDropped.Add(context.Background(), 1, w.metricAttrs)
	count := w.dropCount.Add(1)
	if count > 0 && (count%100 == 0 || count&(count-1) == 0) {
		w.lggr.Warnf("chip ingress emitter buffer full, dropping event (domain=%s, entity=%s, droppedCount=%d)",
			payload.domain, payload.entity, count)
	}
}
