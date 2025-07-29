package logger

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
)

type Config struct {
	Name  string
	Level slog.Leveler

	// (optional) Logger helps convert existing zap.Logger to slog.Logger.
	Logger *zap.Logger
	// TODO: maybe this should be a zapcore.Core? will test it out in core to understand ergonomics.
	Core *zapcore.Core
}

func (c Config) New() *slog.Logger {
	if c.Level == nil {
		c.Level = slog.LevelDebug
	}

	if c.Logger == nil {
		c.Logger = zap.L()
	}

	return slog.New(&zapHandler{
		config: c,
		attributes: []slog.Attr{
			slog.String("version", buildVersion()),
		},
		groups: []string{},
	})
}

func Named(name string, logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return Config{
			Name: name,
		}.New()
	}

	handler := logger.Handler()
	if hndlr, ok := handler.(namedHandler); ok {
		return slog.New(hndlr.WithName(name))
	}

	return logger
}

type namedHandler interface {
	WithName(string) slog.Handler
}

type zapHandler struct {
	config     Config
	attributes []slog.Attr
	groups     []string
}

func New() *slog.Logger {
	cfg := &Config{
		Level: slog.LevelDebug,
	}

	return cfg.New()
}

func (h *zapHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.config.Level.Level()
}

func (h *zapHandler) Handle(ctx context.Context, record slog.Record) error {
	fs := runtime.CallersFrames([]uintptr{record.PC})
	f, _ := fs.Next()
	fields := convert(h.attributes, h.groups, &record)

	checked := h.config.Logger.Core().Check(zapcore.Entry{
		Level:      asZapLevel(record.Level),
		Time:       record.Time,
		LoggerName: h.config.Name,
		Message:    record.Message,
		Caller: zapcore.EntryCaller{
			Defined:  true,
			PC:       f.PC,
			File:     f.File,
			Line:     f.Line,
			Function: f.Function,
		},
		Stack: "", // TODO: add stack trace support
	}, nil)

	if checked != nil {
		return h.config.Logger.Core().Write(checked.Entry, fields)
	}

	h.config.Logger.Log(asZapLevel(record.Level), record.Message, fields...)

	return nil
}

func (h *zapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &zapHandler{
		config:     h.config,
		attributes: appendAttrsToGroup(h.groups, h.attributes, attrs...),
		groups:     h.groups,
	}
}

func (h *zapHandler) WithGroup(name string) slog.Handler {
	return &zapHandler{
		config:     h.config,
		attributes: h.attributes,
		groups:     append(h.groups, name),
	}
}

func (h *zapHandler) WithName(name string) slog.Handler {
	return &zapHandler{
		config: Config{
			Name:   concat(h.config.Name, name),
			Level:  h.config.Level,
			Logger: h.config.Logger,
		},
		attributes: h.attributes,
		groups:     h.groups,
	}
}

func buildVersion() string {
	return fmt.Sprintf("%s@%s", build.Version, build.ChecksumPrefix)
}

func concat(values ...string) string {
	if len(values) == 0 {
		return ""
	}

	if values[0] == "" {
		return strings.Join(values[1:], ".")
	}

	return strings.Join(values, ".")
}

func asZapLevel(slevel slog.Level) zapcore.Level {
	var zlevel zapcore.Level

	switch {
	case slevel < slog.LevelDebug:
		zlevel = zap.DebugLevel - 1
	case slevel < slog.LevelInfo:
		zlevel = zapcore.DebugLevel
	case slevel < slog.LevelWarn:
		zlevel = zapcore.DebugLevel
	case slevel < slog.LevelError:
		zlevel = zapcore.WarnLevel
	default:
		zlevel = zapcore.ErrorLevel
	}

	return zlevel
}
