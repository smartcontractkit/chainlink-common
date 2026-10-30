package beholder

import (
	"sync"
	"time"
)

// errorLogInterval is how often a batch emitter logs a repeating failure for one error code.
const errorLogInterval = time.Minute

// errorLogLimiter rate-limits per-event failure logs: the first failure per key is logged, then
// at most one per interval, reporting how many were suppressed in between. During an outage this
// turns one log line per dropped event into one per error code per interval; the
// chip_ingress.events_dropped metric still counts every event. Keys are bounded (error codes).
type errorLogLimiter struct {
	interval time.Duration

	mu    sync.Mutex
	state map[string]*errorLogState
}

type errorLogState struct {
	lastLogged time.Time
	suppressed int
}

func newErrorLogLimiter(interval time.Duration) *errorLogLimiter {
	return &errorLogLimiter{interval: interval, state: make(map[string]*errorLogState)}
}

// allow reports whether a failure for key should be logged now and, if so, how many were
// suppressed since the last log for that key.
func (l *errorLogLimiter) allow(key string, now time.Time) (ok bool, suppressed int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	s, seen := l.state[key]
	if !seen {
		l.state[key] = &errorLogState{lastLogged: now}
		return true, 0
	}
	if now.Sub(s.lastLogged) < l.interval {
		s.suppressed++
		return false, 0
	}
	suppressed = s.suppressed
	s.lastLogged, s.suppressed = now, 0
	return true, suppressed
}
