package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// ActionMetrics records OTel metrics for v2 capability action lifecycle events.
// Instruments are exported via Beholder OTLP and appear in Prometheus after collector ingestion.
type ActionMetrics interface {
	OnSuccess(ctx context.Context, method string, tsStart, tsEmit time.Time, attrs ...attribute.KeyValue)
	OnError(ctx context.Context, method string, tsStart, tsEmit time.Time, isUserError bool, attrs ...attribute.KeyValue)
}

// NewActionMetrics creates OTel metrics for v2 capability action outcomes.
func NewActionMetrics() ActionMetrics {
	return &actionMetrics{}
}

type actionMetrics struct {
	once        sync.Once
	instruments actionInstruments
	initErr     error
}

type actionInstruments struct {
	count    metric.Int64Counter
	duration metric.Int64Histogram
}

func (m *actionMetrics) OnSuccess(ctx context.Context, method string, tsStart, tsEmit time.Time, attrs ...attribute.KeyValue) {
	m.record(ctx, OutcomeSuccess, tsStart, tsEmit, attrs...)
}

func (m *actionMetrics) OnError(ctx context.Context, method string, tsStart, tsEmit time.Time, isUserError bool, attrs ...attribute.KeyValue) {
	if isUserError {
		return
	}
	m.record(ctx, OutcomeError, tsStart, tsEmit, attrs...)
}

func (m *actionMetrics) record(ctx context.Context, outcome string, tsStart, tsEmit time.Time, attrs ...attribute.KeyValue) {
	instruments, err := m.loadInstruments()
	if err != nil {
		return
	}

	startMs := tsStart.UnixMilli()
	emitMs := tsEmit.UnixMilli()
	recordAttrs := metric.WithAttributes(append(attrs, attribute.String(LabelOutcome, outcome))...)
	instruments.count.Add(ctx, 1, recordAttrs)
	instruments.duration.Record(ctx, emitMs-startMs, recordAttrs)
}

func (m *actionMetrics) loadInstruments() (actionInstruments, error) {
	m.once.Do(func() {
		info := newActionInstrumentInfo()
		meter := beholder.GetMeter()

		count, err := info.count.NewInt64Counter(meter)
		if err != nil {
			m.initErr = fmt.Errorf("failed to create action count counter: %w", err)
			return
		}
		duration, err := info.duration.NewInt64Histogram(meter)
		if err != nil {
			m.initErr = fmt.Errorf("failed to create action duration histogram: %w", err)
			return
		}
		m.instruments = actionInstruments{count: count, duration: duration}
	})
	return m.instruments, m.initErr
}

type noopActionMetrics struct{}

// NoopActionMetrics is a no-op ActionMetrics implementation for tests.
func NoopActionMetrics() ActionMetrics { return noopActionMetrics{} }

func (noopActionMetrics) OnSuccess(context.Context, string, time.Time, time.Time, ...attribute.KeyValue) {}
func (noopActionMetrics) OnError(context.Context, string, time.Time, time.Time, bool, ...attribute.KeyValue) {
}
