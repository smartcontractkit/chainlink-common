package beholder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// ChipIngressBatchEmitter buffers events per (domain, entity) and flushes them
// via chipingress.Client.PublishBatch on a periodic interval.
// Each (domain, entity) pair gets its own batch.Client, providing per-entity
// isolation and independent concurrency scaling.
// It satisfies the Emitter interface so it can be used as a drop-in replacement
// for ChipIngressEmitter.
type ChipIngressBatchEmitter struct {
	services.Service
	eng *services.Engine

	client chipingress.Client

	workers      map[string]*chipIngressEmitterWorker
	workersMutex sync.RWMutex

	bufferSize         int
	maxBatchSize       int
	maxWorkers         int
	maxConcurrentSends int
	sendInterval       time.Duration
	sendTimeout        time.Duration
	drainTimeout       time.Duration

	metrics batchEmitterMetrics
}

type batchEmitterMetrics struct {
	eventsSent    otelmetric.Int64Counter
	eventsDropped otelmetric.Int64Counter
}

// NewChipIngressBatchEmitter creates a batch emitter backed by the given chipingress client.
// Call Start() to begin health monitoring, and Close() to stop all workers.
func NewChipIngressBatchEmitter(client chipingress.Client, cfg Config, lggr logger.Logger) (*ChipIngressBatchEmitter, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	bufferSize := int(cfg.ChipIngressBufferSize)
	if bufferSize == 0 {
		bufferSize = 1000
	}
	maxBatchSize := int(cfg.ChipIngressMaxBatchSize)
	if maxBatchSize == 0 {
		maxBatchSize = 500
	}
	maxWorkers := cfg.ChipIngressMaxWorkers
	if maxWorkers == 0 {
		maxWorkers = defaultMaxWorkers
	}
	maxConcurrentSends := cfg.ChipIngressMaxConcurrentSends
	if maxConcurrentSends == 0 {
		maxConcurrentSends = defaultMaxConcurrentSends
	}
	sendInterval := cfg.ChipIngressSendInterval
	if sendInterval == 0 {
		sendInterval = 100 * time.Millisecond
	}
	sendTimeout := cfg.ChipIngressSendTimeout
	if sendTimeout == 0 {
		sendTimeout = 3 * time.Second
	}
	drainTimeout := cfg.ChipIngressDrainTimeout
	if drainTimeout == 0 {
		drainTimeout = 10 * time.Second
	}

	meter := otel.Meter("beholder/chip_ingress_batch_emitter")
	metrics, err := newBatchEmitterMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch emitter metrics: %w", err)
	}

	e := &ChipIngressBatchEmitter{
		client:             client,
		workers:            make(map[string]*chipIngressEmitterWorker),
		bufferSize:         bufferSize,
		maxBatchSize:       maxBatchSize,
		maxWorkers:         maxWorkers,
		maxConcurrentSends: maxConcurrentSends,
		sendInterval:       sendInterval,
		sendTimeout:        sendTimeout,
		drainTimeout:       drainTimeout,
		metrics:            metrics,
	}

	e.Service, e.eng = services.Config{
		Name: "ChipIngressBatchEmitter",
	}.NewServiceEngine(lggr)

	return e, nil
}

// Emit queues an event for batched delivery. It returns immediately without blocking.
// If the worker's buffer is full, the event is silently dropped (metric bumped).
// Returns an error only if the emitter is closed or the context is cancelled.
func (e *ChipIngressBatchEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return e.emitInternal(ctx, body, nil, attrKVs...)
}

// EmitWithCallback works like Emit but accepts a callback that is invoked once
// the event's fate is determined: nil on successful PublishBatch delivery, or a
// non-nil error on send failure or buffer-full drop.
//
// Contract:
//   - If EmitWithCallback returns a non-nil error, the callback will NOT be invoked.
//     The caller should handle the error from the return value.
//   - If EmitWithCallback returns nil, the callback is guaranteed to be invoked
//     exactly once — either asynchronously (after the batch is sent) or
//     synchronously (if the event was dropped at enqueue time).
//
// Callers can use this for "synchronous" emission:
//
//	done := make(chan error, 1)
//	if err := emitter.EmitWithCallback(ctx, body, func(err error) { done <- err }, attrKVs...); err != nil {
//	    return err // callback will not fire
//	}
//	err := <-done // safe — callback will fire exactly once
func (e *ChipIngressBatchEmitter) EmitWithCallback(ctx context.Context, body []byte, callback func(error), attrKVs ...any) error {
	return e.emitInternal(ctx, body, callback, attrKVs...)
}

