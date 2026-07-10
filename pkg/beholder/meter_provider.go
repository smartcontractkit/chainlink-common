package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// metricOptions returns cfg-derived sdkmetric.Option values (caller MetricViews
// and the SDK per-instrument cardinality limit). Callers append reader/resource
// options at the call site.
func (cfg Config) metricOptions() []sdkmetric.Option {
	var opts []sdkmetric.Option
	if len(cfg.MetricViews) > 0 {
		opts = append(opts, sdkmetric.WithView(cfg.MetricViews...))
	}
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
