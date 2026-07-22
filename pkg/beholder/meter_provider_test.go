package beholder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/metricviews"
)

func TestConfig_metricOptions_cardinalityLimit(t *testing.T) {
	t.Parallel()

	const (
		uniqueAttributes = 10
		limit            = 5
	)

	reader := sdkmetric.NewManualReader()
	cfg := DefaultConfig()
	cfg.MetricCardinalityLimit = limit
	cfg.metricViewsDisabled = true

	mpOpts := append(cfg.metricOptions(), sdkmetric.WithReader(reader))
	mp := sdkmetric.NewMeterProvider(mpOpts...)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	meter := mp.Meter("test")
	counter, err := meter.Int64Counter("overflow_test_total")
	require.NoError(t, err)

	for i := range uniqueAttributes {
		counter.Add(context.Background(), 1, metric.WithAttributes(attribute.Int("key", i)))
	}

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	sum := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Sum[int64])
	assert.Len(t, sum.DataPoints, limit)

	var total int64
	for _, dp := range sum.DataPoints {
		total += dp.Value
	}
	assert.Equal(t, int64(uniqueAttributes), total)
}

func TestConfig_metricViews_appendsDefaultsAfterCallerViews(t *testing.T) {
	t.Parallel()

	callerView := sdkmetric.NewView(
		sdkmetric.Instrument{Name: "custom_metric"},
		sdkmetric.Stream{},
	)
	cfg := DefaultConfig()
	cfg.MetricViews = []sdkmetric.View{callerView}

	views := cfg.metricViews()
	require.GreaterOrEqual(t, len(views), len(metricviews.Default(nil))+1)
}

func TestConfig_metricViews_includesDenylistDefaultView(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.MetricViewsDenyAttributes = []string{"event_id"}

	views := cfg.metricViews()
	require.Len(t, views, len(metricviews.Default(cfg.MetricViewsDenyAttributes)))
	require.NotEmpty(t, views)
}

// TestConfig_metricViews_emptyDenylistOmitsCatchAll verifies that leaving
// MetricViewsDenyAttributes empty skips only the configurable global "*"
// deny-list view; the fixed PerWorkflow histogram bucket and base-trigger
// allow-list defaults still apply (see metricviews.Default).
func TestConfig_metricViews_emptyDenylistOmitsCatchAll(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	views := cfg.metricViews()
	require.Len(t, views, len(metricviews.Default(nil)))
	require.NotEmpty(t, views)
}
