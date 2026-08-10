package monitoring

import (
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestMetricViews_ActionLatencyBuckets(t *testing.T) {
	views := MetricViews()
	require.Len(t, views, 1)

	stream, ok := metricViewStream(views, ActionDurationMetric)
	require.True(t, ok, "missing metric view for %s", ActionDurationMetric)

	aggregation, ok := stream.Aggregation.(sdkmetric.AggregationExplicitBucketHistogram)
	require.True(t, ok, "expected explicit bucket histogram for %s", ActionDurationMetric)
	require.Equal(t, ActionLatencyBucketBoundariesMs, aggregation.Boundaries)
}

func metricViewStream(views []sdkmetric.View, name string) (sdkmetric.Stream, bool) {
	for _, view := range views {
		if stream, ok := view(sdkmetric.Instrument{Name: name}); ok {
			return stream, true
		}
	}
	return sdkmetric.Stream{}, false
}
