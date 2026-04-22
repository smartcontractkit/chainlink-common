package host

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type ModuleAndHandler struct {
	ModuleV2
	RequirementsHandler
}

func NewRequirementSelectingModule(moduleAndHandlers []ModuleAndHandler) ModuleV2 {
	return &requirementSelectingModule{
		moduleAndHandler: moduleAndHandlers,
		runOn:            -1,
	}
}

type requirementSelectingModule struct {
	moduleAndHandler []ModuleAndHandler
	runOn            int
	started          atomic.Bool
	findMutex        sync.Mutex
}

func (r *requirementSelectingModule) Start() {
	r.started.Store(true)
	r.moduleAndHandler[0].Start()
}

func (r *requirementSelectingModule) Close() {
	r.findMutex.Lock()
	defer r.findMutex.Unlock()
	if r.runOn == -1 {
		r.moduleAndHandler[0].Close()
	} else {
		r.moduleAndHandler[r.runOn].Close()
	}
}

func (r *requirementSelectingModule) IsLegacyDAG() bool {
	return r.moduleAndHandler[0].IsLegacyDAG()
}

func (r *requirementSelectingModule) Execute(ctx context.Context, request *sdk.ExecuteRequest, handler ExecutionHelper) (*sdk.ExecutionResult, error) {
	if r.runOn >= 0 {
		return r.moduleAndHandler[r.runOn].Execute(ctx, request, handler)
	}

	r.findMutex.Lock()
	defer r.findMutex.Unlock()
	result, err := r.moduleAndHandler[0].Execute(ctx, request, handler)
	if err == nil {
		r.runOn = 0
		return result, nil
	}

	rerun := &RequirementsRerun{}
	if !errors.As(err, &rerun) {
		return nil, err
	}

	numHandlers := len(r.moduleAndHandler)
	for i := 1; i < numHandlers; i++ {
		item := r.moduleAndHandler[i]
		if CheckRequirements(item.RequirementsHandler, (*sdk.Requirements)(rerun)) {
			r.runOn = i
			if r.started.Load() {
				item.Start()
			}
			return item.Execute(ctx, request, handler)
		}
	}

	return nil, fmt.Errorf("cannot find a runner that can satisfy the requirements %+v\n", rerun)
}

var _ ModuleV2 = &requirementSelectingModule{}
