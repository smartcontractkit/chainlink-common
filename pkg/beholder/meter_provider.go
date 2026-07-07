package beholder

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

func mergeMetricViews(cfg Config) []sdkmetric.View {
	if cfg.MetricViewsDisabled {
		return cfg.MetricViews
	}
	return append(metricviews.DefaultViews(), cfg.MetricViews...)
}

func appendMeterProviderOptions(cfg Config, opts ...sdkmetric.Option) []sdkmetric.Option {
	opts = append(opts, sdkmetric.WithView(mergeMetricViews(cfg)...))
	if cfg.MetricCardinalityLimit > 0 {
		opts = append(opts, sdkmetric.WithCardinalityLimit(cfg.MetricCardinalityLimit))
	}
	return opts
}
