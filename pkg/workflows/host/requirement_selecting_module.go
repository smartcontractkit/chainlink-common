package host

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

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
				r.cache.Store(uint64(i), triggerInfo{moduleIdx: j, requirements: sub.Requirements})
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("cannot find a runner that can satisfy the requirements for trigger %d", i)
		}
	}

	return result, nil
}

func (r *requirementSelectingModule) trigger(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	trigger := request.GetTrigger()
	if val, cached := r.cache.Load(trigger.Id); cached {
		info := val.(triggerInfo)
		m := r.modules[info.moduleIdx]
		if rem, ok := m.Module.(RequirementEnforcingModule); ok && info.requirements != nil {
			rem.SetRequirements(handler.GetWorkflowExecutionID(), info.requirements)
		}

		return m.Execute(ctx, request, handler)
	}
	return nil, errors.New("cannot trigger before gathering subscriptions")
}

var _ Module = &requirementSelectingModule{}
