package ocr3cap

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

// Note this isn't generated because generics isn't supported in json schema

type IdenticalConsensusConfig[T any] struct {
	Encoder       Encoder
	EncoderConfig EncoderConfig
	ReportID      ReportId
}

func (c IdenticalConsensusConfig[T]) New(w *sdk.WorkflowSpecFactory, ref string, input IdenticalConsensusInput[T]) SignedReportCap {
	def := sdk.StepDefinition{
		ID:     "offchain_reporting@1.0.0",
		Ref:    ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"encoder":            c.Encoder,
			"encoder_config":     c.EncoderConfig,
			"aggregation_method": "identical",
			"report_id":          c.ReportID,
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := &sdk.Step[SignedReport]{Definition: def}
	return SignedReportWrapper(step.AddTo(w))
}

type IdenticalConsensusInput[T any] struct {
	Observation   sdk.CapDefinition[T]
	Encoder       Encoder
	EncoderConfig EncoderConfig
}

func (input IdenticalConsensusInput[T]) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"observations":  sdk.ListOf(input.Observation).Ref(),
			"encoder":       input.Encoder,
			"encoderConfig": input.EncoderConfig,
		},
	}
}
