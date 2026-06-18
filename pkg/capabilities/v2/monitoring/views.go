package monitoring

import sdkmetric "go.opentelemetry.io/otel/sdk/metric"

// MetricViews returns OTel views for v2 capability action metrics.
// Register via loop.WithOtelViews before the Beholder client creates instruments.
func MetricViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{Name: ActionDurationMetric},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: ActionLatencyBucketBoundariesMs,
			}},
		),
	}
}
