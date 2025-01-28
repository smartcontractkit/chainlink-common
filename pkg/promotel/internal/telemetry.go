package internal

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

// NewTelemetrySettings returns a new telemetry settings for Create* functions.
func NewTelemetrySettings(logger *zap.Logger) component.TelemetrySettings {
	l := zap.NewNop()
	if logger != nil {
		l = logger
	}
	return component.TelemetrySettings{
		Logger:         l,
		TracerProvider: nooptrace.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		MetricsLevel:   configtelemetry.LevelNone,
		Resource:       pcommon.NewResource(),
	}
}
