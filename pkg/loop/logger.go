package loop

import (
	"io"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

// hclogLogger returns an [hclog.Logger] backed by the given [logger.Logger].
func hclogLogger(l logger.Logger) hclog.Logger {
	hcl := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Output: io.Discard, // only write through p.Logger Sink
	})
	hcl.RegisterSink(&hclSinkAdapter{l: l})
	return hcl
}

var _ hclog.SinkAdapter = (*hclSinkAdapter)(nil)

// hclSinkAdapter implements [hclog.SinkAdapter] with a [logger.Logger].
type hclSinkAdapter struct {
	l logger.Logger
}

func (h *hclSinkAdapter) Accept(name string, level hclog.Level, msg string, args ...interface{}) {
	// name is ignored because hclog does not promote the logger name field like with level, message, and timestamp.
	switch level {
	case hclog.NoLevel:
	case hclog.Debug, hclog.Trace:
		h.l.Debugw(msg, args...)
	case hclog.Info:
		h.l.Infow(msg, args...)
	case hclog.Warn:
		h.l.Warnw(msg, args...)
	case hclog.Error:
		h.l.Errorw(msg, args...)
	}
}

// NewLogger returns a new [logger.Logger] configured to encode [hclog] compatible JSON.
func NewLogger() (logger.Logger, error) {
	return logger.NewWith(func(cfg *zap.Config) {
		cfg.Level.SetLevel(zap.DebugLevel)
		cfg.EncoderConfig.LevelKey = "@level"
		cfg.EncoderConfig.MessageKey = "@message"
		cfg.EncoderConfig.TimeKey = "@timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000Z07:00")
	})
}
