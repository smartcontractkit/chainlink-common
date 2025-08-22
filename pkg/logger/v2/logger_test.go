package logger_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	logger "github.com/smartcontractkit/chainlink-common/pkg/logger/v2"
)

func TestZapHandler(t *testing.T) {
	// this is an slog text logger
	textLogger := slog.New(slog.NewTextHandler(&testWriter{t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	jsonLogger := slog.New(slog.NewJSONHandler(&testWriter{t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// a basic zap logger that mimics the current logger
	zapLogger, observed := testObserved(t, zapcore.DebugLevel)

	// Baseline:
	// zap logs a message without grouping. the function `With` adds key/value pairs. `Named` concats the provided name
	// to the existing name.
	//
	// ex: logger.go:146: 2025-07-28T12:01:06.516-0500 INFO    ZapLogger.Test        v2/logger_test.go:29    direct zap log  {"key": "value", "other_key": "other_value"}
	zapLogger.Named("ZapLogger").
		With(zap.String("key", "value")).
		Named("Test").
		With(zap.String("other_key", "other_value")).
		Info("direct zap log")

	// Example slog usage #1:
	// slog logs a message with grouping. the function `WithGroup` adds a group to the logger and concats the provided
	// name to the key/value pairs within the group. `With` adds key/value pairs.
	//
	// ex: logger_test.go:52: time=2025-07-28T12:01:06.517-05:00 level=INFO msg=TextHandler ZapLogger.key=value TestHandler.Test.other_key=other_value
	textLogger.WithGroup("TextHandler").
		With(slog.String("key", "value")).
		WithGroup("Test").
		With(slog.String("other_key", "other_value")).
		InfoContext(t.Context(), "TextHandler log")

	// Example slog usage #2:
	// slog logs a message with grouping. the function `WithGroup` adds a group to the logger and nests the grouped
	// pairs within the total json output. `With` adds key/value pairs.
	//
	// ex: logger_test.go:85: {"time":"2025-07-28T12:42:13.933841-05:00","level":"INFO","msg":"JSONHandler log","JSONHandler":{"key":"value","Test":{"other_key":"other_value"}}}
	jsonLogger.WithGroup("JSONHandler").
		With(slog.String("key", "value")).
		WithGroup("Test").
		With(slog.String("other_key", "other_value")).
		InfoContext(t.Context(), "JSONHandler log")

	// Example New Handler usage:
	// for our custom handler, we can use `WithGroup` to concat the name that gets passed to the zap logger or use it
	// to concat the group name to the key/value pairs like the text or json handlers.
	// if we want to retain the expected slog behavior of `WithGroup` grouping the key/value pairs, we would need
	// another way to create new loggers with a new name from an existing logger.
	//
	// ex: logger.go:146: 2025-07-28T12:01:06.516-0500 INFO    ZapLogger.Test        v2/logger_test.go:29    direct zap log  {"group":{"key": "value", "other_key": "other_value"}}
	// to retain key grouping, we need a way to add `ZapLogger.Test` as a name to the logger
	lggr := logger.Config{
		Name:   "ZapHandler",
		Level:  slog.LevelDebug,
		Logger: zapLogger,
	}.New()

	logger.Named("Test", lggr).
		WithGroup("actual_group").
		With(slog.String("key", "value")).
		InfoContext(t.Context(), "ZapHandler log here")

	require.Len(t, observed.TakeAll(), 2)
}

func testObserved(tb testing.TB, lvl zapcore.Level) (*zap.Logger, *observer.ObservedLogs) {
	oCore, logs := observer.New(lvl)
	observe := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, oCore)
	})

	return zaptest.NewLogger(tb, zaptest.WrapOptions(observe, zap.AddCaller())), logs
}

type testWriter struct {
	t *testing.T
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.t.Logf("%s", p)

	return len(p), nil
}
