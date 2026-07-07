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

func TestDefaultViews_stoppedResendingDropsHighCardinalityKeys(t *testing.T) {
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
	assert.NotContains(t, keys, attribute.Key("event_id"))
	assert.NotContains(t, keys, attribute.Key("attempts"))
}

func TestDefaultViews_count(t *testing.T) {
	t.Parallel()
	assert.Len(t, metricviews.DefaultViews(), 7)
}

func TestDefaultViews_perWorkflowHistogramBuckets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		instrument string
		unit       string
		record     func(t *testing.T, meter metric.Meter)
		wantBounds []float64
	}{
		{
			name:       "bytes usage",
			instrument: "bound.PerWorkflow.WASMBinarySizeLimit.usage",
			unit:       "By",
			record: func(t *testing.T, meter metric.Meter) {
				t.Helper()
				h, err := meter.Int64Histogram("bound.PerWorkflow.WASMBinarySizeLimit.usage", metric.WithUnit("By"))
				require.NoError(t, err)
				h.Record(context.Background(), 512*1024, metric.WithAttributes(attribute.String("workflow", "wf-1")))
			},
			wantBounds: []float64{0, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8},
		},
		{
			name:       "seconds runtime",
			instrument: "time.PerWorkflow.ExecutionTimeout.runtime",
			unit:       "s",
			record: func(t *testing.T, meter metric.Meter) {
				t.Helper()
				h, err := meter.Float64Histogram("time.PerWorkflow.ExecutionTimeout.runtime", metric.WithUnit("s"))
				require.NoError(t, err)
				h.Record(context.Background(), 42.5, metric.WithAttributes(attribute.String("workflow", "wf-1")))
			},
			wantBounds: []float64{0, 1, 10, 60, 300, 900, 3600},
		},
		{
			name:       "gas usage",
			instrument: "bound.PerWorkflow.ChainWrite.EVM.GasLimit.usage",
			unit:       "{gas}",
			record: func(t *testing.T, meter metric.Meter) {
				t.Helper()
				h, err := meter.Int64Histogram("bound.PerWorkflow.ChainWrite.EVM.GasLimit.usage", metric.WithUnit("{gas}"))
				require.NoError(t, err)
				// pkg/settings/cresettings ChainWrite gas defaults: Solana
				// 300_000, Aptos 2_000_000, EVM 5_000_000, up to 50_000_000
				// for per-chain-selector overrides. All must land in a finite
				// bucket, not overflow to +Inf.
				for _, gas := range []int64{300_000, 2_000_000, 5_000_000, 10_000_000, 50_000_000} {
					h.Record(context.Background(), gas, metric.WithAttributes(attribute.String("workflow", "wf-1")))
				}
			},
			wantBounds: []float64{0, 1e5, 5e5, 1e6, 5e6, 1e7, 5e7},
		},
		{
			name:       "count usage",
			instrument: "bound.PerWorkflow.TriggerSubscriptionLimit.usage",
			record: func(t *testing.T, meter metric.Meter) {
				t.Helper()
				h, err := meter.Int64Histogram("bound.PerWorkflow.TriggerSubscriptionLimit.usage")
				require.NoError(t, err)
				h.Record(context.Background(), 3, metric.WithAttributes(attribute.String("workflow", "wf-1")))
			},
			wantBounds: []float64{0, 1, 10, 100, 1e3, 1e4, 1e5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(
				sdkmetric.WithReader(reader),
				sdkmetric.WithView(metricviews.DefaultViews()...),
			)
			t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

			meter := mp.Meter("test")
			tt.record(t, meter)

			var rm metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(context.Background(), &rm))

			bounds := histogramBounds(t, rm, tt.instrument)
			assert.Equal(t, tt.wantBounds, bounds)
			assert.Len(t, bounds, 7, "expected 7 boundaries (8 Prometheus buckets including +Inf)")

			if tt.name == "gas usage" {
				counts := histogramBucketCounts(t, rm, tt.instrument)
				overflow := counts[len(counts)-1]
				assert.Zero(t, overflow, "gas observations up to the documented ChainWrite defaults must not collapse into the +Inf bucket")
			}
		})
	}
}

// TestDefaultViews_perWorkflowHistogramDropsHighCardinalityKeys guards
// against the bucket-override view for a PerWorkflow histogram claiming the
// stream identity ahead of the global "*" deny-filter view and, as a result,
// silently bypassing it (the SDK dedupes matching views by stream identity
// and keeps only the first match's Stream mask). The bucket-boundary views
// must carry the deny filter themselves.
func TestDefaultViews_perWorkflowHistogramDropsHighCardinalityKeys(t *testing.T) {
	t.Parallel()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricviews.DefaultViews()...),
	)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	h, err := meter.Int64Histogram("bound.PerWorkflow.WASMBinarySizeLimit.usage", metric.WithUnit("By"))
	require.NoError(t, err)
	h.Record(context.Background(), 512*1024, metric.WithAttributes(
		attribute.String("workflow_execution_id", "wf-exec-1"),
		attribute.String("event_id", "evt-1"),
		attribute.String("workflow", "wf-1"),
	))

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)
	data, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[int64])
	require.True(t, ok)
	require.Len(t, data.DataPoints, 1)

	keys := keysFromSet(data.DataPoints[0].Attributes)
	assert.Contains(t, keys, attribute.Key("workflow"))
	assert.NotContains(t, keys, attribute.Key("workflow_execution_id"))
	assert.NotContains(t, keys, attribute.Key("event_id"))
}

func histogramBounds(t *testing.T, rm metricdata.ResourceMetrics, name string) []float64 {
	t.Helper()
	require.Len(t, rm.ScopeMetrics, 1)
	for _, m := range rm.ScopeMetrics[0].Metrics {
		if m.Name != name {
			continue
		}
		switch data := m.Data.(type) {
		case metricdata.Histogram[int64]:
			require.Len(t, data.DataPoints, 1)
			return data.DataPoints[0].Bounds
		case metricdata.Histogram[float64]:
			require.Len(t, data.DataPoints, 1)
			return data.DataPoints[0].Bounds
		default:
			t.Fatalf("unexpected metric data type %T for %s", m.Data, name)
		}
	}
	t.Fatalf("metric %q not found", name)
	return nil
}

func histogramBucketCounts(t *testing.T, rm metricdata.ResourceMetrics, name string) []uint64 {
	t.Helper()
	require.Len(t, rm.ScopeMetrics, 1)
	for _, m := range rm.ScopeMetrics[0].Metrics {
		if m.Name != name {
			continue
		}
		switch data := m.Data.(type) {
		case metricdata.Histogram[int64]:
			require.Len(t, data.DataPoints, 1)
			return data.DataPoints[0].BucketCounts
		case metricdata.Histogram[float64]:
			require.Len(t, data.DataPoints, 1)
			return data.DataPoints[0].BucketCounts
		default:
			t.Fatalf("unexpected metric data type %T for %s", m.Data, name)
		}
	}
	t.Fatalf("metric %q not found", name)
	return nil
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
