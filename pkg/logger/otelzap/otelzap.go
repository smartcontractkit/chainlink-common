package otelzap

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"go.uber.org/zap/zapcore"
)

// OtelZapCore is a zapcore.Core implementation that exports logs to OpenTelemetry
// It implements the zapcore.Core interface and uses OpenTelemetry's logging API
type OtelZapCore struct {
	logger       otellog.Logger
	fields       []zapcore.Field
	levelEnabler zapcore.LevelEnabler
}

type Option func(c *OtelZapCore)

// NewOtelCore initializes an OpenTelemetry Core for exporting logs in OTLP format
func NewCore(logger otellog.Logger, opts ...Option) zapcore.Core {

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
	newFields := append([]zapcore.Field{}, o.fields...)
	newFields = append(newFields, fields...)

	return &OtelZapCore{
		logger:       o.logger,
		fields:       newFields,
		levelEnabler: o.levelEnabler,
	}
}

// Check checks if the given log entry is enabled for the OpenTelemetry Core
func (o OtelZapCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if o.Enabled(entry.Level) {
		checked = checked.AddCore(entry, o)
	}
	return checked
}

func (o OtelZapCore) Sync() error {
	// If no zap core is set, we don't need to sync anything
	// as OpenTelemetry Core does not have a sync operation.
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
	encoder := &otelAttrEncoder{}
	var spanCtx *oteltrace.SpanContext

	// Add core-attached fields
	for _, f := range o.fields {
		if f.Key == "context" {
			if ctxValue, ok := f.Interface.(oteltrace.SpanContext); ok {
				spanCtx = &ctxValue
			}
		} else {
			f.AddTo(encoder)
		}
	}

	// Add fields passed during log call
	for _, f := range fields {
		f.AddTo(encoder)
	}

	// Start with encoder attributes
	attributes := encoder.attributes

	// Add caller information if available
	if entry.Caller.Defined {
		attributes = append(attributes, attribute.String("caller", entry.Caller.String()))
	}

	// Add exception metadata for error levels
	if entry.Level > zapcore.InfoLevel {
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
