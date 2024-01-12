package timeutil

import (
	"time"
)

// Ticker is like time.Ticker, but with a variable period.
type Ticker struct {
	C    <-chan time.Time
	stop chan struct{}
}

// NewTicker returns a started Ticker which calls nextDur for each period.
func NewTicker(nextDur func() time.Duration) *Ticker {
	c := make(chan time.Time) // unbuffered so we block and delay if not being handled
	t := Ticker{C: c, stop: make(chan struct{})}
	go t.run(c, nextDur)
	return &t
}

func (t *Ticker) run(c chan<- time.Time, nextDur func() time.Duration) {
	var timer *time.Timer
	defer timer.Stop()
	for {
		timer = time.NewTimer(nextDur())
		select {
		case <-t.stop:
			return

		case <-timer.C:
			timer.Stop()
			select {
			case <-t.stop:
				return
			case c <- time.Now():
			}
		}
	}
}

func (t *Ticker) Stop() { close(t.stop) }
