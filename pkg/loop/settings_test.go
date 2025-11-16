package loop

import (
	"testing"
	"time"

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
