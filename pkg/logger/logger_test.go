package logger

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestWith(t *testing.T) {
	prod, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		logger Logger
	}{
		{
			name:   "test",
			logger: Test(t),
		},
		{
			name:   "nop",
			logger: Nop(),
		},
		{
			name:   "prod",
			logger: prod,
		},
		{
			name:   "other",
			logger: &other{zaptest.NewLogger(t).Sugar()},
		},
		{
			name:   "different",
			logger: &different{zaptest.NewLogger(t).Sugar()},
		},
		{
			name:   "missing",
			logger: &mismatch{zaptest.NewLogger(t).Sugar()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := With(tt.logger, "foo", "bar")
			if got == tt.logger {
				t.Error("expected a new logger with foo==bar, but got same")
			}
		})
	}
}

func TestNamed(t *testing.T) {
	prod, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		logger       Logger
		expectedName string
	}{
		{
			expectedName: "test.test1",
			logger:       Named(Named(Test(t), "test"), "test1"),
		},
		{
			expectedName: "nop.nested",
			logger:       Named(Named(Nop(), "nop"), "nested"),
		},
		{
			expectedName: "prod",
			logger:       Named(prod, "prod"),
		},
		{
			expectedName: "initialized",
			logger:       &other{zaptest.NewLogger(t).Sugar().Named("initialized")},
		},
		{
			expectedName: "different.should_still_work",
			logger:       Named(&different{zaptest.NewLogger(t).Named("different").Sugar()}, "should_still_work"),
		},
		{
			expectedName: "mismatch",
			logger:       Named(&mismatch{zaptest.NewLogger(t).Named("mismatch").Sugar()}, "should_not_work"),
		},
	} {
		t.Run(fmt.Sprintf("test_logger_name_expect_%s", tt.expectedName), func(t *testing.T) {
			require.Equal(t, tt.expectedName, tt.logger.Name())
		})
	}
}

func TestHelper(t *testing.T) {
	prod, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		logger Logger
	}{
		{
			name:   "test",
			logger: Test(t),
		},
		{
			name:   "nop",
			logger: Nop(),
		},
		{
			name:   "prod",
			logger: prod,
		},
		{
			name:   "other",
			logger: &other{zaptest.NewLogger(t).Sugar()},
		},
		{
			name:   "different",
			logger: &different{zaptest.NewLogger(t).Sugar()},
		},
		{
			name:   "missing",
			logger: &mismatch{zaptest.NewLogger(t).Sugar()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Helper(tt.logger, 1)
			if got == tt.logger {
				t.Error("expected a new logger with foo==bar, but got same")
			}
		})
	}
}

func TestCritical(t *testing.T) {
	lggr, observed := TestObserved(t, zap.DebugLevel)
	testCritical(t, lggr, observed, "foobar", zap.DPanicLevel)

	var sl *zap.SugaredLogger
	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &other{sl}
	testCritical(t, lggr, observed, "foobar", zap.DPanicLevel)

	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &mismatch{sl}
	testCritical(t, lggr, observed, "[crit] foobar", zap.ErrorLevel)
}

func testCritical(t *testing.T, lggr Logger, observed *observer.ObservedLogs, msg string, lvl zapcore.Level) {
	Critical(lggr, "foo", "bar")
	_, filename, lineNum, ok := runtime.Caller(0)
	require.True(t, ok)
	lineNum--

	all := observed.TakeAll()
	require.Len(t, all, 1)
	line := all[0]
	assert.Equal(t, lvl, line.Level, "expected %q but got %q", lvl, line.Level)
	assert.Equal(t, msg, line.Message)
	assert.Equal(t, filename, line.Caller.File)
	assert.Equal(t, lineNum, line.Caller.Line)
}

func TestCriticalw(t *testing.T) {
	lggr, observed := TestObserved(t, zap.DebugLevel)
	testCriticalw(t, lggr, observed, "msg", zap.DPanicLevel)

	var sl *zap.SugaredLogger
	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &other{sl}
	testCriticalw(t, lggr, observed, "msg", zap.DPanicLevel)

	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &mismatch{sl}
	testCriticalw(t, lggr, observed, "[crit] msg", zap.ErrorLevel)
}

