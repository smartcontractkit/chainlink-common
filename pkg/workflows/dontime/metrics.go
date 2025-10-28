package dontime

import (
	"context"
	"fmt"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promDONTimeClockDrift = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "don_time_clock_drift_ms",
		Help: "Drift between node local time and generated DON Time",
	}, []string{"oracleID"})
)

type GenericDONTimeMetrics interface {
	RecordClockDrift(ctx context.Context, drift int64)
}

var _ GenericDONTimeMetrics = &donTimeMetrics{}

type donTimeMetrics struct {
	oracleID     string
	donTimeCount metric.Int64Counter
	clockDrift   metric.Int64Gauge
}

func NewGenericDONTimeMetrics(oracleID string) (GenericDONTimeMetrics, error) {
	clockDrift, err := beholder.GetMeter().Int64Gauge("don_time_clock_drift_ms")
	if err != nil {
		return nil, fmt.Errorf("failed to register clock drift metric: %w", err)
	}

	return &donTimeMetrics{
		oracleID:   oracleID,
		clockDrift: clockDrift,
	}, nil
}

func (m *donTimeMetrics) RecordClockDrift(ctx context.Context, drift int64) {
	promDONTimeClockDrift.WithLabelValues(m.oracleID).Set(float64(drift))
	m.clockDrift.Record(ctx, drift, metric.WithAttributes(
		attribute.String("oracleID", m.oracleID)))
}
