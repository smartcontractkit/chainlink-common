package limits

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
)

type metricsChecker struct {
	exp exporter
	*sdkmetric.MeterProvider
}

func newMetricsChecker(t *testing.T) *metricsChecker {
	var c metricsChecker
	c.exp.t = t
	c.MeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				&c.exp,
				sdkmetric.WithInterval(time.Second),
			),
		),
	)
	return &c
}

func (mc *metricsChecker) lastResourceFirstScopeMetric(t *testing.T) metrics {
	require.NoError(t, mc.MeterProvider.ForceFlush(t.Context()))
	return mc.exp.lastResourceFirstScopeMetric(t)
}

var _ sdkmetric.Exporter = &exporter{}

type exporter struct {
	t    *testing.T
	temp metricdata.Temporality
	agg  sdkmetric.Aggregation

	mu  sync.Mutex
	rms []*metricdata.ResourceMetrics
}

func (e *exporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.temp
}

func (e *exporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return e.agg
}

func (e *exporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	metrics = cloneResourceMetrics(e.t, metrics)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rms = append(e.rms, metrics)
	return nil
}

func cloneResourceMetrics(t *testing.T, rm *metricdata.ResourceMetrics) *metricdata.ResourceMetrics {
	return &metricdata.ResourceMetrics{
		Resource:     cloneResource(t, rm.Resource),
		ScopeMetrics: cloneScopeMetrics(t, rm.ScopeMetrics),
	}
}

func cloneResource(t *testing.T, r *resource.Resource) *resource.Resource {
	r2, err := resource.Merge(resource.Empty(), r)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Equal(r2) || !r2.Equal(r) {
		t.Fatalf("clone is not equal: orig %v \n clone %v", r, r2)
	}
	return r2
}

func cloneScopeMetrics(t *testing.T, sm []metricdata.ScopeMetrics) []metricdata.ScopeMetrics {
	out := make([]metricdata.ScopeMetrics, len(sm))
	for i := range sm {
		out[i] = cloneScopeMetric(t, sm[i])
	}
	return out
}

func cloneScopeMetric(t *testing.T, sm metricdata.ScopeMetrics) metricdata.ScopeMetrics {
	return metricdata.ScopeMetrics{
		Scope:   sm.Scope,
		Metrics: cloneMetrics(t, sm.Metrics),
	}
}

func cloneMetrics(t *testing.T, ms []metricdata.Metrics) []metricdata.Metrics {
	out := make([]metricdata.Metrics, len(ms))
	for i := range ms {
		out[i] = cloneMetric(t, ms[i])
	}
	return out
}

func cloneMetric(t *testing.T, sm metricdata.Metrics) metricdata.Metrics {
	return metricdata.Metrics{
		Name:        sm.Name,
		Description: sm.Description,
		Unit:        sm.Unit,
		Data:        redactAggregationTimestamps(t, sm.Data),
	}
}

func (e *exporter) ForceFlush(ctx context.Context) error { return nil }

func (e *exporter) Shutdown(ctx context.Context) error { return nil }

func (e *exporter) lastResourceFirstScopeMetric(t *testing.T) metrics {
	e.mu.Lock()
	defer e.mu.Unlock()
	require.NotEmpty(t, e.rms)
	sc := e.rms[len(e.rms)-1].ScopeMetrics
	require.NotEmpty(t, sc)
	return sc[0].Metrics
}

type metrics []metricdata.Metrics

func redactHistogramVals[N int64 | float64](t *testing.T, ms metrics, name string) {
	t.Helper()
	i := ms.forName(name)
	if i < 0 {
		t.Fatalf("failed to find histogram named: %s", name)
		return
	}
	h, ok := ms[i].Data.(metricdata.Histogram[N])
	if !ok {
		t.Fatalf("failed to find histogram named: %s: %#v", name, ms)
		return
	}
	for j := range h.DataPoints {
		h.DataPoints[j].Min = metricdata.Extrema[N]{}
		h.DataPoints[j].Max = metricdata.Extrema[N]{}
		h.DataPoints[j].Sum = 0
	}
}

func (ms metrics) forName(name string) int {
	for i := range ms {
		if ms[i].Name == name {
			return i
		}
	}
	return -1
}

func redactAggregationTimestamps(t *testing.T, orig metricdata.Aggregation) metricdata.Aggregation {
	switch a := orig.(type) {
	case metricdata.Sum[float64]:
		return metricdata.Sum[float64]{
			Temporality: a.Temporality,
			DataPoints:  redactDataPointTimestamps(a.DataPoints),
			IsMonotonic: a.IsMonotonic,
		}
	case metricdata.Sum[int64]:
		return metricdata.Sum[int64]{
			Temporality: a.Temporality,
			DataPoints:  redactDataPointTimestamps(a.DataPoints),
			IsMonotonic: a.IsMonotonic,
		}
	case metricdata.Gauge[float64]:
		return metricdata.Gauge[float64]{
			DataPoints: redactDataPointTimestamps(a.DataPoints),
		}
	case metricdata.Gauge[int64]:
		return metricdata.Gauge[int64]{
			DataPoints: redactDataPointTimestamps(a.DataPoints),
		}
	case metricdata.Histogram[int64]:
		return metricdata.Histogram[int64]{
			Temporality: a.Temporality,
			DataPoints:  redactHistogramTimestamps(a.DataPoints),
		}
	case metricdata.Histogram[float64]:
		return metricdata.Histogram[float64]{
			Temporality: a.Temporality,
			DataPoints:  redactHistogramTimestamps(a.DataPoints),
		}
	default:
		t.Errorf("%T", a)
		return orig
	}
}

func redactHistogramTimestamps[T int64 | float64](hdp []metricdata.HistogramDataPoint[T]) []metricdata.HistogramDataPoint[T] {
	out := make([]metricdata.HistogramDataPoint[T], len(hdp))
	for i, dp := range hdp {
		out[i] = metricdata.HistogramDataPoint[T]{
			Attributes:   attribute.NewSet(dp.Attributes.ToSlice()...),
			Count:        dp.Count,
			Sum:          dp.Sum,
			Bounds:       slices.Clone(dp.Bounds),
			BucketCounts: slices.Clone(dp.BucketCounts),
			Min:          dp.Min,
			Max:          dp.Max,
		}
	}
	return out
}

func redactDataPointTimestamps[T int64 | float64](sdp []metricdata.DataPoint[T]) []metricdata.DataPoint[T] {
	out := make([]metricdata.DataPoint[T], len(sdp))
	for i, dp := range sdp {
		out[i] = metricdata.DataPoint[T]{
			Attributes: attribute.NewSet(dp.Attributes.ToSlice()...),
			Value:      dp.Value,
		}
	}
	return out
}
