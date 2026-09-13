package registry_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type mockCapability struct {
	capabilities.CapabilityInfo
}

func (m *mockCapability) Execute(_ context.Context, _ capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	return capabilities.CapabilityResponse{}, nil
}

func (m *mockCapability) RegisterToWorkflow(_ context.Context, _ capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (m *mockCapability) UnregisterFromWorkflow(_ context.Context, _ capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func TestRegistry(t *testing.T) {
	r := registry.NewBaseRegistry(logger.Test(t))
	ctx := t.Context()

	id := "capability-1@1.0.0"
	ci, err := capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeAction,
		"capability-1-description",
	)
	require.NoError(t, err)

	c := &mockCapability{CapabilityInfo: ci}
	err = r.Add(ctx, c)
	require.NoError(t, err)

	gc, err := r.Get(ctx, id)
	require.NoError(t, err)
	info, err := gc.Info(t.Context())
	require.NoError(t, err)

	assert.Equal(t, c.CapabilityInfo, info)

	cs, err := r.List(ctx)
	require.NoError(t, err)
	assert.Len(t, cs, 1)
	info, err = cs[0].Info(t.Context())
	require.NoError(t, err)
	assert.Equal(t, c.CapabilityInfo, info)
}

func TestRegistryCompatibleVersions(t *testing.T) {
	ctx := t.Context()

	t.Run("Compatible minor version", func(t *testing.T) {
		r := registry.NewBaseRegistry(logger.Test(t))
		id := "capability-1@1.5.0"
		ci, err := capabilities.NewCapabilityInfo(
			id,
			capabilities.CapabilityTypeAction,
			"capability-1-description",
		)
		require.NoError(t, err)

		c := &mockCapability{CapabilityInfo: ci}
		err = r.Add(ctx, c)
		require.NoError(t, err)
		_, err = r.Get(ctx, "capability-1@1.0.0")
		require.NoError(t, err)
	})

	t.Run("Incompatible minor version", func(t *testing.T) {
		r := registry.NewBaseRegistry(logger.Test(t))
		id := "capability-1@1.1.0"
		ci, err := capabilities.NewCapabilityInfo(
			id,
			capabilities.CapabilityTypeAction,
			"capability-1-description",
		)
		require.NoError(t, err)

		c := &mockCapability{CapabilityInfo: ci}
		err = r.Add(ctx, c)
		require.NoError(t, err)
		_, err = r.Get(ctx, "capability-1@1.2.0")
		require.Error(t, err)
	})

	t.Run("Incompatible major version", func(t *testing.T) {
		r := registry.NewBaseRegistry(logger.Test(t))
		id := "capability-1@2.0.0"
		ci, err := capabilities.NewCapabilityInfo(
			id,
			capabilities.CapabilityTypeAction,
			"capability-1-description",
		)
		require.NoError(t, err)

		c := &mockCapability{CapabilityInfo: ci}
		err = r.Add(ctx, c)
		require.NoError(t, err)
		_, err = r.Get(ctx, "capability-1@1.0.0")
		require.Error(t, err)
	})

	t.Run("Don't match pre-release tags if requested version if not pre-release", func(t *testing.T) {
		r := registry.NewBaseRegistry(logger.Test(t))
		id := "capability-1@1.5.0-alpha"
		ci, err := capabilities.NewCapabilityInfo(
			id,
			capabilities.CapabilityTypeAction,
			"capability-1-description",
		)
		require.NoError(t, err)

		c := &mockCapability{CapabilityInfo: ci}
		err = r.Add(ctx, c)
		require.NoError(t, err)
		_, err = r.Get(ctx, "capability-1@1.0.0")
		require.Error(t, err)
	})

	t.Run("Match pre-release tags if requested version is pre-release", func(t *testing.T) {
		r := registry.NewBaseRegistry(logger.Test(t))
		id := "capability-1@1.5.0-alpha"
		ci, err := capabilities.NewCapabilityInfo(
			id,
			capabilities.CapabilityTypeAction,
			"capability-1-description",
		)
		require.NoError(t, err)

		c := &mockCapability{CapabilityInfo: ci}
		err = r.Add(ctx, c)
		require.NoError(t, err)
		_, err = r.Get(ctx, "capability-1@1.0.0-alpha")
		require.NoError(t, err)
	})
}

func TestRegistry_NoDuplicateIDs(t *testing.T) {
	r := registry.NewBaseRegistry(logger.Test(t))
	ctx := t.Context()

	id := "capability-1@1.0.0"
	ci, err := capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeAction,
		"capability-1-description",
	)
	require.NoError(t, err)

	c := &mockCapability{CapabilityInfo: ci}
	err = r.Add(ctx, c)
	require.NoError(t, err)

	ci, err = capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeConsensus,
		"capability-2-description",
	)
	require.NoError(t, err)
	c2 := &mockCapability{CapabilityInfo: ci}

	err = r.Add(ctx, c2)
	assert.ErrorIs(t, err, registry.ErrCapabilityAlreadyExists)
}

