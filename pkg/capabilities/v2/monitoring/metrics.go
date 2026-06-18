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
	success sync.Map // method -> metricsCapBasic
	error   sync.Map // method -> metricsCapBasic
}

type metricsCapBasic struct {
	count             metric.Int64Counter
	capTimestampStart metric.Int64Gauge
	capTimestampEmit  metric.Int64Gauge
	capDuration       metric.Int64Histogram
}

type metricsInfoCapBasic struct {
	count             beholder.MetricInfo
	capTimestampStart beholder.MetricInfo
	capTimestampEmit  beholder.MetricInfo
	capDuration       beholder.MetricInfo
}

func (m *actionMetrics) OnSuccess(ctx context.Context, method string, tsStart, tsEmit time.Time, attrs ...attribute.KeyValue) {
	instruments, err := m.successInstruments(method)
	if err != nil {
		return
	}
	instruments.recordEmit(ctx, tsStart, tsEmit, attrs...)
}

func (m *actionMetrics) OnError(ctx context.Context, method string, tsStart, tsEmit time.Time, isUserError bool, attrs ...attribute.KeyValue) {
	if isUserError {
		return
	}
	instruments, err := m.errorInstruments(method)
	if err != nil {
		return
	}
	instruments.recordEmit(ctx, tsStart, tsEmit, attrs...)
}

func (m *actionMetrics) successInstruments(method string) (metricsCapBasic, error) {
	return m.loadInstruments(&m.success, method, OutcomeSuccess)
}

func (m *actionMetrics) errorInstruments(method string) (metricsCapBasic, error) {
	return m.loadInstruments(&m.error, method, OutcomeError)
}

func (m *actionMetrics) loadInstruments(store *sync.Map, method, outcome string) (metricsCapBasic, error) {
	if cached, ok := store.Load(method); ok {
		return cached.(metricsCapBasic), nil
	}

	instruments, err := newMetricsCapBasic(ActionMetricInfo(method, outcome))
	if err != nil {
		return metricsCapBasic{}, err
	}

	actual, _ := store.LoadOrStore(method, instruments)
	return actual.(metricsCapBasic), nil
}

func newMetricsCapBasic(info metricsInfoCapBasic) (metricsCapBasic, error) {
	meter := beholder.GetMeter()
	set := metricsCapBasic{}
	var err error

	set.count, err = info.count.NewInt64Counter(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create counter: %w", err)
	}
	set.capTimestampStart, err = info.capTimestampStart.NewInt64Gauge(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create start gauge: %w", err)
	}
	set.capTimestampEmit, err = info.capTimestampEmit.NewInt64Gauge(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create emit gauge: %w", err)
	}
	set.capDuration, err = info.capDuration.NewInt64Histogram(meter)
	if err != nil {
		return set, fmt.Errorf("failed to create duration histogram: %w", err)
	}
	return set, nil
}

func (m *metricsCapBasic) recordEmit(ctx context.Context, tsStart, tsEmit time.Time, attrKVs ...attribute.KeyValue) {
	startMs := tsStart.UnixMilli()
	emitMs := tsEmit.UnixMilli()
	attrs := metric.WithAttributes(attrKVs...)
	m.count.Add(ctx, 1, attrs)
	m.capTimestampStart.Record(ctx, startMs, attrs)
	m.capTimestampEmit.Record(ctx, emitMs, attrs)
	m.capDuration.Record(ctx, emitMs-startMs, attrs)
}

type noopActionMetrics struct{}

// NoopActionMetrics is a no-op ActionMetrics implementation for tests.
func NoopActionMetrics() ActionMetrics { return noopActionMetrics{} }

func (noopActionMetrics) OnSuccess(context.Context, string, time.Time, time.Time, ...attribute.KeyValue) {}
func (noopActionMetrics) OnError(context.Context, string, time.Time, time.Time, bool, ...attribute.KeyValue) {
}
