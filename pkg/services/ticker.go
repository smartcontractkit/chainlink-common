package services

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// NewTicker returns a new timeutil.Ticker configured to:
// - fire the first tick immediately
// - apply jitter to each period via timeutil.WithJitter
func NewTicker(d time.Duration) *timeutil.Ticker {
	first := true
	return timeutil.NewTicker(func() time.Duration {
		if first {
			first = false
			return 0
		}
		return timeutil.WithJitter(d)
	})
}
