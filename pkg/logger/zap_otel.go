package logger

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"go.opencensus.io/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	otel "go.opentelemetry.io/otel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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
	return o.levelEnabler.Enabled(level)
}

// With returns a new OpenTelemetry Core with the given fields added to the log entry
func (o OtelZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &OtelZapCore{
		logger:       o.logger,
		fields:       append(o.fields, fields...),
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
	var spanCtx *trace.SpanContext

	// Add core-attached fields
	for _, f := range o.fields {
		if f.Key == "context" {
			if ctxValue, ok := f.Interface.(trace.SpanContext); ok {
				spanCtx = &ctxValue
			}
		} else {
			attributes = append(attributes, zapFieldToAttr(f))
		}
	}

	// Add fields passed during log call
	for _, f := range fields {
		attributes = append(attributes, zapFieldToAttr(f))
	}

	// Add severity and basic metadata
	attributes = append(attributes,
		attribute.String("log.severity", entry.Level.String()),
		attribute.String("log.message", entry.Message),
		attribute.String("logger.name", entry.LoggerName),
	)

	// Add caller info
	if entry.Caller.Defined {
		attributes = append(attributes,
			attribute.String("code.filepath", entry.Caller.File),
			attribute.Int("code.line_number", entry.Caller.Line),
			attribute.String("code.function", entry.Caller.Function),
		)
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
	otelAttrs := make([]otel.KeyValue, len(attributes))
	for i, attr := range attributes {
		otelAttrs[i] = otel.KeyValue(attr)
	}

	o.logger.Emit(context.Background(), []byte(entry.Message), otelAttrs)

	return nil
}

func zapFieldToAttr(f zapcore.Field) attribute.KeyValue {
	switch f.Type {
	case zapcore.StringType:
		return attribute.String(f.Key, f.String)
	case zapcore.BoolType:
		return attribute.Bool(f.Key, f.Integer == 1)
	case zapcore.Int64Type:
		return attribute.Int64(f.Key, f.Integer)
	case zapcore.Uint64Type:
		return attribute.Int64(f.Key, int64(f.Integer))
	case zapcore.Float64Type:
		return attribute.Float64(f.Key, math.Float64frombits(uint64(f.Integer)))
	case zapcore.ErrorType:
		return attribute.String(f.Key, f.Interface.(error).Error())
	case zapcore.StringerType:
		return attribute.String(f.Key, f.Interface.(fmt.Stringer).String())
	case zapcore.ReflectType, zapcore.ObjectMarshalerType:
		return attribute.String(f.Key, fmt.Sprintf("%v", f.Interface))
	default:
		return attribute.String(f.Key, f.String)
	}
}

func convertLevel(level zapcore.Level) log.Severity {
	switch level {
	case zapcore.DebugLevel:
		return log.SeverityDebug
	case zapcore.InfoLevel:
		return log.SeverityInfo
	case zapcore.WarnLevel:
		return log.SeverityWarn
	case zapcore.ErrorLevel:
		return log.SeverityError
	case zapcore.DPanicLevel:
		return log.SeverityFatal1
	case zapcore.PanicLevel:
		return log.SeverityFatal2
	case zapcore.FatalLevel:
		return log.SeverityFatal3
	default:
		return log.SeverityUndefined
	}
}

func convertFields(fields []zapcore.Field) []log.KeyValue {
	kvs := make([]log.KeyValue, 0, len(fields)+numExtraAttr)
	for _, field := range fields {
		kvs = appendField(kvs, field)
	}
	return kvs
}

func appendField(kvs []log.KeyValue, f zapcore.Field) []log.KeyValue {
	switch f.Type {
	case zapcore.BoolType:
		return append(kvs, log.Bool(f.Key, f.Integer == 1))

	case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type,
		zapcore.Uint32Type, zapcore.Uint8Type, zapcore.Uint16Type, zapcore.Uint64Type,
		zapcore.UintptrType:
		return append(kvs, log.Int64(f.Key, f.Integer))

	case zapcore.Float64Type:
		num := math.Float64frombits(uint64(f.Integer))
		return append(kvs, log.Float64(f.Key, num))
	case zapcore.Float32Type:
		num := math.Float32frombits(uint32(f.Integer))
		return append(kvs, log.Float64(f.Key, float64(num)))

	case zapcore.Complex64Type:
		str := strconv.FormatComplex(complex128(f.Interface.(complex64)), 'E', -1, 64)
		return append(kvs, log.String(f.Key, str))
	case zapcore.Complex128Type:
		str := strconv.FormatComplex(f.Interface.(complex128), 'E', -1, 128)
		return append(kvs, log.String(f.Key, str))

	case zapcore.StringType:
		return append(kvs, log.String(f.Key, f.String))
	case zapcore.BinaryType, zapcore.ByteStringType:
		bs := f.Interface.([]byte)
		return append(kvs, log.Bytes(f.Key, bs))
	case zapcore.StringerType:
		str := f.Interface.(fmt.Stringer).String()
		return append(kvs, log.String(f.Key, str))

	case zapcore.DurationType, zapcore.TimeType:
		return append(kvs, log.Int64(f.Key, f.Integer))
	case zapcore.TimeFullType:
		str := f.Interface.(time.Time).Format(time.RFC3339Nano)
		return append(kvs, log.String(f.Key, str))
	case zapcore.ErrorType:
		err := f.Interface.(error)
		typ := reflect.TypeOf(err).String()
		kvs = append(kvs, log.String("exception.type", typ))
		kvs = append(kvs, log.String("exception.message", err.Error()))
		return kvs
	case zapcore.ReflectType:
		str := fmt.Sprint(f.Interface)
		return append(kvs, log.String(f.Key, str))
	case zapcore.SkipType:
		return kvs

	case zapcore.ArrayMarshalerType:
		kv := log.String(f.Key+"_error", "otelzap: zapcore.ArrayMarshalerType is not implemented")
		return append(kvs, kv)
	case zapcore.ObjectMarshalerType:
		kv := log.String(f.Key+"_error", "otelzap: zapcore.ObjectMarshalerType is not implemented")
		return append(kvs, kv)

	default:
		kv := log.String(f.Key+"_error", fmt.Sprintf("otelzap: unknown field type: %v", f))
		return append(kvs, kv)
	}
}
