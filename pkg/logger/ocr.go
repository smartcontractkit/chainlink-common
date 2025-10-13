package logger

import (
	ocrtypes "github.com/smartcontractkit/libocr/commontypes"
)

var _ ocrtypes.Logger = &ocrWrapper{}

type ocrWrapper struct {
	l         Logger
	trace     bool
	saveError func(string)
}

// NewOCRWrapper returns a new [ocrtypes.Logger] backed by the given Logger.
// Note: trace logs are written at debug level, regardless of any build tags.
func NewOCRWrapper(l Logger, trace bool, saveError func(string)) ocrtypes.Logger {
	return &ocrWrapper{
		// Skip an extra level since we are passed along to another wrapper.
		l:         Helper(l, 2),
		trace:     trace,
		saveError: saveError,
	}
}
func (o *ocrWrapper) Trace(msg string, fields ocrtypes.LogFields) {
	if o.trace {
		o.l.Debugw(msg, toKeysAndValues(fields)...)
	}
}

func (o *ocrWrapper) Debug(msg string, fields ocrtypes.LogFields) {
	o.l.Debugw(msg, toKeysAndValues(fields)...)
}

func (o *ocrWrapper) Info(msg string, fields ocrtypes.LogFields) {
	o.l.Infow(msg, toKeysAndValues(fields)...)
}

func (o *ocrWrapper) Warn(msg string, fields ocrtypes.LogFields) {
	o.l.Warnw(msg, toKeysAndValues(fields)...)
}

// Note that the structured fields may contain dynamic data (timestamps etc.)
// So when saving the error, we only save the top level string, details
// are included in the log.
func (o *ocrWrapper) Error(msg string, fields ocrtypes.LogFields) {
	o.saveError(msg)
	o.l.Errorw(msg, toKeysAndValues(fields)...)
}

func (o *ocrWrapper) Critical(msg string, fields ocrtypes.LogFields) {
	Criticalw(o.l, msg, toKeysAndValues(fields)...)
}

func toKeysAndValues(fields ocrtypes.LogFields) []any {
	out := []any{}
	for key, val := range fields {
		out = append(out, key, val)
	}
	return out
}
