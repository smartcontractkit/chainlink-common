package beholder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// ChipIngressBatchEmitter buffers events per (domain, entity) and flushes them
// via chipingress.Client.PublishBatch on a periodic interval.
// It satisfies the Emitter interface so it can be used as a drop-in replacement
// for ChipIngressEmitter.
type ChipIngressBatchEmitter struct {
	services.Service
	eng *services.Engine

	client chipingress.Client

	workers      map[string]*chipIngressEmitterWorker
	workersMutex sync.RWMutex

	bufferSize   uint
	maxBatchSize uint
	sendInterval time.Duration
	sendTimeout  time.Duration
	retryCfg     *RetryConfig
	drainTimeout time.Duration

	metrics batchEmitterMetrics
}

type batchEmitterMetrics struct {
	eventsSent      otelmetric.Int64Counter
	eventsDropped   otelmetric.Int64Counter
	batchRetries    otelmetric.Int64Counter
	batchFailures   otelmetric.Int64Counter
	eventsDrained   otelmetric.Int64Counter
}

// NewChipIngressBatchEmitter creates a batch emitter backed by the given chipingress client.
// Call Start() to begin health monitoring, and Close() to stop all workers.
func NewChipIngressBatchEmitter(client chipingress.Client, cfg Config, lggr logger.Logger) (*ChipIngressBatchEmitter, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	bufferSize := cfg.ChipIngressBufferSize
	if bufferSize == 0 {
		bufferSize = 100
	}
	maxBatchSize := cfg.ChipIngressMaxBatchSize
	if maxBatchSize == 0 {
		maxBatchSize = 50
	}
	sendInterval := cfg.ChipIngressSendInterval
	if sendInterval == 0 {
		sendInterval = 500 * time.Millisecond
	}
	sendTimeout := cfg.ChipIngressSendTimeout
	if sendTimeout == 0 {
		sendTimeout = 5 * time.Second
	}
	drainTimeout := cfg.ChipIngressDrainTimeout
	if drainTimeout == 0 {
		drainTimeout = 5 * time.Second
	}

	meter := otel.Meter("beholder/chip_ingress_batch_emitter")
	metrics, err := newBatchEmitterMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch emitter metrics: %w", err)
	}

	e := &ChipIngressBatchEmitter{
		client:       client,
		workers:      make(map[string]*chipIngressEmitterWorker),
		bufferSize:   bufferSize,
		maxBatchSize: maxBatchSize,
		sendInterval: sendInterval,
		sendTimeout:  sendTimeout,
		retryCfg:     cfg.ChipIngressRetryConfig,
		drainTimeout: drainTimeout,
		metrics:      metrics,
	}

	e.Service, e.eng = services.Config{
		Name: "ChipIngressBatchEmitter",
	}.NewServiceEngine(lggr)

	return e, nil
}

// Emit extracts (domain, entity) from the attributes, routes the event to the
// appropriate per-(domain, entity) worker, and returns immediately.
// If the worker's buffer is full, the event is dropped and a warning is logged.
// Returns an error if the emitter has been closed.
func (e *ChipIngressBatchEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return e.eng.IfNotStopped(func() error {
		domain, entity, err := ExtractSourceAndType(attrKVs...)
		if err != nil {
			return err
		}

		attributes := newAttributes(attrKVs...)

		worker := e.findOrCreateWorker(domain, entity)

		payload := emitterPayload{
			body:       body,
			attributes: attributes,
			domain:     domain,
			entity:     entity,
		}

		select {
		case worker.ch <- payload:
			// Intentionally racy with logBufferFullWithExpBackoff — only affects log frequency, not correctness.
			worker.dropCount.Store(0)
		case <-ctx.Done():
			return ctx.Err()
		default:
			worker.logBufferFullWithExpBackoff(payload)
		}

		return nil
	})
}

// findOrCreateWorker returns the worker for the given (domain, entity) pair,
// creating one with a new buffered channel and flush goroutine if it doesn't exist.
func (e *ChipIngressBatchEmitter) findOrCreateWorker(domain, entity string) *chipIngressEmitterWorker {
	workerKey := fmt.Sprintf("%s_%s", domain, entity)

	e.workersMutex.RLock()
	worker, found := e.workers[workerKey]
	e.workersMutex.RUnlock()

	if found {
		return worker
	}

	e.workersMutex.Lock()
	defer e.workersMutex.Unlock()

	// Double-check after acquiring write lock
	if worker, found = e.workers[workerKey]; found {
		return worker
	}

	worker = newChipIngressEmitterWorker(
		e.client,
		make(chan emitterPayload, e.bufferSize),
		domain,
		entity,
		e.maxBatchSize,
		e.sendTimeout,
		e.retryCfg,
		e.metrics,
		e.eng,
	)

	sendInterval := e.sendInterval
	drainTimeout := e.drainTimeout
	e.eng.Go(func(ctx context.Context) {
		ticker := time.NewTicker(sendInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				worker.Send(ctx)
			case <-ctx.Done():
				worker.drain(drainTimeout)
				return
			}
		}
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
		otelmetric.WithDescription("Total events dropped (buffer full or retries exhausted)"),
		otelmetric.WithUnit("{event}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	batchRetries, err := meter.Int64Counter("chip_ingress.batch_retries",
		otelmetric.WithDescription("Total PublishBatch retry attempts"),
		otelmetric.WithUnit("{attempt}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	batchFailures, err := meter.Int64Counter("chip_ingress.batch_failures",
		otelmetric.WithDescription("Total batches that failed after all retries"),
		otelmetric.WithUnit("{batch}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	eventsDrained, err := meter.Int64Counter("chip_ingress.events_drained",
		otelmetric.WithDescription("Total events flushed during graceful shutdown"),
		otelmetric.WithUnit("{event}"))
	if err != nil {
		return batchEmitterMetrics{}, err
	}

	return batchEmitterMetrics{
		eventsSent:    eventsSent,
		eventsDropped: eventsDropped,
		batchRetries:  batchRetries,
		batchFailures: batchFailures,
		eventsDrained: eventsDrained,
	}, nil
}
