package beholder

import (
	"time"
)

type Config struct {
	OtelExporterGRPCEndpoint string
	PackageName              string
	EventEmitterRetryCount   uint
	EventEmitterRetryDelay   time.Duration
}

func DefaultBeholderConfig() Config {
	return Config{
		OtelExporterGRPCEndpoint: "localhost:4317",
		PackageName:              "beholder",
		EventEmitterRetryCount:   3,
		EventEmitterRetryDelay:   100 * time.Millisecond,
	}
}

func (c Config) Attributes() map[string]interface{} {
	return map[string]interface{}{
		"otel_exporter_otlp_endpoint": c.OtelExporterGRPCEndpoint,
	}
}
