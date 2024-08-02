package logger

import (
	"log"
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO(gg): use or remove this package

//go:generate mockery --name Logger --output ./mocks/ --case=underscore
type Logger interface {
	Info(args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Infof(format string, values ...interface{})

	Debug(args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Debugf(format string, values ...interface{})

	Warn(args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Warnf(format string, values ...interface{})

	Error(args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Errorf(format string, values ...interface{})

	Panic(args ...interface{})
	Panicf(format string, values ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, values ...interface{})

	Sync() error
}

func New() Logger {
	return newZapLogger().Sugar()
}

// NewOtelzapLogger records Zap log messages as events on the existing span that must be passed in a context.Context as a first argument.
// It does not record anything if the context does not contain a span.
// See more details [here](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelzap)
func NewOtelzapLogger() *otelzap.Logger {
	zl := newZapLogger()
	return otelzap.New(newZapLogger(), otelzap.WithMinLevel(zl.Level()))
}

func newZapLogger() *zap.Logger {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	if err != nil {
		log.Fatal(err)
	}
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	zl, err := config.Build(zap.AddCallerSkip(0))
	if err != nil {
		log.Fatal(err)
	}
	return zl
}
