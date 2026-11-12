package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAtomic(t *testing.T) {
	const (
		initial = `
Foo = "initial"`
		first = `
Foo = "first"`
		replaced = `
Foo = "replaced"`
		final = `
Foo = "final"`
	)
	var as Atomic
	require.NoError(t, as.Store(Snapshot{
		Settings: initial, Hash: "initial",
	}))

	ch, err := as.Subscribe(t.Context())
	require.NoError(t, err)

	waitFor := func(s, h string) {
		select {
		case got := <-ch:
			require.Equal(t, Snapshot{
				Settings: s, Hash: h,
			}, got)
			return
		case <-time.After(time.Second):
		case <-t.Context().Done():
		}
		t.Fatalf("timed out waiting for %s update", s)
	}

	// receive initial
	waitFor(initial, "initial")

	// send update
	require.NoError(t, as.Store(Snapshot{
		Settings: first, Hash: "first",
	}))
	// receive update
	waitFor(first, "first")

	// send
	require.NoError(t, as.Store(Snapshot{
		Settings: replaced, Hash: "replaced",
	}))
	// replace
	require.NoError(t, as.Store(Snapshot{
		Settings: final, Hash: "final",
	}))

	// receive replace
	waitFor(final, "final")
}

func TestAtomic_Unsubscribe(t *testing.T) {
	var as Atomic
	ch, err := as.Subscribe(t.Context())
	require.NoError(t, err)
	other, err := as.Subscribe(t.Context())
	require.NoError(t, err)

	as.Unsubscribe(ch)

	select {
	case _, ok := <-ch:
		require.False(t, ok, "expected the channel to be closed")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the channel to close")
	}
	as.mu.RLock()
	require.Len(t, as.subs, 1)
	as.mu.RUnlock()

	as.Unsubscribe(ch) // a second call is a no-op

	require.NoError(t, as.Store(Snapshot{Settings: `Foo = "bar"`, Hash: "bar"}))
	select {
	case got := <-other:
		require.Equal(t, "bar", got.Hash)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the remaining subscriber")
	}
}
