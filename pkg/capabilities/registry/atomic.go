package registry

import (
	"context"
	"errors"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// atomicBaseCapability extends [capabilities.BaseCapability] to support atomic updates and forward client state checks.
type atomicBaseCapability interface {
	capabilities.BaseCapability
	// Update puts the capability in the wrapper, or removes the wrapped one when it is given nil.
	Update(context.Context, capabilities.BaseCapability) error
	// Replace is Update for a wrapper that may still be in use: it fails with a
	// *stillUpError, and changes nothing, when the connection of the wrapped capability is up.
	Replace(context.Context, capabilities.BaseCapability) error
	StateGetter
}

var _ StateGetter = (*grpc.ClientConn)(nil)

// StateGetter is implemented by GRPC client connections.
type StateGetter interface {
	GetState() connectivity.State
}

// replaceTimings are the limits that apply when a wrapped capability is replaced.
type replaceTimings struct {
	grace  time.Duration // see defaultReplaceGracePeriod
	replay time.Duration // see defaultReplayTimeout
}

// triggerRegistrations is the part that the wrappers of capabilities with triggers share.
type triggerRegistrations struct {
	// lock serialises registration changes. It is held across the call into the
	// capability, so waiting for it observes the caller's context.
	lock    ctxMutex
	manager *triggerRegistrationManager
}

func newTriggerRegistrations(lggr logger.Logger, replayTimeout time.Duration) triggerRegistrations {
	return triggerRegistrations{lock: newCtxMutex(), manager: newTriggerRegistrationManager(lggr, replayTimeout)}
}

var _ capabilities.TriggerCapability = &atomicTriggerCapability{}

type atomicTriggerCapability struct {
	*slot[capabilities.TriggerCapability]
	registrations triggerRegistrations
	timings       replaceTimings
}

func newAtomicTriggerCapability(lggr logger.Logger, timings replaceTimings) *atomicTriggerCapability {
	return &atomicTriggerCapability{
		slot:          newSlot[capabilities.TriggerCapability](),
		registrations: newTriggerRegistrations(lggr, timings.replay),
		timings:       timings,
	}
}

func (a *atomicTriggerCapability) Update(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, false)
}

func (a *atomicTriggerCapability) Replace(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, true)
}

func (a *atomicTriggerCapability) update(ctx context.Context, c capabilities.BaseCapability, onlyIfDown bool) error {
	var tc capabilities.TriggerCapability
	if c != nil {
		var ok bool
		if tc, ok = c.(capabilities.TriggerCapability); !ok {
			return errors.New("trigger capability does not satisfy TriggerCapability interface")
		}
	}
	if err := a.replace(ctx, tc, c != nil, onlyIfDown, a.registrations.lock, a.timings.grace); err != nil {
		return err
	}
	_, scope, _ := a.load()
	a.registrations.manager.rebind(a.registrations.lock, scope, tc)
	return nil
}

func (a *atomicTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}
	defer done()
	return current.Info(ctx)
}

func (a *atomicTriggerCapability) GetState() connectivity.State { return a.state() }

func (a *atomicTriggerCapability) AckEvent(ctx context.Context, triggerID string, eventID string, method string) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.AckEvent(ctx, triggerID, eventID, method)
}

func (a *atomicTriggerCapability) Load() *capabilities.TriggerCapability {
	current, _, ok := a.load()
	if !ok {
		return nil
	}
	return &current
}

func (a *atomicTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	if err := a.registrations.lock.lock(ctx); err != nil {
		return nil, err
	}
	defer a.registrations.lock.unlock()
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return a.registrations.manager.register(ctx, current, request)
}

func (a *atomicTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	if err := a.registrations.lock.lock(ctx); err != nil {
		return err
	}
	defer a.registrations.lock.unlock()
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return a.registrations.manager.unregister(ctx, current, request)
}

var _ capabilities.ExecutableCapability = &atomicExecuteCapability{}

type atomicExecuteCapability struct {
	*slot[capabilities.ExecutableCapability]
	timings replaceTimings
}

