package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var (
	ErrCapabilityAlreadyExists = errors.New("capability already exists")
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

type baseRegistry struct {
	m    map[string]atomicBaseCapability
	lggr logger.Logger
	mu   sync.RWMutex
}

var _ core.CapabilitiesRegistryBase = (*baseRegistry)(nil)

func NewBaseRegistry(lggr logger.Logger) core.CapabilitiesRegistryBase {
	return &baseRegistry{
		m:    map[string]atomicBaseCapability{},
		lggr: logger.Named(lggr, "registries.basic"),
	}
}

// Get gets a capability from the registry.
func (r *baseRegistry) Get(_ context.Context, id string) (capabilities.BaseCapability, error) {
	r.lggr.Debugw("get capability", "id", id)
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.m[id]
	if ok {
		return c, nil
	}

	// Find compatible version (>= requested version with same major)
	parts := strings.Split(id, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid capability id format: %s", id)
	}
	name, verStr := parts[0], parts[1]

	reqVer, err := semver.NewVersion(verStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in capability id %q: %w", id, err)
	}
	reqIsPrerelease := reqVer.Prerelease() != ""

	var bestCap capabilities.BaseCapability
	var bestVer *semver.Version
	for key, cap := range r.m {
		p := strings.Split(key, "@")
		if len(p) != 2 {
			continue
		}
		if p[0] != name {
			continue
		}
		v, err := semver.NewVersion(p[1])
		if err != nil {
			continue
		}
		if v.Major() != reqVer.Major() {
			continue
		}
		// If the request is stable, skip pre-release candidates
		if !reqIsPrerelease && v.Prerelease() != "" {
			continue
		}

		if v.GreaterThan(reqVer) {
			if bestVer == nil || v.LessThan(bestVer) {
				bestCap = cap
				bestVer = v
			}
		}
	}

	if bestCap != nil {
		r.lggr.Debugw("found compatible capability", "id", name+"@"+bestVer.String())
		return bestCap, nil
	}
	return nil, fmt.Errorf("no compatible capability found for id %s", id)
}

// GetTrigger gets a capability from the registry and tries to coerce it to the TriggerCapability interface.
func (r *baseRegistry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	tc, ok := c.(capabilities.TriggerCapability)
	if !ok {
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", id)
	}

	return tc, nil
}

// GetExecutable gets a capability from the registry and tries to coerce it to the ExecutableCapability interface.
func (r *baseRegistry) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	ac, ok := c.(capabilities.ExecutableCapability)
	if !ok {
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", id)
	}

	return ac, nil
}

func (r *baseRegistry) List(_ context.Context) ([]capabilities.BaseCapability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var cl []capabilities.BaseCapability
	for _, v := range r.m {
		cl = append(cl, v)
	}

	return cl, nil
}

func (r *baseRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	if c == nil {
		return errors.New("cannot add a nil capability to the registry")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}

	id := info.ID
	bc, ok := r.m[id]
	if ok {
		switch state := bc.GetState(); state {
		case connectivity.Shutdown, connectivity.TransientFailure, connectivity.Idle:
			// allow replace
		default:
			return fmt.Errorf("%w: id %s found in registry: state %s", ErrCapabilityAlreadyExists, id, state)
		}
		if err := bc.Update(c); err != nil {
			return fmt.Errorf("failed to update capability %s: %w", id, err)
		}
	} else {
		var ac atomicBaseCapability
		lggr := logger.With(r.lggr, "capabilityID", id)
		switch info.CapabilityType {
		case capabilities.CapabilityTypeTrigger:
			ac = newAtomicTriggerCapability(lggr)
		case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget:
			ac = &atomicExecuteCapability{}
		case capabilities.CapabilityTypeCombined:
			ac = newAtomicExecuteAndTriggerCapability(lggr)
		default:
			return fmt.Errorf("unknown capability type: %s", info.CapabilityType)
		}
		if err := ac.Update(c); err != nil {
			return err
		}
		r.m[id] = ac
	}
	r.lggr.Infow("capability added", "id", id, "type", info.CapabilityType, "description", info.Description, "version", info.Version())
	return nil
}

func (r *baseRegistry) Remove(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ac, ok := r.m[id]
	if !ok {
		return fmt.Errorf("unable to remove, capability not found: %s", id)
	}
	if err := ac.Update(nil); err != nil {
		return fmt.Errorf("failed to remove capability %s: %w", id, err)
	}
	r.lggr.Infow("capability removed", "id", id)
	return nil
}

