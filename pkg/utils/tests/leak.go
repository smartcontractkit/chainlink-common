package tests

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"go.uber.org/goleak"
)

// VerifyNoLeaks verifies that the test does not leak any goroutines.
// Must be called at the start of the test, before any goroutines have spawned.
// Cannot be used from parallel tests.
func VerifyNoLeaks(t testing.TB) {
	// Set a random environment variable to trigger testing.checkParallel()
	t.Setenv(strconv.Itoa(rand.Int()), strconv.Itoa(rand.Int()))
	current := goleak.IgnoreCurrent()
	t.Cleanup(func() {
		if t.Failed() {
			t.Log("Test failed - skipping goroutine leak check")
			return
		}
		goleak.VerifyNone(t, current)
	})
}
