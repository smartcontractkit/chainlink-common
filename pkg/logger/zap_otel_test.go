package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

type stringerMock struct{}

func (s stringerMock) String() string { return "stringer-value" }

func Test_mapZapField(t *testing.T) {
	now := time.Now()
	duration := time.Second * 42

	tests := []struct {
		name     string
		field    zapcore.Field
		expected attribute.KeyValue
	}{
		{
			name:     "StringType",
			field:    zapcore.Field{Key: "str", Type: zapcore.StringType, String: "foo"},
			expected: attribute.String("str", "foo"),
		},
		{
			name:     "Int64Type",
			field:    zapcore.Field{Key: "int64", Type: zapcore.Int64Type, Integer: 42},
			expected: attribute.Int64("int64", 42),
		},
		{
			name:     "Uint64Type",
			field:    zapcore.Field{Key: "uint64", Type: zapcore.Uint64Type, Integer: 99},
			expected: attribute.Int64("uint64", 99),
		},
		{
			name:     "BoolType true",
			field:    zapcore.Field{Key: "bool", Type: zapcore.BoolType, Integer: 1},
			expected: attribute.Bool("bool", true),
		},
		{
			name:     "BoolType false",
			field:    zapcore.Field{Key: "bool", Type: zapcore.BoolType, Integer: 0},
			expected: attribute.Bool("bool", false),
		},
		{
			name:     "Float64Type",
			field:    zapcore.Field{Key: "float", Type: zapcore.Float64Type, Integer: int64(math.Float64bits(3.14))},
			expected: attribute.Float64("float", 3.14),
		},
		{
			name:     "ErrorType",
			field:    zapcore.Field{Key: "err", Type: zapcore.ErrorType, Interface: errors.New("fail")},
			expected: attribute.String("err", "fail"),
		},
		{
			name:     "ErrorType invalid",
			field:    zapcore.Field{Key: "err", Type: zapcore.ErrorType, Interface: "not-an-error"},
			expected: attribute.String("err", "invalid error field"),
		},
		{
			name:     "StringerType",
			field:    zapcore.Field{Key: "stringer", Type: zapcore.StringerType, Interface: stringerMock{}},
			expected: attribute.String("stringer", "stringer-value"),
		},
		{
			name:     "TimeType valid",
			field:    zapcore.Field{Key: "time", Type: zapcore.TimeType, Interface: now},
			expected: attribute.String("time", now.Format(time.RFC3339)),
		},
		{
			name:     "TimeType invalid",
			field:    zapcore.Field{Key: "time", Type: zapcore.TimeType, Interface: "not-a-time"},
			expected: attribute.String("time", "invalid time: not-a-time"),
		},
		{
			name:     "DurationType valid",
			field:    zapcore.Field{Key: "dur", Type: zapcore.DurationType, Interface: duration},
			expected: attribute.String("dur", duration.String()),
		},
		{
			name:     "DurationType invalid",
			field:    zapcore.Field{Key: "dur", Type: zapcore.DurationType, Interface: "not-a-duration"},
			expected: attribute.String("dur", "invalid duration: not-a-duration"),
		},
		{
			name:     "BinaryType valid",
			field:    zapcore.Field{Key: "bin", Type: zapcore.BinaryType, Interface: []byte{0x1, 0x2}},
			expected: attribute.String("bin", "binary data: 0102"),
		},
		{
			name:     "BinaryType invalid",
			field:    zapcore.Field{Key: "bin", Type: zapcore.BinaryType, Interface: "not-bytes"},
			expected: attribute.String("bin", "invalid binary: not-bytes"),
		},
		{
			name:     "ByteStringType valid",
			field:    zapcore.Field{Key: "bs", Type: zapcore.ByteStringType, Interface: []byte{0x3, 0x4}},
			expected: attribute.String("bs", "byte string: 0304"),
		},
		{
			name:     "ByteStringType invalid",
			field:    zapcore.Field{Key: "bs", Type: zapcore.ByteStringType, Interface: "not-bytes"},
			expected: attribute.String("bs", "invalid byte string: not-bytes"),
		},
		{
			name:     "Default",
			field:    zapcore.Field{Key: "def", Type: zapcore.UnknownType, String: "default"},
			expected: attribute.String("def", "default"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapZapField(tt.field)
			assert.Equal(t, tt.expected.Key, got.Key)
			assert.Equal(t, tt.expected.Value.Type(), got.Value.Type())
			assert.Equal(t, tt.expected.Value.AsInterface(), got.Value.AsInterface())
		})
	}
}

