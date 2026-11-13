package settings

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jpillora/backoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestFollow_AppliesUpdates(t *testing.T) {
	t.Parallel()
	dst := &Atomic{}
	host := newFakeBroadcaster()
	runFollow(t, host, dst)

	host.send(t, Snapshot{Settings: `Foo = "a"`, Hash: "a"})
	requireHash(t, dst, "a")
	host.send(t, Snapshot{Settings: `Foo = "b"`, Hash: "b"})
	requireHash(t, dst, "b")
}

func TestFollow_KeepsThePreviousSettingsOnFailure(t *testing.T) {
	t.Parallel()
	dst := &Atomic{}
	host := newFakeBroadcaster()
	runFollow(t, host, dst)

	host.send(t, Snapshot{Settings: `Foo = "a"`, Hash: "a"})
	requireHash(t, dst, "a")
	host.send(t, Snapshot{Settings: "not a settings payload", Hash: "bad"})
	host.send(t, Snapshot{Settings: `Foo = "c"`, Hash: "c"})
	requireHash(t, dst, "c")
}

func TestFollow_Resubscribes(t *testing.T) {
	t.Parallel()
	dst := &Atomic{}
	host := newFakeBroadcaster()
	host.failFirst.Store(true)
	runFollow(t, host, dst)

	host.send(t, Snapshot{Settings: `Foo = "a"`, Hash: "a"})
	requireHash(t, dst, "a")
	host.closeStream()
	host.send(t, Snapshot{Settings: `Foo = "b"`, Hash: "b"})
	requireHash(t, dst, "b")

	assert.GreaterOrEqual(t, host.subscriptions.Load(), int32(3), "one failed subscription, then one per stream")
}

func TestFollow_BacksOffWhenStreamsEndImmediately(t *testing.T) {
	t.Parallel()
	host := &endingBroadcaster{}
	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()

	follow(ctx, logger.Test(t), host, &Atomic{}, &backoff.Backoff{Min: 20 * time.Millisecond, Max: time.Second, Factor: 2})

	// 20+40+80+160ms of backoff fits about five subscriptions in the window; without backoff it is thousands.
	assert.LessOrEqual(t, host.subscriptions.Load(), int32(6))
}

func TestFollow_ReturnsOnCancelWithAnOpenStream(t *testing.T) {
	t.Parallel()
	host := &Atomic{}
	require.NoError(t, host.Store(Snapshot{Settings: `Foo = "a"`, Hash: "a"}))
	dst := &Atomic{}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		Follow(ctx, logger.Test(t), host, dst)
	}()
	requireHash(t, dst, "a")

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Follow did not return after its context was cancelled")
	}
	host.mu.RLock()
	defer host.mu.RUnlock()
	assert.Empty(t, host.subs, "Follow should unsubscribe from a local broadcaster")
}

// runFollow runs follow with a short backoff until the test ends, and waits for it to return.
func runFollow(t *testing.T, b Broadcaster, dst *Atomic) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		follow(ctx, logger.Test(t), b, dst, &backoff.Backoff{Min: time.Millisecond, Max: 10 * time.Millisecond})
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

func requireHash(t *testing.T, a *Atomic, want string) {
	t.Helper()
	require.Eventually(t, func() bool {
		got, err := a.Load()
		return err == nil && got.Hash == want
	}, 5*time.Second, time.Millisecond, "settings %q were not applied", want)
}

// fakeBroadcaster hands out one stream per subscription, and sends on whichever is current.
type fakeBroadcaster struct {
	failFirst     atomic.Bool
	subscriptions atomic.Int32
	streams       chan chan Snapshot
	current       chan Snapshot
}

func newFakeBroadcaster() *fakeBroadcaster {
	return &fakeBroadcaster{streams: make(chan chan Snapshot, 1)}
}

func (f *fakeBroadcaster) Subscribe(ctx context.Context) (<-chan Snapshot, error) {
	f.subscriptions.Add(1)
	if f.failFirst.CompareAndSwap(true, false) {
		return nil, errors.New("host unavailable")
	}
	ch := make(chan Snapshot)
	select {
	case f.streams <- ch:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return ch, nil
}

func (f *fakeBroadcaster) send(t *testing.T, s Snapshot) {
	t.Helper()
	if f.current == nil {
		select {
		case f.current = <-f.streams:
		case <-time.After(5 * time.Second):
			t.Fatal("nothing subscribed")
		}
	}
	select {
	case f.current <- s:
	case <-time.After(5 * time.Second):
		t.Fatal("the update was not received")
	}
}

func (f *fakeBroadcaster) closeStream() {
	close(f.current)
	f.current = nil
}

// endingBroadcaster hands out streams that end before delivering anything.
type endingBroadcaster struct {
	subscriptions atomic.Int32
}

func (e *endingBroadcaster) Subscribe(context.Context) (<-chan Snapshot, error) {
	e.subscriptions.Add(1)
	ch := make(chan Snapshot)
	close(ch)
	return ch, nil
}