func TestRegistry_ChecksExecutionAPIByType(t *testing.T) {
	tcs := []struct {
		name          string
		newCapability func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error)
		getCapability func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error
		errContains   string
	}{
		{
			name: "action",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeAction,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
		{
			name: "target",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeTarget,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
		{
			name: "trigger",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				odt := triggers.NewOnDemand(logger.Test(t))
				info, err := odt.Info(ctx)
				require.NoError(t, err)
				return info.ID, reg.Add(ctx, odt)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetTrigger(ctx, id)
				return err
			},
		},
		{
			name: "consensus",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeConsensus,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
	}

	ctx := t.Context()
	reg := registry.NewBaseRegistry(logger.Test(t))
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.newCapability(ctx, reg)
			require.NoError(t, err)

			err = tc.getCapability(ctx, reg, id)
			require.NoError(t, err)
		})
	}
}

type testTrigger struct {
	capabilities.CapabilityInfo
	mu            sync.Mutex
	registrations map[string]chan capabilities.TriggerResponse
	registerCount atomic.Int32
	inFlight      atomic.Int32
	maxInFlight   atomic.Int32
	failAfter     int32         // fail RegisterTrigger after this many successful calls (-1 = never fail)
	registerDelay time.Duration // delay each RegisterTrigger call to simulate a slow RPC
}

func newTestTrigger(name string) *testTrigger {
	return newTestTriggerWithFailures(name, -1)
}

func newTestTriggerWithFailures(name string, failAfter int32) *testTrigger {
	info := capabilities.MustNewCapabilityInfo(
		name+"@1.0.0",
		capabilities.CapabilityTypeTrigger,
		name,
	)
	return &testTrigger{
		CapabilityInfo: info,
		registrations:  make(map[string]chan capabilities.TriggerResponse),
		failAfter:      failAfter,
	}
}

func (t *testTrigger) RegisterTrigger(_ context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	count := t.registerCount.Add(1)

	// track how many RegisterTrigger calls are in flight concurrently
	inFlight := t.inFlight.Add(1)
	for {
		cur := t.maxInFlight.Load()
		if inFlight <= cur || t.maxInFlight.CompareAndSwap(cur, inFlight) {
			break
		}
	}
	defer t.inFlight.Add(-1)

	if t.registerDelay > 0 {
		time.Sleep(t.registerDelay)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.failAfter >= 0 && count > t.failAfter {
		return nil, errors.New("simulated registration failure")
	}
	ch := make(chan capabilities.TriggerResponse, 10)
	t.registrations[req.TriggerID] = ch
	return ch, nil
}

func (t *testTrigger) UnregisterTrigger(_ context.Context, req capabilities.TriggerRegistrationRequest) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if ch, ok := t.registrations[req.TriggerID]; ok {
		close(ch)
		delete(t.registrations, req.TriggerID)
	}
	return nil
}

