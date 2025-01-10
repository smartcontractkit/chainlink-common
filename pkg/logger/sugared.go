package logger

// SugaredLogger extends the base Logger interface with syntactic sugar, similar to zap.SugaredLogger, include two new levels.
//   - Critical: Requires quick action from the node op, obviously these should happen extremely rarely. Example: failed to listen on TCP port
//   - Trace: Only included if compiled with the trace tag. For example: go test -tags trace ...
type SugaredLogger interface {
	Logger

	// AssumptionViolation variants log at error level with the message prefix "AssumptionViolation: ".
	AssumptionViolation(args ...interface{})
	AssumptionViolationf(format string, vals ...interface{})
	AssumptionViolationw(msg string, keysAndVals ...interface{})

	// ErrorIf logs the error if present.
	ErrorIf(err error, msg string)
	// ErrorIfFn calls fn() and logs any returned error along with msg.
	// Unlike ErrorIf, this can be deffered inline, since the function call is delayed:
	//
	//	defer lggr.ErrorIfFn(resource.Close, "Failed to close resource")
	ErrorIfFn(fn func() error, msg string)

	// Critical emits critical level logs (a remapping of [zap.DPanicLevel]) or falls back to error level with a '[crit]' prefix.
	Critical(args ...interface{})
	Criticalf(format string, vals ...interface{})
	Criticalw(msg string, keysAndVals ...interface{})

	// Trace emits logs only when built with the 'trace' tag.
	//
	//	go test -tags trace ./foo -run TestBar
	Trace(args ...interface{})
	Tracef(format string, vals ...interface{})
	Tracew(msg string, keysAndVals ...interface{})

	// Named creates a new Logger sub-scoped with name.
	// Names are inherited and dot-separated.
	//   a := l.Named("A") // logger=A
	//   b := a.Named("A") // logger=A.B
	// Names are generally `MixedCaps`, without spaces, like Go names. `Foo.Bar.HTTPBaz`
	Named(string) SugaredLogger
	// With returns a new Logger with the given arguments.
	With(keyvals ...any) SugaredLogger
	// Helper returns a new logger with the number of callers skipped by caller annotation increased by skip.
	// This allows wrappers and helpers to point higher up the stack (like testing.T.Helper()).
	Helper(skip int) SugaredLogger
}

// Sugared returns a new SugaredLogger wrapping the given Logger.
// Prefer to store the SugaredLogger for reuse, instead of recreating it as needed.
func Sugared(l Logger) SugaredLogger {
	if sl, ok := l.(SugaredLogger); ok {
		return sl
	}
	return &sugared{
		Logger: l,
		h:      Helper(l, 1),
	}
}

type sugared struct {
	Logger
	h Logger // helper with stack trace skip level
}

func (s *sugared) ErrorIf(err error, msg string) {
	if err != nil {
		s.h.Errorw(msg, "err", err)
	}
}

func (s *sugared) ErrorIfFn(fn func() error, msg string) {
	if err := fn(); err != nil {
		s.h.Errorw(msg, "err", err)
	}
}

const assumptionViolationPrefix = "AssumptionViolation: "

func (s *sugared) AssumptionViolation(args ...interface{}) {
	s.h.Error(append([]interface{}{assumptionViolationPrefix}, args...))
}

func (s *sugared) AssumptionViolationf(format string, vals ...interface{}) {
	s.h.Errorf(assumptionViolationPrefix+format, vals...)
}

func (s *sugared) AssumptionViolationw(msg string, keyvals ...interface{}) {
	s.h.Errorw(assumptionViolationPrefix+msg, keyvals...)
}

const critPrefix = "[crit] "

func (s *sugared) Critical(args ...interface{}) {
	switch t := s.h.(type) {
	case *logger:
		t.DPanic(args...)
		return
	}
	c, ok := s.h.(interface {
		Critical(args ...interface{})
	})
	if ok {
		c.Critical(args...)
		return
	}
	s.h.Error(append([]any{critPrefix}, args...)...)
}

func (s *sugared) Criticalf(format string, values ...interface{}) {
	switch t := s.h.(type) {
	case *logger:
		t.DPanicf(format, values...)
		return
	}
	c, ok := s.h.(interface {
		Criticalf(format string, values ...interface{})
	})
	if ok {
		c.Criticalf(format, values...)
		return
	}
	s.h.Errorf(critPrefix+format, values...)
}

func (s *sugared) Criticalw(msg string, keysAndValues ...interface{}) {
	switch t := s.h.(type) {
	case *logger:
		t.DPanicw(msg, keysAndValues...)
		return
	}
	c, ok := s.h.(interface {
		Criticalw(msg string, keysAndValues ...interface{})
	})
	if ok {
		c.Criticalw(msg, keysAndValues...)
		return
	}
	s.h.Errorw(critPrefix+msg, keysAndValues...)
}

func (s *sugared) Named(n string) SugaredLogger {
	return Sugared(Named(s.Logger, n))
}

func (s *sugared) With(keyvals ...interface{}) SugaredLogger {
	return Sugared(With(s.Logger, keyvals...))
}

func (s *sugared) Helper(skip int) SugaredLogger {
	return Sugared(Helper(s.Logger, skip))
}
