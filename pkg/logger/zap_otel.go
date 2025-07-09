package logger

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

const instrumentationLibName = "github.com/smartcontractkit/chainlink-common"

// OtelZapCore is a zapcore.Core implementation that exports logs to OpenTelemetry
// It implements the zapcore.Core interface and uses OpenTelemetry's logging API
type OtelZapCore struct {
	zapcore.Core

	logger       otellog.Logger
	fields       []zapcore.Field
	levelEnabler zapcore.LevelEnabler
}

type Option func(c *OtelZapCore)

// NewOtelCore initializes an OpenTelemetry Core for exporting logs in OTLP format
func NewOtelZapCore(loggerProvider otellog.LoggerProvider, opts ...Option) zapcore.Core {
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
	return o.levelEnabler.Enabled(level)
}

// With returns a new OpenTelemetry Core with the given fields added to the log entry
func (o OtelZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &OtelZapCore{
		logger:       o.logger,
		attributes:       append(o.fields, fields...),
		levelEnabler: o.levelEnabler,
	}
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
	var attributes []attribute.KeyValue
	var spanCtx *oteltrace.SpanContext

	// Add core-attached fields
	for _, f := range o.fields {
		if f.Key == "context" {
			if ctxValue, ok := f.Interface.(oteltrace.SpanContext); ok {
				spanCtx = &ctxValue
			}
		} else {
			attributes = append(attributes, mapZapField(f))
		}
	}

	// Add fields passed during log call
	for _, f := range fields {
		attributes = append(attributes, mapZapField(f))
	}

	// Add exception metadata
	if entry.Level > zapcore.InfoLevel {
		if entry.Caller.Defined {
			attributes = append(attributes, semconv.ExceptionType(entry.Caller.String()))
		}
		if entry.Stack != "" {
			attributes = append(attributes, semconv.ExceptionStacktrace(entry.Stack))
		}
	}

	// Add span context attributes if available
	if spanCtx != nil {
		if spanCtx.TraceID().IsValid() {
			attributes = append(attributes, attribute.String("trace_id", spanCtx.TraceID().String()))
		}
		if spanCtx.SpanID().IsValid() {
			attributes = append(attributes, attribute.String("span_id", spanCtx.SpanID().String()))
		}
		attributes = append(attributes, attribute.String("trace_flags", spanCtx.TraceFlags().String()))
	}

	// Convert to OpenTelemetry KeyValues and emit
	otelAttrs := make([]otellog.KeyValue, len(attributes))
	for i, attr := range attributes {
		otelAttrs[i] = otellog.KeyValue{
			Key:   string(attr.Key),
			Value: mapAttrValueToLogAttrValue(attr.Value),
		}
	}

	logRecord := otellog.Record{}
	logRecord.SetBody(otellog.StringValue(entry.Message))
	logRecord.SetSeverity(mapZapSeverity(entry.Level))
	logRecord.SetSeverityText(entry.Level.String())
	logRecord.SetTimestamp(entry.Time)
	logRecord.SetObservedTimestamp(time.Now())
	logRecord.AddAttributes(otelAttrs...)

	o.logger.Emit(context.Background(), logRecord)

	return nil
}

func mapZapField(f zapcore.Field) attribute.KeyValue {
	switch f.Type {
	case zapcore.StringType:
		return attribute.String(f.Key, f.String)

	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return attribute.Int64(f.Key, f.Integer)

	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type, zapcore.UintptrType:
		return attribute.Int64(f.Key, int64(f.Integer))

	case zapcore.BoolType:
		return attribute.Bool(f.Key, f.Integer == 1)

	case zapcore.Float64Type:
		return attribute.Float64(f.Key, math.Float64frombits(uint64(f.Integer)))

	case zapcore.ErrorType:
		if err, ok := f.Interface.(error); ok {
			return attribute.String(f.Key, err.Error())
		}
		return attribute.String(f.Key, "invalid error field")

	case zapcore.StringerType:
		return attribute.String(f.Key, f.Interface.(fmt.Stringer).String())

	case zapcore.TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return attribute.String(f.Key, t.Format(time.RFC3339))
		}
		return attribute.String(f.Key, fmt.Sprintf("invalid time: %v", f.Interface))

	case zapcore.DurationType:
		if d, ok := f.Interface.(time.Duration); ok {
			return attribute.String(f.Key, d.String())
		}
		return attribute.String(f.Key, fmt.Sprintf("invalid duration: %v", f.Interface))

	case zapcore.BinaryType:
		if b, ok := f.Interface.([]byte); ok {
			return attribute.String(f.Key, fmt.Sprintf("binary data: %x", b))
		}
		return attribute.String(f.Key, fmt.Sprintf("invalid binary: %v", f.Interface))

	case zapcore.ByteStringType:
		if b, ok := f.Interface.([]byte); ok {
			return attribute.String(f.Key, fmt.Sprintf("byte string: %x", b))
		}
		return attribute.String(f.Key, fmt.Sprintf("invalid byte string: %v", f.Interface))

	default:
		return attribute.String(f.Key, f.String)
	}
}

func mapZapSeverity(level zapcore.Level) otellog.Severity {
	switch level {
	case zapcore.DebugLevel:
		return otellog.SeverityDebug
	case zapcore.InfoLevel:
		return otellog.SeverityInfo
	case zapcore.WarnLevel:
		return otellog.SeverityWarn
	case zapcore.ErrorLevel:
		return otellog.SeverityError
	case zapcore.DPanicLevel:
		return otellog.SeverityFatal1
	case zapcore.PanicLevel:
		return otellog.SeverityFatal2
	case zapcore.FatalLevel:
		return otellog.SeverityFatal3
	default:
		return otellog.SeverityUndefined
	}
}

func mapAttrValueToLogAttrValue(v attribute.Value) otellog.Value {
	switch v.Type() {
	case attribute.STRING:
		return otellog.StringValue(v.AsString())
	case attribute.BOOL:
		return otellog.BoolValue(v.AsBool())
	case attribute.INT64:
		return otellog.Int64Value(v.AsInt64())
	case attribute.FLOAT64:
		return otellog.Float64Value(v.AsFloat64())
	case attribute.INVALID:
		return otellog.StringValue("invalid")
	case attribute.STRINGSLICE, attribute.BOOLSLICE, attribute.INT64SLICE, attribute.FLOAT64SLICE:
		return otellog.StringValue(fmt.Sprintf("%v", v.AsInterface()))
	default:
		return otellog.StringValue(fmt.Sprintf("%v", v.AsInterface()))
	}
}