func (t *testTrigger) AckEvent(_ context.Context, triggerID string, eventID string, method string) error {
	return nil
}

func (t *testTrigger) SendEvent(triggerID string, resp capabilities.TriggerResponse) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	ch, ok := t.registrations[triggerID]
	if !ok {
		return false
	}
	select {
	case ch <- resp:
		return true
	default:
		return false
	}
}

func (t *testTrigger) GetRegistrationCount() int32 {
	return t.registerCount.Load()
}

// GetMaxInFlight returns the highest number of RegisterTrigger calls observed concurrently.
func (t *testTrigger) GetMaxInFlight() int32 {
	return t.maxInFlight.Load()
}

func (t *testTrigger) GetState() connectivity.State {
	return connectivity.Shutdown
}

func TestAtomicTrigger_RegistrationsReplayed(t *testing.T) {
	ctx := t.Context()
	r := registry.NewBaseRegistry(logger.Test(t))

	trigger1 := newTestTriggerWithFailures("trigger", 1) // fails on 2nd call to RegisterTrigger
	require.NoError(t, r.Add(ctx, trigger1))

	tc, err := r.GetTrigger(ctx, "trigger@1.0.0")
	require.NoError(t, err)

	outCh, err := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "reg1"})
	require.NoError(t, err)
	require.NotNil(t, outCh)

	_, err = tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "reg2"})
	require.Error(t, err)

	trigger2 := newTestTrigger("trigger")
	require.NoError(t, r.Add(ctx, trigger2))                    // replace with a new trigger
	require.Equal(t, int32(1), trigger2.GetRegistrationCount()) // only successful registration replayed

	trigger2.SendEvent("reg1", capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: "event1"}})

	resp1 := <-outCh
	assert.Equal(t, "event1", resp1.Event.ID)
}

func TestAtomicTrigger_ConcurrentRegistrations(t *testing.T) {
	ctx := t.Context()
	r := registry.NewBaseRegistry(logger.Test(t))

	const (
		nTriggers = 10
		delay     = 100 * time.Millisecond
	)

	trigger := newTestTrigger("trigger")
	trigger.registerDelay = delay
	require.NoError(t, r.Add(ctx, trigger))

	tc, err := r.GetTrigger(ctx, "trigger@1.0.0")
	require.NoError(t, err)

	outChs := make([]<-chan capabilities.TriggerResponse, nTriggers)
	start := time.Now()
	var wg sync.WaitGroup
	for i := range nTriggers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outCh, err := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: fmt.Sprintf("reg%d", i)})
			assert.NoError(t, err)
			outChs[i] = outCh
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	require.Equal(t, int32(nTriggers), trigger.GetRegistrationCount())
	// Registrations of different triggers must not serialize on a single lock: each call to the
	// underlying capability takes delay, so a serialized implementation would need at least
	// nTriggers*delay in total.
	assert.GreaterOrEqual(t, trigger.GetMaxInFlight(), int32(2), "registrations should run concurrently in the underlying capability")
	assert.Less(t, elapsed, time.Duration(nTriggers)*delay, "concurrent registrations should complete faster than serial ones")

	// events flow through every registration's channel
	for i := range nTriggers {
		require.NotNil(t, outChs[i])
		require.True(t, trigger.SendEvent(fmt.Sprintf("reg%d", i), capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: fmt.Sprintf("event%d", i)}}))
	}
	for i := range nTriggers {
		select {
		case resp := <-outChs[i]:
			assert.Equal(t, fmt.Sprintf("event%d", i), resp.Event.ID)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for event on registration %d", i)
		}
	}
}

