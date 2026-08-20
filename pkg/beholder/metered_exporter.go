package beholder

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/stats"
)

const (
	exportBytesMetric    = "beholder.export.bytes"
	exportDurationMetric = "beholder.export.duration"
)

const (
	signalLogs    = "logs"
	signalMetrics = "metrics"
	signalTraces  = "traces"
)

const csaPublicKeyNotConfigured = "not-configured"

// exportMetrics holds the instruments shared by the export stats handler and
// the metered exporters. They live on the beholder MeterProvider and are
// distinguished per signal.
type exportMetrics struct {
	bytes    otelmetric.Int64Counter
	duration otelmetric.Float64Histogram
}

// newExportMetrics creates the instruments shared by all metered exporters.
func newExportMetrics(meter otelmetric.Meter) (exportMetrics, error) {
	bytes, err := meter.Int64Counter(
		exportBytesMetric,
		otelmetric.WithDescription("Uncompressed OTLP proto size in bytes of each export batch. Recorded once per batch, on success only; retry attempts are not summed."),
		otelmetric.WithUnit("By"),
	)
	if err != nil {
		return exportMetrics{}, err
	}
	duration, err := meter.Float64Histogram(
		exportDurationMetric,
		otelmetric.WithDescription("Wall-clock duration in seconds of each OTLP export batch, covering all retry attempts and backoff. Recorded once per batch, on both success and failure."),
		otelmetric.WithUnit("s"),
		// Sized for network exports: sub-10ms to a 60s deadline. The SDK defaults
		// are millisecond-scaled, so nearly every export would land in bucket one.
		otelmetric.WithExplicitBucketBoundaries(
			0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60,
		),
	)
	if err != nil {
		return exportMetrics{}, err
	}
	return exportMetrics{bytes: bytes, duration: duration}, nil
}

// exportAttrs builds the attribute set identifying one signal's exports, plus
// any per-measurement extras.
func exportAttrs(signal, csaPublicKeyHex string, extra ...attribute.KeyValue) otelmetric.MeasurementOption {
	if csaPublicKeyHex == "" {
		csaPublicKeyHex = csaPublicKeyNotConfigured
	}
	attrs := make([]attribute.KeyValue, 0, 2+len(extra))
	attrs = append(attrs,
		attribute.String("otel_signal", signal),
		attribute.String("csa_public_key", csaPublicKeyHex),
	)
	return otelmetric.WithAttributes(append(attrs, extra...)...)
}

// meteredExporter holds the shared metering logic for all signals.
type meteredExporter struct {
	metrics exportMetrics

	okAttrs  otelmetric.MeasurementOption // otel_signal, csa_public_key, error=false
	errAttrs otelmetric.MeasurementOption // otel_signal, csa_public_key, error=true
}

func newBaseExporter(metrics exportMetrics, signal, csaPublicKeyHex string) meteredExporter {
	return meteredExporter{
		metrics:  metrics,
		okAttrs:  exportAttrs(signal, csaPublicKeyHex, attribute.Bool("error", false)),
		errAttrs: exportAttrs(signal, csaPublicKeyHex, attribute.Bool("error", true)),
	}
}

func (m meteredExporter) record(ctx context.Context, export func() error) error {
	start := time.Now()
	err := export()
	elapsed := time.Since(start).Seconds()

	if err == nil {
		m.metrics.duration.Record(ctx, elapsed, m.okAttrs)
	} else {
		m.metrics.duration.Record(ctx, elapsed, m.errAttrs)
	}
	return err
}

// meteredLogExporter wraps an sdklog.Exporter and records each export batch's
// duration. It sits above the otlploggrpc retry loop, so Export is called once
// per logical batch and the duration covers the whole retry sequence. Bytes are
// recorded by exportStatsHandler inside the gRPC stack.
type meteredLogExporter struct {
	meteredExporter
	inner sdklog.Exporter
}

func newMeteredLogExporter(inner sdklog.Exporter, metrics exportMetrics, csaPublicKeyHex string) *meteredLogExporter {
	return &meteredLogExporter{
		meteredExporter: newBaseExporter(metrics, signalLogs, csaPublicKeyHex),
		inner:           inner,
	}
}

