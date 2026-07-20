package beholder

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc/stats"
)

// --- test doubles ---------------------------------------------------------

// fakeLogExporter stands in for the otlploggrpc exporter. On Export it mimics
// what the real gRPC stack does
type fakeLogExporter struct {
	size            int           // OutPayload length to report
	sizeFromRecords bool          // if true, report len(records) instead of size
	fireN           int           // number of OutPayload events (retry simulation); 0 => 1
	delay           time.Duration // widen the store→read window for the race test
	err             error

	shutdownCalled   atomic.Bool
	forceFlushCalled atomic.Bool
}

func (f *fakeLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	sz := f.size
	if f.sizeFromRecords {
		sz = len(records)
	}
	n := f.fireN
	if n == 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		exportSizeHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: sz})
	}
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	return f.err
}

func (f *fakeLogExporter) Shutdown(context.Context) error { f.shutdownCalled.Store(true); return nil }
func (f *fakeLogExporter) ForceFlush(context.Context) error {
	f.forceFlushCalled.Store(true)
	return nil
}

// fakeMetricExporter stands in for the otlpmetricgrpc exporter.
type fakeMetricExporter struct {
	size int
	err  error

	temporalityCalled atomic.Bool
	aggregationCalled atomic.Bool
	shutdownCalled    atomic.Bool
	forceFlushCalled  atomic.Bool
}

func (f *fakeMetricExporter) Temporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	f.temporalityCalled.Store(true)
	return metricdata.CumulativeTemporality
}

func (f *fakeMetricExporter) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	f.aggregationCalled.Store(true)
	return sdkmetric.DefaultAggregationSelector(k)
}

func (f *fakeMetricExporter) Export(ctx context.Context, _ *metricdata.ResourceMetrics) error {
	exportSizeHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: f.size})
	return f.err
}

func (f *fakeMetricExporter) ForceFlush(context.Context) error {
	f.forceFlushCalled.Store(true)
	return nil
}
func (f *fakeMetricExporter) Shutdown(context.Context) error {
	f.shutdownCalled.Store(true)
	return nil
}

// --- helpers --------------------------------------------------------------

// newTestCounter wires beholder.export.bytes to an in-memory ManualReader and
// returns the counter plus a collect func that reads back the recorded metrics.
func newTestCounter(t *testing.T) (otelmetric.Int64Counter, func() []metricdata.Metrics) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	counter, err := newExportBytesCounter(mp.Meter("test"))
	require.NoError(t, err)

	collect := func() []metricdata.Metrics {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &rm))
		var out []metricdata.Metrics
		for _, sm := range rm.ScopeMetrics {
			out = append(out, sm.Metrics...)
		}
		return out
	}
	return counter, collect
}

func dpForSignal(t *testing.T, ms []metricdata.Metrics, signal string) (metricdata.DataPoint[int64], bool) {
	t.Helper()
	for _, m := range ms {
		if m.Name != exportBytesMetric {
			continue
		}
		sum, ok := m.Data.(metricdata.Sum[int64])
		require.True(t, ok, "expected beholder.export.bytes to be Sum[int64]")
		for _, dp := range sum.DataPoints {
			if v, ok := dp.Attributes.Value(attribute.Key("otel_signal")); ok && v.AsString() == signal {
				return dp, true
			}
		}
	}
	return metricdata.DataPoint[int64]{}, false
}

// --- exportSizeHandler ---------------------------------------------------

func TestexportSizeHandler_StoresOutPayloadLength(t *testing.T) {
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)

	exportSizeHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: 1234})

	assert.Equal(t, int64(1234), holder.Load())
}

func TestexportSizeHandler_IgnoresNonOutPayload(t *testing.T) {
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)

	exportSizeHandler{}.HandleRPC(ctx, &stats.InPayload{Length: 999})
	exportSizeHandler{}.HandleRPC(ctx, &stats.Begin{})
	exportSizeHandler{}.HandleRPC(ctx, &stats.End{})

	assert.Equal(t, int64(0), holder.Load())
}

func TestexportSizeHandler_LastWriteWins(t *testing.T) {
	// Retries resend the same proto, so each OutPayload reports the same length.
	// Store means the holder ends at that length, not a multiple of it.
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)
	h := exportSizeHandler{}

	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})

	assert.Equal(t, int64(700), holder.Load())
}

func TestexportSizeHandler_NoHolderInContextIsNoop(t *testing.T) {
	assert.NotPanics(t, func() {
		exportSizeHandler{}.HandleRPC(context.Background(), &stats.OutPayload{Length: 5})
	})
}

func TestexportSizeHandler_TagAndConnAreInert(t *testing.T) {
	h := exportSizeHandler{}
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &atomic.Int64{})

	assert.Equal(t, ctx, h.TagRPC(ctx, &stats.RPCTagInfo{}))
	assert.Equal(t, ctx, h.TagConn(ctx, &stats.ConnTagInfo{}))
	assert.NotPanics(t, func() { h.HandleConn(ctx, &stats.ConnBegin{}) })
}

// --- meteredLogExporter ---------------------------------------------------

