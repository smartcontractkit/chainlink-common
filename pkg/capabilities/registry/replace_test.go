package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

const (
	testGrace = 50 * time.Millisecond
	// notYet is how long a test waits to see that something does not happen.
	notYet = 4 * testGrace
)

func newTestRegistry(t *testing.T) *baseRegistry {
	r := NewBaseRegistry(logger.Test(t)).(*baseRegistry)
	r.timings = replaceTimings{grace: testGrace, replay: tests.WaitTimeout(t)}
	return r
}

// fakeCapability is a combined capability that can stop answering, as the client of a
// plugin does when the plugin is gone: a call then lasts until its context ends.
type fakeCapability struct {
	capabilities.CapabilityInfo

	mu            sync.Mutex
	state         connectivity.State
	hold          chan struct{} // while not nil, calls wait for it to be closed
	ignoreContext bool          // a waiting call does not return when its context ends
	// linger is how long a waiting call takes to return after its context ended. The client
	// of a plugin that is gone finishes an attempt to connect before it observes its context.
	linger        time.Duration
	inFlight      int
	registerCalls int
	registered    map[string]chan capabilities.TriggerResponse
}

func newFakeCapability(id string, typ capabilities.CapabilityType) *fakeCapability {
	return &fakeCapability{
		CapabilityInfo: capabilities.MustNewCapabilityInfo(id, typ, "fake"),
		state:          connectivity.Ready,
		registered:     map[string]chan capabilities.TriggerResponse{},
	}
}

// stopAnswering makes every later call wait. The returned func lets them go on.
func (f *fakeCapability) stopAnswering(state connectivity.State) (resume func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	hold := make(chan struct{})
	f.hold, f.state = hold, state
	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.hold == hold {
			close(hold)
			f.hold = nil
		}
	}
}

func (f *fakeCapability) call(ctx context.Context) error {
	f.mu.Lock()
	hold, ignore, linger := f.hold, f.ignoreContext, f.linger
	f.inFlight++
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		f.inFlight--
		f.mu.Unlock()
	}()
	if hold == nil {
		return ctx.Err()
	}
	if ignore {
		<-hold
		return nil
	}
	select {
	case <-hold:
		return nil
	case <-ctx.Done():
		time.Sleep(linger)
		return context.Cause(ctx)
	}
}

func (f *fakeCapability) callsInFlight() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.inFlight
}

func (f *fakeCapability) registerTriggerCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.registerCalls
}

func (f *fakeCapability) isRegistered(triggerID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.registered[triggerID]
	return ok
}

func (f *fakeCapability) sendEvent(triggerID, eventID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	ch, ok := f.registered[triggerID]
	if !ok {
		return false
	}
	select {
	case ch <- capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: eventID}}:
		return true
	default:
		return false
	}
}

func (f *fakeCapability) GetState() connectivity.State {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}

func (f *fakeCapability) RegisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	f.mu.Lock()
	f.registerCalls++
	f.mu.Unlock()
	if err := f.call(ctx); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	ch := make(chan capabilities.TriggerResponse, 10)
	f.registered[req.TriggerID] = ch
	return ch, nil
}

func (f *fakeCapability) UnregisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) error {
	if err := f.call(ctx); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.registered, req.TriggerID)
	return nil
}

func (f *fakeCapability) AckEvent(ctx context.Context, _, _, _ string) error { return f.call(ctx) }

func (f *fakeCapability) RegisterToWorkflow(ctx context.Context, _ capabilities.RegisterToWorkflowRequest) error {
	return f.call(ctx)
}

func (f *fakeCapability) UnregisterFromWorkflow(ctx context.Context, _ capabilities.UnregisterFromWorkflowRequest) error {
	return f.call(ctx)
}

func (f *fakeCapability) Execute(ctx context.Context, _ capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	return capabilities.CapabilityResponse{}, f.call(ctx)
}

func triggerRequest(id string) capabilities.TriggerRegistrationRequest {
	return capabilities.TriggerRegistrationRequest{TriggerID: id}
}

// goCall runs fn in the background and returns the channel its error arrives on.
func goCall(fn func() error) <-chan error {
	errCh := make(chan error, 1)
	go func() { errCh <- fn() }()
	return errCh
}

func requireInFlight(t *testing.T, f *fakeCapability, n int) {
	t.Helper()
	require.Eventually(t, func() bool { return f.callsInFlight() == n }, tests.WaitTimeout(t), time.Millisecond)
}