func TestOtelZapCore_Write(t *testing.T) {
	var buf bytes.Buffer

	// Create a stdout exporter for OpenTelemetry logs
	// This is used to capture the output of the OtelZapCore.
	exporter, err := stdoutlog.New(stdoutlog.WithWriter(&buf))
	require.NoError(t, err)

	// Create a simple processor for the exporter
	// This processor will handle the logs and send them to the exporter.
	processor := sdklog.NewSimpleProcessor(exporter)
	// Create a logger provider with the processor
	// This provider will be used by the OtelZapCore to emit logs.
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
	)

	core := NewOtelZapCore(provider).(*OtelZapCore)

	tests := []struct {
		name        string
		entry       zapcore.Entry
		fields      []zapcore.Field
		coreFields  []zapcore.Field
		wantMessage string
		wantAttrs   map[string]string
	}{
		{
			name: "basic info log",
			entry: zapcore.Entry{
				Message: "hello world",
				Level:   zapcore.InfoLevel,
				Time:    time.Now(),
			},
			fields: []zapcore.Field{
				{Key: "foo", Type: zapcore.StringType, String: "bar"},
			},
			wantMessage: "hello world",
			wantAttrs:   map[string]string{"foo": "bar"},
		},
		{
			name: "error log with stack and caller",
			entry: zapcore.Entry{
				Message: "fail",
				Level:   zapcore.ErrorLevel,
				Time:    time.Now(),
				Stack:   "stacktrace",
				Caller:  zapcore.EntryCaller{Defined: true, File: "file.go", Line: 42},
			},
			fields: []zapcore.Field{
				{Key: "err", Type: zapcore.ErrorType, Interface: errors.New("fail")},
			},
			wantMessage: "fail",
			wantAttrs: map[string]string{
				"err":                  "fail",
				"exception.type":       "file.go:42",
				"exception.stacktrace": "stacktrace",
			},
		},
		{
			name: "core fields and span context",
			entry: zapcore.Entry{
				Message: "with span",
				Level:   zapcore.InfoLevel,
				Time:    time.Now(),
			},
			coreFields: []zapcore.Field{
				{Key: "context", Type: zapcore.ReflectType, Interface: trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    [16]byte{1, 2, 3},
					SpanID:     [8]byte{4, 5, 6},
					TraceFlags: 1,
				})},
			},
			fields: []zapcore.Field{
				{Key: "foo", Type: zapcore.StringType, String: "bar"},
			},
			wantMessage: "with span",
			wantAttrs: map[string]string{
				"foo":         "bar",
				"trace_id":    "01020300000000000000000000000000",
				"span_id":     "0405060000000000",
				"trace_flags": "01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			c := *core
			if len(tt.coreFields) > 0 {
				c.fields = tt.coreFields
			}

			err := c.Write(tt.entry, tt.fields)
			require.NoError(t, err)

			var logEntry struct {
				Body struct {
					Value string `json:"Value"`
				} `json:"Body"`
				Attributes []struct {
					Key   string `json:"Key"`
					Value struct {
						Value string `json:"Value"`
					} `json:"Value"`
				} `json:"Attributes"`
			}

			err = json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err, "failed to parse OTEL JSON log output")

			assert.Equal(t, tt.wantMessage, logEntry.Body.Value)

			got := map[string]string{}
			for _, attr := range logEntry.Attributes {
				got[attr.Key] = attr.Value.Value
			}

			for k, v := range tt.wantAttrs {
				assert.Contains(t, got, k, "expected key %q", k)
				assert.Equal(t, v, got[k], "mismatch for key %q", k)
			}
		})
	}
}
