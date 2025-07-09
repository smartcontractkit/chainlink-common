package custmsg

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const (
	LogLevelKey                  = "log_level"
	defaultBeholderEmitTimeoutMs = 3000
)

type beholderLogger struct {
	underlying  logger.Logger
	emitter     MessageEmitter
	emitTimeout time.Duration
}

func NewBeholderLogger(underlying logger.Logger, emitter MessageEmitter) *beholderLogger {
	return &beholderLogger{
		underlying:  logger.Helper(underlying, 1),
		emitter:     emitter,
		emitTimeout: defaultBeholderEmitTimeoutMs * time.Millisecond,
	}
}

func (b *beholderLogger) With(args ...any) *beholderLogger {
	return &beholderLogger{
		underlying:  logger.With(b.underlying, args...),
		emitter:     b.emitter.WithMapLabels(getLabelMap(args...)),
		emitTimeout: b.emitTimeout,
	}
}

func (b *beholderLogger) WithEmitTimeout(name string, emitTimeout time.Duration) *beholderLogger {
	return &beholderLogger{
		underlying:  b.underlying,
		emitter:     b.emitter,
		emitTimeout: emitTimeout,
	}
}

func (b *beholderLogger) Named(name string) *beholderLogger {
	return &beholderLogger{
		underlying:  logger.Named(b.underlying, name),
		emitter:     b.emitter,
		emitTimeout: b.emitTimeout,
	}
}

func (b *beholderLogger) Debug(args ...any) {
	b.underlying.Debug(args...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "debug").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Info(args ...any) {
	b.underlying.Info(args...)
	ctx, cancel := getBeholderCallContext()
	defer cancel()
	err := b.emitter.With(LogLevelKey, "info").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Warn(args ...any) {
	b.underlying.Warn(args...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "warn").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Error(args ...any) {
	b.underlying.Error(args...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "error").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Panic(args ...any) {
	b.underlying.Panic(args...)
	ctx, _ := getBeholderCallContext()
	//	defer cancel()
	err := b.emitter.With(LogLevelKey, "panic").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Fatal(args ...any) {
	b.underlying.Fatal(args...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "fatal").Emit(ctx, fmt.Sprint(args...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Debugf(format string, values ...any) {
	b.underlying.Debugf(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "debug").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Infof(format string, values ...any) {
	b.underlying.Infof(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "info").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Warnf(format string, values ...any) {
	b.underlying.Warnf(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "warn").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Errorf(format string, values ...any) {
	b.underlying.Errorf(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "error").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Panicf(format string, values ...any) {
	b.underlying.Panicf(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "panic").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Fatalf(format string, values ...any) {
	b.underlying.Fatalf(format, values...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "fatal").Emit(ctx, fmt.Sprintf(format, values...))
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Debugw(msg string, keysAndValues ...any) {
	b.underlying.Debugw(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "debug").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Infow(msg string, keysAndValues ...any) {
	b.underlying.Infow(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "info").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Warnw(msg string, keysAndValues ...any) {
	b.underlying.Warnw(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "warn").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Errorw(msg string, keysAndValues ...any) {
	b.underlying.Errorw(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "error").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Panicw(msg string, keysAndValues ...any) {
	b.underlying.Panicw(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "panic").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
}

func (b *beholderLogger) Fatalw(msg string, keysAndValues ...any) {
	b.underlying.Fatalw(msg, keysAndValues...)
	ctx, _ := getBeholderCallContext()
	//defer cancel()
	err := b.emitter.With(LogLevelKey, "fatal").WithMapLabels(getLabelMap(keysAndValues...)).Emit(ctx, msg)
	if err != nil {
		b.underlying.Errorw("error emitting log to Beholder", "error", err)
	}
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

func getBeholderCallContext() (context.Context, context.CancelFunc) {
	// Beholder Emit() should not block for logs wrapped in a custom event but let's be safe
	return context.WithTimeout(context.Background(), defaultBeholderEmitTimeoutMs*time.Millisecond)
}
