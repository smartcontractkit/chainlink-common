package testutils

import (
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func NewRunner() *Runner {
	return &Runner{registry: map[string]CapabilityMock{}}
}

type ConsensusMock interface {
	CapabilityMock
	MultiplexObservations(value values.Value) (*values.List, error)
}

type Runner struct {
	registry map[string]CapabilityMock
	errors   []error
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

	// TODO https://smartcontract-it.atlassian.net/browse/KS-442, implement this function
	_ = spec
	return nil
}

// MockCapability registers a new capability mock with the runner
// if the step is not nil, the capability will be registered for that step
// If a step is explicitly mocked, that will take priority over a mock of the entire capability.
// This is best used with generated code to ensure correctness
func (r *Runner) MockCapability(name string, step *string, capability CapabilityMock) {
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

func (r *Runner) GetRegisteredMock(name string, step string) CapabilityMock {
	fullName := getFullName(name, &step)
	if c, ok := r.registry[fullName]; ok {
		return c
	}

	return r.registry[name]
}

func getFullName(name string, step *string) string {
	fullName := name
	if step != nil {
		fullName += fmt.Sprintf("@@@%s", *step)
	}
	return fullName
}
