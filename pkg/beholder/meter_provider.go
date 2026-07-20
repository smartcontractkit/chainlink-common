package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

// mergeMetricViews builds the view list passed to sdkmetric.WithView.
//
// Final order:
//  1. cfg.MetricViews — caller overrides (e.g. chainlink metricViews() histogram buckets)
//  2. metricviews.DefaultViews() — attribute filters for high-cardinality labels
//
// Caller views must come first: when multiple views match an instrument and
// resolve to the same output stream (same name/description/unit/kind), the SDK
// keeps only the first in registration order and drops the rest. If the default
// "*" denylist ran first, caller histogram-bucket views would be silently dropped.
//
// A consequence is that filters do not compose: once a caller view wins for a
// stream, the default attribute filter no longer applies to it. That is acceptable
// because caller views target metrics that do not emit the denied labels; capability
// trigger metrics carry no caller view and fall through to the defaults below.
func mergeMetricViews(cfg Config) []sdkmetric.View {
	if cfg.MetricViewsDisabled {
		return cfg.MetricViews
	}
	return append(append([]sdkmetric.View{}, cfg.MetricViews...), metricviews.DefaultViews()...)
}

// metricOptions returns cfg-derived sdkmetric.Option values (the merged
// MetricViews and the SDK per-instrument cardinality limit). Callers append
// reader/resource options at the call site.
func (cfg Config) metricOptions() []sdkmetric.Option {
	var opts []sdkmetric.Option
	if views := mergeMetricViews(cfg); len(views) > 0 {
		opts = append(opts, sdkmetric.WithView(views...))
	}
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
