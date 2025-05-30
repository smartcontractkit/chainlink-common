// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package basicconsensus

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func (cfg ConsensusConfig) New(w *sdk.WorkflowSpecFactory, ref string, input ConsensusInput) ConsensusOutputsCap {

	def := sdk.StepDefinition{
		ID: "basic-test-consensus@1.0.0", Ref: ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"name":   cfg.Name,
			"number": cfg.Number,
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := sdk.Step[ConsensusOutputs]{Definition: def}
	raw := step.AddTo(w)
	return ConsensusOutputsWrapper(raw)
}

// ConsensusOutputsWrapper allows access to field from an sdk.CapDefinition[ConsensusOutputs]
func ConsensusOutputsWrapper(raw sdk.CapDefinition[ConsensusOutputs]) ConsensusOutputsCap {
	wrapped, ok := raw.(ConsensusOutputsCap)
	if ok {
		return wrapped
	}
	return &consensusOutputsCap{CapDefinition: raw}
}

type ConsensusOutputsCap interface {
	sdk.CapDefinition[ConsensusOutputs]
	Consensus() sdk.CapDefinition[[]string]
	Sigs() sdk.CapDefinition[[]string]
	private()
}

type consensusOutputsCap struct {
	sdk.CapDefinition[ConsensusOutputs]
}

func (*consensusOutputsCap) private() {}
func (c *consensusOutputsCap) Consensus() sdk.CapDefinition[[]string] {
	return sdk.AccessField[ConsensusOutputs, []string](c.CapDefinition, "consensus")
}
func (c *consensusOutputsCap) Sigs() sdk.CapDefinition[[]string] {
	return sdk.AccessField[ConsensusOutputs, []string](c.CapDefinition, "sigs")
}

func ConstantConsensusOutputs(value ConsensusOutputs) ConsensusOutputsCap {
	return &consensusOutputsCap{CapDefinition: sdk.ConstantDefinition(value)}
}

func NewConsensusOutputsFromFields(
	consensus sdk.CapDefinition[[]string],
	sigs sdk.CapDefinition[[]string]) ConsensusOutputsCap {
	return &simpleConsensusOutputs{
		CapDefinition: sdk.ComponentCapDefinition[ConsensusOutputs]{
			"consensus": consensus.Ref(),
			"sigs":      sigs.Ref(),
		},
		consensus: consensus,
		sigs:      sigs,
	}
}

type simpleConsensusOutputs struct {
	sdk.CapDefinition[ConsensusOutputs]
	consensus sdk.CapDefinition[[]string]
	sigs      sdk.CapDefinition[[]string]
}

func (c *simpleConsensusOutputs) Consensus() sdk.CapDefinition[[]string] {
	return c.consensus
}
func (c *simpleConsensusOutputs) Sigs() sdk.CapDefinition[[]string] {
	return c.sigs
}

func (c *simpleConsensusOutputs) private() {}

type ConsensusInput struct {
	InputThing sdk.CapDefinition[bool]
}

func (input ConsensusInput) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"input_thing": input.InputThing.Ref(),
		},
	}
}
