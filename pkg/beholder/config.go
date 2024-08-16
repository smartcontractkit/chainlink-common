package beholder

import (
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"
)

type Config struct {
	InsecureConnection       bool
	CACertFile               string
	OtelExporterGRPCEndpoint string

	// OTel Resource
	ResourceAttributes []otelattr.KeyValue
	// Message Emitter
	EmitterExportTimeout time.Duration
	// OTel Trace
	TraceSampleRate   float64
	TraceBatchTimeout time.Duration
	// OTel Metric
	MetricReaderInterval time.Duration
	// OTel Log
	LogExportTimeout time.Duration
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
		EmitterExportTimeout: 1 * time.Second,
		// Trace
		TraceSampleRate:   1,
		TraceBatchTimeout: 1 * time.Second,
		// Metric
		MetricReaderInterval: 1 * time.Second,
		// Log
		LogExportTimeout: 1 * time.Second,
	}
}
