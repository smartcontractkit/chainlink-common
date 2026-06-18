package monitoring

import (
	"testing"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestMetricViews_ActionLatencyBuckets(t *testing.T) {
	views := MetricViews()
	require.Len(t, views, 1)

	sampleNames := []string{
		MetricName("WriteReport", OutcomeSuccess, MetricSuffixCapDuration),
		MetricName("GetBalance", OutcomeError, MetricSuffixCapDuration),
	}
	for _, name := range sampleNames {
		stream, ok := metricViewStream(views, name)
		require.True(t, ok, "missing metric view for %s", name)

		aggregation, ok := stream.Aggregation.(sdkmetric.AggregationExplicitBucketHistogram)
		require.True(t, ok, "expected explicit bucket histogram for %s", name)
		require.Equal(t, ActionLatencyBucketBoundariesMs, aggregation.Boundaries)
	}
}

func metricViewStream(views []sdkmetric.View, name string) (sdkmetric.Stream, bool) {
	for _, view := range views {
		if stream, ok := view(sdkmetric.Instrument{Name: name}); ok {
			return stream, true
		}
	}
	return sdkmetric.Stream{}, false
}
