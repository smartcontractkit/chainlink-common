package internal

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var _ sdklog.Exporter = (*otlploggrpc.Exporter)(nil)
var _ OTLPExporter = (*otlploggrpc.Exporter)(nil)

var _ sdklog.Exporter = (*otlploghttp.Exporter)(nil)
var _ OTLPExporter = (*otlploggrpc.Exporter)(nil)

// Copy of sdklog.Exporter interface, used for mocking
type OTLPExporter interface {
	Export(ctx context.Context, records []sdklog.Record) error
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}
