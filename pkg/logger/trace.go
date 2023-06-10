//go:build trace

package logger

const tracePrefix = "[TRACE] "

// Tracew emits trace level logs, which are debug level with a '[trace]' prefix.
func Tracew(l Logger, msg string, keysAndValues ...interface{}) {
	t, ok := l.(interface {
		Tracew(string, ...interface{})
	})
	if ok {
		t.Tracew(msg, keysAndValues...)
		return
	}
	l.Helper(1).Debugw(tracePrefix+msg, keysAndValues...)
}

// Tracef emits trace level logs, which are debug level with a '[trace]' prefix.
func Tracef(l Logger, format string, values ...interface{}) {
	t, ok := l.(interface {
		Tracef(string, ...interface{})
	})
	if ok {
		t.Tracef(msg, keysAndValues...)
		return
	}
	l.Helper(1).Debugf(tracePrefix+format, values...)
}