func testCriticalw(t *testing.T, lggr Logger, observed *observer.ObservedLogs, msg string, lvl zapcore.Level) {
	Criticalw(lggr, "msg", "foo", "bar")
	_, filename, lineNum, ok := runtime.Caller(0)
	require.True(t, ok)
	lineNum--

	all := observed.TakeAll()
	require.Len(t, all, 1)
	line := all[0]
	assert.Equal(t, lvl, line.Level)
	assert.Equal(t, msg, line.Message)
	assert.Equal(t, filename, line.Caller.File)
	assert.Equal(t, lineNum, line.Caller.Line)
	require.Equal(t, "bar", line.ContextMap()["foo"])
}

func TestCriticalf(t *testing.T) {
	lggr, observed := TestObserved(t, zap.DebugLevel)
	testCriticalf(t, lggr, observed, "foo: bar", zap.DPanicLevel)

	var sl *zap.SugaredLogger
	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &other{sl}
	testCriticalf(t, lggr, observed, "foo: bar", zap.DPanicLevel)

	sl, observed = testObserved(t, zap.DebugLevel)
	lggr = &mismatch{sl}
	testCriticalf(t, lggr, observed, "[crit] foo: bar", zap.ErrorLevel)
}

func testCriticalf(t *testing.T, lggr Logger, observed *observer.ObservedLogs, msg string, lvl zapcore.Level) {
	Criticalf(lggr, "foo: %s", "bar")
	_, filename, lineNum, ok := runtime.Caller(0)
	require.True(t, ok)
	lineNum--

	all := observed.TakeAll()
	require.Len(t, all, 1)
	line := all[0]
	assert.Equal(t, lvl, line.Level)
	assert.Equal(t, msg, line.Message)
	assert.Equal(t, filename, line.Caller.File)
	assert.Equal(t, lineNum, line.Caller.Line)
}

type other struct {
	*zap.SugaredLogger
}

func (o *other) With(args ...any) Logger {
	return &other{o.SugaredLogger.With(args...)}
}

func (o *other) Helper(skip int) Logger {
	return &other{o.SugaredLogger.WithOptions(zap.AddCallerSkip(skip))}
}

func (o *other) Name() string {
	return o.Desugar().Name()
}

func (o *other) Named(name string) Logger {
	newLogger := *o
	newLogger.SugaredLogger = o.SugaredLogger.Named(name)
	return &newLogger
}

func (o *other) Critical(args ...any) {
	o.WithOptions(zap.AddCallerSkip(1)).DPanic(args...)
}
func (o *other) Criticalf(format string, values ...any) {
	o.WithOptions(zap.AddCallerSkip(1)).DPanicf(format, values...)
}
func (o *other) Criticalw(msg string, keysAndValues ...any) {
	o.WithOptions(zap.AddCallerSkip(1)).DPanicw(msg, keysAndValues...)
}

type different struct {
	*zap.SugaredLogger
}

func (d *different) With(args ...any) differentLogger {
	return &different{d.SugaredLogger.With(args...)}
}

func (d *different) Helper(skip int) differentLogger {
	return &different{d.SugaredLogger.WithOptions(zap.AddCallerSkip(skip))}
}

func (d *different) Name() string {
	return d.Desugar().Name()
}

func (d *different) Named(name string) Logger {
	newLogger := *d
	newLogger.SugaredLogger = d.SugaredLogger.Named(name)
	return &newLogger
}

type mismatch struct {
	*zap.SugaredLogger
}

func (m *mismatch) With(args ...any) any {
	return &mismatch{m.SugaredLogger.With(args...)}
}

func (m *mismatch) Helper(skip int) any {
	return &mismatch{m.SugaredLogger.WithOptions(zap.AddCallerSkip(skip))}
}

func (m *mismatch) Name() string {
	return m.Desugar().Name()
}

type differentLogger interface {
	Name() string
	Named(string) Logger

	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Panic(args ...any)
	Fatal(args ...any)

	Debugf(format string, values ...any)
	Infof(format string, values ...any)
	Warnf(format string, values ...any)
	Errorf(format string, values ...any)
	Panicf(format string, values ...any)
	Fatalf(format string, values ...any)

	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Panicw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)

	Sync() error
}
