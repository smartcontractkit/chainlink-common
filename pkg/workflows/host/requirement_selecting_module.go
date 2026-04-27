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

// lazyModule wraps a ModuleAndHandler so that Start is called at most once.
type lazyModule struct {
	ModuleAndHandler
	startOnce sync.Once
	started   bool
}

func (l *lazyModule) ensureStarted() {
	l.startOnce.Do(func() {
		l.Module.Start()
		l.started = true
	})
}

func NewRequirementSelectingModule(main ModuleAndHandler, additional []ModuleAndHandler) Module {
	wrapped := make([]*lazyModule, len(additional))
	for i := range additional {
		wrapped[i] = &lazyModule{ModuleAndHandler: additional[i]}
	}
	return &requirementSelectingModule{
		main:       main,
		additional: wrapped,
	}
}

type requirementSelectingModule struct {
	main       ModuleAndHandler
	additional []*lazyModule
	// triggerID → index into additional
	cache sync.Map
}

func (r *requirementSelectingModule) Start() {
	r.main.Start()
}

func (r *requirementSelectingModule) Close() {
	r.main.Close()
	for _, m := range r.additional {
		if m.started {
			m.Close()
		}
	}
}

func (r *requirementSelectingModule) IsLegacyDAG() bool {
	return r.main.IsLegacyDAG()
}

func (r *requirementSelectingModule) Execute(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	triggerID, hasTrigger := extractTriggerID(request)

	if hasTrigger {
		if idx, ok := r.cache.Load(triggerID); ok {
			return r.additional[idx.(int)].Execute(ctx, request, handler)
		}
	}

	result, err := r.main.Execute(ctx, request, handler)
	if err == nil {
		return result, nil
	}

	rerun := &RequirementsRerun{}
	if !errors.As(err, &rerun) {
		return nil, err
	}

	for i, m := range r.additional {
		if CheckRequirements(m.RequirementsHandler, (*sdk.Requirements)(rerun)) {
			m.ensureStarted()
			if hasTrigger {
				r.cache.Store(triggerID, i)
			}
			return m.Execute(ctx, request, handler)
		}
	}

	return nil, fmt.Errorf("cannot find a runner that can satisfy the requirements %+v", rerun)
}

func extractTriggerID(req *sdk.ExecuteRequest) (uint64, bool) {
	if t := req.GetTrigger(); t != nil {
		return t.Id, true
	}
	return 0, false
}

var _ Module = &requirementSelectingModule{}
