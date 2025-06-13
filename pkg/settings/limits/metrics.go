package limits

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

type gauge[N any] interface {
	Record(ctx context.Context, value N, options ...metric.RecordOption)
}

type floatGauge[N Number] struct {
	otel metric.Float64Gauge
}

func (g *floatGauge[N]) Record(ctx context.Context, value N, options ...metric.RecordOption) {
	g.otel.Record(ctx, float64(value), options...)
}

type intGauge[N Number] struct {
	otel metric.Int64Gauge
}

func (g *intGauge[N]) Record(ctx context.Context, value N, options ...metric.RecordOption) {
	g.otel.Record(ctx, int64(value), options...)
}

func withScope(ctx context.Context, scope settings.Scope) metric.MeasurementOption {
	var kvs []attribute.KeyValue
	for s := scope; s > settings.ScopeGlobal; s-- {
		kvs = append(kvs, attribute.String(scope.String(), scope.Value(ctx)))
	}
	return metric.WithAttributes(kvs...)
}
