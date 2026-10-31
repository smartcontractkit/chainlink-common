package loop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

func TestAtomicSettings(t *testing.T) {
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
	var as AtomicSettings
	require.NoError(t, as.Store(core.SettingsUpdate{
		Settings: initial, Hash: "initial",
	}))

	ch, err := as.Subscribe(t.Context())
	require.NoError(t, err)

	waitFor := func(s, h string) {
		select {
		case got := <-ch:
			require.Equal(t, core.SettingsUpdate{
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
	require.NoError(t, as.Store(core.SettingsUpdate{
		Settings: first, Hash: "first",
	}))
	// receive update
	waitFor(first, "first")

	// send
	require.NoError(t, as.Store(core.SettingsUpdate{
		Settings: replaced, Hash: "replaced",
	}))
	// replace
	require.NoError(t, as.Store(core.SettingsUpdate{
		Settings: final, Hash: "final",
	}))

	// receive replace
	waitFor(final, "final")
}

func TestAtomicSettings_Unsubscribe(t *testing.T) {
	t.Parallel()
	var as AtomicSettings
	require.NoError(t, as.Store(core.SettingsUpdate{Settings: `Foo = "initial"`, Hash: "initial"}))

	ch, err := as.Subscribe(t.Context())
	require.NoError(t, err)
	other, err := as.Subscribe(t.Context())
	require.NoError(t, err)

	as.Unsubscribe(ch)
	as.Unsubscribe(ch) // no-op once unsubscribed

	// ch is closed once its helper goroutine exits; it may first deliver a value it already held.
	timeout := time.After(5 * time.Second)
	for closed := false; !closed; {
		select {
		case _, ok := <-ch:
			closed = !ok
		case <-timeout:
			t.Fatal("timed out waiting for unsubscribed channel to close")
		}
	}

	// Later updates must not touch the closed subscription, and must still reach the others.
	final := core.SettingsUpdate{Settings: `Foo = "final"`, Hash: "final"}
	require.NoError(t, as.Store(final))
	for got := (core.SettingsUpdate{}); got != final; {
		select {
		case got = <-other:
		case <-timeout:
			t.Fatal("timed out waiting for final update on remaining subscriber")
		}
	}

	as.mu.RLock()
	defer as.mu.RUnlock()
	assert.Len(t, as.subs, 1)
}