func requireReturns(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(tests.WaitTimeout(t)):
		require.FailNow(t, "the call did not return")
		return nil
	}
}

func requireStillWaiting(t *testing.T, errCh <-chan error) {
	t.Helper()
	select {
	case err := <-errCh:
		require.FailNow(t, "the call returned", "err: %v", err)
	case <-time.After(notYet):
	}
}

// The case this file is about: a plugin is gone, its capability is still in the registry,
// callers without a deadline are inside it, and the plugin's replacement is added.
func TestRegistry_ReplacesCapabilityThatNoLongerAnswers(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	for _, typ := range []capabilities.CapabilityType{capabilities.CapabilityTypeTrigger, capabilities.CapabilityTypeCombined} {
		t.Run(string(typ), func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			r := newTestRegistry(t)

			old := newFakeCapability(id, typ)
			require.NoError(t, r.Add(ctx, old))
			onNode, err := r.GetTrigger(ctx, id)
			require.NoError(t, err)
			events, err := onNode.RegisterTrigger(ctx, triggerRequest("kept"))
			require.NoError(t, err)

			old.stopAnswering(connectivity.TransientFailure)

			// None of these callers has a deadline. The first registration call holds the
			// registrations lock, and the second one waits for that lock.
			noDeadline := context.Background()
			inFlight := []<-chan error{
				goCall(func() error { return onNode.AckEvent(noDeadline, "kept", "event", "method") }),
				goCall(func() error { _, err := onNode.RegisterTrigger(noDeadline, triggerRequest("new")); return err }),
			}
			requireInFlight(t, old, 2)
			queued := goCall(func() error { return onNode.UnregisterTrigger(noDeadline, triggerRequest("other")) })

			replacement := newFakeCapability(id, typ)
			require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, replacement) })))

			for _, errCh := range inFlight {
				require.ErrorIs(t, requireReturns(t, errCh), ErrCapabilityReplaced)
			}
			// The call that waited for the lock returns too. It fails when it was ahead of the
			// replacement, and it goes to the new capability when it was behind.
			if err := requireReturns(t, queued); err != nil {
				require.ErrorIs(t, err, ErrCapabilityReplaced)
			}
			require.Zero(t, old.callsInFlight())

			// The registration made on the old capability is replayed onto the new one, and
			// its events arrive on the channel the caller already has.
			require.Eventually(t, func() bool { return replacement.sendEvent("kept", "from-replacement") },
				tests.WaitTimeout(t), time.Millisecond)
			select {
			case resp := <-events:
				assert.Equal(t, "from-replacement", resp.Event.ID)
			case <-time.After(tests.WaitTimeout(t)):
				require.FailNow(t, "no event from the replacement")
			}
			assert.False(t, replacement.isRegistered("new"), "a registration that failed must not be replayed")
		})
	}
}

// Calls that wait in line for a capability that no longer answers must not each take a
// turn on it once it is being replaced. Each turn lasts as long as one attempt to connect
// to the plugin that is gone, and the node's remote trigger publisher can have 100 calls
// in that line, so the replacement would wait for all of them.
func TestRegistry_CallsInLineForACapabilityThatIsBeingReplacedDoNotReachIt(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	const inLine = 20
	ctx := t.Context()
	r := newTestRegistry(t)

	old := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	old.linger = notYet
	require.NoError(t, r.Add(ctx, old))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)

	old.stopAnswering(connectivity.TransientFailure)
	noDeadline := context.Background()
	register := func(triggerID string) <-chan error {
		return goCall(func() error {
			_, registerErr := onNode.RegisterTrigger(noDeadline, triggerRequest(triggerID))
			return registerErr
		})
	}
	first := register("first")
	requireInFlight(t, old, 1)
	waiting := make([]<-chan error, 0, inLine)
	for i := range inLine {
		waiting = append(waiting, register(fmt.Sprintf("in-line-%d", i)))
	}
	time.Sleep(testGrace) // let them reach the lock, ahead of the replacement

	replacement := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, replacement) })))

	require.ErrorIs(t, requireReturns(t, first), ErrCapabilityReplaced)
	for _, errCh := range waiting {
		// A call that was ahead of the replacement fails. One that was behind it is made on
		// the new capability.
		if waitErr := requireReturns(t, errCh); waitErr != nil {
			require.ErrorIs(t, waitErr, ErrCapabilityReplaced)
		}
	}
	assert.Equal(t, 1, old.registerTriggerCalls(), "only the call that was in flight may reach the capability that was replaced")
}

