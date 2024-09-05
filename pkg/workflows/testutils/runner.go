package testutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/exec"
)

func NewRunner(ctx context.Context) *Runner {
	return &Runner{
		ctx:          ctx,
		registry:     map[string]capabilities.CallbackCapability{},
		results:      runnerResults{},
		idToStep:     map[string]workflows.StepDefinition{},
		dependencies: map[string][]string{},
		sdk:          &SDK{},
	}
}

type ConsensusMock interface {
	capabilities.ConsensusCapability
	SingleToManyObservations(value values.Value) (*values.Map, error)
}

type Runner struct {
	ctx          context.Context
	trigger      capabilities.TriggerCapability
	registry     map[string]capabilities.CallbackCapability
	am           map[string]map[string]graph.Edge[string]
	results      runnerResults
	idToStep     map[string]workflows.StepDefinition
	dependencies map[string][]string
	sdk          workflows.SDK
	errors       []error
}

var _ workflows.Runner = &Runner{}

func (r *Runner) Run(factory *workflows.WorkflowSpecFactory) error {
	if len(r.errors) > 0 {
		return fmt.Errorf("error registering capaiblities: %w", errors.Join(r.errors...))
	}

	spec, err := factory.Spec()
	if err != nil {
		return err
	}

	if err = r.ensureGraph(spec); err != nil {
		return err
	}

	r.setupSteps(factory, spec)

	return r.walk(spec, workflows.KeywordTrigger)
}

func (r *Runner) ensureGraph(spec workflows.WorkflowSpec) error {
	g, err := workflows.BuildDependencyGraph(spec)
	if err != nil {
		return err
	}

	if len(g.Triggers) != 1 {
		return fmt.Errorf("expected exactly 1 trigger, got %d", len(g.Triggers))
	}

	edges, err := g.Edges()
	if err != nil {
		return err
	}

	for _, edge := range edges {
		r.dependencies[edge.Target] = append(r.dependencies[edge.Target], edge.Source)
	}

	r.am, err = g.AdjacencyMap()
	return err
}

func (r *Runner) setupSteps(factory *workflows.WorkflowSpecFactory, spec workflows.WorkflowSpec) {
	for _, step := range spec.Steps() {
		if step.Ref == "" {
			step.Ref = step.ID
		}

		r.idToStep[step.Ref] = step
		if run := factory.GetFn(step.Ref); run != nil {
			compute := &computeCapability{
				sdk:      r.sdk,
				callback: run,
			}
			info, err := compute.Info(r.ctx)
			if err != nil {
				r.errors = append(r.errors, err)
				continue
			}
			r.MockCapability(info.ID, &step.Ref, compute)
		}
	}
	r.idToStep[workflows.KeywordTrigger] = spec.Triggers[0]
}

// MockCapability registers a new capability mock with the runner
// if the step is not nil, the capability will be registered for that step
// If a step is explicitly mocked, that will take priority over a mock of the entire capability.
// This is best used with generated code to ensure correctness
// Note that mocks of custom compute will not be used in place of the user's code
func (r *Runner) MockCapability(name string, step *string, capability capabilities.CallbackCapability) {
	fullName := getFullName(name, step)
	if r.registry[fullName] != nil {
		forSuffix := ""
		if step != nil {
			forSuffix = fmt.Sprintf(" for step %s", *step)
		}
		r.errors = append(r.errors, fmt.Errorf("capability %s already exists in registry%s", name, forSuffix))
	}

	r.registry[fullName] = capability
}

func (r *Runner) MockTrigger(trigger capabilities.TriggerCapability) {
	r.trigger = trigger
}

func (r *Runner) GetRegisteredMock(name string, step string) capabilities.ActionCapability {
	fullName := getFullName(name, &step)
	if c, ok := r.registry[fullName]; ok {
		return c
	}

	return r.registry[name]
}

func (r *Runner) walk(spec workflows.WorkflowSpec, ref string) error {
	capability := r.idToStep[ref]
	mock := r.GetRegisteredMock(capability.ID, ref)
	if mock == nil {
		return fmt.Errorf("no mock found for capability %s on step %s", capability.ID, ref)
	}

	request, err := r.buildRequest(spec, capability)
	if err != nil {
		return err
	}

	if c, ok := mock.(ConsensusMock); ok {
		if request.Inputs, err = c.SingleToManyObservations(request.Inputs); err != nil {
			return err
		}
	}

	resultCh, err := mock.Execute(r.ctx, request)
	if err != nil {
		return err
	}

	results, ok := <-resultCh
	if !ok {
		return nil
	}

	r.results[ref] = &exec.Result{
		Inputs:  request.Inputs,
		Outputs: results.Value,
		Error:   results.Err,
	}

	if results.Err != nil {
		return results.Err
	}

	edges, ok := r.am[ref]
	if !ok {
		return nil
	}

	return r.walkNext(spec, edges)
}

func (r *Runner) buildRequest(spec workflows.WorkflowSpec, capability workflows.StepDefinition) (capabilities.CapabilityRequest, error) {
	conf, err := values.NewMap(capability.Config)
	if err != nil {
		return capabilities.CapabilityRequest{}, err
	}

	inputs, err := r.buildInput(capability)
	if err != nil {
		return capabilities.CapabilityRequest{}, err
	}

	request := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowOwner: spec.Owner,
			WorkflowName:  spec.Name,
			ReferenceID:   capability.Ref,
		},
		Config: conf,
		Inputs: inputs,
	}
	return request, nil
}

func (r *Runner) walkNext(spec workflows.WorkflowSpec, edges map[string]graph.Edge[string]) error {
	var errs []error
	for edgeRef := range edges {
		if r.isReady(edgeRef) {
			if err := r.walk(spec, edgeRef); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (r *Runner) buildInput(capability workflows.StepDefinition) (*values.Map, error) {
	var input any
	if capability.Inputs.OutputRef != "" {
		input = capability.Inputs.OutputRef
	} else {
		input = capability.Inputs.Mapping
	}

	val, err := exec.FindAndInterpolateAllKeys(input, r.results)
	if err != nil {
		return nil, err
	}
	return values.NewMap(val.(map[string]any))
}

func (r *Runner) isReady(ref string) bool {
	for _, dep := range r.dependencies[ref] {
		if _, ok := r.results[dep]; !ok {
			return false
		}
	}

	return true
}

func getFullName(name string, step *string) string {
	fullName := name
	if step != nil {
		fullName += fmt.Sprintf("@@@%s", *step)
	}
	return fullName
}

type runnerResults map[string]*exec.Result

func (f runnerResults) ResultForStep(s string) (*exec.Result, bool) {
	r, ok := f[s]
	return r, ok
}

var _ exec.Results = runnerResults{}
