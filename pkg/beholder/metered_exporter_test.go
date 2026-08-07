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
		beholderStatsHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: sz})
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
	size  int
	delay time.Duration
	err   error

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
	beholderStatsHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: f.size})
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
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

// newTestMetrics wires the export instruments to an in-memory ManualReader and
// returns them plus a collect func that reads back the recorded metrics.
func newTestMetrics(t *testing.T) (exportMetrics, func() []metricdata.Metrics) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	metrics, err := newExportMetrics(mp.Meter("test"))
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
	return metrics, collect
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

// histForSignal finds the beholder.export.duration datapoint for one signal and
// error outcome.
func histForSignal(t *testing.T, ms []metricdata.Metrics, signal string, wantErr bool) (metricdata.HistogramDataPoint[float64], bool) {
	t.Helper()
	for _, m := range ms {
		if m.Name != exportDurationMetric {
			continue
		}
		h, ok := m.Data.(metricdata.Histogram[float64])
		require.True(t, ok, "expected beholder.export.duration to be Histogram[float64]")
		for _, dp := range h.DataPoints {
			sig, ok := dp.Attributes.Value(attribute.Key("otel_signal"))
			if !ok || sig.AsString() != signal {
				continue
			}
			isErr, ok := dp.Attributes.Value(attribute.Key("error"))
			require.True(t, ok, "duration datapoint must carry an error attribute")
			if isErr.AsBool() == wantErr {
				return dp, true
			}
		}
	}
	return metricdata.HistogramDataPoint[float64]{}, false
}

// --- exportSizeHandler ---------------------------------------------------

func TestExportSizeHandler_StoresOutPayloadLength(t *testing.T) {
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)

	beholderStatsHandler{}.HandleRPC(ctx, &stats.OutPayload{Length: 1234})

	assert.Equal(t, int64(1234), holder.Load())
}

func TestExportSizeHandler_IgnoresNonOutPayload(t *testing.T) {
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)

	beholderStatsHandler{}.HandleRPC(ctx, &stats.InPayload{Length: 999})
	beholderStatsHandler{}.HandleRPC(ctx, &stats.Begin{})
	beholderStatsHandler{}.HandleRPC(ctx, &stats.End{})

	assert.Equal(t, int64(0), holder.Load())
}

func TestExportSizeHandler_LastWriteWins(t *testing.T) {
	// Retries resend the same proto, so each OutPayload reports the same length.
	// Store means the holder ends at that length, not a multiple of it.
	var holder atomic.Int64
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &holder)
	h := beholderStatsHandler{}

	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 700})

	assert.Equal(t, int64(700), holder.Load())
}

func TestExportSizeHandler_NoHolderInContextIsNoop(t *testing.T) {
	assert.NotPanics(t, func() {
		beholderStatsHandler{}.HandleRPC(context.Background(), &stats.OutPayload{Length: 5})
	})
}

func TestExportSizeHandler_TagAndConnAreInert(t *testing.T) {
	h := beholderStatsHandler{}
	ctx := context.WithValue(context.Background(), exportSizeKey{}, &atomic.Int64{})

	assert.Equal(t, ctx, h.TagRPC(ctx, &stats.RPCTagInfo{}))
	assert.Equal(t, ctx, h.TagConn(ctx, &stats.ConnTagInfo{}))
	assert.NotPanics(t, func() { h.HandleConn(ctx, &stats.ConnBegin{}) })
}

// --- meteredLogExporter ---------------------------------------------------

