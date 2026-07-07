package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

func mergeMetricViews(cfg Config) []sdkmetric.View {
	if cfg.MetricViewsDisabled {
		return cfg.MetricViews
	}
	// Caller-supplied views must be evaluated before the default cardinality-limiting
	// views: the OTel SDK dedupes views that resolve to the same stream identity
	// (name/description/unit/kind) and keeps only the first match, so a caller
	// customizing a stream (e.g. histogram buckets via WithOtelViews) would
	// otherwise be silently shadowed by the default catch-all view.
	return append(append([]sdkmetric.View{}, cfg.MetricViews...), metricviews.DefaultViews()...)
}

func appendMeterProviderOptions(cfg Config, opts ...sdkmetric.Option) []sdkmetric.Option {
	opts = append(opts, sdkmetric.WithView(mergeMetricViews(cfg)...))
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
