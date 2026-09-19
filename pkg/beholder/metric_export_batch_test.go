package beholder

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestPeriodicReaderMetricExportBatchSize(t *testing.T) {
	testCases := []struct {
		name          string
		batchSize     int
		wantExporters int
		wantMaxPoints int
	}{
		{
			name:          "positive value batches exports",
			batchSize:     2,
			wantExporters: 3,
			wantMaxPoints: 2,
		},
		{
			name:          "zero is unbatched",
			wantExporters: 1,
			wantMaxPoints: 5,
		},
		{
			name:          "negative value is unbatched",
			batchSize:     -1,
			wantExporters: 1,
			wantMaxPoints: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exporter := &metricBatchRecorder{}
			reader := newPeriodicReader(exporter, tc.batchSize, sdkmetric.WithInterval(time.Hour))
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

			meter := provider.Meter("beholder/metric-export-batching-test")
			counter, err := meter.Int64Counter("metric_export_batching_points")
			require.NoError(t, err)
			for i := range 5 {
				counter.Add(context.Background(), 1, metric.WithAttributes(attribute.Int("point", i)))
			}

			require.NoError(t, provider.ForceFlush(context.Background()))

			batches := exporter.batchesSnapshot()
			require.Len(t, batches, tc.wantExporters)
			seen := make(map[int]int, 5)
			for _, batch := range batches {
				require.LessOrEqual(t, len(batch), tc.wantMaxPoints)
				for _, point := range batch {
					seen[point.attribute]++
				}
			}
			require.Len(t, seen, 5)
			for i := range 5 {
				require.Equal(t, 1, seen[i])
			}
		})
	}
}

type metricBatchRecorder struct {
	mu      sync.Mutex
	batches [][]metricBatchPoint
}

type metricBatchPoint struct {
	attribute int
}

var _ sdkmetric.Exporter = (*metricBatchRecorder)(nil)

func (e *metricBatchRecorder) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (e *metricBatchRecorder) Aggregation(sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.AggregationDefault{}
}

func (e *metricBatchRecorder) Export(_ context.Context, metrics *metricdata.ResourceMetrics) error {
	var points []metricBatchPoint
	for _, scope := range metrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			sum, ok := metric.Data.(metricdata.Sum[int64])
			if !ok {
				continue
			}
			for _, point := range sum.DataPoints {
				value, ok := point.Attributes.Value("point")
				if !ok {
					continue
				}
				points = append(points, metricBatchPoint{attribute: int(value.AsInt64())})
			}
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.batches = append(e.batches, points)
	return nil
}

func (e *metricBatchRecorder) ForceFlush(context.Context) error { return nil }

func (e *metricBatchRecorder) Shutdown(context.Context) error { return nil }

func (e *metricBatchRecorder) batchesSnapshot() [][]metricBatchPoint {
	e.mu.Lock()
	defer e.mu.Unlock()

	batches := make([][]metricBatchPoint, len(e.batches))
	for i, batch := range e.batches {
		batches[i] = append([]metricBatchPoint(nil), batch...)
	}
	return batches
}
