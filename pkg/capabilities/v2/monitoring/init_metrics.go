package monitoring

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// InitMetrics records OTel metrics for v2 capability Initialise outcomes.
// Instruments are exported via Beholder OTLP and appear in Prometheus after collector ingestion.
type InitMetrics interface {
	RecordInit(ctx context.Context, err error, attrs ...attribute.KeyValue)
}

// NewInitMetrics creates OTel metrics for v2 capability Initialise outcomes.
func NewInitMetrics() InitMetrics {
	return &initMetrics{}
}

type initMetrics struct {
	once        sync.Once
	instruments initInstruments
	initErr     error
}

type initInstruments struct {
	success metric.Int64Gauge
	failure metric.Int64Gauge
}

func (m *initMetrics) RecordInit(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	instruments, loadErr := m.loadInstruments()
	if loadErr != nil {
		return
	}

	recordAttrs := metric.WithAttributes(attrs...)
	if err != nil {
		instruments.failure.Record(ctx, 1, recordAttrs)
		instruments.success.Record(ctx, 0, recordAttrs)
		return
	}
	instruments.success.Record(ctx, 1, recordAttrs)
	instruments.failure.Record(ctx, 0, recordAttrs)
}

func (m *initMetrics) loadInstruments() (initInstruments, error) {
	m.once.Do(func() {
		info := newInitInstrumentInfo()
		meter := beholder.GetMeter()

		success, err := info.success.NewInt64Gauge(meter)
		if err != nil {
			m.initErr = fmt.Errorf("failed to create chain capability initialization success gauge: %w", err)
			return
		}
		failure, err := info.failure.NewInt64Gauge(meter)
		if err != nil {
			m.initErr = fmt.Errorf("failed to create chain capability initialization failure gauge: %w", err)
			return
		}
		m.instruments = initInstruments{success: success, failure: failure}
	})
	return m.instruments, m.initErr
}
