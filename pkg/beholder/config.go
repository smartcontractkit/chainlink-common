package beholder

import (
	"time"

	otelattr "go.opentelemetry.io/otel/attribute"
)

type Config struct {
	InsecureConnection       bool
	CACertFile               string
	OtelExporterGRPCEndpoint string

	PackageName string
	// OTel Resource
	ResourceAttributes map[string]string
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

var defaultOtelAttributes = map[string]string{
	"package_name": "beholder",
}

func DefaultConfig() Config {
	return Config{
		InsecureConnection:       true,
		CACertFile:               "",
		OtelExporterGRPCEndpoint: "localhost:4317",
		PackageName:              "beholder",
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

func (c Config) Attributes() []otelattr.KeyValue {
	attrs := make([]otelattr.KeyValue, 0, len(c.ResourceAttributes))
	for k, v := range c.ResourceAttributes {
		attrs = append(attrs, otelattr.String(k, v))
	}
	return attrs
}
