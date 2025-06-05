package custmsg

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const LogLevelKey = "log_level"

type beholderLogger struct {
	underlying logger.Logger
	emitter    MessageEmitter
}

func NewBeholderLogger(underlying logger.Logger, emitter MessageEmitter) *beholderLogger {
	return &beholderLogger{
		underlying: logger.Helper(underlying, 1),
		emitter:    emitter,
	}
}

func (b *beholderLogger) With(args ...any) *beholderLogger {
	emitterWithLabels := b.emitter.WithMapLabels(getLabelMap(args...))
	underlyingWith := logger.With(b.underlying, args...)
	return NewBeholderLogger(underlyingWith, emitterWithLabels)
}

func (b *beholderLogger) Named(name string) *beholderLogger {
	return &beholderLogger{
		underlying: logger.Named(b.underlying, name),
		emitter:    b.emitter,
	}
}

func (b *beholderLogger) Debug(args ...any) {
	_ = b.emitter.With(LogLevelKey, "debug").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Debug(args...)
}

func (b *beholderLogger) Info(args ...any) {
	_ = b.emitter.With(LogLevelKey, "info").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Info(args...)
}

func (b *beholderLogger) Warn(args ...any) {
	_ = b.emitter.With(LogLevelKey, "warn").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Warn(args...)
}

func (b *beholderLogger) Error(args ...any) {
	_ = b.emitter.With(LogLevelKey, "error").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Error(args...)
}

func (b *beholderLogger) Panic(args ...any) {
	_ = b.emitter.With(LogLevelKey, "panic").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Panic(args...)
}

func (b *beholderLogger) Fatal(args ...any) {
	_ = b.emitter.With(LogLevelKey, "fatal").Emit(context.Background(), fmt.Sprint(args...))
	b.underlying.Fatal(args...)
}

func (b *beholderLogger) Debugf(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "debug").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Debugf(format, values...)
}

func (b *beholderLogger) Infof(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "info").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Infof(format, values...)
}

func (b *beholderLogger) Warnf(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "warn").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Warnf(format, values...)
}

func (b *beholderLogger) Errorf(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "error").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Errorf(format, values...)
}

func (b *beholderLogger) Panicf(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "panic").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Panicf(format, values...)
}

func (b *beholderLogger) Fatalf(format string, values ...any) {
	_ = b.emitter.With(LogLevelKey, "fatal").Emit(context.Background(), fmt.Sprintf(format, values...))
	b.underlying.Fatalf(format, values...)
}

func (b *beholderLogger) Debugw(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "debug").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Debugw(msg, keysAndValues...)
}

func (b *beholderLogger) Infow(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "info").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Infow(msg, keysAndValues...)
}

func (b *beholderLogger) Warnw(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "warn").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Warnw(msg, keysAndValues...)
}

func (b *beholderLogger) Errorw(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "error").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Errorw(msg, keysAndValues...)
}

func (b *beholderLogger) Panicw(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "panic").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Panicw(msg, keysAndValues...)
}

func (b *beholderLogger) Fatalw(msg string, keysAndValues ...any) {
	_ = b.emitter.With(LogLevelKey, "fatal").WithMapLabels(getLabelMap(keysAndValues...)).Emit(context.Background(), msg)
	b.underlying.Fatalw(msg, keysAndValues...)
}

func (b *beholderLogger) Sync() error {
	return b.underlying.Sync()
}

func (b *beholderLogger) Name() string {
	return b.underlying.Name()
}

func getLabelMap(keysAndValues ...any) map[string]string {
	labels := make(map[string]string)
	for i := 0; i < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		var value string
		if i+1 < len(keysAndValues) {
			value = fmt.Sprintf("%v", keysAndValues[i+1])
		} else {
			value = "MISSING"
		}
		labels[key] = value
	}
	return labels
}
