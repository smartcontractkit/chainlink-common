package jwt

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type slogLoggerAdapter struct {
	logger *slog.Logger
}

func newSlogLoggerAdapter(l *slog.Logger) logger.Logger {
	if l == nil {
		return logger.Nop()
	}
	return &slogLoggerAdapter{logger: l}
}

func (s *slogLoggerAdapter) Name() string {
	return ""
}

func (s *slogLoggerAdapter) Debug(args ...any) { s.log(slog.LevelDebug, fmt.Sprint(args...)) }
func (s *slogLoggerAdapter) Info(args ...any)  { s.log(slog.LevelInfo, fmt.Sprint(args...)) }
func (s *slogLoggerAdapter) Warn(args ...any)  { s.log(slog.LevelWarn, fmt.Sprint(args...)) }
func (s *slogLoggerAdapter) Error(args ...any) { s.log(slog.LevelError, fmt.Sprint(args...)) }

func (s *slogLoggerAdapter) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	s.log(slog.LevelError, msg)
	panic(msg)
}

func (s *slogLoggerAdapter) Fatal(args ...any) {
	s.log(slog.LevelError, fmt.Sprint(args...))
	os.Exit(1)
}

func (s *slogLoggerAdapter) Debugf(format string, values ...any) {
	s.log(slog.LevelDebug, fmt.Sprintf(format, values...))
}

func (s *slogLoggerAdapter) Infof(format string, values ...any) {
	s.log(slog.LevelInfo, fmt.Sprintf(format, values...))
}

func (s *slogLoggerAdapter) Warnf(format string, values ...any) {
	s.log(slog.LevelWarn, fmt.Sprintf(format, values...))
}

func (s *slogLoggerAdapter) Errorf(format string, values ...any) {
	s.log(slog.LevelError, fmt.Sprintf(format, values...))
}

func (s *slogLoggerAdapter) Panicf(format string, values ...any) {
	msg := fmt.Sprintf(format, values...)
	s.log(slog.LevelError, msg)
	panic(msg)
}

func (s *slogLoggerAdapter) Fatalf(format string, values ...any) {
	s.log(slog.LevelError, fmt.Sprintf(format, values...))
	os.Exit(1)
}

func (s *slogLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
	s.log(slog.LevelDebug, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Infow(msg string, keysAndValues ...any) {
	s.log(slog.LevelInfo, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
	s.log(slog.LevelWarn, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
	s.log(slog.LevelError, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Panicw(msg string, keysAndValues ...any) {
	s.log(slog.LevelError, msg, keysAndValues...)
	panic(msg)
}

func (s *slogLoggerAdapter) Fatalw(msg string, keysAndValues ...any) {
	s.log(slog.LevelError, msg, keysAndValues...)
	os.Exit(1)
}

func (s *slogLoggerAdapter) Sync() error {
	return nil
}

func (s *slogLoggerAdapter) log(level slog.Level, msg string, args ...any) {
	s.logger.Log(context.Background(), level, msg, args...)
}
