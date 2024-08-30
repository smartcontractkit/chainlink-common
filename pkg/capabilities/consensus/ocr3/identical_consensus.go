package ocr3

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

// Note this isn't generated because generics isn't supported in json schema

type IdenticalConsensusConfig[T any] struct {
	Encoder       Encoder
	EncoderConfig EncoderConfig
}

func (c IdenticalConsensusConfig[T]) New(w *workflows.WorkflowSpecFactory, ref string, input IdenticalConsensusInput[T]) SignedReportCap {
	def := workflows.StepDefinition{
		ID:     "offchain_reporting@1.0.0",
		Ref:    ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"encoder": c.Encoder, "encoder_config": c.EncoderConfig, "aggregation_method": "identical",
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := workflows.Step[SignedReport]{Definition: def}
	return SignedReportCapFromStep(w, step)
}

type IdenticalConsensusInput[T any] struct {
	Observations workflows.CapDefinition[T]
}

func (input IdenticalConsensusInput[T]) ToSteps() workflows.StepInputs {
	return workflows.StepInputs{
		Mapping: map[string]any{
			"observations": input.Observations.Ref(),
		},
	}
}
