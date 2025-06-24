package logger

import (
	otel "go.opentelemetry.io/otel/log"
	"go.uber.org/zap/zapcore"
)

const instrumentationLibName = "github.com/smartcontractkit/chainlink-common"

// todo - do we need this interface?
type ZapOtelLogger interface {
	Logger
}

type OtelZapCore struct {
	zapcore.Core

	logger       otel.Logger
	fields       []zapcore.Field
	levelEnabler zapcore.LevelEnabler
}

type Option func(c *OtelZapCore)

// NewOtelCore initializes an OpenTelemetry Core for exporting logs in OTLP format
func NewOtelZapCore(loggerProvider otel.LoggerProvider, opts ...Option) zapcore.Core {
	logger := loggerProvider.Logger(instrumentationLibName)

	c := &OtelZapCore{
		logger:       logger,
		levelEnabler: zapcore.InfoLevel,
	}
	for _, apply := range opts {
		apply(c)
	}

	return c
}

// Enabled checks if the given log level is enabled for the OpenTelemetry Core
func (o OtelZapCore) Enabled(level zapcore.Level) bool {
	//TODO implement me
	panic("implement me")
}

// With returns a new OpenTelemetry Core with the given fields added to the log entry
func (o OtelZapCore) With(fields []zapcore.Field) zapcore.Core {
	core := OtelZapCore{
		logger:       o.logger,
		fields:       append(o.fields, fields...),
		levelEnabler: o.levelEnabler,
	}
	return &core
}

// Check checks if the given log entry is enabled for the OpenTelemetry Core
func (o OtelZapCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if o.Enabled(entry.Level) {
		return checked.AddCore(entry, o)
	}
	return checked
}

func (o OtelZapCore) Sync() error {
	// OpenTelemetry does not require a sync operation like zap does
	return nil
}

// WithLevel sets the minimum level for the OpenTelemetry Core log to be exported
func WithLevel(level zapcore.Level) Option {
	return Option(func(o *OtelZapCore) {
		o.levelEnabler = level
	})
}

// WithLevelEnabler sets the zapcore.LevelEnabler for determining which log levels to export
func WithLevelEnabler(levelEnabler zapcore.LevelEnabler) Option {
	return Option(func(o *OtelZapCore) {
		o.levelEnabler = levelEnabler
	})
}

func (o OtelZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	//TODO implement me
	panic("implement me")
}
