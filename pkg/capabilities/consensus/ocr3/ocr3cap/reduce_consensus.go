package ocr3cap

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

// Note this isn't generated because generics isn't supported in json schema

type ReduceConsensusConfig[T any] struct {
	Encoder       Encoder
	EncoderConfig EncoderConfig
	ReportID      ReportId
}

func (c ReduceConsensusConfig[T]) New(w *sdk.WorkflowSpecFactory, ref string, input ReduceConsensusInput[T]) SignedReportCap {
	def := sdk.StepDefinition{
		ID:     "offchain_reporting@1.0.0",
		Ref:    ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"encoder":            c.Encoder,
			"encoder_config":     c.EncoderConfig,
			"aggregation_method": "reduce",
			"report_id":          c.ReportID,
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := sdk.Step[SignedReport]{Definition: def}
	return SignedReportCapFromStep(w, step)
}

type ReduceConsensusInput[T any] struct {
	Observation   sdk.CapDefinition[T]
	Encoder       Encoder
	EncoderConfig EncoderConfig
}

func (input ReduceConsensusInput[T]) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"observations":  sdk.ListOf(input.Observation).Ref(),
			"encoder":       input.Encoder,
			"encoderConfig": input.EncoderConfig,
		},
	}
}
