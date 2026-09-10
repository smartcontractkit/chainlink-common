package registry

import (
	"context"
	"errors"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// atomicBaseCapability extends [capabilities.BaseCapability] to support atomic updates and forward client state checks.
type atomicBaseCapability interface {
	capabilities.BaseCapability
	Update(capabilities.BaseCapability) error
	StateGetter
}

var _ StateGetter = (*grpc.ClientConn)(nil)

// StateGetter is implemented by GRPC client connections.
type StateGetter interface {
	GetState() connectivity.State
}

var _ capabilities.TriggerCapability = &atomicTriggerCapability{}

type atomicTriggerCapability struct {
	mu            sync.RWMutex
	cap           capabilities.TriggerCapability
	registrations *triggerRegistrationManager
}

func newAtomicTriggerCapability(lggr logger.Logger) *atomicTriggerCapability {
	return &atomicTriggerCapability{
		registrations: newTriggerRegistrationManager(lggr),
	}
}

func (a *atomicTriggerCapability) Update(c capabilities.BaseCapability) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c == nil {
		a.cap = nil
		a.registrations.rebind(nil)
		return nil
	}
	tc, ok := c.(capabilities.TriggerCapability)
	if !ok {
		return errors.New("trigger capability does not satisfy TriggerCapability interface")
	}
	a.cap = tc
	a.registrations.rebind(tc)
	return nil
}

func (a *atomicTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return cap.Info(ctx)
}

func (a *atomicTriggerCapability) GetState() connectivity.State {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := cap.(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicTriggerCapability) AckEvent(ctx context.Context, triggerID string, eventID string, method string) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.AckEvent(ctx, triggerID, eventID, method)
}

func (a *atomicTriggerCapability) Load() *capabilities.TriggerCapability {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return nil
	}
	cap := a.cap
	return &cap
}

func (a *atomicTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cap == nil {
		return nil, errors.New("capability unavailable")
	}
	return a.registrations.register(ctx, a.cap, request)
}

func (a *atomicTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.registrations.unregister(ctx, a.cap, request)
}

var _ capabilities.ExecutableCapability = &atomicExecuteCapability{}

type atomicExecuteCapability struct {
	mu  sync.RWMutex
	cap capabilities.ExecutableCapability
}

func (a *atomicExecuteCapability) Update(c capabilities.BaseCapability) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c == nil {
		a.cap = nil
		return nil
	}
	tc, ok := c.(capabilities.ExecutableCapability)
	if !ok {
		return errors.New("action does not satisfy ExecutableCapability interface")
	}
	a.cap = tc
	return nil
}

func (a *atomicExecuteCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return cap.Info(ctx)
}

func (a *atomicExecuteCapability) GetState() connectivity.State {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := cap.(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicExecuteCapability) Load() *capabilities.ExecutableCapability {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return nil
	}
	cap := a.cap
	return &cap
}

func (a *atomicExecuteCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return cap.Execute(ctx, request)
}

var _ capabilities.ExecutableAndTriggerCapability = &atomicExecuteAndTriggerCapability{}

type atomicExecuteAndTriggerCapability struct {
	mu            sync.RWMutex
	cap           capabilities.ExecutableAndTriggerCapability
	registrations *triggerRegistrationManager
}

func newAtomicExecuteAndTriggerCapability(lggr logger.Logger) *atomicExecuteAndTriggerCapability {
	return &atomicExecuteAndTriggerCapability{
		registrations: newTriggerRegistrationManager(lggr),
	}
}

func (a *atomicExecuteAndTriggerCapability) Update(c capabilities.BaseCapability) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c == nil {
		a.cap = nil
		a.registrations.rebind(nil)
		return nil
	}
	tc, ok := c.(capabilities.ExecutableAndTriggerCapability)
	if !ok {
		return errors.New("target capability does not satisfy ExecutableAndTriggerCapability interface")
	}
	a.cap = tc
	a.registrations.rebind(tc)
	return nil
}

func (a *atomicExecuteAndTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return cap.Info(ctx)
}

func (a *atomicExecuteAndTriggerCapability) GetState() connectivity.State {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if a.cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := cap.(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicExecuteAndTriggerCapability) AckEvent(ctx context.Context, triggerID string, eventID string, method string) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.AckEvent(ctx, triggerID, eventID, method)
}

func (a *atomicExecuteAndTriggerCapability) Load() *capabilities.ExecutableAndTriggerCapability {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return nil
	}
	cap := a.cap
	return &cap
}

func (a *atomicExecuteAndTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cap == nil {
		return nil, errors.New("capability unavailable")
	}
	return a.registrations.register(ctx, a.cap, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.registrations.unregister(ctx, a.cap, request)
}

func (a *atomicExecuteAndTriggerCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return errors.New("capability unavailable")
	}
	return cap.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	a.mu.RLock()
	cap := a.cap
	a.mu.RUnlock()
	if cap == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return cap.Execute(ctx, request)
}
