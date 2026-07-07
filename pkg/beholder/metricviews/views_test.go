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

func TestDefaultViews_dropsEventIDFromBaseTriggerRetry(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	counter, err := meter.Int64Counter("capabilities_base_trigger_retry_total")
	require.NoError(t, err)

	counter.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", "cap-a"),
			attribute.String("trigger_id", "trig-a"),
			attribute.String("event_id", "ev-1"),
		),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	keys := attributeKeysFromSum(t, rm)
	assert.Contains(t, keys, attribute.Key("capability_id"))
	assert.NotContains(t, keys, attribute.Key("trigger_id"))
	assert.NotContains(t, keys, attribute.Key("event_id"))
}

func TestDefaultViews_dropsHighCardinalityKeysGlobally(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	counter, err := meter.Int64Counter("some_other_metric_total")
	require.NoError(t, err)

	counter.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("service", "foo"),
			attribute.String("trigger_id", "trig-a"),
			attribute.String("workflow_execution_id", "exec-1"),
		),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	keys := attributeKeysFromSum(t, rm)
	assert.Contains(t, keys, attribute.Key("service"))
	assert.NotContains(t, keys, attribute.Key("trigger_id"))
	assert.NotContains(t, keys, attribute.Key("workflow_execution_id"))
}

func TestDefaultViews_stoppedResendingDropsAttempts(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	gauge, err := meter.Int64Gauge("capabilities_base_trigger_stopped_resending_timestamp")
	require.NoError(t, err)

	gauge.Record(context.Background(), 123,
		metric.WithAttributes(
			attribute.String("capability_id", "cap-a"),
			attribute.String("trigger_id", "trig-a"),
			attribute.String("event_id", "ev-1"),
			attribute.Int("attempts", 20),
		),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	keys := attributeKeysFromGauge(t, rm)
	assert.Contains(t, keys, attribute.Key("capability_id"))
	assert.Contains(t, keys, attribute.Key("trigger_id"))
	assert.Contains(t, keys, attribute.Key("event_id"))
	assert.NotContains(t, keys, attribute.Key("attempts"))
}

func TestDefaultViews_count(t *testing.T) {
	t.Parallel()
	assert.Len(t, metricviews.DefaultViews(), 3)
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

func attributeKeysFromGauge(t *testing.T, rm metricdata.ResourceMetrics) []attribute.Key {
	t.Helper()
	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)
	gauge, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, gauge.DataPoints, 1)
	return keysFromSet(gauge.DataPoints[0].Attributes)
}

func keysFromSet(set attribute.Set) []attribute.Key {
	keys := make([]attribute.Key, 0, set.Len())
	for _, kv := range set.ToSlice() {
		keys = append(keys, kv.Key)
	}
	slices.Sort(keys)
	return keys
}
