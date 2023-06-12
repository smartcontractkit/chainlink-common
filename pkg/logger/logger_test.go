package logger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
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
			logger: &other{zaptest.NewLogger(t).Sugar(), ""},
		},
		{
			name:   "different",
			logger: &different{zaptest.NewLogger(t).Sugar(), ""},
		},
		{
			name:   "missing",
			logger: &mismatch{zaptest.NewLogger(t).Sugar(), ""},
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
			logger:       &other{zaptest.NewLogger(t).Sugar(), "initialized"},
		},
		{
			expectedName: "different.should_still_work",
			logger:       Named(&different{zaptest.NewLogger(t).Sugar(), "different"}, "should_still_work"),
		},
		{
			expectedName: "mismatch",
			logger:       Named(&mismatch{zaptest.NewLogger(t).Sugar(), "mismatch"}, "should_not_work"),
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
			logger: &other{zaptest.NewLogger(t).Sugar(), ""},
		},
		{
			name:   "different",
			logger: &different{zaptest.NewLogger(t).Sugar(), ""},
		},
		{
			name:   "missing",
			logger: &mismatch{zaptest.NewLogger(t).Sugar(), ""},
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

type other struct {
	*zap.SugaredLogger
	name string
}

func (o *other) With(args ...interface{}) Logger {
	return &other{o.SugaredLogger.With(args...), ""}
}

func (o *other) Helper(skip int) Logger {
	return &other{o.SugaredLogger.With(zap.AddCallerSkip(skip)), ""}
}

func (o *other) Name() string {
	return o.name
}

func (o *other) Named(name string) Logger {
	newLogger := *o
	newLogger.name = joinName(o.name, name)
	newLogger.SugaredLogger = o.SugaredLogger.Named(name)
	return &newLogger
}

type different struct {
	*zap.SugaredLogger
	name string
}

func (d *different) With(args ...interface{}) differentLogger {
	return &different{d.SugaredLogger.With(args...), ""}
}

func (d *different) Helper(skip int) differentLogger {
	return &other{d.SugaredLogger.With(zap.AddCallerSkip(skip)), ""}
}

func (d *different) Name() string {
	return d.name
}

func (d *different) Named(name string) Logger {
	newLogger := *d
	newLogger.name = joinName(d.name, name)
	newLogger.SugaredLogger = d.SugaredLogger.Named(name)
	return &newLogger
}

type mismatch struct {
	*zap.SugaredLogger
	name string
}

func (m *mismatch) With(args ...interface{}) interface{} {
	return &mismatch{m.SugaredLogger.With(args...), ""}
}

func (m *mismatch) Helper(skip int) interface{} {
	return &other{m.SugaredLogger.With(zap.AddCallerSkip(skip)), ""}
}

func (m *mismatch) Name() string {
	return m.name
}

type differentLogger interface {
	Name() string
	Named(string) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, values ...interface{})
	Infof(format string, values ...interface{})
	Warnf(format string, values ...interface{})
	Errorf(format string, values ...interface{})
	Panicf(format string, values ...interface{})
	Fatalf(format string, values ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Sync() error
}
