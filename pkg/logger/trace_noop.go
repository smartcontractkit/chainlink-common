//go:build !trace

package logger

func (s *sugared) Trace(args ...any) {}

func (s *sugared) Tracef(format string, vals ...any) {}

func (s *sugared) Tracew(msg string, keysAndValues ...any) {}

// Deprecated: instead use [SugaredLogger.Trace]:
//
//	Sugared(l).Trace(args...)
func Trace(l Logger, args ...any) {}

// Deprecated: instead use [SugaredLogger.Tracef]:
//
//	Sugared(l).Tracef(args...)
func Tracef(l Logger, format string, values ...any) {}

// Deprecated: instead use [SugaredLogger.Tracew]:
//
//	Sugared(l).Tracew(args...)
func Tracew(l Logger, msg string, keysAndValues ...any) {}
