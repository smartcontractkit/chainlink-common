package monitoring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestInitMetrics_RecordInitSuccess(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewInitMetrics()
	// 4051577828743386545 is the chain selector for Polygon mainnet (chain ID 137).
	attrs := []attribute.KeyValue{
		attribute.String("capability", "evm@1.0.0"),
		attribute.String("ChainSelector", "4051577828743386545"),
	}

	metrics.RecordInit(t.Context(), nil, attrs...)

	rm := collectMetrics(t, reader)

	success := mustMetric(t, rm, InitSuccessMetric)
	successGauge, ok := success.Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, successGauge.DataPoints, 1)
	assert.Equal(t, int64(1), successGauge.DataPoints[0].Value)
	requireAttrValue(t, successGauge.DataPoints[0].Attributes, "ChainSelector", "4051577828743386545")

	failure := mustMetric(t, rm, InitFailureMetric)
	failureGauge, ok := failure.Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, failureGauge.DataPoints, 1)
	assert.Equal(t, int64(0), failureGauge.DataPoints[0].Value)
}

func TestInitMetrics_RecordInitFailure(t *testing.T) {
	reader, cleanup := useTestMeterProvider(t)
	defer cleanup()

	metrics := NewInitMetrics()
	attrs := []attribute.KeyValue{
		attribute.String("capability", "solana@1.0.0"),
		attribute.String("ChainSelector", "42"),
	}

	metrics.RecordInit(t.Context(), errors.New("boom"), attrs...)

	rm := collectMetrics(t, reader)

	failure := mustMetric(t, rm, InitFailureMetric)
	failureGauge, ok := failure.Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, failureGauge.DataPoints, 1)
	assert.Equal(t, int64(1), failureGauge.DataPoints[0].Value)
	requireAttrValue(t, failureGauge.DataPoints[0].Attributes, "ChainSelector", "42")

	success := mustMetric(t, rm, InitSuccessMetric)
	successGauge, ok := success.Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, successGauge.DataPoints, 1)
	assert.Equal(t, int64(0), successGauge.DataPoints[0].Value)
}
