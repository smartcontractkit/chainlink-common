package beholder

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc/stats"
)

// exportSizeKey is the context key under which a metered exporter stashes a
// per-export byte holder for sizeCaptureHandler to fill in.
type exportSizeKey struct{}

// sizeCaptureHandler is a minimal, stateless gRPC stats.Handler that records the
// uncompressed proto size of each outbound message.
type sizeCaptureHandler struct{}

func (sizeCaptureHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (sizeCaptureHandler) HandleConn(context.Context, stats.ConnStats) {}

func (sizeCaptureHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC fires on every gRPC stats event. On OutPayload it stores the
// uncompressed message length, the same field otelgrpc used for
// rpc.client.request.size
func (sizeCaptureHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	op, ok := rs.(*stats.OutPayload)
	if !ok {
		return
	}
	if holder, ok := ctx.Value(exportSizeKey{}).(*atomic.Int64); ok {
		holder.Store(int64(op.Length))
	}
}

const exportBytesMetric = "beholder.export.bytes"

// newExportBytesCounter creates the counter shared by all metered exporters.
func newExportBytesCounter(meter otelmetric.Meter) (otelmetric.Int64Counter, error) {
	return meter.Int64Counter(
		exportBytesMetric,
		otelmetric.WithDescription("Uncompressed OTLP proto size in bytes of each successful export batch, by signal."),
		otelmetric.WithUnit("By"),
	)
}

func exportAttrs(signal, csaPublicKeyHex string) otelmetric.MeasurementOption {
	return otelmetric.WithAttributes(
		attribute.String("otel_signal", signal),
		attribute.String("csa_public_key", csaPublicKeyHex),
	)
}

// meteredExporter holds the shared metering logic: run an export with a per-call
// size holder in the context, then record the captured OutPayload size on
// success.
type meteredExporter struct {
	counter otelmetric.Int64Counter
	attrs   otelmetric.MeasurementOption
}

func (m meteredExporter) record(ctx context.Context, export func(context.Context) error) error {
	var size atomic.Int64
	err := export(context.WithValue(ctx, exportSizeKey{}, &size))
	if err == nil {
		m.counter.Add(ctx, size.Load(), m.attrs)
	}
	return err
}

// meteredLogExporter wraps an sdklog.Exporter and records each export batch's
// uncompressed proto size. It sits above the otlploggrpc retry loop, so Export is
// called once per logical batch and bytes are counted only on success.
type meteredLogExporter struct {
	meteredExporter
	inner sdklog.Exporter
}

func newMeteredLogExporter(inner sdklog.Exporter, counter otelmetric.Int64Counter, csaPublicKeyHex string) *meteredLogExporter {
	return &meteredLogExporter{
		meteredExporter: meteredExporter{counter: counter, attrs: exportAttrs("logs", csaPublicKeyHex)},
		inner:           inner,
	}
}

func (e *meteredLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	return e.record(ctx, func(c context.Context) error { return e.inner.Export(c, records) })
}

func (e *meteredLogExporter) Shutdown(ctx context.Context) error { return e.inner.Shutdown(ctx) }

func (e *meteredLogExporter) ForceFlush(ctx context.Context) error { return e.inner.ForceFlush(ctx) }

// meteredMetricExporter wraps an sdkmetric.Exporter.
// It is created by the MeterProvider and has no access to the counter until
// the MeterProvider exits and calls attachCounter.
type meteredMetricExporter struct {
	sdkmetric.Exporter
	base atomic.Pointer[meteredExporter]
}

func newMeteredMetricExporter(inner sdkmetric.Exporter) *meteredMetricExporter {
	return &meteredMetricExporter{Exporter: inner}
}

// attachCounter wires the size counter once the MeterProvider exits.
func (e *meteredMetricExporter) attachCounter(counter otelmetric.Int64Counter, csaPublicKeyHex string) {
	e.base.Store(&meteredExporter{counter: counter, attrs: exportAttrs("metrics", csaPublicKeyHex)})
}

func (e *meteredMetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	base := e.base.Load()
	if base == nil {
		return e.Exporter.Export(ctx, rm)
	}
	return base.record(ctx, func(c context.Context) error { return e.Exporter.Export(c, rm) })
}
