// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package referenceaction

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func (cfg SomeConfig) New(w *workflows.WorkflowSpecFactory, ref string, input ActionInput) SomeOutputsCap {

	def := workflows.StepDefinition{
		ID: "reference-test-action@1.0.0", Ref: ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"name":   cfg.Name,
			"number": cfg.Number,
		},
		CapabilityType: capabilities.CapabilityTypeAction,
	}

	step := workflows.Step[SomeOutputs]{Definition: def}
	return SomeOutputsCapFromStep(w, step)
}

type SomeOutputsCap interface {
	workflows.CapDefinition[SomeOutputs]
	AdaptedThing() workflows.CapDefinition[string]
	private()
}

// SomeOutputsCapFromStep should only be called from generated code to assure type safety
func SomeOutputsCapFromStep(w *workflows.WorkflowSpecFactory, step workflows.Step[SomeOutputs]) SomeOutputsCap {
	raw := step.AddTo(w)
	return &someOutputs{CapDefinition: raw}
}

type someOutputs struct {
	workflows.CapDefinition[SomeOutputs]
}

func (*someOutputs) private() {}
func (c *someOutputs) AdaptedThing() workflows.CapDefinition[string] {
	return workflows.AccessField[SomeOutputs, string](c.CapDefinition, "adapted_thing")
}

func NewSomeOutputsFromFields(
	adaptedThing workflows.CapDefinition[string]) SomeOutputsCap {
	return &simpleSomeOutputs{
		CapDefinition: workflows.ComponentCapDefinition[SomeOutputs]{
			"adapted_thing": adaptedThing.Ref(),
		},
		adaptedThing: adaptedThing,
	}
}

type simpleSomeOutputs struct {
	workflows.CapDefinition[SomeOutputs]
	adaptedThing workflows.CapDefinition[string]
}

func (c *simpleSomeOutputs) AdaptedThing() workflows.CapDefinition[string] {
	return c.adaptedThing
}

func (c *simpleSomeOutputs) private() {}

type ActionInput struct {
	InputThing workflows.CapDefinition[bool]
}

func (input ActionInput) ToSteps() workflows.StepInputs {
	return workflows.StepInputs{
		Mapping: map[string]any{
			"input_thing": input.InputThing.Ref(),
		},
	}
}
