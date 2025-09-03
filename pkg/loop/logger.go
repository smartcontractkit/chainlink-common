package loop

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	otellog "go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger/otelzap"
)

// HCLogLogger returns an [hclog.Logger] backed by the given [logger.Logger].
func HCLogLogger(l logger.Logger) hclog.Logger {
	hcl := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Output: io.Discard, // only write through p.Logger Sink
		Level:  hclog.Debug,
	})
	hcl.RegisterSink(&hclSinkAdapter{l: logger.Sugared(l)})
	return hcl
}

var _ hclog.SinkAdapter = (*hclSinkAdapter)(nil)

// hclSinkAdapter implements [hclog.SinkAdapter] with a [logger.Logger].
type hclSinkAdapter struct {
	l logger.SugaredLogger
	m sync.Map // [string]func() l.Logger
}

func (h *hclSinkAdapter) named(name string) logger.SugaredLogger {
	onceVal := onceValue(func() logger.SugaredLogger {
		return h.l.Named(name)
	})
	v, _ := h.m.LoadOrStore(name, onceVal)
	return v.(func() logger.SugaredLogger)()
}

func removeArg(args []any, key string) ([]any, string) {
	if len(args) < 2 {
		return args, ""
	}
	for i := 0; i+1 < len(args); i += 2 {
		if args[i] == key {
			if v, ok := (args)[i+1].(string); ok {
				return slices.Delete(args, i, i+2), v
			}
			break
		}
	}
	return args, ""
}

// logMessage is the JSON payload that gets sent to Stderr from the plugin to the host
type logMessage struct {
	Message   string                 `json:"@message"`
	Level     string                 `json:"@level"`
	Timestamp time.Time              `json:"timestamp"`
	ExtraArgs []*LogMessageExtraArgs `json:"extra_args"`
}

// LogMessageExtraArgs is a key value pair within the Output payload
type LogMessageExtraArgs struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// flattenExtraArgs is used to flatten arguments of the log message
func flattenExtraArgs(le *logMessage) []any {
	var result []any
	result = append(result, "level")
	result = append(result, le.Level)
	result = append(result, "timestamp")
	result = append(result, le.Timestamp)
	for _, kv := range le.ExtraArgs {
		result = append(result, kv.Key)
		result = append(result, kv.Value)
	}

	return result
}

func parseJSON(input string) (*logMessage, error) {
	var raw map[string]any
	entry := &logMessage{}

	err := json.Unmarshal([]byte(input), &raw)
	if err != nil {
		return nil, err
	}

	if v, ok := raw["@message"]; ok {
		entry.Message = v.(string)
		delete(raw, "@message")
	}

	if v, ok := raw["@level"]; ok {
		entry.Level = v.(string)
		delete(raw, "@level")
	}

	for k, v := range raw {
		entry.ExtraArgs = append(entry.ExtraArgs, &LogMessageExtraArgs{
			Key:   k,
			Value: v,
		})
	}

	return entry, nil
}

// maybeCritical recognizes panic, fatal or critical log message, and logs them again at critical level.
func maybeCritical(msg string, l logger.SugaredLogger, args ...any) bool {
	if strings.HasPrefix(msg, "panic:") {
		l.Criticalw(fmt.Sprintf("[PANIC] %s", msg), args...)
		return true
	} else if log, err := parseJSON(msg); err == nil {
		switch log.Level {
		case "fatal":
			l.Criticalw(fmt.Sprintf("[FATAL] %s", log.Message), flattenExtraArgs(log)...)
			return true
		case "critical", "dpanic":
			l.Criticalw(log.Message, flattenExtraArgs(log)...)
			return true
		}
	}
	return false
}

func (h *hclSinkAdapter) Accept(_ string, level hclog.Level, msg string, args ...any) {
	if level == hclog.Off {
		return
	}

	l := h.l
	var name string
	if args, name = removeArg(args, "logger"); name != "" {
		l = h.named(name)
	}
	switch level {
	case hclog.Off:
		return // unreachable, but satisfies linter
	case hclog.NoLevel:
	case hclog.Trace:
		l.Debugw(msg, args...)
	case hclog.Debug:
		if !maybeCritical(msg, l, args...) { // panic/fatal/critical used to come through as stderr/out which maps to debug
			l.Debugw(msg, args...)
		}
	case hclog.Info:
		l.Infow(msg, args...)
	case hclog.Warn:
		l.Warnw(msg, args...)
	case hclog.Error:
		if !maybeCritical(msg, l, args...) { // as of v1.7.0, panic/fatal errors are send at error level
			l.Errorw(msg, args...)
		}
	}
}

// NewLogger returns a new [logger.Logger] configured to encode [hclog] compatible JSON.
func NewLogger() (logger.Logger, error) {
	return logger.NewWith(configureHCLogEncoder)
}

// configureHCLogEncoder mutates cfg to use hclog-compatible field names and timestamp format.
// NOTE: It also sets the log level to Debug to preserve prior behavior where each caller
// manually set Debug before applying identical encoder tweaks. Centralizing avoids drift.
// If a different level is desired, callers should override cfg.Level AFTER calling this helper.
func configureHCLogEncoder(cfg *zap.Config) {
	cfg.Level.SetLevel(zap.DebugLevel)
	cfg.EncoderConfig.LevelKey = "@level"
	cfg.EncoderConfig.MessageKey = "@message"
	cfg.EncoderConfig.TimeKey = "@timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000Z07:00")
}

// NewOtelLogger returns a logger with two cores:
// 1. Primary JSON core with hclog-compatible encoder keys (@level, @message, @timestamp)
// 2. OTEL core (otelzap.NewCore) that exports logs to OpenTelemetry at the specified level
//
// The encoder config only affects the primary core's JSON output.
// The OTEL core extracts data directly from zap.Entry and fields, independent of encoder settings.
func NewOtelLogger(otelLogger otellog.Logger, level zapcore.Level) (logger.Logger, error) {
	primaryCore, err := logger.NewCore(configureHCLogEncoder)
	if err != nil {
		return nil, err
	}
	return logger.NewWithCores(primaryCore, otelzap.NewCore(otelLogger, otelzap.WithLevel(level))), nil
}

// onceValue returns a function that invokes f only once and returns the value
// returned by f. The returned function may be called concurrently.
//
// If f panics, the returned function will panic with the same value on every call.
//
// Note: Copied from sync.OnceValue in upcoming 1.21 release. Can be removed after upgrading.
func onceValue[T any](f func() T) func() T {
	var (
		once   sync.Once
		valid  bool
		p      any
		result T
	)
	g := func() {
		defer func() {
			p = recover()
			if !valid {
				panic(p)
			}
		}()
		result = f()
		valid = true
	}
	return func() T {
		once.Do(g)
		if !valid {
			panic(p)
		}
		return result
	}
}
