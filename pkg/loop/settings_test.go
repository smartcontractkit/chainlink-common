package loop

import (
	"fmt"
	"sync"
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

// TestAtomicSettings_SubscribeConcurrentStore checks that a Store racing with
// Subscribe is never lost: each subscriber must end up with the latest value,
// either as its seed or as a later fan-out.
func TestAtomicSettings_SubscribeConcurrentStore(t *testing.T) {
	t.Parallel()
	const (
		rounds              = 50
		subscribersPerRound = 20
		storesPerRound      = 5
	)
	var as AtomicSettings

	update := func(round, i int) core.SettingsUpdate {
		return core.SettingsUpdate{
			Settings: fmt.Sprintf(`Foo = "%d-%d"`, round, i),
			Hash:     fmt.Sprintf("%d-%d", round, i),
		}
	}
	type subscriber struct {
		ch     <-chan core.SettingsUpdate
		latest core.SettingsUpdate
	}
	waitFor := func(s *subscriber, want core.SettingsUpdate) {
		t.Helper()
		timeout := time.After(5 * time.Second)
		for s.latest != want {
			select {
			case s.latest = <-s.ch:
			case <-timeout:
				t.Fatalf("subscriber never observed update %q; latest was %q", want.Hash, s.latest.Hash)
			}
		}
	}

	var all []*subscriber
	var final core.SettingsUpdate
	for round := range rounds {
		final = update(round, storesPerRound-1)
		subs := make([]*subscriber, subscribersPerRound)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := range subs {
			wg.Go(func() {
				<-start
				ch, err := as.Subscribe(t.Context())
				assert.NoError(t, err)
				subs[i] = &subscriber{ch: ch}
			})
		}
		wg.Go(func() {
			<-start
			for i := range storesPerRound {
				assert.NoError(t, as.Store(update(round, i)))
			}
		})
		close(start)
		wg.Wait()

		for _, s := range subs {
			require.NotNil(t, s.ch)
			waitFor(s, final)
		}
		all = append(all, subs...)
	}

	// Subscribers from earlier rounds must also have caught up via fan-out.
	for _, s := range all {
		waitFor(s, final)
	}
}
