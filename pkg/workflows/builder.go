package workflows

import (
	"fmt"
	"strconv"
)

// 1. Capability defines JSON schema for inputs and outputs of a capability.
// Trigger: triggerOutputType := workflowBuilder.addTrigger(DataStreamsTrigger.Config{})
// Adds metadata to the builder. Returns output type.
// 2. Consensus: consensusOutputType := workflowBuilder.addConsensus(ConsensusConfig{
// 	Inputs: triggerOutputType,
// })

type Workflow struct {
	spec *WorkflowSpec
}

type Trigger[O any] struct {
	Definition StepDefinition
	Output     O
}

type Consensus[O any] struct {
	Definition StepDefinition
	Output     O
}

type CapabilityDefinition[O any] struct {
	Ref    string
	Output O
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

func AddTrigger[O any](w *Workflow, trigger Trigger[O]) CapabilityDefinition[O] {
	// Add ref to trigger.Definition
	trigger.Definition.Ref = fmt.Sprintf("trigger-%s", strconv.Itoa((len(w.spec.Triggers))))
	w.spec.Triggers = append(w.spec.Triggers, trigger.Definition)

	return CapabilityDefinition[O]{
		Output: trigger.Output,
		Ref:    trigger.Definition.Ref,
	}
}

func AddConsensus[O any](w *Workflow, consensus Consensus[O]) CapabilityDefinition[O] {
	// Add ref to trigger.Definition
	consensus.Definition.Ref = fmt.Sprintf("consensus-%s", strconv.Itoa((len(w.spec.Consensus))))
	w.spec.Consensus = append(w.spec.Consensus, consensus.Definition)

	return CapabilityDefinition[O]{
		Output: consensus.Output,
		Ref:    consensus.Definition.Ref,
	}
}

func (w Workflow) Spec() WorkflowSpec {
	return *w.spec
}