func TestAtomicTrigger_ConcurrentSameTriggerRegistrations(t *testing.T) {
	ctx := t.Context()
	r := registry.NewBaseRegistry(logger.Test(t))

	trigger := newTestTrigger("trigger")
	trigger.registerDelay = 10 * time.Millisecond
	require.NoError(t, r.Add(ctx, trigger))

	tc, err := r.GetTrigger(ctx, "trigger@1.0.0")
	require.NoError(t, err)

	const n = 5
	outChs := make([]<-chan capabilities.TriggerResponse, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outCh, err := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "same-reg"})
			assert.NoError(t, err)
			outChs[i] = outCh
		}()
	}
	wg.Wait()

	// every call registers with the underlying capability and all callers of the same trigger
	// share one channel, matching the sequential behaviour
	require.Equal(t, int32(n), trigger.GetRegistrationCount())
	for i := range n {
		require.NotNil(t, outChs[i])
		assert.Equal(t, outChs[0], outChs[i])
	}

	require.True(t, trigger.SendEvent("same-reg", capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: "event1"}}))
	select {
	case resp := <-outChs[0]:
		assert.Equal(t, "event1", resp.Event.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestAtomicTrigger_Unregister(t *testing.T) {
	ctx := t.Context()
	r := registry.NewBaseRegistry(logger.Test(t))

	trigger := newTestTrigger("trigger")
	require.NoError(t, r.Add(ctx, trigger))

	tc, err := r.GetTrigger(ctx, "trigger@1.0.0")
	require.NoError(t, err)

	outCh, err := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "reg1"})
	require.NoError(t, err)

	// unregistering an unknown trigger is a no-op on the cache and still hits the underlying capability
	require.NoError(t, tc.UnregisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "unknown"}))

	require.NoError(t, tc.UnregisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "reg1"}))

	// the caller's channel is closed and the registration is not replayed on rebind
	select {
	case _, ok := <-outCh:
		require.False(t, ok, "expected the registration channel to be closed")
	case <-time.After(2 * time.Second):
		t.Fatal("expected the registration channel to be closed")
	}

	trigger2 := newTestTrigger("trigger")
	require.NoError(t, r.Add(ctx, trigger2))
	assert.Equal(t, int32(0), trigger2.GetRegistrationCount())

	// re-registering after an unregister works and gets a fresh channel
	outCh2, err := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "reg1"})
	require.NoError(t, err)
	require.NotNil(t, outCh2)
	assert.NotEqual(t, outCh, outCh2)

	require.True(t, trigger2.SendEvent("reg1", capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: "event2"}}))
	select {
	case resp := <-outCh2:
		assert.Equal(t, "event2", resp.Event.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestAtomicTrigger_ConcurrentRegisterUnregisterUpdate(t *testing.T) {
	ctx := t.Context()
	r := registry.NewBaseRegistry(logger.Test(t))

	trigger := newTestTrigger("trigger")
	trigger.registerDelay = time.Millisecond
	require.NoError(t, r.Add(ctx, trigger))

	tc, err := r.GetTrigger(ctx, "trigger@1.0.0")
	require.NoError(t, err)

	const (
		nGoroutines = 8
		nIterations = 25
		nTriggers   = 5
	)

	var wg sync.WaitGroup
	for g := range nGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range nIterations {
				id := fmt.Sprintf("reg%d", (g+it)%nTriggers)
				switch it % 3 {
				case 0, 1:
					outCh, regErr := tc.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: id})
					assert.NoError(t, regErr)
					if outCh != nil {
						// the channel may already be closed by a concurrent unregister
						select {
						case _, ok := <-outCh:
							assert.False(t, ok)
						default:
						}
					}
				case 2:
					assert.NoError(t, tc.UnregisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: id}))
				}
				if it%7 == 0 {
					// swap the underlying capability to exercise rebind concurrently
					assert.NoError(t, r.Add(ctx, newTestTrigger("trigger")))
				}
			}
		}()
	}
	wg.Wait()

	// after the storm, unregister everything and swap in a fresh capability: nothing should be replayed
	for i := range nTriggers {
		require.NoError(t, tc.UnregisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: fmt.Sprintf("reg%d", i)}))
	}
	final := newTestTrigger("trigger")
	require.NoError(t, r.Add(ctx, final))
	assert.Equal(t, int32(0), final.GetRegistrationCount())
}
