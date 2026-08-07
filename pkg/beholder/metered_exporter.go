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
	"google.golang.org/grpc/stats"
)

// exportSizeKey is the context key under which a metered exporter stashes a
// per-export byte holder for beholderStatsHandler to fill in.
type exportSizeKey struct{}

// beholderStatsHandler is a minimal, stateless gRPC stats.Handler that records the
// uncompressed proto size of each outbound message.
type beholderStatsHandler struct{}

func (beholderStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (beholderStatsHandler) HandleConn(context.Context, stats.ConnStats) {}

func (beholderStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC fires on every gRPC stats event. On OutPayload it stores the
// uncompressed message length, the same field otelgrpc used for
// rpc.client.request.size
func (beholderStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	op, ok := rs.(*stats.OutPayload)
	if !ok {
		return
	}
	if holder, ok := ctx.Value(exportSizeKey{}).(*atomic.Int64); ok {
		holder.Store(int64(op.Length))
	}
}

const (
	exportBytesMetric    = "beholder.export.bytes"
	exportDurationMetric = "beholder.export.duration"
)

// exportMetrics holds the instruments shared by all metered exporters. They live
// on the beholder MeterProvider and are distinguished per exporter only by
// attributes, so one set covers every signal.
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
	attrs := make([]attribute.KeyValue, 0, 2+len(extra))
	attrs = append(attrs,
		attribute.String("otel_signal", signal),
		attribute.String("csa_public_key", csaPublicKeyHex),
	)
	return otelmetric.WithAttributes(append(attrs, extra...)...)
}

// meteredExporter holds the shared metering logic: run an export with a per-call
// size holder in the context, then record the captured OutPayload size on
// success and the duration either way. Attribute options are
// precomputed so the export path allocates nothing per call.
type meteredExporter struct {
	metrics exportMetrics

	byteAttrs otelmetric.MeasurementOption // otel_signal, csa_public_key
	okAttrs   otelmetric.MeasurementOption // + error=false
	errAttrs  otelmetric.MeasurementOption // + error=true
}

func newBaseExporter(metrics exportMetrics, signal, csaPublicKeyHex string) meteredExporter {
	return meteredExporter{
		metrics:   metrics,
		byteAttrs: exportAttrs(signal, csaPublicKeyHex),
		okAttrs:   exportAttrs(signal, csaPublicKeyHex, attribute.Bool("error", false)),
		errAttrs:  exportAttrs(signal, csaPublicKeyHex, attribute.Bool("error", true)),
	}
}

func (m meteredExporter) record(ctx context.Context, export func(context.Context) error) error {
	var size atomic.Int64
	start := time.Now()
	err := export(context.WithValue(ctx, exportSizeKey{}, &size))
	elapsed := time.Since(start).Seconds()

	// Bytes are only meaningful for a batch that landed; duration is recorded
	// either way
	if err == nil {
		m.metrics.bytes.Add(ctx, size.Load(), m.byteAttrs)
		m.metrics.duration.Record(ctx, elapsed, m.okAttrs)
	} else {
		m.metrics.duration.Record(ctx, elapsed, m.errAttrs)
	}
	return err
}

// meteredLogExporter wraps an sdklog.Exporter and records each export batch's
// uncompressed proto size and duration. It sits above the otlploggrpc retry
// loop, so Export is called once per logical batch: bytes are counted only on
// success, and the duration covers the whole retry sequence.
type meteredLogExporter struct {
	meteredExporter
	inner sdklog.Exporter
}

func newMeteredLogExporter(inner sdklog.Exporter, metrics exportMetrics, csaPublicKeyHex string) *meteredLogExporter {
	return &meteredLogExporter{
		meteredExporter: newBaseExporter(metrics, "logs", csaPublicKeyHex),
		inner:           inner,
	}
}

func (e *meteredLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	return e.record(ctx, func(c context.Context) error { return e.inner.Export(c, records) })
}

func (e *meteredLogExporter) Shutdown(ctx context.Context) error { return e.inner.Shutdown(ctx) }

func (e *meteredLogExporter) ForceFlush(ctx context.Context) error { return e.inner.ForceFlush(ctx) }

// meteredMetricExporter wraps an sdkmetric.Exporter.
// It is created by the MeterProvider and has no access to the instruments until
// the MeterProvider exists and calls attachMetrics.
type meteredMetricExporter struct {
	sdkmetric.Exporter
	base atomic.Pointer[meteredExporter]
}

func newMeteredMetricExporter(inner sdkmetric.Exporter) *meteredMetricExporter {
	return &meteredMetricExporter{Exporter: inner}
}

// attachMetrics wires the export instruments once the MeterProvider exists.
func (e *meteredMetricExporter) attachMetrics(metrics exportMetrics, csaPublicKeyHex string) {
	base := newBaseExporter(metrics, "metrics", csaPublicKeyHex)
	e.base.Store(&base)
}

func (e *meteredMetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	base := e.base.Load()
	if base == nil {
		return e.Exporter.Export(ctx, rm)
	}
	return base.record(ctx, func(c context.Context) error { return e.Exporter.Export(c, rm) })
}