// A capability whose calls never return, not even when they are cancelled, must hold up
// only its own replacement, and only for as long as the caller of Add allows.
func TestRegistry_CapabilityThatNeverReturnsDoesNotBlockTheRegistry(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)

	stuck := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	stuck.ignoreContext = true
	require.NoError(t, r.Add(ctx, stuck))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)

	resume := stuck.stopAnswering(connectivity.TransientFailure)
	registration := goCall(func() error {
		_, registerErr := onNode.RegisterTrigger(context.Background(), triggerRequest("t"))
		return registerErr
	})
	requireInFlight(t, stuck, 1)

	addCtx, cancel := context.WithCancel(ctx)
	replace := goCall(func() error { return r.Add(addCtx, newFakeCapability(id, capabilities.CapabilityTypeCombined)) })
	requireStillWaiting(t, replace)
	remove := goCall(func() error { return r.Remove(addCtx, id) })
	requireStillWaiting(t, remove)

	// The rest of the registry works while that Add waits.
	other := newFakeCapability("evm:ChainSelector:2@1.0.0", capabilities.CapabilityTypeCombined)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, other) })))
	require.NoError(t, requireReturns(t, goCall(func() error { _, getErr := r.Get(ctx, other.ID); return getErr })))
	require.NoError(t, requireReturns(t, goCall(func() error { _, listErr := r.List(ctx); return listErr })))
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Remove(ctx, other.ID) })))
	// So do the calls of the stuck entry that need no lock, and they answer at once.
	assert.Equal(t, connectivity.TransientFailure, onNode.(StateGetter).GetState())
	// A registration call with a deadline gives up at its deadline.
	shortCtx, cancelShort := context.WithTimeout(ctx, testGrace)
	defer cancelShort()
	_, err = onNode.RegisterTrigger(shortCtx, triggerRequest("u"))
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// Add and Remove give up when their caller does, and the entry is not changed.
	cancel()
	require.ErrorIs(t, requireReturns(t, replace), context.Canceled)
	require.ErrorIs(t, requireReturns(t, remove), context.Canceled)
	got, err := r.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, connectivity.TransientFailure, got.(StateGetter).GetState())

	// The replacement was given up, so a call made from now on is not ended by it: a
	// slow call is waited for.
	resume()
	require.NoError(t, requireReturns(t, registration))
	stuck.mu.Lock()
	stuck.ignoreContext = false
	stuck.mu.Unlock()
	resume = stuck.stopAnswering(connectivity.TransientFailure)
	later := goCall(func() error { return onNode.AckEvent(context.Background(), "t", "event", "method") })
	requireStillWaiting(t, later)
	resume()
	require.NoError(t, requireReturns(t, later))
}

// Add returns once the capability is in place. It does not wait for the registrations to
// be replayed, so its duration does not grow with their number or with a slow plugin.
func TestRegistry_AddDoesNotWaitForTheReplay(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)

	old := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	require.NoError(t, r.Add(ctx, old))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)
	triggerIDs := []string{"a", "b", "c"}
	for _, triggerID := range triggerIDs {
		_, err = onNode.RegisterTrigger(ctx, triggerRequest(triggerID))
		require.NoError(t, err)
	}
	old.stopAnswering(connectivity.TransientFailure)

	slow := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	resume := slow.stopAnswering(connectivity.Ready)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, slow) })))
	requireInFlight(t, slow, 1) // the replay has started, and waits in the first registration

	// A registration change waits for the replay, for as long as its caller allows.
	shortCtx, cancel := context.WithTimeout(ctx, testGrace)
	defer cancel()
	_, err = onNode.RegisterTrigger(shortCtx, triggerRequest("d"))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	// A call that needs no registration is not held up: it reaches the new capability.
	ack := goCall(func() error { return onNode.AckEvent(ctx, "a", "event", "method") })
	requireInFlight(t, slow, 2)

	resume()
	require.NoError(t, requireReturns(t, ack))
	for _, triggerID := range triggerIDs {
		require.Eventually(t, func() bool { return slow.isRegistered(triggerID) }, tests.WaitTimeout(t), time.Millisecond)
	}
}