func (e *meteredLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	return e.record(ctx, func() error { return e.inner.Export(ctx, records) })
}

func (e *meteredLogExporter) Shutdown(ctx context.Context) error { return e.inner.Shutdown(ctx) }

func (e *meteredLogExporter) ForceFlush(ctx context.Context) error { return e.inner.ForceFlush(ctx) }

// lazyMetered defers duration instrument wiring for exporters constructed
// before the beholder MeterProvider exists
type lazyMetered struct {
	signal string
	base   atomic.Pointer[meteredExporter]
}

// attachMetrics wires the export instruments once the MeterProvider exists.
func (l *lazyMetered) attachMetrics(metrics exportMetrics, csaPublicKeyHex string) {
	base := newBaseExporter(metrics, l.signal, csaPublicKeyHex)
	l.base.Store(&base)
}

func (l *lazyMetered) record(ctx context.Context, export func() error) error {
	base := l.base.Load()
	if base == nil {
		return export()
	}
	return base.record(ctx, export)
}

// meteredMetricExporter wraps an sdkmetric.Exporter.
// It is created by the MeterProvider and has no access to the instruments until
// the MeterProvider exists and calls attachMetrics.
type meteredMetricExporter struct {
	sdkmetric.Exporter
	lazyMetered
}

func newMeteredMetricExporter(inner sdkmetric.Exporter) *meteredMetricExporter {
	return &meteredMetricExporter{Exporter: inner, lazyMetered: lazyMetered{signal: signalMetrics}}
}

func (e *meteredMetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	return e.record(ctx, func() error { return e.Exporter.Export(ctx, rm) })
}

type meteredTraceExporter struct {
	sdktrace.SpanExporter
	lazyMetered
}

func newMeteredTraceExporter(inner sdktrace.SpanExporter) *meteredTraceExporter {
	return &meteredTraceExporter{SpanExporter: inner, lazyMetered: lazyMetered{signal: signalTraces}}
}

func (e *meteredTraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return e.record(ctx, func() error { return e.SpanExporter.ExportSpans(ctx, spans) })
}

// rpcStateKey is the handler's private context key.
type rpcStateKey struct{}

// rpcState accumulates one RPC's outbound bytes.
type rpcState struct {
	outBytes atomic.Int64
}

// boundInstruments is the handler's metric wiring.
type boundInstruments struct {
	bytes otelmetric.Int64Counter
	attrs otelmetric.MeasurementOption
}

type exportStatsHandler struct {
	signal          string
	csaPublicKeyHex string

	// nil until attachMetrics; the handler is registered at dial time but the
	// instruments only exist once the beholder MeterProvider does.
	instruments atomic.Pointer[boundInstruments]
}

func newExportStatsHandler(signal, csaPublicKeyHex string) *exportStatsHandler {
	return &exportStatsHandler{signal: signal, csaPublicKeyHex: csaPublicKeyHex}
}

// attachMetrics wires the bytes instrument once the MeterProvider exists. Until
// it is called the handler is an inert no-op.
func (h *exportStatsHandler) attachMetrics(metrics exportMetrics) {
	h.instruments.Store(&boundInstruments{
		bytes: metrics.bytes,
		attrs: exportAttrs(h.signal, h.csaPublicKeyHex),
	})
}

// TagRPC allocates this RPC's byte accumulator. grpc-go threads the returned
// context through every subsequent HandleRPC call for this RPC.
func (h *exportStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return context.WithValue(ctx, rpcStateKey{}, &rpcState{})
}

// HandleRPC accumulates outbound proto length per RPC and records it when the
// RPC ends successfully.
func (h *exportStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	state, ok := ctx.Value(rpcStateKey{}).(*rpcState)
	if !ok {
		return
	}
	switch e := rs.(type) {
	case *stats.OutPayload:
		state.outBytes.Add(int64(e.Length))
	case *stats.End:
		// Swap keeps this idempotent if End is ever delivered twice.
		n := state.outBytes.Swap(0)
		if e.Error != nil || n == 0 {
			return
		}
		if inst := h.instruments.Load(); inst != nil {
			inst.bytes.Add(ctx, n, inst.attrs)
		}
	}
}

func (h *exportStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (h *exportStatsHandler) HandleConn(context.Context, stats.ConnStats) {}