func TestMeteredLogExporter_RecordsBytesOnSuccess(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	inner := &fakeLogExporter{size: 4096}
	exp := newMeteredLogExporter(inner, metrics, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), nil))

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok, "expected a logs datapoint")
	assert.Equal(t, int64(4096), dp.Value)

	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredLogExporter_NoRecordOnError(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	// Handler still fires, but Export returns an error.
	inner := &fakeLogExporter{size: 4096, err: errors.New("boom")}
	exp := newMeteredLogExporter(inner, metrics, "csa")

	require.Error(t, exp.Export(context.Background(), nil))

	ms := collect()
	_, ok := dpForSignal(t, ms, "logs")
	assert.False(t, ok, "no bytes should be recorded when export fails")
	// Duration is still recorded, labelled error=true.
	hdp, ok := histForSignal(t, ms, "logs", true)
	require.True(t, ok, "expected an error-labelled logs duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	_, ok = histForSignal(t, ms, "logs", false)
	assert.False(t, ok, "a failed export must not be labelled error=false")
}

func TestMeteredLogExporter_RetriesCountedOnce(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	// Three OutPayload events for one batch, all the same size.
	inner := &fakeLogExporter{size: 500, fireN: 3}
	exp := newMeteredLogExporter(inner, metrics, "csa")

	require.NoError(t, exp.Export(context.Background(), nil))

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, int64(500), dp.Value, "retries of the same batch must be counted once")
}

func TestMeteredLogExporter_Passthrough(t *testing.T) {
	inner := &fakeLogExporter{}
	exp := newMeteredLogExporter(inner, exportMetrics{}, "csa")

	require.NoError(t, exp.Shutdown(context.Background()))
	require.NoError(t, exp.ForceFlush(context.Background()))

	assert.True(t, inner.shutdownCalled.Load())
	assert.True(t, inner.forceFlushCalled.Load())
}

