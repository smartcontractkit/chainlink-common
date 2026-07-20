package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

// mergeMetricViews builds the view list passed to sdkmetric.WithView: caller
// cfg.MetricViews first, then metricviews.DefaultViews(). Caller views must come
// first so a caller's histogram-bucket override wins over the default "*" deny
// view for the same instrument (see the view-precedence rule in the metricviews
// package doc). A caller view's attribute filter, if any, therefore replaces
// rather than composes with the defaults—acceptable because caller views target
// metrics that do not emit the denied labels.
func mergeMetricViews(cfg Config) []sdkmetric.View {
	if cfg.metricViewsDisabled {
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