func TestMeteredLogExporter_RecordsBytesOnSuccess(t *testing.T) {
	counter, collect := newTestCounter(t)
	inner := &fakeLogExporter{size: 4096}
	exp := newMeteredLogExporter(inner, counter, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), nil))

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok, "expected a logs datapoint")
	assert.Equal(t, int64(4096), dp.Value)

	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredLogExporter_NoRecordOnError(t *testing.T) {
	counter, collect := newTestCounter(t)
	// Handler still fires, but Export returns an error.
	inner := &fakeLogExporter{size: 4096, err: errors.New("boom")}
	exp := newMeteredLogExporter(inner, counter, "csa")

	require.Error(t, exp.Export(context.Background(), nil))

	_, ok := dpForSignal(t, collect(), "logs")
	assert.False(t, ok, "nothing should be recorded when export fails")
}

func TestMeteredLogExporter_RetriesCountedOnce(t *testing.T) {
	counter, collect := newTestCounter(t)
	// Three OutPayload events for one batch, all the same size.
	inner := &fakeLogExporter{size: 500, fireN: 3}
	exp := newMeteredLogExporter(inner, counter, "csa")

	require.NoError(t, exp.Export(context.Background(), nil))

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, int64(500), dp.Value, "retries of the same batch must be counted once")
}

func TestMeteredLogExporter_Passthrough(t *testing.T) {
	inner := &fakeLogExporter{}
	exp := newMeteredLogExporter(inner, nil, "csa")

	require.NoError(t, exp.Shutdown(context.Background()))
	require.NoError(t, exp.ForceFlush(context.Background()))

	assert.True(t, inner.shutdownCalled.Load())
	assert.True(t, inner.forceFlushCalled.Load())
}

// TestMeteredExporter_ConcurrentExportsIsolated is the regression guard for the
// per-call context holder
func TestMeteredExporter_ConcurrentExportsIsolated(t *testing.T) {
	counter, collect := newTestCounter(t)
	// delay widens the window between the handler storing and record reading,
	// so a broken implementation would reliably mis-attribute.
	inner := &fakeLogExporter{sizeFromRecords: true, delay: 200 * time.Microsecond}
	exp := newMeteredLogExporter(inner, counter, "csa")

	const n = 50
	var wg sync.WaitGroup
	var want int64
	for i := 1; i <= n; i++ {
		size := i
		want += int64(size)
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, exp.Export(context.Background(), make([]sdklog.Record, size)))
		}()
	}
	wg.Wait()

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, want, dp.Value)
}

// --- meteredMetricExporter ------------------------------------------------

func TestMeteredMetricExporter_RecordsBytesOnSuccess(t *testing.T) {
	counter, collect := newTestCounter(t)
	inner := &fakeMetricExporter{size: 8192}
	exp := newMeteredMetricExporter(inner)
	exp.attachCounter(counter, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	dp, ok := dpForSignal(t, collect(), "metrics")
	require.True(t, ok, "expected a metrics datapoint")
	assert.Equal(t, int64(8192), dp.Value)

	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredMetricExporter_UnmeteredBeforeAttach(t *testing.T) {
	counter, collect := newTestCounter(t)
	inner := &fakeMetricExporter{size: 8192}
	exp := newMeteredMetricExporter(inner) // no attachCounter

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	_, ok := dpForSignal(t, collect(), "metrics")
	assert.False(t, ok, "no metric should be recorded before the counter is attached")
	_ = counter
}

func TestMeteredMetricExporter_NoRecordOnError(t *testing.T) {
	counter, collect := newTestCounter(t)
	inner := &fakeMetricExporter{size: 8192, err: errors.New("boom")}
	exp := newMeteredMetricExporter(inner)
	exp.attachCounter(counter, "csa")

	require.Error(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	_, ok := dpForSignal(t, collect(), "metrics")
	assert.False(t, ok)
}

func TestMeteredMetricExporter_Passthrough(t *testing.T) {
	inner := &fakeMetricExporter{}
	exp := newMeteredMetricExporter(inner)

	assert.Equal(t, metricdata.CumulativeTemporality, exp.Temporality(sdkmetric.InstrumentKindCounter))
	assert.NotNil(t, exp.Aggregation(sdkmetric.InstrumentKindCounter))
	require.NoError(t, exp.ForceFlush(context.Background()))
	require.NoError(t, exp.Shutdown(context.Background()))

	assert.True(t, inner.temporalityCalled.Load())
	assert.True(t, inner.aggregationCalled.Load())
	assert.True(t, inner.forceFlushCalled.Load())
	assert.True(t, inner.shutdownCalled.Load())
}

// --- shared naming --------------------------------------------------------

func TestMeteredExporters_ShareOneMetricBySignal(t *testing.T) {
	counter, collect := newTestCounter(t)
	logs := newMeteredLogExporter(&fakeLogExporter{size: 100}, counter, "csa")
	metrics := newMeteredMetricExporter(&fakeMetricExporter{size: 200})
	metrics.attachCounter(counter, "csa")

	require.NoError(t, logs.Export(context.Background(), nil))
	require.NoError(t, metrics.Export(context.Background(), &metricdata.ResourceMetrics{}))

	ms := collect()
	logDP, ok := dpForSignal(t, ms, "logs")
	require.True(t, ok)
	assert.Equal(t, int64(100), logDP.Value)
	metricDP, ok := dpForSignal(t, ms, "metrics")
	require.True(t, ok)
	assert.Equal(t, int64(200), metricDP.Value)

	// Both are datapoints of the same instrument, distinguished only by otel_signal.
	for _, m := range ms {
		if m.Name == exportBytesMetric {
			assert.Equal(t, "By", m.Unit)
		}
	}
}
