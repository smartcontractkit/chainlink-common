package beholder

import (
	"context"
	"fmt"
	"math/rand/v2"
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
type chipIngressEmitterWorker struct {
	client       chipingress.Client
	ch           chan emitterPayload
	domain       string
	entity       string
	maxBatchSize uint
	sendTimeout  time.Duration
	lggr         logger.Logger
	dropCount    atomic.Uint32

	retryInitialInterval time.Duration
	retryMaxInterval     time.Duration
	retryMaxElapsed      time.Duration

	metrics batchEmitterMetrics
}

func newChipIngressEmitterWorker(
	client chipingress.Client,
	ch chan emitterPayload,
	domain string,
	entity string,
	maxBatchSize uint,
	sendTimeout time.Duration,
	retryCfg *RetryConfig,
	metrics batchEmitterMetrics,
	lggr logger.Logger,
) *chipIngressEmitterWorker {
	w := &chipIngressEmitterWorker{
		client:       client,
		ch:           ch,
		domain:       domain,
		entity:       entity,
		maxBatchSize: maxBatchSize,
		sendTimeout:  sendTimeout,
		lggr:         logger.Named(lggr, "ChipIngressEmitterWorker"),
		metrics:      metrics,
	}
	if retryCfg != nil && retryCfg.Enabled() {
		w.retryInitialInterval = retryCfg.InitialInterval
		w.retryMaxInterval = retryCfg.MaxInterval
		w.retryMaxElapsed = retryCfg.MaxElapsedTime
	}
	return w
}

// Send drains the channel and sends a batch with retry on failure.
// Called periodically by the tick loop.
func (w *chipIngressEmitterWorker) Send(ctx context.Context) {
	if len(w.ch) == 0 {
		return
	}

	batch := w.buildBatch()
	if batch == nil || len(batch.Events) == 0 {
		return
	}

	if w.retryMaxElapsed == 0 {
		w.publishOnce(ctx, batch)
		return
	}

	backoff := w.retryInitialInterval
	deadline := time.Now().Add(w.retryMaxElapsed)

	batchSize := int64(len(batch.Events))
	metricAttrs := otelmetric.WithAttributeSet(w.metricAttributes())

	for attempt := 0; ; attempt++ {
		sendCtx, cancel := context.WithTimeout(ctx, w.sendTimeout)
		_, err := w.client.PublishBatch(sendCtx, batch)
		cancel()

		if err == nil {
			w.metrics.eventsSent.Add(context.Background(), batchSize, metricAttrs)
			return
		}

		w.lggr.Warnf("PublishBatch failed (attempt %d, domain=%s, entity=%s): %v",
			attempt+1, w.domain, w.entity, err)
		w.metrics.batchRetries.Add(context.Background(), 1, metricAttrs)

		if time.Now().Add(backoff).After(deadline) {
			w.lggr.Warnf("PublishBatch retries exhausted, dropping %d events (domain=%s, entity=%s)",
				len(batch.Events), w.domain, w.entity)
			w.metrics.batchFailures.Add(context.Background(), 1, metricAttrs)
			w.metrics.eventsDropped.Add(context.Background(), batchSize, metricAttrs)
			return
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			w.lggr.Warnf("context cancelled during retry, dropping %d events (domain=%s, entity=%s)",
				len(batch.Events), w.domain, w.entity)
			w.metrics.eventsDropped.Add(context.Background(), batchSize, metricAttrs)
			return
		case <-timer.C:
		}

		backoff = min(backoff*2, w.retryMaxInterval)
		jitter := time.Duration(rand.Int64N(int64(backoff) / 5)) //nolint:gosec
		backoff += jitter
	}
}

func (w *chipIngressEmitterWorker) publishOnce(ctx context.Context, batch *chipingress.CloudEventBatch) {
	sendCtx, cancel := context.WithTimeout(ctx, w.sendTimeout)
	defer cancel()

	batchSize := int64(len(batch.Events))
	metricAttrs := otelmetric.WithAttributeSet(w.metricAttributes())
	_, err := w.client.PublishBatch(sendCtx, batch)
	if err != nil {
		w.lggr.Warnf("could not send batch via chip ingress (domain=%s, entity=%s): %v",
			w.domain, w.entity, err)
		w.metrics.batchFailures.Add(context.Background(), 1, metricAttrs)
		w.metrics.eventsDropped.Add(context.Background(), batchSize, metricAttrs)
		return
	}
	w.metrics.eventsSent.Add(context.Background(), batchSize, metricAttrs)
}

// buildBatch drains the channel up to maxBatchSize and converts payloads to a CloudEventBatch.
func (w *chipIngressEmitterWorker) buildBatch() *chipingress.CloudEventBatch {
	var events []chipingress.CloudEvent
	metricAttrs := otelmetric.WithAttributeSet(w.metricAttributes())

	for len(w.ch) > 0 && len(events) < int(w.maxBatchSize) { // #nosec G115
		payload := <-w.ch
		event, err := w.payloadToEvent(payload)
		if err != nil {
			w.lggr.Warnf("failed to build CloudEvent, dropping: %v", err)
			w.metrics.eventsDropped.Add(context.Background(), 1, metricAttrs)
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
		w.metrics.eventsDropped.Add(context.Background(), int64(len(events)), metricAttrs)
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

	metricAttrs := otelmetric.WithAttributeSet(w.metricAttributes())
	w.lggr.Infof("draining %d buffered events (domain=%s, entity=%s)", len(w.ch), w.domain, w.entity)

	for len(w.ch) > 0 {
		if ctx.Err() != nil {
			remaining := len(w.ch)
			if remaining > 0 {
				w.lggr.Warnf("drain timeout exceeded, dropping %d remaining events (domain=%s, entity=%s)",
					remaining, w.domain, w.entity)
				w.metrics.eventsDropped.Add(context.Background(), int64(remaining), metricAttrs)
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
			w.metrics.eventsDropped.Add(context.Background(), batchSize, metricAttrs)
			continue
		}

		w.metrics.eventsDrained.Add(context.Background(), batchSize, metricAttrs)
	}
}

// logBufferFullWithExpBackoff logs at 1, 2, 4, 8, 16, 32, 64, 100, 200, 300, ...
// to avoid flooding logs when the buffer is persistently full.
func (w *chipIngressEmitterWorker) logBufferFullWithExpBackoff(payload emitterPayload) {
	w.metrics.eventsDropped.Add(context.Background(), 1, otelmetric.WithAttributeSet(w.metricAttributes()))
	count := w.dropCount.Add(1)
	if count > 0 && (count%100 == 0 || count&(count-1) == 0) {
		w.lggr.Warnf("chip ingress emitter buffer full, dropping event (domain=%s, entity=%s, droppedCount=%d)",
			payload.domain, payload.entity, count)
	}
}

func (w *chipIngressEmitterWorker) metricAttributes() attribute.Set {
	return attribute.NewSet(
		attribute.String("domain", w.domain),
		attribute.String("entity", w.entity),
	)
}
