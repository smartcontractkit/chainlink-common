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
	Definition() StepDefinition
	Output() O
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

func AddStep[O any](w *Workflow, step CapabilityDefinition[O]) CapabilityDefinition[O] {
	stepDefinition := step.Definition()

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

	return step
}

func AddTrigger[O any](w *Workflow, trigger CapabilityDefinition[O]) CapabilityDefinition[O] {
	return trigger
}

func AddConsensus[O any](w *Workflow, consensus CapabilityDefinition[O]) CapabilityDefinition[O] {
	w.spec.Consensus = append(w.spec.Consensus, consensus.Definition())
	return consensus
}

func (w Workflow) Spec() WorkflowSpec {
	return *w.spec
}
