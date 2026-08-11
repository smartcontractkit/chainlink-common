package beholder

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc/credentials/insecure"

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

// capturingMetricExporter records the ResourceMetrics handed to Export so a test
// can inspect what a MeterProvider's reader actually collected.
type capturingMetricExporter struct {
	mu     sync.Mutex
	rm     metricdata.ResourceMetrics
	called bool
}

func (c *capturingMetricExporter) Temporality(k sdkmetric.InstrumentKind) metricdata.Temporality {
	return sdkmetric.DefaultTemporalitySelector(k)
}

func (c *capturingMetricExporter) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(k)
}

func (c *capturingMetricExporter) Export(_ context.Context, rm *metricdata.ResourceMetrics) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rm, c.called = *rm, true
	return nil
}

func (c *capturingMetricExporter) ForceFlush(context.Context) error { return nil }
func (c *capturingMetricExporter) Shutdown(context.Context) error   { return nil }

func (c *capturingMetricExporter) collected() (metricdata.ResourceMetrics, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rm, c.called
}

// TestNewMeterProvider_AppliesCardinalityLimit is the regression guard for the
// gRPC meter provider bypassing Config.metricOptions. Asserting on
// metricOptions alone cannot catch that, since the bug is the call site not
// using it.
func TestNewMeterProvider_AppliesCardinalityLimit(t *testing.T) {
	t.Parallel()

	const (
		uniqueAttributes = 10
		limit            = 5
	)

	cfg := TestDefaultConfig()
	cfg.MetricCardinalityLimit = limit
	// Isolate the assertion from metricviews.Default, same as
	// TestConfig_metricOptions_cardinalityLimit above.
	cfg.metricViewsDisabled = true
	// Long interval so only the explicit ForceFlush below triggers a collection.
	cfg.MetricReaderInterval = time.Hour

	resource, err := newOtelResource(cfg)
	require.NoError(t, err)

	mp, metered, err := newMeterProvider(cfg, resource, nil, insecure.NewCredentials())
	require.NoError(t, err)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	// Swap the real OTLP exporter for one we can read. Shut the original down
	// separately, since mp.Shutdown will now reach the capturing exporter.
	capture := &capturingMetricExporter{}
	original := metered.Exporter
	metered.Exporter = capture
	t.Cleanup(func() { _ = original.Shutdown(context.Background()) })

	counter, err := mp.Meter("test").Int64Counter("overflow_test_total")
	require.NoError(t, err)
	for i := range uniqueAttributes {
		counter.Add(context.Background(), 1, metric.WithAttributes(attribute.Int("key", i)))
	}

	require.NoError(t, mp.ForceFlush(context.Background()))

	rm, called := capture.collected()
	require.True(t, called, "ForceFlush should have exported a batch")

	var sum metricdata.Sum[int64]
	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "overflow_test_total" {
				sum, found = m.Data.(metricdata.Sum[int64]), true
			}
		}
	}
	require.True(t, found, "expected overflow_test_total in the collected batch")
	assert.Len(t, sum.DataPoints, limit,
		"newMeterProvider must honour MetricCardinalityLimit via Config.metricOptions")
}