func newAtomicExecuteCapability(timings replaceTimings) *atomicExecuteCapability {
	return &atomicExecuteCapability{slot: newSlot[capabilities.ExecutableCapability](), timings: timings}
}

func (a *atomicExecuteCapability) Update(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, false)
}

func (a *atomicExecuteCapability) Replace(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, true)
}

func (a *atomicExecuteCapability) update(ctx context.Context, c capabilities.BaseCapability, onlyIfDown bool) error {
	var ec capabilities.ExecutableCapability
	if c != nil {
		var ok bool
		if ec, ok = c.(capabilities.ExecutableCapability); !ok {
			return errors.New("action does not satisfy ExecutableCapability interface")
		}
	}
	return a.replace(ctx, ec, c != nil, onlyIfDown, nil, a.timings.grace)
}

func (a *atomicExecuteCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}
	defer done()
	return current.Info(ctx)
}

func (a *atomicExecuteCapability) GetState() connectivity.State { return a.state() }

func (a *atomicExecuteCapability) Load() *capabilities.ExecutableCapability {
	current, _, ok := a.load()
	if !ok {
		return nil
	}
	return &current
}

func (a *atomicExecuteCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}
	defer done()
	return current.Execute(ctx, request)
}

var _ capabilities.ExecutableAndTriggerCapability = &atomicExecuteAndTriggerCapability{}

type atomicExecuteAndTriggerCapability struct {
	*slot[capabilities.ExecutableAndTriggerCapability]
	registrations triggerRegistrations
	timings       replaceTimings
}

func newAtomicExecuteAndTriggerCapability(lggr logger.Logger, timings replaceTimings) *atomicExecuteAndTriggerCapability {
	return &atomicExecuteAndTriggerCapability{
		slot:          newSlot[capabilities.ExecutableAndTriggerCapability](),
		registrations: newTriggerRegistrations(lggr, timings.replay),
		timings:       timings,
	}
}

func (a *atomicExecuteAndTriggerCapability) Update(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, false)
}

func (a *atomicExecuteAndTriggerCapability) Replace(ctx context.Context, c capabilities.BaseCapability) error {
	return a.update(ctx, c, true)
}

func (a *atomicExecuteAndTriggerCapability) update(ctx context.Context, c capabilities.BaseCapability, onlyIfDown bool) error {
	var tc capabilities.ExecutableAndTriggerCapability
	if c != nil {
		var ok bool
		if tc, ok = c.(capabilities.ExecutableAndTriggerCapability); !ok {
			return errors.New("target capability does not satisfy ExecutableAndTriggerCapability interface")
		}
	}
	if err := a.replace(ctx, tc, c != nil, onlyIfDown, a.registrations.lock, a.timings.grace); err != nil {
		return err
	}
	_, scope, _ := a.load()
	a.registrations.manager.rebind(a.registrations.lock, scope, tc)
	return nil
}

func (a *atomicExecuteAndTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return capabilities.CapabilityInfo{}, err
	}
	defer done()
	return current.Info(ctx)
}

func (a *atomicExecuteAndTriggerCapability) GetState() connectivity.State { return a.state() }

func (a *atomicExecuteAndTriggerCapability) AckEvent(ctx context.Context, triggerID string, eventID string, method string) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.AckEvent(ctx, triggerID, eventID, method)
}

func (a *atomicExecuteAndTriggerCapability) Load() *capabilities.ExecutableAndTriggerCapability {
	current, _, ok := a.load()
	if !ok {
		return nil
	}
	return &current
}

func (a *atomicExecuteAndTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	if err := a.registrations.lock.lock(ctx); err != nil {
		return nil, err
	}
	defer a.registrations.lock.unlock()
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return a.registrations.manager.register(ctx, current, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	if err := a.registrations.lock.lock(ctx); err != nil {
		return err
	}
	defer a.registrations.lock.unlock()
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return a.registrations.manager.unregister(ctx, current, request)
}

func (a *atomicExecuteAndTriggerCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return err
	}
	defer done()
	return current.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	current, ctx, done, err := a.call(ctx)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}
	defer done()
	return current.Execute(ctx, request)
}
