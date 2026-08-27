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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/stats"
)

// --- test doubles ---------------------------------------------------------

// fakeLogExporter stands in for the otlploggrpc exporter. Bytes are recorded by
// exportStatsHandler inside the gRPC stack, so this fake only controls the two
// things the wrapper observes: how long Export takes and whether it fails.
type fakeLogExporter struct {
	delay time.Duration
	err   error

	shutdownCalled   atomic.Bool
	forceFlushCalled atomic.Bool
}

func (f *fakeLogExporter) Export(context.Context, []sdklog.Record) error {
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

func (f *fakeMetricExporter) Export(context.Context, *metricdata.ResourceMetrics) error {
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

// fakeTraceExporter stands in for the otlptracegrpc exporter.
type fakeTraceExporter struct {
	delay time.Duration
	err   error

	shutdownCalled atomic.Bool
}

func (f *fakeTraceExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	return f.err
}

func (f *fakeTraceExporter) Shutdown(context.Context) error {
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

// --- meteredLogExporter ---------------------------------------------------

func TestMeteredLogExporter_RecordsErrorDurationOnFailure(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	inner := &fakeLogExporter{err: errors.New("boom")}
	exp := newMeteredLogExporter(inner, metrics, "csa")

	require.Error(t, exp.Export(context.Background(), nil))

	ms := collect()
	hdp, ok := histForSignal(t, ms, "logs", true)
	require.True(t, ok, "expected an error-labelled logs duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	_, ok = histForSignal(t, ms, "logs", false)
	assert.False(t, ok, "a failed export must not be labelled error=false")
}

// TestMeteredLogExporter_ZeroValueMetricsIsPassthrough guards the nil-check in
// record: a wrapper built without instruments must export and propagate the
// inner error rather than panic.
func TestMeteredLogExporter_ZeroValueMetricsIsPassthrough(t *testing.T) {
	exp := newMeteredLogExporter(&fakeLogExporter{err: errors.New("boom")}, exportMetrics{}, "csa")

	require.Error(t, exp.Export(context.Background(), nil))
}

func TestMeteredLogExporter_Passthrough(t *testing.T) {
	inner := &fakeLogExporter{}
	exp := newMeteredLogExporter(inner, exportMetrics{}, "csa")

	require.NoError(t, exp.Shutdown(context.Background()))
	require.NoError(t, exp.ForceFlush(context.Background()))

	assert.True(t, inner.shutdownCalled.Load())
	assert.True(t, inner.forceFlushCalled.Load())
}

// --- meteredMetricExporter ------------------------------------------------

func TestMeteredMetricExporter_UnmeteredBeforeAttach(t *testing.T) {
	_, collect := newTestMetrics(t)
	inner := &fakeMetricExporter{}
	exp := newMeteredMetricExporter(inner) // no attachMetrics

	require.NoError(t, exp.Export(context.Background(), &metricdata.ResourceMetrics{}))

	ms := collect()
	_, ok := histForSignal(t, ms, "metrics", false)
	assert.False(t, ok, "no duration should be recorded before the instruments are attached")
}

func TestMeteredMetricExporter_RecordsErrorDurationOnFailure(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	exp := newMeteredMetricExporter(&fakeMetricExporter{err: errors.New("boom")})
	exp.attachMetrics(metrics, "csa")

	require.Error(t, exp.Export(context.Background(), nil))

	ms := collect()
	hdp, ok := histForSignal(t, ms, "metrics", true)
	require.True(t, ok, "expected an error-labelled metrics duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	_, ok = histForSignal(t, ms, "metrics", false)
	assert.False(t, ok, "a failed export must not be labelled error=false")
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

// --- meteredTraceExporter -------------------------------------------------

func TestMeteredTraceExporter_RecordsDurationOnSuccess(t *testing.T) {
	const delay = 2 * time.Millisecond
	inst, collect := newTestMetrics(t)
	exp := newMeteredTraceExporter(&fakeTraceExporter{delay: delay})
	exp.attachMetrics(inst, "csa-pub-hex")

	require.NoError(t, exp.ExportSpans(context.Background(), nil))

	hdp, ok := histForSignal(t, collect(), "traces", false)
	require.True(t, ok, "expected a traces duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	assert.GreaterOrEqual(t, hdp.Sum, delay.Seconds())

	csa, ok := hdp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestMeteredTraceExporter_RecordsErrorDurationOnFailure(t *testing.T) {
	inst, collect := newTestMetrics(t)
	exp := newMeteredTraceExporter(&fakeTraceExporter{err: errors.New("boom")})
	exp.attachMetrics(inst, "csa")

	require.Error(t, exp.ExportSpans(context.Background(), nil))

	ms := collect()
	hdp, ok := histForSignal(t, ms, "traces", true)
	require.True(t, ok, "expected an error-labelled traces duration datapoint")
	assert.Equal(t, uint64(1), hdp.Count)
	_, ok = histForSignal(t, ms, "traces", false)
	assert.False(t, ok, "a failed export must not be labelled error=false")
}

// TestMeteredTraceExporter_UnmeteredBeforeAttach is the guard for the deferred
// attach: newTracerProvider runs before the MeterProvider exists, so an
// unattached exporter must still export rather than panic.
func TestMeteredTraceExporter_UnmeteredBeforeAttach(t *testing.T) {
	_, collect := newTestMetrics(t)
	inner := &fakeTraceExporter{}
	exp := newMeteredTraceExporter(inner) // no attachMetrics

	require.NoError(t, exp.ExportSpans(context.Background(), nil))

	ms := collect()
	_, ok := histForSignal(t, ms, "traces", false)
	assert.False(t, ok, "no duration before the instruments are attached")
}

func TestMeteredTraceExporter_Passthrough(t *testing.T) {
	inner := &fakeTraceExporter{}
	exp := newMeteredTraceExporter(inner)

	require.NoError(t, exp.Shutdown(context.Background()))

	assert.True(t, inner.shutdownCalled.Load())
}

// --- shared naming --------------------------------------------------------

// TestMeteredExporters_ShareOneMetricBySignal asserts the three wrappers write
// to one shared duration histogram, separated only by otel_signal.
func TestMeteredExporters_ShareOneMetricBySignal(t *testing.T) {
	inst, collect := newTestMetrics(t)

	logs := newMeteredLogExporter(&fakeLogExporter{}, inst, "csa")
	metrics := newMeteredMetricExporter(&fakeMetricExporter{})
	metrics.attachMetrics(inst, "csa")
	traces := newMeteredTraceExporter(&fakeTraceExporter{})
	traces.attachMetrics(inst, "csa")

	require.NoError(t, logs.Export(context.Background(), nil))
	require.NoError(t, metrics.Export(context.Background(), nil))
	require.NoError(t, traces.ExportSpans(context.Background(), nil))

	ms := collect()
	for _, signal := range []string{"logs", "metrics", "traces"} {
		hdp, ok := histForSignal(t, ms, signal, false)
		require.True(t, ok, "expected a %s duration datapoint", signal)
		assert.Equal(t, uint64(1), hdp.Count, "one observation per export for %s", signal)
	}
}

// --- beholder.export.duration ---------------------------------------------

func TestMeteredLogExporter_RecordsDurationOnSuccess(t *testing.T) {
	inst, collect := newTestMetrics(t)
	const delay = 20 * time.Millisecond
	inner := &fakeLogExporter{delay: delay}
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
	exp := newMeteredMetricExporter(&fakeMetricExporter{delay: delay})
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

// TestMeteredExporter_DurationRetriesCountedOnce: the wrapper sits above the
// otlp retry loop, so one logical batch is one observation covering the whole
// retry sequence.
func TestMeteredExporter_DurationRetriesCountedOnce(t *testing.T) {
	inst, collect := newTestMetrics(t)
	inner := &fakeLogExporter{}
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
	ok1 := newMeteredLogExporter(&fakeLogExporter{}, inst, "csa")
	bad := newMeteredLogExporter(&fakeLogExporter{err: errors.New("boom")}, inst, "csa")

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

// --- exportStatsHandler ---------------------------------------------------

// simulateRPC drives one full unary gRPC lifecycle through the handler, in the
// same order grpc-go fires the events. Nothing outside the handler touches the
// context: TagRPC allocates the state, HandleRPC reads it back.
func simulateRPC(h *exportStatsHandler, parent context.Context, size int, rpcErr error) {
	ctx := h.TagRPC(parent, &stats.RPCTagInfo{FullMethodName: "/test.Service/Export"})
	h.HandleRPC(ctx, &stats.Begin{})
	h.HandleRPC(ctx, &stats.OutPayload{Length: size})
	h.HandleRPC(ctx, &stats.End{Error: rpcErr})
}

func TestExportStatsHandler_RecordsBytesOnSuccessfulRPC(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa-pub-hex")
	h.attachMetrics(metrics)

	simulateRPC(h, context.Background(), 4096, nil)

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok, "expected a logs datapoint")
	assert.Equal(t, int64(4096), dp.Value)

	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "csa-pub-hex", csa.AsString())
}

func TestExportStatsHandler_NoBytesOnFailedRPC(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	simulateRPC(h, context.Background(), 4096, errors.New("unavailable"))

	_, ok := dpForSignal(t, collect(), "logs")
	assert.False(t, ok, "a failed RPC must not contribute bytes")
}

// TestExportStatsHandler_RetryAttemptsCountedOnce is the semantic guard: the
// OTLP exporters retry by issuing a fresh unary RPC per attempt, so the handler
// sees three independent RPCs carrying the same proto. Only the one that lands
// may be counted.
func TestExportStatsHandler_RetryAttemptsCountedOnce(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	simulateRPC(h, context.Background(), 500, errors.New("unavailable"))
	simulateRPC(h, context.Background(), 500, errors.New("unavailable"))
	simulateRPC(h, context.Background(), 500, nil)

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, int64(500), dp.Value, "retries of the same batch must be counted once")
}

// TestExportStatsHandler_ConcurrentRPCsIsolated replaces the old
// TestMeteredExporter_ConcurrentExportsIsolated. State is per-RPC via TagRPC,
// so interleaved RPCs cannot overwrite each other.
func TestExportStatsHandler_ConcurrentRPCsIsolated(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	const n = 50
	var wg sync.WaitGroup
	var want int64
	for i := 1; i <= n; i++ {
		size := i
		want += int64(size)
		wg.Go(func() {
			simulateRPC(h, context.Background(), size, nil)
		})
	}
	wg.Wait()

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, want, dp.Value)
}

func TestExportStatsHandler_AccumulatesMultipleOutPayloads(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	ctx := h.TagRPC(context.Background(), &stats.RPCTagInfo{FullMethodName: "/test.Service/Export"})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 100})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 250})
	h.HandleRPC(ctx, &stats.End{})

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, int64(350), dp.Value, "all outbound payloads of one RPC belong to that RPC")
}

func TestExportStatsHandler_IgnoresInboundAndEmptyRPCs(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	// Inbound payload only: nothing was sent, so nothing is counted.
	ctx := h.TagRPC(context.Background(), &stats.RPCTagInfo{FullMethodName: "/test.Service/Export"})
	h.HandleRPC(ctx, &stats.InPayload{Length: 999})
	h.HandleRPC(ctx, &stats.End{})

	_, ok := dpForSignal(t, collect(), "logs")
	assert.False(t, ok, "an RPC with no outbound payload must not record a zero datapoint")
}

func TestExportStatsHandler_SecondEndIsIdempotent(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	ctx := h.TagRPC(context.Background(), &stats.RPCTagInfo{FullMethodName: "/test.Service/Export"})
	h.HandleRPC(ctx, &stats.OutPayload{Length: 64})
	h.HandleRPC(ctx, &stats.End{})
	h.HandleRPC(ctx, &stats.End{})

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	assert.Equal(t, int64(64), dp.Value)
}

func TestExportStatsHandler_NoStateInContextIsNoop(t *testing.T) {
	metrics, _ := newTestMetrics(t)
	h := newExportStatsHandler("logs", "csa")
	h.attachMetrics(metrics)

	assert.NotPanics(t, func() {
		h.HandleRPC(context.Background(), &stats.OutPayload{Length: 5})
		h.HandleRPC(context.Background(), &stats.End{})
	})
}

// TestExportStatsHandler_UnattachedIsNoop covers the window before the
// MeterProvider exists: the handler is registered at dial time, instruments
// arrive later.
func TestExportStatsHandler_UnattachedIsNoop(t *testing.T) {
	h := newExportStatsHandler("metrics", "csa")

	assert.NotPanics(t, func() {
		simulateRPC(h, context.Background(), 128, nil)
	})
}

func TestExportStatsHandler_ConnHooksAreInert(t *testing.T) {
	h := newExportStatsHandler("logs", "csa")
	ctx := context.Background()

	assert.Equal(t, ctx, h.TagConn(ctx, &stats.ConnTagInfo{}))
	assert.NotPanics(t, func() { h.HandleConn(ctx, &stats.ConnBegin{}) })
}

func TestExportStatsHandler_CSAKeyFallbackWhenEmpty(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	h := newExportStatsHandler("logs", "")
	h.attachMetrics(metrics)

	simulateRPC(h, context.Background(), 64, nil)

	dp, ok := dpForSignal(t, collect(), "logs")
	require.True(t, ok)
	csa, ok := dp.Attributes.Value(attribute.Key("csa_public_key"))
	require.True(t, ok)
	assert.Equal(t, "not-configured", csa.AsString())
}

func TestExportStatsHandler_PerSignalAttribution(t *testing.T) {
	metrics, collect := newTestMetrics(t)
	logs := newExportStatsHandler("logs", "csa")
	logs.attachMetrics(metrics)
	mtrx := newExportStatsHandler("metrics", "csa")
	mtrx.attachMetrics(metrics)

	simulateRPC(logs, context.Background(), 100, nil)
	simulateRPC(mtrx, context.Background(), 200, nil)

	ms := collect()
	logDP, ok := dpForSignal(t, ms, "logs")
	require.True(t, ok)
	assert.Equal(t, int64(100), logDP.Value)
	metricDP, ok := dpForSignal(t, ms, "metrics")
	require.True(t, ok)
	assert.Equal(t, int64(200), metricDP.Value)
}
