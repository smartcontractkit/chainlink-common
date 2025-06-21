package limits

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// GaugeFactory is a func that constructs a Gauge for the given Key.
type GaugeFactory[N Number] func(key string) (Gauge[N], error)

// Gauge is an adapter that forwards to an otel gauge.
type Gauge[N Number] func(ctx context.Context, value N, options ...metric.RecordOption)

// IntGaugeFactory returns a GaugeFactory for the given meter based on [metric.Int64Gauge].
func IntGaugeFactory[N Number](meter metric.Meter) GaugeFactory[N] {
	return func(key string) (usage Gauge[N], err error) {
		meter, err := meter.Int64Gauge(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create int gauge for key %s: %w", key, err)
		}
		return func(ctx context.Context, value N, options ...metric.RecordOption) {
			meter.Record(ctx, int64(value), options...)
		}, nil
	}
}

// FloatGaugeFactory returns a GaugeFactory for the given meter based on [metric.Float64Gauge].
func FloatGaugeFactory[N Number](meter metric.Meter) GaugeFactory[N] {
	return func(key string) (usage Gauge[N], err error) {
		meter, err := meter.Float64Gauge(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create float gauge for key %s: %w", key, err)
		}
		return func(ctx context.Context, value N, options ...metric.RecordOption) {
			meter.Record(ctx, float64(value), options...)
		}, nil
	}
}
