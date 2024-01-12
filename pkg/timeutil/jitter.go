package timeutil

import (
	"math"
	mrand "math/rand"
	"time"
)

// WithJitter adds +/- 10% to a duration.
// TODO better name? JitterDuration; ApplyJitter
// TODO customizable percentage?
func WithJitter(d time.Duration) time.Duration {
	// #nosec
	if d == 0 {
		return 0
	}
	// ensure non-zero arg to Intn to avoid panic
	max := math.Max(float64(d.Abs())/5.0, 1.)
	// #nosec - non critical randomness
	jitter := mrand.Intn(int(max)) / 2
	return time.Duration(int(d) + jitter)
}
