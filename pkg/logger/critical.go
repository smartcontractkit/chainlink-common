package logger

import "go.uber.org/zap"

func (l *logger) Criticalf(format string, values ...interface{}) {
	l.WithOptions(zap.AddCallerSkip(1)).DPanicf(format, values...)
}
