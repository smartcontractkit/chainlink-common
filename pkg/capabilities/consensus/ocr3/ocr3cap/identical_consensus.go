package ocr3cap

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

// Note this isn't generated because generics isn't supported in json schema

type IdenticalConsensusConfig[T any] struct {
	Encoder       Encoder
	EncoderConfig EncoderConfig
}

func (c IdenticalConsensusConfig[T]) New(w *sdk.WorkflowSpecFactory, ref string, input IdenticalConsensusInput[T]) SignedReportCap {
	def := sdk.StepDefinition{
		ID:             "offchain_reporting@1.0.0",
		Ref:            ref,
		Inputs:         input.ToSteps(),
		Config:         map[string]any{"encoder": c.Encoder, "encoder_config": c.EncoderConfig},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := sdk.Step[SignedReport]{Definition: def}
	return SignedReportCapFromStep(w, step)
}

type IdenticalConsensusInput[T any] struct {
	Observations sdk.CapDefinition[T]
}

type IdenticalConsensusMergedInput[T any] struct {
	Observations []T
}

func (input IdenticalConsensusInput[T]) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"observations": input.Observations.Ref(),
		},
	}
}
