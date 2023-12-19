package utils_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

func TestAllEqual(t *testing.T) {
	t.Parallel()

	require.False(t, utils.AllEqual(1, 2, 3, 4, 5))
	require.True(t, utils.AllEqual(1, 1, 1, 1, 1))
	require.False(t, utils.AllEqual(1, 1, 1, 2, 1, 1, 1))
}

func TestWaitGroupChan(t *testing.T) {
	t.Parallel()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	ch := utils.WaitGroupChan(wg)

	select {
	case <-ch:
		t.Fatal("should not fire immediately")
	default:
	}

	wg.Done()

	select {
	case <-ch:
		t.Fatal("should not fire until finished")
	default:
	}

	go func() {
		time.Sleep(2 * time.Second)
		wg.Done()
	}()

	callbackOrTimeout(t, "WaitGroupChan fires", func() {
		<-ch
	}, 5*time.Second)
}

func TestDependentAwaiter(t *testing.T) {
	t.Parallel()

	da := utils.NewDependentAwaiter()
	da.AddDependents(2)

	select {
	case <-da.AwaitDependents():
		t.Fatal("should not fire immediately")
	default:
	}

	da.DependentReady()

	select {
	case <-da.AwaitDependents():
		t.Fatal("should not fire until finished")
	default:
	}

	go func() {
		time.Sleep(2 * time.Second)
		da.DependentReady()
	}()

	callbackOrTimeout(t, "dependents are now ready", func() {
		<-da.AwaitDependents()
	}, 5*time.Second)
}

func callbackOrTimeout(t testing.TB, msg string, callback func(), durationParams ...time.Duration) {
	t.Helper()

	duration := 100 * time.Millisecond
	if len(durationParams) > 0 {
		duration = durationParams[0]
	}

	done := make(chan struct{})
	go func() {
		callback()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(duration):
		t.Fatalf("CallbackOrTimeout: %s timed out", msg)
	}
}
