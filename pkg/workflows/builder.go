package workflows

import "github.com/smartcontractkit/chainlink-common/pkg/capabilities"

// 1. Capability defines JSON schema for inputs and outputs of a capability.
// Trigger: triggerOutputType := workflowBuilder.addTrigger(DataStreamsTrigger.Config{})
// Adds metadata to the builder. Returns output type.
// 2. Consensus: consensusOutputType := workflowBuilder.addConsensus(ConsensusConfig{
// 	Inputs: triggerOutputType,
// })

type Workflow struct {
	spec *WorkflowSpec
}

type CapabilityDefinition[O any] interface {
	Ref() string
	impl() capabilityDefinitionImpl
}

type Step[O any] struct {
	Ref        string
	Definition StepDefinition
}

type capabilityDefinitionImpl struct {
	ref string
}

func (c *capabilityDefinitionImpl) Ref() string {
	return c.ref
}

func (c *capabilityDefinitionImpl) impl() capabilityDefinitionImpl {
	return *c
}

type NewWorkflowParams struct {
	Owner string
	Name  string
}

func NewWorkflow(
	params NewWorkflowParams,
) *Workflow {
	return &Workflow{
		spec: &WorkflowSpec{
			Owner: params.Owner,
			Name:  params.Name,
		},
	}
}

func AddStep[O any](w *Workflow, step Step[O]) (CapabilityDefinition[O], error) {
	// TODO should return error if the name is already used
	stepDefinition := step.Definition

	switch stepDefinition.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		w.spec.Triggers = append(w.spec.Triggers, stepDefinition)
	case capabilities.CapabilityTypeAction:
		w.spec.Actions = append(w.spec.Actions, stepDefinition)
	case capabilities.CapabilityTypeConsensus:
		w.spec.Consensus = append(w.spec.Consensus, stepDefinition)
	case capabilities.CapabilityTypeTarget:
		w.spec.Targets = append(w.spec.Targets, stepDefinition)
	}

	return &capabilityDefinitionImpl{ref: step.Ref}, nil
}

// AccessField is meant to be used by generated code
func AccessField[I, O any](c CapabilityDefinition[I], fieldName string) CapabilityDefinition[O] {
	return &capabilityDefinitionImpl{ref: c.Ref() + "." + fieldName}
}

func (w Workflow) Spec() WorkflowSpec {
	return *w.spec
}