// A replay that is in progress stops when its capability is replaced in turn, and the
// next capability gets every registration.
func TestRegistry_ReplayStopsWhenTheCapabilityIsReplacedAgain(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)

	first := newFakeCapability(id, capabilities.CapabilityTypeTrigger)
	require.NoError(t, r.Add(ctx, first))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)
	triggerIDs := []string{"a", "b", "c"}
	for _, triggerID := range triggerIDs {
		_, err = onNode.RegisterTrigger(ctx, triggerRequest(triggerID))
		require.NoError(t, err)
	}
	first.stopAnswering(connectivity.TransientFailure)

	second := newFakeCapability(id, capabilities.CapabilityTypeTrigger)
	second.stopAnswering(connectivity.TransientFailure) // it is gone before it has answered the replay
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, second) })))
	requireInFlight(t, second, 1)

	third := newFakeCapability(id, capabilities.CapabilityTypeTrigger)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, third) })))
	for _, triggerID := range triggerIDs {
		require.Eventually(t, func() bool { return third.isRegistered(triggerID) }, tests.WaitTimeout(t), time.Millisecond)
	}
	require.Eventually(t, func() bool { return second.callsInFlight() == 0 }, tests.WaitTimeout(t), time.Millisecond)
	for _, triggerID := range triggerIDs {
		assert.False(t, second.isRegistered(triggerID))
	}
	assert.Equal(t, 1, second.registerTriggerCalls(), "the replay must stop, not go on with the other registrations")
}

// A new capability that does not answer the replay holds the registrations lock for a
// bounded time for each registration, not forever.
func TestRegistry_ReplayIsBoundedForEachRegistration(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)
	r.timings.replay = testGrace

	old := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	require.NoError(t, r.Add(ctx, old))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)
	for _, triggerID := range []string{"a", "b"} {
		_, err = onNode.RegisterTrigger(ctx, triggerRequest(triggerID))
		require.NoError(t, err)
	}
	old.stopAnswering(connectivity.TransientFailure)

	silent := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	silent.stopAnswering(connectivity.Ready)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, silent) })))

	lock := r.m[id].(*atomicExecuteAndTriggerCapability).registrations.lock
	require.NoError(t, requireReturns(t, goCall(func() error { return lock.lock(ctx) })), "the replay must end and release the lock")
	lock.unlock()
	assert.Equal(t, 2, silent.registerTriggerCalls())
}

// Which connection states of the current capability allow its replacement.
func TestRegistry_AddReplacesOnlyACapabilityThatIsNotConnected(t *testing.T) {
	t.Parallel()
	const id = "write_evm@1.0.0"
	for state, replaced := range map[connectivity.State]bool{
		connectivity.Shutdown:         true,
		connectivity.TransientFailure: true,
		connectivity.Idle:             true,
		// A client whose peer is gone alternates between TransientFailure and Connecting.
		connectivity.Connecting: true,
		connectivity.Ready:      false,
	} {
		t.Run(state.String(), func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			r := newTestRegistry(t)
			current := newFakeCapability(id, capabilities.CapabilityTypeTarget)
			current.state = state
			require.NoError(t, r.Add(ctx, current))

			err := r.Add(ctx, newFakeCapability(id, capabilities.CapabilityTypeTarget))
			if replaced {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrCapabilityAlreadyExists)
			}
		})
	}
}

// Add for an ID whose capability is still up fails as before, and must not disturb the
// calls in flight on that capability.
func TestRegistry_AddOfDuplicateLeavesTheLiveCapabilityAlone(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	for _, typ := range []capabilities.CapabilityType{
		capabilities.CapabilityTypeTrigger, capabilities.CapabilityTypeAction, capabilities.CapabilityTypeCombined,
	} {
		t.Run(string(typ), func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			r := newTestRegistry(t)

			live := newFakeCapability(id, typ)
			require.NoError(t, r.Add(ctx, live))
			resume := live.stopAnswering(connectivity.Ready) // slow, but connected
			var slowCall <-chan error
			if typ == capabilities.CapabilityTypeAction {
				onNode, err := r.GetExecutable(ctx, id)
				require.NoError(t, err)
				slowCall = goCall(func() error {
					_, err := onNode.Execute(context.Background(), capabilities.CapabilityRequest{})
					return err
				})
			} else {
				onNode, err := r.GetTrigger(ctx, id)
				require.NoError(t, err)
				slowCall = goCall(func() error {
					_, err := onNode.RegisterTrigger(context.Background(), triggerRequest("t"))
					return err
				})
			}
			requireInFlight(t, live, 1)

			err := requireReturns(t, goCall(func() error { return r.Add(ctx, newFakeCapability(id, typ)) }))
			require.ErrorIs(t, err, ErrCapabilityAlreadyExists)
			require.ErrorContains(t, err, "id "+id+" found in registry: state READY")

			requireStillWaiting(t, slowCall)
			resume()
			require.NoError(t, requireReturns(t, slowCall))
		})
	}
}

