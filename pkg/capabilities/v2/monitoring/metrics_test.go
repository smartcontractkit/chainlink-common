package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestActionMetrics_OnSuccessRecordsMetrics(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewActionMetrics()
	start := time.Now().Add(-50 * time.Millisecond)
	emit := time.Now()
	attrs := ActionMetricAttributes(
		"WriteReport",
		capabilities.RequestMetadata{WorkflowDonID: 1},
		func() []attribute.KeyValue {
			return []attribute.KeyValue{attribute.String(LabelCapabilityID, "solana:ChainSelector:1@1.0.0")}
		},
		nil,
	)

	metrics.OnSuccess(t.Context(), "WriteReport", start, emit, attrs...)

	rm := collectMetrics(t, reader)
	metric := mustMetric(t, rm, ActionCountMetric)
	sum, ok := metric.Data.(metricdata.Sum[int64])
	require.True(t, ok)
	require.Len(t, sum.DataPoints, 1)
	assert.Equal(t, int64(1), sum.DataPoints[0].Value)
	requireAttrValue(t, sum.DataPoints[0].Attributes, LabelMethod, "WriteReport")
	requireAttrValue(t, sum.DataPoints[0].Attributes, LabelOutcome, OutcomeSuccess)
	requireAttrValue(t, sum.DataPoints[0].Attributes, LabelCapabilityID, "solana:ChainSelector:1@1.0.0")

	durationMetric := mustMetric(t, rm, ActionDurationMetric)
	hist, ok := durationMetric.Data.(metricdata.Histogram[int64])
	require.True(t, ok)
	require.Len(t, hist.DataPoints, 1)
	requireAttrValue(t, hist.DataPoints[0].Attributes, LabelOutcome, OutcomeSuccess)
}

func TestActionMetrics_OnErrorSkipsUserErrors(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewActionMetrics()
	start := time.Now()
	emit := time.Now()

	metrics.OnError(t.Context(), "WriteReport", start, emit, true)

	rm := collectMetrics(t, reader)
	assert.Empty(t, rm.ScopeMetrics)
}

func TestActionMetrics_OnErrorRecordsSystemErrors(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewActionMetrics()
	start := time.Now()
	emit := time.Now()
	attrs := ActionMetricAttributes(
		"GetBalance",
		capabilities.RequestMetadata{WorkflowDonID: 1},
		func() []attribute.KeyValue {
			return []attribute.KeyValue{attribute.String(LabelCapabilityID, "solana:ChainSelector:1@1.0.0")}
		},
		nil,
	)

	metrics.OnError(t.Context(), "GetBalance", start, emit, false, attrs...)

	rm := collectMetrics(t, reader)
	metric := mustMetric(t, rm, ActionCountMetric)
	sum, ok := metric.Data.(metricdata.Sum[int64])
	require.True(t, ok)
	require.Len(t, sum.DataPoints, 1)
	assert.Equal(t, int64(1), sum.DataPoints[0].Value)
	requireAttrValue(t, sum.DataPoints[0].Attributes, LabelMethod, "GetBalance")
	requireAttrValue(t, sum.DataPoints[0].Attributes, LabelOutcome, OutcomeError)
}

func TestActionMetrics_UsesSingleCounterAcrossMethods(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewActionMetrics()
	now := time.Now()
	capAttrs := func() []attribute.KeyValue {
		return []attribute.KeyValue{attribute.String(LabelCapabilityID, "solana:ChainSelector:1@1.0.0")}
	}

	metrics.OnSuccess(t.Context(), "GetBalance", now, now, ActionMetricAttributes("GetBalance", capabilities.RequestMetadata{}, capAttrs, nil)...)
	metrics.OnSuccess(t.Context(), "WriteReport", now, now, ActionMetricAttributes("WriteReport", capabilities.RequestMetadata{}, capAttrs, nil)...)

	rm := collectMetrics(t, reader)
	countMetrics := 0
	for _, sm := range rm.ScopeMetrics {
		for _, metric := range sm.Metrics {
			if metric.Name == ActionCountMetric {
				countMetrics++
			}
		}
	}
	assert.Equal(t, 1, countMetrics)
}

func TestNoopActionMetrics(t *testing.T) {
	t.Parallel()

	metrics := NoopActionMetrics()
	now := time.Now()
	assert.NotPanics(t, func() {
		metrics.OnSuccess(context.Background(), "WriteReport", now, now)
		metrics.OnError(context.Background(), "WriteReport", now, now, false)
	})
}

func useTestMeterProvider(t *testing.T) (*sdkmetric.ManualReader, func()) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	client := beholder.NoopClientConfig{Lggr: logger.Test(t)}.New()
	client.MeterProvider = provider
	client.Meter = provider.Meter("test")
	prev := beholder.GetClient()
	beholder.SetClient(client)

	return reader, func() {
		require.NoError(t, provider.Shutdown(t.Context()))
		beholder.SetClient(prev)
	}
}

func collectMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &rm))
	return rm
}

func mustMetric(t *testing.T, rm metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, metric := range sm.Metrics {
			if metric.Name == name {
				return metric
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func requireAttrValue(t *testing.T, set attribute.Set, key, want string) {
	t.Helper()
	val, ok := set.Value(attribute.Key(key))
	require.True(t, ok, "missing attribute %q", key)
	assert.Equal(t, want, val.AsString())
}
