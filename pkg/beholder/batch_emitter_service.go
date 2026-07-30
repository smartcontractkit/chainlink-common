package beholder

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// ChipIngressBatchEmitterService batches events and sends them via chipingress.Client.PublishBatch.
// It implements the Emitter interface.
type ChipIngressBatchEmitterService struct {
	services.Service
	eng *services.Engine

	batchClient *batch.Client

	metricAttrsCache sync.Map // map[string]otelmetric.MeasurementOption
	metrics          batchEmitterMetrics
}

type batchEmitterMetrics struct {
	eventsSent    otelmetric.Int64Counter
	eventsDropped otelmetric.Int64Counter
}

// NewChipIngressBatchEmitterService creates a batch emitter service backed by the given chipingress client.
func NewChipIngressBatchEmitterService(client chipingress.Client, cfg Config, lggr logger.Logger) (*ChipIngressBatchEmitterService, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	defaults := DefaultConfig()
	bufferSize := int(cfg.ChipIngressBufferSize)
	if bufferSize == 0 {
		bufferSize = int(defaults.ChipIngressBufferSize)
	}
	maxBatchSize := int(cfg.ChipIngressMaxBatchSize)
	if maxBatchSize == 0 {
		maxBatchSize = int(defaults.ChipIngressMaxBatchSize)
	}
	maxConcurrentSends := cfg.ChipIngressMaxConcurrentSends
	if maxConcurrentSends == 0 {
		maxConcurrentSends = defaults.ChipIngressMaxConcurrentSends
	}
	sendInterval := cfg.ChipIngressSendInterval
	if sendInterval == 0 {
		sendInterval = defaults.ChipIngressSendInterval
	}
	sendTimeout := cfg.ChipIngressSendTimeout
	if sendTimeout == 0 {
		sendTimeout = defaults.ChipIngressSendTimeout
	}
	drainTimeout := cfg.ChipIngressDrainTimeout
	if drainTimeout == 0 {
		drainTimeout = defaults.ChipIngressDrainTimeout
	}
	maxGRPCRequestSize := cfg.ChipIngressMaxGRPCRequestSize
	if maxGRPCRequestSize == 0 {
		maxGRPCRequestSize = defaults.ChipIngressMaxGRPCRequestSize
	}

	meter := otel.Meter("beholder/chip_ingress_batch_emitter")
	metrics, err := newBatchEmitterMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch emitter metrics: %w", err)
	}

	batchClient, err := batch.NewBatchClient(client,
		batch.WithBatchSize(maxBatchSize),
		batch.WithMessageBuffer(bufferSize),
		batch.WithBatchInterval(sendInterval),
		batch.WithMaxPublishTimeout(sendTimeout),
		batch.WithShutdownTimeout(drainTimeout),
		batch.WithMaxConcurrentSends(maxConcurrentSends),
		batch.WithMaxGRPCRequestSize(maxGRPCRequestSize),
		batch.WithEventClone(false),
		batch.WithClientName(batch.ClientNameBeholder),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch client: %w", err)
	}

	e := &ChipIngressBatchEmitterService{
		batchClient: batchClient,
		metrics:     metrics,
	}

	e.Service, e.eng = services.Config{
		Name:  "ChipIngressBatchEmitterService",
		Start: e.start,
		Close: e.stop,
	}.NewServiceEngine(lggr)

	return e, nil
}

func (e *ChipIngressBatchEmitterService) start(_ context.Context) error {
	// Do not pass the startup ctx — the services contract forbids retaining it
	// after Start returns. Use the engine's lifecycle context so the batcher
	// is cancelled when the service shuts down (StopChan closes before stop() runs).
	ctx, _ := e.eng.NewCtx()
	e.batchClient.Start(ctx)
	return nil
}

func (e *ChipIngressBatchEmitterService) stop() error {
	e.batchClient.Stop()
	return nil
}

// Emit queues an event for batched delivery without blocking.
// Returns an error if the emitter is stopped or the context is cancelled.
// If the buffer is full, the event is silently dropped.
func (e *ChipIngressBatchEmitterService) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return e.emitInternal(ctx, body, nil, attrKVs...)
}

