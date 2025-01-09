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
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func NewRunner(ctx context.Context) *Runner {
	return &Runner{
		ctx:          ctx,
		registry:     map[string]capabilities.ExecutableCapability{},
		results:      runnerResults{},
		idToStep:     map[string]sdk.StepDefinition{},
		dependencies: map[string][]string{},
		runtime:      &NoopRuntime{},
	}
}

type Runner struct {
	RawConfig []byte
	Secrets   map[string]string
	// Context is held in this runner because it's for testing and capability calls are made by it.
	// The real SDK implementation will be for the WASM guest and will make host calls, and callbacks to the program.
	// nolint
	ctx          context.Context
	trigger      capabilities.TriggerCapability
	registry     map[string]capabilities.ExecutableCapability
	am           map[string]map[string]graph.Edge[string]
	results      runnerResults
	idToStep     map[string]sdk.StepDefinition
	dependencies map[string][]string
	runtime      sdk.Runtime
	errors       []error
}

var _ sdk.Runner = &Runner{}

func (r *Runner) Config() []byte {
	return r.RawConfig
}

func (r *Runner) Run(factory *sdk.WorkflowSpecFactory) {
	spec, err := factory.Spec()
	if err != nil {
		r.errors = append(r.errors, err)
		return
	}

	if err = r.ensureGraph(spec); err != nil {
		r.errors = append(r.errors, err)
		return
	}

	r.setupSteps(factory, spec)

	err = r.walk(spec, workflows.KeywordTrigger)
	if err != nil {
		r.errors = append(r.errors, err)
	}

	r.unregisterWorkflow(spec)
}

func (r *Runner) Err() error {
	return errors.Join(r.errors...)
}

func (r *Runner) ensureGraph(spec sdk.WorkflowSpec) error {
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

func (r *Runner) setupSteps(factory *sdk.WorkflowSpecFactory, spec sdk.WorkflowSpec) {
	for _, step := range spec.Steps() {
		if step.Ref == "" {
			step.Ref = step.ID
		}

		r.idToStep[step.Ref] = step

		// if the factory has a method, it's custom compute, we'll run that compute.
		if run := factory.GetFn(step.Ref); run != nil {
			compute := &computeCapability{
				sdk:      r.runtime,
				callback: run,
			}
			info, err := compute.Info(r.ctx)
			if err != nil {
				r.errors = append(r.errors, err)
				continue
			}
			r.MockCapability(info.ID, &step.Ref, compute)
		}

		r.registerStep(step)
	}

	r.registerStep(spec.Triggers[0])
	r.idToStep[workflows.KeywordTrigger] = spec.Triggers[0]
}

// MockCapability registers a new capability mock with the runner
// if the step is not nil, the capability will be registered for that step
// If a step is explicitly mocked, that will take priority over a mock of the entire capability.
// This is best used with generated code to ensure correctness
// Note that mocks of custom compute will not be used in place of the user's code
func (r *Runner) MockCapability(name string, step *string, capability capabilities.ExecutableCapability) {
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

func (r *Runner) walk(spec sdk.WorkflowSpec, ref string) error {
	capability := r.idToStep[ref]
	mock := r.GetRegisteredMock(capability.ID, ref)
	if mock == nil {
		return fmt.Errorf("no mock found for capability %s on step %s", capability.ID, ref)
	}

	request, err := r.buildRequest(spec, capability)
	if err != nil {
		return err
	}

	results, err := mock.Execute(r.ctx, request)
	if err != nil {
		return err
	}

	r.results[ref] = &exec.Result{
		Inputs:  request.Inputs,
		Outputs: results.Value,
	}

	edges, ok := r.am[ref]
	if !ok {
		return nil
	}

	return r.walkNext(spec, edges)
}

func (r *Runner) buildRequest(spec sdk.WorkflowSpec, capability sdk.StepDefinition) (capabilities.CapabilityRequest, error) {
	env := exec.Env{
		Config:  r.RawConfig,
		Binary:  []byte{},
		Secrets: r.Secrets,
	}
	config, err := exec.FindAndInterpolateEnvVars(capability.Config, env)
	if err != nil {
		return capabilities.CapabilityRequest{}, err
	}

	conf, err := values.NewMap(config.(map[string]any))
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

func (r *Runner) walkNext(spec sdk.WorkflowSpec, edges map[string]graph.Edge[string]) error {
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

func (r *Runner) buildInput(capability sdk.StepDefinition) (*values.Map, error) {
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

func (r *Runner) registerStep(step sdk.StepDefinition) {
	mock := r.GetRegisteredMock(step.ID, step.Ref)
	if mock == nil {
		// It's possible the workflow intentionally doesn't reach this step, so it's ok to not register a mock.
		// If it does reach the step, the user will get an error then.
		return
	}

	config, err := values.NewMap(step.Config)
	if err != nil {
		r.errors = append(r.errors, err)
		return
	}

	if err = mock.RegisterToWorkflow(r.ctx, capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:    "test",
			WorkflowOwner: "test",
		},
		Config: config,
	}); err != nil {
		r.errors = append(r.errors, err)
	}
}

func (r *Runner) unregisterStep(step sdk.StepDefinition) {
	mock := r.GetRegisteredMock(step.ID, step.Ref)
	if mock == nil {
		return
	}

	config, err := values.NewMap(step.Config)
	if err != nil {
		r.errors = append(r.errors, err)
		return
	}

	if err := mock.UnregisterFromWorkflow(r.ctx, capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:    "test",
			WorkflowOwner: "test",
		},
		Config: config,
	}); err != nil {
		r.errors = append(r.errors, err)
	}
}

func (r *Runner) unregisterWorkflow(spec sdk.WorkflowSpec) {
	for _, step := range spec.Steps() {
		r.unregisterStep(step)
	}

	r.unregisterStep(spec.Triggers[0])
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
