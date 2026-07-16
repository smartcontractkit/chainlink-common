package host

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// ErrRunnerUnavailable indicates that a trigger's requirements are handled by a
// known runner type, but no runner can currently satisfy them, e.g. a capability
// that gates the runner is temporarily unavailable or not yet activated across the
// DON. It is distinct from a permanent "no such runner is configured": callers may
// treat it as transient and retry rather than failing workflow initialization.
var ErrRunnerUnavailable = errors.New("no runner can currently satisfy the trigger requirements")

type ModuleAndHandler struct {
	Module
	RequirementsHandler
}

// lazyModule wraps a ModuleAndHandler so that Start is called at most once
// and Close only fires for modules that were actually started. The mutex
// serializes start/close so a concurrent Close cannot race past an in-flight
// Start (leaving a started module unclosed) and vice versa.
type lazyModule struct {
	ModuleAndHandler
	mu      sync.Mutex
	started bool
	closed  bool
}

func (l *lazyModule) ensureStarted() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.started || l.closed {
		return
	}
	l.Module.Start()
	l.started = true
}

func (l *lazyModule) ensureClosed() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.closed = true
	if l.started {
		l.Module.Close()
	}
}

// NewRequirementSelectingModule creates a module that routes trigger executions
// based on subscription requirements. main is prepended as modules[0]; additional
// modules follow. Subscribe always runs on modules[0].
func NewRequirementSelectingModule(main ModuleAndHandler, additional []ModuleAndHandler) Module {
	modules := make([]*lazyModule, 1+len(additional))
	modules[0] = &lazyModule{ModuleAndHandler: main}
	for i, a := range additional {
		modules[1+i] = &lazyModule{ModuleAndHandler: a}
	}
	return &requirementSelectingModule{modules: modules}
}

type triggerInfo struct {
	moduleIdx    int
	preHook      bool
	requirements *sdk.Requirements
}

type requirementSelectingModule struct {
	modules []*lazyModule
	// triggerID → triggerInfo
	cache sync.Map
}

func (r *requirementSelectingModule) Start() {
	r.modules[0].ensureStarted()
}

func (r *requirementSelectingModule) Close() {
	for _, m := range r.modules {
		m.ensureClosed()
	}
}

func (r *requirementSelectingModule) IsLegacyDAG() bool {
	return r.modules[0].IsLegacyDAG()
}

func (r *requirementSelectingModule) Execute(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	if request.GetTrigger() == nil {
		return r.subscribe(ctx, request, handler)
	}
	return r.trigger(ctx, request, handler)
}

func (r *requirementSelectingModule) subscribe(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	result, err := r.modules[0].Execute(ctx, request, handler)
	if err != nil {
		return nil, err
	}

	for i, sub := range result.GetTriggerSubscriptions().GetSubscriptions() {
		matched := false
		for j, m := range r.modules {
			if CheckRequirements(ctx, m.RequirementsHandler, sub.Requirements) {
				m.ensureStarted()
				r.cache.Store(uint64(i), triggerInfo{moduleIdx: j, requirements: sub.Requirements, preHook: sub.PreHook})
				matched = true
				break
			}
		}
		if !matched {
			if r.runnerTypeAvailable(sub.Requirements) {
				// A runner of the right type exists but cannot currently satisfy the
				// requirements (e.g. a gated or temporarily unavailable capability).
				// Surface this as transient so callers can hold and retry rather than
				// failing workflow initialization outright.
				return nil, fmt.Errorf("%w for trigger %d", ErrRunnerUnavailable, i)
			}
			return nil, fmt.Errorf("cannot find a runner that can satisfy the requirements for trigger %d", i)
		}
	}

	return result, nil
}

// runnerTypeAvailable reports whether any module is the right type of runner for req
// (it handles every requirement field), even if none can currently satisfy the
// requirement values. Used to distinguish a transient ErrRunnerUnavailable from a
// permanent "no such runner".
func (r *requirementSelectingModule) runnerTypeAvailable(req *sdk.Requirements) bool {
	for _, m := range r.modules {
		if HandlesRequirements(m.RequirementsHandler, req) {
			return true
		}
	}
	return false
}

func (r *requirementSelectingModule) trigger(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	trigger := request.GetTrigger()
	if val, cached := r.cache.Load(trigger.Id); cached {
		info := val.(triggerInfo)

		m := r.modules[info.moduleIdx]
		if info.preHook {
			prehook := &sdk.ExecuteRequest{Request: &sdk.ExecuteRequest_PreHook{PreHook: trigger}}
			preHookResult, err := r.modules[0].Execute(ctx, prehook, handler)
			if err != nil {
				return nil, fmt.Errorf("pre-hook execution failed: %w", err)
			}

			switch preHookResult.Result.(type) {
			case *sdk.ExecutionResult_Error:
				return preHookResult, nil
			}

			restrictions := preHookResult.GetRestrictions()

			handler = NewRestrictedExecutionHelper(handler, restrictions)
			if rem, ok := m.Module.(RestrictionAwareModule); ok {
				rem.SetRestrictions(handler.GetWorkflowExecutionID(), restrictions)
			}
		}

		if rem, ok := m.Module.(RequirementEnforcingModule); ok && info.requirements != nil {
			rem.SetRequirements(handler.GetWorkflowExecutionID(), info.requirements)
		}

		return m.Execute(ctx, request, handler)
	}

	return nil, errors.New("cannot trigger before gathering subscriptions")
}

var _ Module = &requirementSelectingModule{}
