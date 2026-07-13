package metricviews_test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

func TestDefaultViews_emptyBlacklist(t *testing.T) {
	t.Parallel()
	assert.Nil(t, metricviews.DefaultViews(nil))

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews(nil)...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	counter, err := meter.Int64Counter("some_metric_total")
	require.NoError(t, err)

	counter.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("service", "foo"),
			attribute.String("event_id", "ev-1"),
			attribute.String("workflow_execution_id", "exec-1"),
		),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	keys := attributeKeysFromSum(t, rm)
	assert.Contains(t, keys, attribute.Key("service"))
	assert.Contains(t, keys, attribute.Key("event_id"))
	assert.Contains(t, keys, attribute.Key("workflow_execution_id"))
}

func TestDefaultViews_dropsBlacklistedAttributes(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews([]string{"event_id"})...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	counter, err := meter.Int64Counter("some_metric_total")
	require.NoError(t, err)

	counter.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("service", "foo"),
			attribute.String("event_id", "ev-1"),
			attribute.String("workflow_execution_id", "exec-1"),
		),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	keys := attributeKeysFromSum(t, rm)
	assert.Contains(t, keys, attribute.Key("service"))
	assert.NotContains(t, keys, attribute.Key("event_id"))
	assert.Contains(t, keys, attribute.Key("workflow_execution_id"))
}

func attributeKeysFromSum(t *testing.T, rm metricdata.ResourceMetrics) []attribute.Key {
	t.Helper()
	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)
	sum, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Sum[int64])
	require.True(t, ok)
	require.Len(t, sum.DataPoints, 1)
	return keysFromSet(sum.DataPoints[0].Attributes)
}

func keysFromSet(set attribute.Set) []attribute.Key {
	keys := make([]attribute.Key, 0, set.Len())
	for _, kv := range set.ToSlice() {
		keys = append(keys, kv.Key)
	}
	slices.Sort(keys)
	return keys
}
