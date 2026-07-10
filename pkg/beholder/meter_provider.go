package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// metricOptions returns sdkmetric.Option values for caller MetricViews and the
// SDK per-instrument cardinality limit.
func (cfg Config) metricOptions(opts ...sdkmetric.Option) []sdkmetric.Option {
	if len(cfg.MetricViews) > 0 {
		opts = append(opts, sdkmetric.WithView(cfg.MetricViews...))
	}
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
