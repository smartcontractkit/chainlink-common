// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package arrayaction

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
)

func (cfg ActionConfig) New(w *sdk.WorkflowSpecFactory, ref string, input ActionInput) sdk.CapDefinition[[]ActionOutputsElem] {

	def := sdk.StepDefinition{
		ID: "array-test-action@1.0.0", Ref: ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"details": cfg.Details,
		},
		CapabilityType: capabilities.CapabilityTypeAction,
	}

	step := sdk.Step[[]ActionOutputsElem]{Definition: def}
	return step.AddTo(w)
}

// ActionOutputsElemWrapper allows access to field from an sdk.CapDefinition[ActionOutputsElem]
func ActionOutputsElemWrapper(raw sdk.CapDefinition[ActionOutputsElem]) ActionOutputsElemCap {
	wrapped, ok := raw.(ActionOutputsElemCap)
	if ok {
		return wrapped
	}
	return &actionOutputsElemCap{CapDefinition: raw}
}

type ActionOutputsElemCap interface {
	sdk.CapDefinition[ActionOutputsElem]
	Results() ActionOutputsElemResultsCap
	private()
}

type actionOutputsElemCap struct {
	sdk.CapDefinition[ActionOutputsElem]
}

func (*actionOutputsElemCap) private() {}
func (c *actionOutputsElemCap) Results() ActionOutputsElemResultsCap {
	return ActionOutputsElemResultsWrapper(sdk.AccessField[ActionOutputsElem, ActionOutputsElemResults](c.CapDefinition, "results"))
}

func ConstantActionOutputsElem(value ActionOutputsElem) ActionOutputsElemCap {
	return &actionOutputsElemCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewActionOutputsElemFromFields(
	results ActionOutputsElemResultsCap) ActionOutputsElemCap {
	return &simpleActionOutputsElem{
		CapDefinition: sdk.ComponentCapDefinition[ActionOutputsElem]{
			"results": results.Ref(),
		},
		results: results,
	}
}

type simpleActionOutputsElem struct {
	sdk.CapDefinition[ActionOutputsElem]
	results ActionOutputsElemResultsCap
}

func (c *simpleActionOutputsElem) Results() ActionOutputsElemResultsCap {
	return c.results
}

func (c *simpleActionOutputsElem) private() {}

// ActionOutputsElemResultsWrapper allows access to field from an sdk.CapDefinition[ActionOutputsElemResults]
func ActionOutputsElemResultsWrapper(raw sdk.CapDefinition[ActionOutputsElemResults]) ActionOutputsElemResultsCap {
	wrapped, ok := raw.(ActionOutputsElemResultsCap)
	if ok {
		return wrapped
	}
	return &actionOutputsElemResultsCap{CapDefinition: raw}
}

type ActionOutputsElemResultsCap interface {
	sdk.CapDefinition[ActionOutputsElemResults]
	AdaptedThing() sdk.CapDefinition[string]
	private()
}

type actionOutputsElemResultsCap struct {
	sdk.CapDefinition[ActionOutputsElemResults]
}

func (*actionOutputsElemResultsCap) private() {}
func (c *actionOutputsElemResultsCap) AdaptedThing() sdk.CapDefinition[string] {
	return sdk.AccessField[ActionOutputsElemResults, string](c.CapDefinition, "adapted_thing")
}

func ConstantActionOutputsElemResults(value ActionOutputsElemResults) ActionOutputsElemResultsCap {
	return &actionOutputsElemResultsCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewActionOutputsElemResultsFromFields(
	adaptedThing sdk.CapDefinition[string]) ActionOutputsElemResultsCap {
	return &simpleActionOutputsElemResults{
		CapDefinition: sdk.ComponentCapDefinition[ActionOutputsElemResults]{
			"adapted_thing": adaptedThing.Ref(),
		},
		adaptedThing: adaptedThing,
	}
}

type simpleActionOutputsElemResults struct {
	sdk.CapDefinition[ActionOutputsElemResults]
	adaptedThing sdk.CapDefinition[string]
}

func (c *simpleActionOutputsElemResults) AdaptedThing() sdk.CapDefinition[string] {
	return c.adaptedThing
}

func (c *simpleActionOutputsElemResults) private() {}

type ActionInput struct {
	Metadata sdk.CapDefinition[ActionInputsMetadata]
}

func (input ActionInput) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"metadata": input.Metadata.Ref(),
		},
	}
}
