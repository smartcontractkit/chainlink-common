package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

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
		switch info.CapabilityType {
		case capabilities.CapabilityTypeTrigger:
			ac = &atomicTriggerCapability{}
		case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget:
			ac = &atomicExecuteCapability{}
		case capabilities.CapabilityTypeCombined:
			ac = &atomicExecuteAndTriggerCapability{}
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

var _ capabilities.TriggerCapability = &atomicTriggerCapability{}

type atomicTriggerCapability struct {
	atomic.Pointer[capabilities.TriggerCapability]
}

func (a *atomicTriggerCapability) Update(c capabilities.BaseCapability) error {
	if c == nil {
		a.Store(nil)
		return nil
	}
	tc, ok := c.(capabilities.TriggerCapability)
	if !ok {
		return errors.New("trigger capability does not satisfy TriggerCapability interface")
	}
	a.Store(&tc)
	return nil
}

func (a *atomicTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	c := a.Load()
	if c == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return (*c).Info(ctx)
}

func (a *atomicTriggerCapability) GetState() connectivity.State {
	c := a.Load()
	if c == nil {
		return connectivity.Shutdown
	}
	if sg, ok := (*c).(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicTriggerCapability) AckEvent(ctx context.Context, eventId string) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).AckEvent(ctx, eventId)
}

func (a *atomicTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	c := a.Load()
	if c == nil {
		return nil, errors.New("capability unavailable")
	}
	return (*c).RegisterTrigger(ctx, request)
}

func (a *atomicTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).UnregisterTrigger(ctx, request)
}

var _ capabilities.ExecutableCapability = &atomicExecuteCapability{}

type atomicExecuteCapability struct {
	atomic.Pointer[capabilities.ExecutableCapability]
}

func (a *atomicExecuteCapability) Update(c capabilities.BaseCapability) error {
	if c == nil {
		a.Store(nil)
		return nil
	}
	tc, ok := c.(capabilities.ExecutableCapability)
	if !ok {
		return errors.New("action does not satisfy ExecutableCapability interface")
	}
	a.Store(&tc)
	return nil
}

func (a *atomicExecuteCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	c := a.Load()
	if c == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return (*c).Info(ctx)
}

func (a *atomicExecuteCapability) GetState() connectivity.State {
	c := a.Load()
	if c == nil {
		return connectivity.Shutdown
	}
	if sg, ok := (*c).(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicExecuteCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	c := a.Load()
	if c == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return (*c).Execute(ctx, request)
}

var _ capabilities.ExecutableAndTriggerCapability = &atomicExecuteAndTriggerCapability{}

type atomicExecuteAndTriggerCapability struct {
	atomic.Pointer[capabilities.ExecutableAndTriggerCapability]
}

func (a *atomicExecuteAndTriggerCapability) Update(c capabilities.BaseCapability) error {
	if c == nil {
		a.Store(nil)
		return nil
	}
	tc, ok := c.(capabilities.ExecutableAndTriggerCapability)
	if !ok {
		return errors.New("target capability does not satisfy ExecutableAndTriggerCapability interface")
	}
	a.Store(&tc)
	return nil
}

func (a *atomicExecuteAndTriggerCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	c := a.Load()
	if c == nil {
		return capabilities.CapabilityInfo{}, errors.New("capability unavailable")
	}
	return (*c).Info(ctx)
}

func (a *atomicExecuteAndTriggerCapability) GetState() connectivity.State {
	c := a.Load()
	if c == nil {
		return connectivity.Shutdown
	}
	if sg, ok := (*c).(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

func (a *atomicExecuteAndTriggerCapability) AckEvent(ctx context.Context, eventId string) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).AckEvent(ctx, eventId)
}

func (a *atomicExecuteAndTriggerCapability) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	c := a.Load()
	if c == nil {
		return nil, errors.New("capability unavailable")
	}
	return (*c).RegisterTrigger(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).UnregisterTrigger(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).RegisterToWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	c := a.Load()
	if c == nil {
		return errors.New("capability unavailable")
	}
	return (*c).UnregisterFromWorkflow(ctx, request)
}

func (a *atomicExecuteAndTriggerCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	c := a.Load()
	if c == nil {
		return capabilities.CapabilityResponse{}, errors.New("capability unavailable")
	}
	return (*c).Execute(ctx, request)
}
