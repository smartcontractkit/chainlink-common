package internal

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// NewTelemetrySettings returns a new telemetry settings for Create* functions.
func NewTelemetrySettings(lggr logger.Logger) component.TelemetrySettings {
	return component.TelemetrySettings{
		Logger:         mustZapLogger(lggr),
		TracerProvider: nooptrace.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		MetricsLevel:   configtelemetry.LevelNone,
		Resource:       pcommon.NewResource(),
	}
}

type zapGetter interface {
	ToZapLogger(lggr logger.Logger) *zap.Logger
}

func mustZapLogger(lggr logger.Logger) *zap.Logger {
	if g, ok := lggr.(zapGetter); ok {
		return g.ToZapLogger(lggr)
	}
	l, err := zap.NewProduction()
	if err != nil {
		return zap.NewNop()
	}
	return l
}
