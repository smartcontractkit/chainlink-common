package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

// metricViews returns the view list passed to sdkmetric.WithView.
//
// Final order:
//  1. cfg.MetricViews — caller overrides (e.g. chainlink metricViews() histogram buckets)
//  2. metricviews.Default(cfg.MetricViewsDenyAttributes) — PerWorkflow histogram
//     bucket overrides, base-trigger attribute allow-lists, and (when configured)
//     the global deny-list catch-all
//
// Caller views must come first: when multiple views match an instrument and
// resolve to the same output stream (same name/description/unit/kind), the SDK
// keeps only the first in registration order and drops the rest.
func (cfg Config) metricViews() []sdkmetric.View {
	if cfg.metricViewsDisabled {
		return cfg.MetricViews
	}
	return append(cfg.MetricViews, metricviews.Default(cfg.MetricViewsDenyAttributes)...)
}

// metricOptions returns cfg-derived sdkmetric.Option values (metric views and the
// SDK per-instrument cardinality limit). Callers append reader/resource options
// at the call site.
func (cfg Config) metricOptions() []sdkmetric.Option {
	views := cfg.metricViews()
	var opts []sdkmetric.Option
	if len(views) > 0 {
		opts = append(opts, sdkmetric.WithView(views...))
	}
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