// TestMeteredExporter_ConcurrentExportsIsolated is the regression guard for the
// per-call context holder
func TestMeteredExporter_ConcurrentExportsIsolated(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	// delay widens the window between the handler storing and record reading,
	// so a broken implementation would reliably mis-attribute.
	inner := &fakeLogExporter{sizeFromRecords: true, delay: 200 * time.Microsecond}
	exp := newMeteredLogExporter(inner, metrics, "csa")

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
	metrics, collect := newTestMetrics(t)
	inner := &fakeMetricExporter{size: 8192}
	exp := newMeteredMetricExporter(inner)
	exp.attachMetrics(metrics, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	dp, ok := dpForSignal(t, collect(), "metrics")
	require.True(t, ok, "expected a metrics datapoint")
	assert.Equal(t, int64(8192), dp.Value)

	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredMetricExporter_UnmeteredBeforeAttach(t *testing.T) {
	_, collect := newTestMetrics(t)
	inner := &fakeMetricExporter{size: 8192}
	exp := newMeteredMetricExporter(inner) // no attachMetrics

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	ms := collect()
	_, ok := dpForSignal(t, ms, "metrics")
	assert.False(t, ok, "no bytes should be recorded before the instruments are attached")
	_, ok = histForSignal(t, ms, "metrics", false)
	assert.False(t, ok, "no duration should be recorded before the instruments are attached")
}

func TestMeteredMetricExporter_NoRecordOnError(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	inner := &fakeMetricExporter{size: 8192, err: errors.New("boom")}
	exp := newMeteredMetricExporter(inner)
	exp.attachMetrics(metrics, "csa")

	require.Error(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	ms := collect()
	_, ok := dpForSignal(t, ms, "metrics")
	assert.False(t, ok)
	hdp, ok := histForSignal(t, ms, "metrics", true)
	require.True(t, ok, "expected an error-labelled metrics duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
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
	inst, collect := newTestMetrics(t)
	logs := newMeteredLogExporter(&fakeLogExporter{size: 100}, inst, "csa")
	metrics := newMeteredMetricExporter(&fakeMetricExporter{size: 200})
	metrics.attachMetrics(inst, "csa")

	require.NoError(t, logs.Export(context.Background(), nil))
	require.NoError(t, metrics.Export(context.Background(), &metricdata.ResourceMetrics{}))

	ms := collect()
	logDP, ok := dpForSignal(t, ms, "logs")
	require.True(t, ok)
	assert.Equal(t, int64(100), logDP.Value)
	metricDP, ok := dpForSignal(t, ms, "metrics")
	require.True(t, ok)
	assert.Equal(t, int64(200), metricDP.Value)

	// Each signal gets its own duration datapoint on the same instrument too.
	for _, signal := range []string{"logs", "metrics"} {
		hdp, ok := histForSignal(t, ms, signal, false)
		require.True(t, ok, "expected a %s duration datapoint", signal)
		assert.Equal(t, uint64(1), hdp.Count)
	}

	// Both are datapoints of the same instrument, distinguished only by otel_signal.
	for _, m := range ms {
		switch m.Name {
		case exportBytesMetric:
			assert.Equal(t, "By", m.Unit)
		case exportDurationMetric:
			assert.Equal(t, "s", m.Unit)
		}
	}
}

// --- beholder.export.duration ---------------------------------------------

func TestMeteredLogExporter_RecordsDurationOnSuccess(t *testing.T) {
	inst, collect := newTestMetrics(t)
	const delay = 20 * time.Millisecond
	inner := &fakeLogExporter{size: 4096, delay: delay}
	exp := newMeteredLogExporter(inner, inst, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), nil))

	hdp, ok := histForSignal(t, collect(), "logs", false)
	require.True(t, ok, "expected a logs duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	assert.GreaterOrEqual(t, hdp.Sum, delay.Seconds(),
		"recorded duration must cover the inner export")
	assert.Less(t, hdp.Sum, 10.0, "recorded duration should be in seconds, not another unit")

	csa, ok := hdp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredMetricExporter_RecordsDurationOnSuccess(t *testing.T) {
	inst, collect := newTestMetrics(t)
	const delay = 20 * time.Millisecond
	exp := newMeteredMetricExporter(&fakeMetricExporter{size: 8192, delay: delay})
	exp.attachMetrics(inst, "csa-pub-hex")

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	hdp, ok := histForSignal(t, collect(), "metrics", false)
	require.True(t, ok, "expected a metrics duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	assert.GreaterOrEqual(t, hdp.Sum, delay.Seconds())

	csa, ok := hdp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

// TestMeteredExporter_DurationRetriesCountedOnce mirrors the bytes behaviour:
// the wrapper sits above the otlp retry loop, so one logical batch is one
// observation covering the whole retry sequence.
func TestMeteredExporter_DurationRetriesCountedOnce(t *testing.T) {
	inst, collect := newTestMetrics(t)
	inner := &fakeLogExporter{size: 500, fireN: 3}
	exp := newMeteredLogExporter(inner, inst, "csa")

	require.NoError(t, exp.Export(context.Background(), nil))

	hdp, ok := histForSignal(t, collect(), "logs", false)
	require.True(t, ok)
	assert.Equal(t, uint64(1), hdp.Count, "one batch must be one duration observation")
}

func TestExportDuration_UsesSecondScaledBuckets(t *testing.T) {
	inst, collect := newTestMetrics(t)
	exp := newMeteredLogExporter(&fakeLogExporter{}, inst, "csa")

	require.NoError(t, exp.Export(context.Background(), nil))

	for _, m := range collect() {
		if m.Name != exportDurationMetric {
			continue
		}
		h, ok := m.Data.(metricdata.Histogram[float64])
		require.True(t, ok)
		require.NotEmpty(t, h.DataPoints)
		assert.Equal(t,
			[]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
			h.DataPoints[0].Bounds,
			"explicit second-scaled boundaries must survive to the reader")
		return
	}
	t.Fatalf("no %s metric collected", exportDurationMetric)
}

func TestMeteredExporter_DurationRecordedPerOutcome(t *testing.T) {
	inst, collect := newTestMetrics(t)
	ok1 := newMeteredLogExporter(&fakeLogExporter{size: 10}, inst, "csa")
	bad := newMeteredLogExporter(&fakeLogExporter{size: 10, err: errors.New("boom")}, inst, "csa")

	require.NoError(t, ok1.Export(context.Background(), nil))
	require.NoError(t, ok1.Export(context.Background(), nil))
	require.Error(t, bad.Export(context.Background(), nil))

	ms := collect()
	okDP, found := histForSignal(t, ms, "logs", false)
	require.True(t, found)
	assert.Equal(t, uint64(2), okDP.Count)

	errDP, found := histForSignal(t, ms, "logs", true)
	require.True(t, found)
	assert.Equal(t, uint64(1), errDP.Count)
}
