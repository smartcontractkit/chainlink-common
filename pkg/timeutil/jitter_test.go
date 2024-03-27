package timeutil

import (
	"testing"
	"time"
)

func TestWithJitter(t *testing.T) {
	for _, tt := range []struct {
		dur      time.Duration
		from, to time.Duration
	}{
		{0, 0, 0},
		{time.Second, 900 * time.Millisecond, 1100 * time.Millisecond},
		{time.Minute, 54 * time.Second, 66 * time.Second},
		{24 * time.Hour, 21*time.Hour + 36*time.Minute, 26*time.Hour + 24*time.Minute},
	} {
		t.Run(tt.dur.String(), func(t *testing.T) {
			for i := 0; i < 100; i++ {
				got := WithJitter(tt.dur)
				t.Logf("%d: %s", i, got)
				if got < tt.from || got > tt.to {
					t.Errorf("expected duration %s with jitter to be between (%s, %s) but got: %s", tt.dur, tt.from, tt.to, got)
				}
			}
		})
	}
}
