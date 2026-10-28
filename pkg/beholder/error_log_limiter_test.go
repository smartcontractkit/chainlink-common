package beholder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorLogLimiter(t *testing.T) {
	start := time.Unix(1000, 0)
	l := newErrorLogLimiter(time.Minute)

	ok, suppressed := l.allow("Unavailable", start)
	assert.True(t, ok, "the first failure for a code is logged")
	assert.Zero(t, suppressed)

	for i := range 3 {
		ok, _ = l.allow("Unavailable", start.Add(time.Duration(i+1)*time.Second))
		assert.False(t, ok, "repeats within the interval are suppressed")
	}

	ok, suppressed = l.allow("DeadlineExceeded", start.Add(5*time.Second))
	assert.True(t, ok, "each code is limited independently")
	assert.Zero(t, suppressed)

	ok, suppressed = l.allow("Unavailable", start.Add(time.Minute))
	assert.True(t, ok, "the next failure after the interval is logged")
	assert.Equal(t, 3, suppressed, "and reports how many were suppressed since the last log")

	ok, _ = l.allow("Unavailable", start.Add(time.Minute+time.Second))
	assert.False(t, ok, "the interval restarts from the last log")
}
