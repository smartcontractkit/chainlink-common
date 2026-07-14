package beholder

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc/stats"
)

// sizeCapture is shared between sizeCaptureHandler and meteredLogsExporter.
type sizeCapture struct {
	val atomic.Int64
}

// sizeCaptureHandler is a minimal gRPC stats.Handler that records the
// uncompressed proto size of each outbound message. It reproduces the exact
// measurement that otelgrpc made via OutPayload.Length before semconv v1.40.0
// removed the rpc.client.request.size metric.
type sizeCaptureHandler struct {
	capture *sizeCapture
}

func (h *sizeCaptureHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (*sizeCaptureHandler) HandleConn(context.Context, stats.ConnStats) {}

func (h *sizeCaptureHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC fires on every gRPC stats event. On OutPayload it stores the
// uncompressed message length, the same field otelgrpc used for
// rpc.client.request.size.
func (h *sizeCaptureHandler) HandleRPC(_ context.Context, rs stats.RPCStats) {
	op, ok := rs.(*stats.OutPayload)
	if !ok {
		return
	}
	h.capture.val.Store(int64(op.Length))
}

// meteredLogsExporter wraps an sdklog.Exporter and records the uncompressed
// OTLP proto size of each export batch as a metric.
type meteredLogsExporter struct {
	inner   sdklog.Exporter
	counter otelmetric.Int64Counter
	attrs   otelmetric.MeasurementOption
	capture *sizeCapture
}

func newMeteredLogsExporter(
	inner sdklog.Exporter,
	meter otelmetric.Meter,
	csaPublicKeyHex string,
	capture *sizeCapture,
) (*meteredLogsExporter, error) {
	counter, err := meter.Int64Counter(
		"beholder.logs.export.bytes",
		otelmetric.WithDescription("Uncompressed OTLP proto size in bytes of each export batch of beholder logs."),
		otelmetric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}
	return &meteredLogsExporter{
		inner:   inner,
		counter: counter,
		attrs:   otelmetric.WithAttributes(attribute.String("csa_public_key", csaPublicKeyHex)),
		capture: capture,
	}, nil
}

func (e *meteredLogsExporter) Export(ctx context.Context, records []sdklog.Record) error {
	err := e.inner.Export(ctx, records)
	if err == nil {
		e.counter.Add(ctx, e.capture.val.Load(), e.attrs)
	}
	return err
}

func (e *meteredLogsExporter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

func (e *meteredLogsExporter) ForceFlush(ctx context.Context) error {
	return e.inner.ForceFlush(ctx)
}
