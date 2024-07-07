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
	spec *workflowSpecYaml
}

type Trigger[O any] struct {
	Definition TriggerDefinitionYaml
	Output     O
}

type TriggerDefinition[O any] struct {
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
		spec: &workflowSpecYaml{
			Owner: params.Owner,
			Name:  params.Name,
		},
	}
}

func AddTrigger[O any](b *Workflow, trigger Trigger[O]) TriggerDefinition[O] {
	// Add ref to trigger.Definition
	trigger.Definition.Ref = fmt.Sprintf("trigger-%s", strconv.Itoa((len(b.spec.Triggers))))
	b.spec.Triggers = append(b.spec.Triggers, trigger.Definition)

	return TriggerDefinition[O]{
		Output: trigger.Output,
		Ref:    trigger.Definition.Ref,
	}
}

func (b Workflow) Spec() WorkflowSpec {
	return b.spec.toWorkflowSpec()
}
