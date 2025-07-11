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

type histogram[N any] interface {
	Record(ctx context.Context, value N, options ...metric.RecordOption)
}

type floatRecorder[N Number] struct {
	// gauge or histogram
	otel interface {
		Record(ctx context.Context, value float64, options ...metric.RecordOption)
	}
}

func (g *floatRecorder[N]) Record(ctx context.Context, value N, options ...metric.RecordOption) {
	g.otel.Record(ctx, float64(value), options...)
}

type intRecorder[N Number] struct {
	// gauge or histogram
	otel interface {
		Record(ctx context.Context, value int64, options ...metric.RecordOption)
	}
}

func (g *intRecorder[N]) Record(ctx context.Context, value N, options ...metric.RecordOption) {
	g.otel.Record(ctx, int64(value), options...)
}

func withScope(ctx context.Context, scope settings.Scope) metric.MeasurementOption {
	return metric.WithAttributes(kvsFromScope(ctx, scope)...)
}

func kvsFromScope(ctx context.Context, scope settings.Scope) []attribute.KeyValue {
	var kvs []attribute.KeyValue
	for s := scope; s > settings.ScopeGlobal; s-- {
		kvs = append(kvs, attribute.String(scope.String(), scope.Value(ctx)))
	}
	return kvs
}
