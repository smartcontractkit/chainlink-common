package tests

import (
	"strings"
	"testing"

	"go.uber.org/zap/zaptest/observer"
)

// AssertLogEventually waits until at least one log message containing the
// specified msg is emitted.
// NOTE: This does not "pop" messages so it cannot be used multiple times to
// check for new instances of the same msg. See AssertLogCountEventually instead.
//
// Get a *observer.ObservedLogs like so:
//
//	observedZapCore, observedLogs := observer.New(zap.DebugLevel)
//	lggr := logger.TestLogger(t, observedZapCore)
func AssertLogEventually(t *testing.T, observedLogs *observer.ObservedLogs, msg string) {
	AssertLogCountEventually(t, observedLogs, msg, 1)
}

// AssertLogCountEventually waits until at least count log message containing the
// specified msg is emitted
func AssertLogCountEventually(t *testing.T, observedLogs *observer.ObservedLogs, msg string, count int) {
	AssertEventually(t, func() bool {
		i := 0
		for _, l := range observedLogs.All() {
			if strings.Contains(l.Message, msg) {
				i++
				if i >= count {
					return true
				}
			}
		}
		return false
	})
}

// RequireLogMessage fails the test if emitted logs don't contain the given message
func RequireLogMessage(t *testing.T, observedLogs *observer.ObservedLogs, msg string) {
	for _, l := range observedLogs.All() {
		if strings.Contains(l.Message, msg) {
			return
		}
	}
	t.Log("observed logs", observedLogs.All())
	t.Fatalf("expected observed logs to contain msg %q, but it didn't", msg)
}
