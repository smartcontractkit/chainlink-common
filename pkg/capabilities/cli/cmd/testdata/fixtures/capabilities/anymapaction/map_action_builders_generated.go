// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package anymapaction

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func (cfg MapActionConfig) New(w *sdk.WorkflowSpecFactory, ref string, input MapActionInput) MapActionOutputsCap {

	def := sdk.StepDefinition{
		ID: "anymapaction@1.0.0", Ref: ref,
		Inputs:         input.ToSteps(),
		Config:         map[string]any{},
		CapabilityType: capabilities.CapabilityTypeAction,
	}

	step := sdk.Step[MapActionOutputs]{Definition: def}
	raw := step.AddTo(w)
	return MapActionOutputsWrapper(raw)
}

// MapActionOutputsWrapper allows access to field from an sdk.CapDefinition[MapActionOutputs]
func MapActionOutputsWrapper(raw sdk.CapDefinition[MapActionOutputs]) MapActionOutputsCap {
	wrapped, ok := raw.(MapActionOutputsCap)
	if ok {
		return wrapped
	}
	return &mapActionOutputsCap{CapDefinition: raw}
}

type MapActionOutputsCap interface {
	sdk.CapDefinition[MapActionOutputs]
	Payload() MapActionOutputsPayloadCap
	private()
}

type mapActionOutputsCap struct {
	sdk.CapDefinition[MapActionOutputs]
}

func (*mapActionOutputsCap) private() {}
func (c *mapActionOutputsCap) Payload() MapActionOutputsPayloadCap {
	return MapActionOutputsPayloadWrapper(sdk.AccessField[MapActionOutputs, MapActionOutputsPayload](c.CapDefinition, "payload"))
}

func ConstantMapActionOutputs(value MapActionOutputs) MapActionOutputsCap {
	return &mapActionOutputsCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewMapActionOutputsFromFields(
	payload MapActionOutputsPayloadCap) MapActionOutputsCap {
	return &simpleMapActionOutputs{
		CapDefinition: sdk.ComponentCapDefinition[MapActionOutputs]{
			"payload": payload.Ref(),
		},
		payload: payload,
	}
}

type simpleMapActionOutputs struct {
	sdk.CapDefinition[MapActionOutputs]
	payload MapActionOutputsPayloadCap
}

func (c *simpleMapActionOutputs) Payload() MapActionOutputsPayloadCap {
	return c.payload
}

func (c *simpleMapActionOutputs) private() {}

// MapActionOutputsPayloadWrapper allows access to field from an sdk.CapDefinition[MapActionOutputsPayload]
func MapActionOutputsPayloadWrapper(raw sdk.CapDefinition[MapActionOutputsPayload]) MapActionOutputsPayloadCap {
	wrapped, ok := raw.(MapActionOutputsPayloadCap)
	if ok {
		return wrapped
	}
	return MapActionOutputsPayloadCap(raw)
}

type MapActionOutputsPayloadCap sdk.CapDefinition[MapActionOutputsPayload]

type MapActionInput struct {
	Payload sdk.CapDefinition[MapActionInputsPayload]
}

func (input MapActionInput) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"payload": input.Payload.Ref(),
		},
	}
}