// Caches all trigger registrations and replays them when the underlying capability is updated.
// Owns channels passed to the higher layer (Engine or Don2Don) and goroutines forwarding events
// from the underlying capability.
// NOT thread-safe - the caller is responsible for locking.
type triggerRegistrationManager struct {
	lggr logger.Logger
	regs map[string]*triggerRegistration
}

type triggerRegistration struct {
	request capabilities.TriggerRegistrationRequest
	outCh   chan capabilities.TriggerResponse
	cancel  context.CancelFunc // used to shut down the forwarding goroutine when the trigger is unregistered
}

func newTriggerRegistrationManager(lggr logger.Logger) *triggerRegistrationManager {
	return &triggerRegistrationManager{
		lggr: lggr,
		regs: make(map[string]*triggerRegistration),
	}
}

func (m *triggerRegistrationManager) register(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	in, err := underlying.RegisterTrigger(ctx, req)
	if err != nil {
		return nil, err
	}
	return m.upsertRegistration(req, nil, in), nil
}

func (m *triggerRegistrationManager) unregister(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) error {
	var out chan capabilities.TriggerResponse
	if reg, ok := m.regs[req.TriggerID]; ok {
		if reg.cancel != nil {
			reg.cancel()
		}
		out = reg.outCh
		delete(m.regs, req.TriggerID)
	}

	if out != nil {
		close(out)
	}
	return underlying.UnregisterTrigger(ctx, req)
}

func (m *triggerRegistrationManager) upsertRegistration(req capabilities.TriggerRegistrationRequest, outCh chan capabilities.TriggerResponse, in <-chan capabilities.TriggerResponse) chan capabilities.TriggerResponse {
	registrationID := req.TriggerID // Engine sets it to (workflowID, triggerIndex)
	regInMap, ok := m.regs[registrationID]
	if !ok {
		if outCh == nil {
			outCh = make(chan capabilities.TriggerResponse)
		}
		regInMap = &triggerRegistration{
			request: req,
			outCh:   outCh,
		}
		m.regs[registrationID] = regInMap
	} else {
		regInMap.request = req
		if outCh != nil {
			regInMap.outCh = outCh
		}
		if regInMap.cancel != nil {
			regInMap.cancel() // shuts down the previous forwarding goroutine
			regInMap.cancel = nil
		}
	}
	if in != nil {
		ctxForward, cancel := context.WithCancel(context.Background())
		regInMap.cancel = cancel
		go forwardTriggerResponses(ctxForward, in, regInMap.outCh)
	}
	return regInMap.outCh
}

func (m *triggerRegistrationManager) rebind(newUnderlying capabilities.TriggerExecutable) {
	for _, reg := range m.regs {
		var in <-chan capabilities.TriggerResponse
		var err error
		if newUnderlying != nil {
			// NOTE: this is tricky - if an existing registration fails on rebind, there is no way to notify the user ...
			in, err = newUnderlying.RegisterTrigger(context.Background(), reg.request)
		}
		if err != nil {
			m.lggr.Errorw("failed to rebind trigger registration", "triggerID", reg.request.TriggerID, "err", err)
		} else {
			m.lggr.Debugw("rebind trigger registration", "triggerID", reg.request.TriggerID)
		}
		_ = m.upsertRegistration(reg.request, reg.outCh, in) // user already has this channel
	}
}

func forwardTriggerResponses(ctx context.Context, in <-chan capabilities.TriggerResponse, out chan<- capabilities.TriggerResponse) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-in:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- resp:
			}
		}
	}
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
	defer a.mu.RUnlock()
	if a.cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return a.cap.Info(ctx)
}

func (a *atomicTriggerCapability) GetState() connectivity.State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := a.cap.(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
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
	defer a.mu.RUnlock()
	if a.cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return a.cap.Info(ctx)
}

func (a *atomicExecuteCapability) GetState() connectivity.State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := a.cap.(StateGetter); ok {
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
	defer a.mu.RUnlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.cap.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.cap.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return a.cap.Execute(ctx, request)
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
	defer a.mu.RUnlock()
	if a.cap == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return a.cap.Info(ctx)
}

func (a *atomicExecuteAndTriggerCapability) GetState() connectivity.State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return connectivity.Shutdown
	}
	if sg, ok := a.cap.(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
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
	defer a.mu.RUnlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.cap.RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return errors.New("capability unavailable")
	}
	return a.cap.UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cap == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return a.cap.Execute(ctx, request)
}
