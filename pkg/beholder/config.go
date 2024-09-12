package beholder

import (
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	InsecureConnection       bool
	CACertFile               string
	OtelExporterGRPCEndpoint string

	// OTel Resource
	ResourceAttributes []otelattr.KeyValue
	// Message Emitter
	EmitterExportTimeout time.Duration
	// Batch processing is enabled by default
	// Disable it only for testing
	EmitterBatchProcessor bool
	// OTel Trace
	TraceSampleRatio  float64
	TraceBatchTimeout time.Duration
	TraceSpanExporter sdktrace.SpanExporter // optional additional exporter
	// OTel Metric
	MetricReaderInterval time.Duration
	// OTel Log
	LogExportTimeout time.Duration
	// Batch processing is enabled by default
	// Disable it only for testing
	LogBatchProcessor bool
}

const (
	defaultPackageName = "beholder"
)

var defaultOtelAttributes = []otelattr.KeyValue{
	otelattr.String("package_name", "beholder"),
}

func DefaultConfig() Config {
	return Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		// Resource
		ResourceAttributes: defaultOtelAttributes,
		// Message Emitter
		EmitterExportTimeout:  1 * time.Second,
		EmitterBatchProcessor: true,
		// Trace
		TraceSampleRatio:  1,
		TraceBatchTimeout: 1 * time.Second,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// Log
		LogExportTimeout:  1 * time.Second,
		LogBatchProcessor: true,
	}
}

func TestDefaultConfig() Config {
	config := DefaultConfig()
	// Should be only disabled for testing
	config.EmitterBatchProcessor = false
	config.LogBatchProcessor = false
	return config
}
