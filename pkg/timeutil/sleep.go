package timeutil

import "time"

// Sleep is like [time.Sleep], but supports [context.Context] and more by short-circuiting on done.
// Returns true, unless short-circuited by done.
func Sleep(done <-chan struct{}, duration time.Duration) bool {
	select {
	case <-done:
		return false
	case <-time.After(duration):
		return true
	}
}
