package jwt

import (
	"context"
	"log/slog"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type slogLoggerAdapter struct {
	logger *slog.Logger
}

func newSlogLoggerAdapter(l *slog.Logger) nodeJWTLogger {
	if l == nil {
		return nopNodeJWTLogger{}
	}
	return &slogLoggerAdapter{logger: l}
}

func (s *slogLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
	s.log(slog.LevelDebug, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
	s.log(slog.LevelWarn, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
	s.log(slog.LevelError, msg, keysAndValues...)
}

func (s *slogLoggerAdapter) log(level slog.Level, msg string, args ...any) {
	s.logger.Log(context.Background(), level, msg, args...)
}

type chainlinkLoggerAdapter struct {
	logger logger.Logger
}

func newChainlinkLoggerAdapter(l logger.Logger) nodeJWTLogger {
	if l == nil {
		return nopNodeJWTLogger{}
	}
	return &chainlinkLoggerAdapter{logger: l}
}

func (c *chainlinkLoggerAdapter) Debugw(msg string, keysAndValues ...any) {
	c.logger.Debugw(msg, keysAndValues...)
}

func (c *chainlinkLoggerAdapter) Warnw(msg string, keysAndValues ...any) {
	c.logger.Warnw(msg, keysAndValues...)
}

func (c *chainlinkLoggerAdapter) Errorw(msg string, keysAndValues ...any) {
	c.logger.Errorw(msg, keysAndValues...)
}

type nopNodeJWTLogger struct{}

func (nopNodeJWTLogger) Debugw(string, ...any) {}
func (nopNodeJWTLogger) Warnw(string, ...any)  {}
func (nopNodeJWTLogger) Errorw(string, ...any) {}
