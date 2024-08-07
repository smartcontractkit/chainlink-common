package beholder

import (
	"time"
)

type Config struct {
	OtelExporterGRPCEndpoint string
	PackageName              string
	MessageEmitterRetryCount uint
	MessageEmitterRetryDelay time.Duration
}

func DefaultBeholderConfig() Config {
	return Config{
		OtelExporterGRPCEndpoint: "localhost:4317",
		PackageName:              "beholder",
		MessageEmitterRetryCount: 3,
		MessageEmitterRetryDelay: 100 * time.Millisecond,
	}
}

func (c Config) Attributes() map[string]interface{} {
	return map[string]interface{}{
		"otel_exporter_otlp_endpoint": c.OtelExporterGRPCEndpoint,
	}
}