// EmitWithCallback works like Emit but invokes callback once the event's fate
// is determined (nil on success, non-nil on failure or buffer-full drop).
//
// If EmitWithCallback returns a non-nil error, the callback will NOT be invoked.
// If it returns nil, the callback is guaranteed to fire exactly once.
func (e *ChipIngressBatchEmitterService) EmitWithCallback(ctx context.Context, body []byte, callback func(error), attrKVs ...any) error {
	return e.emitInternal(ctx, body, callback, attrKVs...)
}

func (e *ChipIngressBatchEmitterService) emitInternal(ctx context.Context, body []byte, callback func(error), attrKVs ...any) error {
	return e.eng.IfStarted(func() error {
		domain, entity, err := ExtractSourceAndType(attrKVs...)
		if err != nil {
			return err
		}

		attributes := newAttributes(attrKVs...)

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

		metricAttrs := e.metricAttrsFor(domain, entity)

		queueErr := e.batchClient.QueueMessage(eventPb, func(sendErr error) {
			// The callback fires asynchronously after the batch is sent,
			// so the caller's ctx may already be cancelled. Use ctx directly
			// for metric recording — OTel Add is non-blocking and tolerates
			// cancelled contexts.
			if sendErr != nil {
				errorCode := batch.ErrorCodeFor(sendErr)
				e.metrics.eventsDropped.Add(ctx, 1, e.dropMetricAttrsFor(domain, entity, errorCode))
				// Partial delivery is not logged: it is a per-event, often persistent
				// server-side condition (e.g. missing schema) that would otherwise log on
				// every dropped event. chip_ingress.events_dropped (error_code) already
				// captures that it's happening and roughly why; the full reason isn't
				// needed at fleet-wide log volume.
				if _, ok := errors.AsType[*batch.PublishError](sendErr); !ok {
					e.eng.Errorw("failed to emit to chip ingress",
						"error", sendErr,
						"error_code", errorCode,
						"error_reason", sendErr.Error(),
						"domain", domain,
						"entity", entity,
					)
				}
			} else {
				e.metrics.eventsSent.Add(ctx, 1, metricAttrs)
			}
			if callback != nil {
				callback(sendErr)
			}
		})
		if queueErr != nil {
			errorCode := batch.ErrorCodeFor(queueErr)
			e.metrics.eventsDropped.Add(ctx, 1, e.dropMetricAttrsFor(domain, entity, errorCode))
			e.eng.Errorw("failed to queue message for chip ingress",
				"error", queueErr,
				"error_code", errorCode,
				"error_reason", queueErr.Error(),
				"domain", domain,
				"entity", entity,
			)
			if callback != nil {
				callback(queueErr)
			}
		}

		return nil
	})
}

func (e *ChipIngressBatchEmitterService) metricAttrsFor(domain, entity string) otelmetric.MeasurementOption {
	key := domain + "\x00" + entity
	if v, ok := e.metricAttrsCache.Load(key); ok {
		return v.(otelmetric.MeasurementOption)
	}
	attrs := otelmetric.WithAttributeSet(attribute.NewSet(
		attribute.String("domain", domain),
		attribute.String("entity", entity),
		attribute.String("client_name", batch.ClientNameBeholder),
	))
	v, _ := e.metricAttrsCache.LoadOrStore(key, attrs)
	return v.(otelmetric.MeasurementOption)
}

// dropMetricAttrsFor returns a measurement option for the eventsDropped counter.
// Not cached — drop paths are not on the hot path.
//
// error_reason is deliberately excluded: it is free-form text from the server/gRPC
// stack (e.g. validation messages, status details) and is not a bounded value, so using
// it as a metric attribute would create unbounded cardinality. domain/entity/client_name/error_code
// are all closed, bounded sets. error_reason is still available on the corresponding
// log line.
func (e *ChipIngressBatchEmitterService) dropMetricAttrsFor(domain, entity, errorCode string) otelmetric.MeasurementOption {
	attrs := []attribute.KeyValue{
		attribute.String("domain", domain),
		attribute.String("entity", entity),
		attribute.String("client_name", batch.ClientNameBeholder),
	}
	if errorCode != "" {
		attrs = append(attrs, attribute.String("error_code", errorCode))
	}
	return otelmetric.WithAttributeSet(attribute.NewSet(attrs...))
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