func (e *ChipIngressBatchEmitter) emitInternal(ctx context.Context, body []byte, callback func(error), attrKVs ...any) error {
	return e.eng.IfNotStopped(func() error {
		domain, entity, err := ExtractSourceAndType(attrKVs...)
		if err != nil {
			return err
		}

		attributes := newAttributes(attrKVs...)

		worker := e.findOrCreateWorker(domain, entity)
		if worker == nil {
			if callback != nil {
				callback(fmt.Errorf("max workers reached, event dropped"))
			}
			return nil
		}

		event, err := chipingress.NewEvent(domain, entity, body, attributes)
		if err != nil {
			return fmt.Errorf("failed to create CloudEvent: %w", err)
		}
		eventPb, err := chipingress.EventToProto(event)
		if err != nil {
			return fmt.Errorf("failed to convert to proto: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		queueErr := worker.batchClient.QueueMessage(eventPb, func(sendErr error) {
			if sendErr != nil {
				e.metrics.eventsDropped.Add(context.Background(), 1, worker.metricAttrs)
			} else {
				e.metrics.eventsSent.Add(context.Background(), 1, worker.metricAttrs)
			}
			if callback != nil {
				callback(sendErr)
			}
		})
		if queueErr != nil {
			e.metrics.eventsDropped.Add(context.Background(), 1, worker.metricAttrs)
			if callback != nil {
				callback(queueErr)
			}
		}

		return nil
	})
}

// findOrCreateWorker returns the worker for the given (domain, entity) pair,
// creating one backed by a new batch.Client if it doesn't exist.
func (e *ChipIngressBatchEmitter) findOrCreateWorker(domain, entity string) *chipIngressEmitterWorker {
	workerKey := domain + "\x00" + entity

	e.workersMutex.RLock()
	worker, found := e.workers[workerKey]
	e.workersMutex.RUnlock()

	if found {
		return worker
	}

	e.workersMutex.Lock()
	defer e.workersMutex.Unlock()

	if worker, found = e.workers[workerKey]; found {
		return worker
	}

	if len(e.workers) >= e.maxWorkers {
		e.eng.Warnf("chip ingress batch emitter: max workers (%d) reached, dropping event for %s/%s", e.maxWorkers, domain, entity)
		e.metrics.eventsDropped.Add(context.Background(), 1, otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("domain", domain),
			attribute.String("entity", entity),
		)))
		return nil
	}

	batchClient, err := batch.NewBatchClient(e.client,
		batch.WithBatchSize(e.maxBatchSize),
		batch.WithMessageBuffer(e.bufferSize),
		batch.WithBatchInterval(e.sendInterval),
		batch.WithMaxPublishTimeout(e.sendTimeout),
		batch.WithShutdownTimeout(e.drainTimeout),
		batch.WithMaxConcurrentSends(e.maxConcurrentSends),
		batch.WithEventClone(false),
	)
	if err != nil {
		e.eng.Errorf("chip ingress batch emitter: failed to create batch client for %s: %v", workerKey, err)
		return nil
	}

	worker = newChipIngressEmitterWorker(batchClient, domain, entity)

	e.eng.Go(func(ctx context.Context) {
		worker.batchClient.Start(ctx)
		<-ctx.Done()
		worker.batchClient.Stop()
	})

	e.workers[workerKey] = worker
	return worker
}

func newBatchEmitterMetrics(meter otelmetric.Meter) (batchEmitterMetrics, error) {
	eventsSent, err := meter.Int64Counter("chip_ingress.events_sent",
		otelmetric.WithDescription("Total events successfully sent via PublishBatch"),
		otelmetric.WithUnit("{event}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	eventsDropped, err := meter.Int64Counter("chip_ingress.events_dropped",
		otelmetric.WithDescription("Total events dropped (buffer full or send failure)"),
		otelmetric.WithUnit("{event}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	return batchEmitterMetrics{
		eventsSent:    eventsSent,
		eventsDropped: eventsDropped,
	}, nil
}