// Remove while the capability is in use: its calls get the grace period, then they are
// ended, and the entry can be filled again.
func TestRegistry_RemoveEndsTheCallsInFlightAfterTheGracePeriod(t *testing.T) {
	t.Parallel()
	const id = "write_evm@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)
	// A long grace period, to tell "ended by the grace period" from "ended at once".
	r.timings.grace = notYet + 2*testGrace

	old := newFakeCapability(id, capabilities.CapabilityTypeTarget)
	require.NoError(t, r.Add(ctx, old))
	onNode, err := r.GetExecutable(ctx, id)
	require.NoError(t, err)

	old.stopAnswering(connectivity.Ready)
	execute := goCall(func() error {
		_, executeErr := onNode.Execute(context.Background(), capabilities.CapabilityRequest{})
		return executeErr
	})
	requireInFlight(t, old, 1)

	require.NoError(t, requireReturns(t, goCall(func() error { return r.Remove(ctx, id) })))
	_, err = onNode.Execute(ctx, capabilities.CapabilityRequest{})
	require.ErrorContains(t, err, "capability unavailable")
	assert.Equal(t, connectivity.Shutdown, onNode.(StateGetter).GetState())

	requireStillWaiting(t, execute)
	require.ErrorIs(t, requireReturns(t, execute), ErrCapabilityReplaced)

	replacement := newFakeCapability(id, capabilities.CapabilityTypeTarget)
	require.NoError(t, r.Add(ctx, replacement))
	_, err = onNode.Execute(ctx, capabilities.CapabilityRequest{})
	require.NoError(t, err)
}

// Remove keeps the trigger registrations of the entry, and the capability that is added
// under the same ID next gets them.
func TestRegistry_RemoveThenAddKeepsTheTriggerRegistrations(t *testing.T) {
	t.Parallel()
	const id = "evm:ChainSelector:1@1.0.0"
	ctx := t.Context()
	r := newTestRegistry(t)

	old := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	require.NoError(t, r.Add(ctx, old))
	onNode, err := r.GetTrigger(ctx, id)
	require.NoError(t, err)
	events, err := onNode.RegisterTrigger(ctx, triggerRequest("kept"))
	require.NoError(t, err)

	require.NoError(t, requireReturns(t, goCall(func() error { return r.Remove(ctx, id) })))
	_, err = onNode.RegisterTrigger(ctx, triggerRequest("new"))
	require.ErrorContains(t, err, "capability unavailable")

	replacement := newFakeCapability(id, capabilities.CapabilityTypeCombined)
	require.NoError(t, requireReturns(t, goCall(func() error { return r.Add(ctx, replacement) })))
	require.Eventually(t, func() bool { return replacement.sendEvent("kept", "from-replacement") },
		tests.WaitTimeout(t), time.Millisecond)
	select {
	case resp := <-events:
		assert.Equal(t, "from-replacement", resp.Event.ID)
	case <-time.After(tests.WaitTimeout(t)):
		require.FailNow(t, "no event from the replacement")
	}
}

func TestCtxMutex(t *testing.T) {
	t.Parallel()
	m := newCtxMutex()
	require.NoError(t, m.lock(t.Context()))

	ctx, cancel := context.WithTimeout(t.Context(), testGrace)
	defer cancel()
	require.ErrorIs(t, m.lock(ctx), context.DeadlineExceeded)

	// Waiters get the lock in the order they arrived.
	var order []int
	var wg sync.WaitGroup
	for i := range 5 {
		waiting := make(chan struct{})
		wg.Go(func() {
			close(waiting)
			require.NoError(t, m.lock(t.Context()))
			order = append(order, i)
			m.unlock()
		})
		<-waiting
		time.Sleep(testGrace / 5) // let the waiter reach the lock before the next one starts
	}
	m.unlock()
	wg.Wait()
	assert.Equal(t, []int{0, 1, 2, 3, 4}, order)

	// A lock that is free is taken even with a context that has ended.
	require.NoError(t, m.lock(ctx))
}
