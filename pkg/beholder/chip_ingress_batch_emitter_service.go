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

	bufferSize := int(cfg.ChipIngressBufferSize)
	if bufferSize == 0 {
		bufferSize = 1000
	}
	maxBatchSize := int(cfg.ChipIngressMaxBatchSize)
	if maxBatchSize == 0 {
		maxBatchSize = 500
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

	batchClient, err := batch.NewBatchClient(client,
		batch.WithBatchSize(maxBatchSize),
		batch.WithMessageBuffer(bufferSize),
		batch.WithBatchInterval(sendInterval),
		batch.WithMaxPublishTimeout(sendTimeout),
		batch.WithShutdownTimeout(drainTimeout),
		batch.WithMaxConcurrentSends(maxConcurrentSends),
		batch.WithEventClone(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch client: %w", err)
	}

	e := &ChipIngressBatchEmitterService{
		batchClient: batchClient,
		metrics:     metrics,
	}

	e.Service, e.eng = services.Config{
		Name: "ChipIngressBatchEmitterService",
	}.NewServiceEngine(lggr)

	e.eng.Go(func(ctx context.Context) {
		batchClient.Start(ctx)
		<-ctx.Done()
		batchClient.Stop()
	})

	return e, nil
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
	return e.eng.IfNotStopped(func() error {
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
			if sendErr != nil {
				e.metrics.eventsDropped.Add(ctx, 1, metricAttrs)
			} else {
				e.metrics.eventsSent.Add(ctx, 1, metricAttrs)
			}
			if callback != nil {
				callback(sendErr)
			}
		})
		if queueErr != nil {
			e.metrics.eventsDropped.Add(ctx, 1, metricAttrs)
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
	))
	v, _ := e.metricAttrsCache.LoadOrStore(key, attrs)
	return v.(otelmetric.MeasurementOption)
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
