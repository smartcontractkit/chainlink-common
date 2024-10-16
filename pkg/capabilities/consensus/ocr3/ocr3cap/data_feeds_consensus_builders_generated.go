// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package ocr3cap

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	streams "github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func (cfg DataFeedsConsensusConfig) New(w *sdk.WorkflowSpecFactory, ref string, input DataFeedsConsensusInput) SignedReportCap {

	def := sdk.StepDefinition{
		ID: "offchain_reporting@1.0.0", Ref: ref,
		Inputs: input.ToSteps(),
		Config: map[string]any{
			"aggregation_config": cfg.AggregationConfig,
			"aggregation_method": cfg.AggregationMethod,
			"encoder":            cfg.Encoder,
			"encoder_config":     cfg.EncoderConfig,
			"report_id":          cfg.ReportId,
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := sdk.Step[SignedReport]{Definition: def}
	return SignedReportCapFromStep(w, step)
}

type FeedValueCap interface {
	sdk.CapDefinition[FeedValue]
	Deviation() sdk.CapDefinition[string]
	Heartbeat() sdk.CapDefinition[uint64]
	RemappedID() sdk.CapDefinition[string]
	private()
}

// FeedValueCapFromStep should only be called from generated code to assure type safety
func FeedValueCapFromStep(w *sdk.WorkflowSpecFactory, step sdk.Step[FeedValue]) FeedValueCap {
	raw := step.AddTo(w)
	return &feedValue{CapDefinition: raw}
}

type feedValue struct {
	sdk.CapDefinition[FeedValue]
}

func (*feedValue) private() {}
func (c *feedValue) Deviation() sdk.CapDefinition[string] {
	return sdk.AccessField[FeedValue, string](c.CapDefinition, "deviation")
}
func (c *feedValue) Heartbeat() sdk.CapDefinition[uint64] {
	return sdk.AccessField[FeedValue, uint64](c.CapDefinition, "heartbeat")
}
func (c *feedValue) RemappedID() sdk.CapDefinition[string] {
	return sdk.AccessField[FeedValue, string](c.CapDefinition, "remappedID")
}

func NewFeedValueFromFields(
	deviation sdk.CapDefinition[string],
	heartbeat sdk.CapDefinition[uint64],
	remappedID sdk.CapDefinition[string]) FeedValueCap {
	return &simpleFeedValue{
		CapDefinition: sdk.ComponentCapDefinition[FeedValue]{
			"deviation":  deviation.Ref(),
			"heartbeat":  heartbeat.Ref(),
			"remappedID": remappedID.Ref(),
		},
		deviation:  deviation,
		heartbeat:  heartbeat,
		remappedID: remappedID,
	}
}

type simpleFeedValue struct {
	sdk.CapDefinition[FeedValue]
	deviation  sdk.CapDefinition[string]
	heartbeat  sdk.CapDefinition[uint64]
	remappedID sdk.CapDefinition[string]
}

func (c *simpleFeedValue) Deviation() sdk.CapDefinition[string] {
	return c.deviation
}
func (c *simpleFeedValue) Heartbeat() sdk.CapDefinition[uint64] {
	return c.heartbeat
}
func (c *simpleFeedValue) RemappedID() sdk.CapDefinition[string] {
	return c.remappedID
}

func (c *simpleFeedValue) private() {}

type DataFeedsConsensusInput struct {
	Observations sdk.CapDefinition[[]streams.Feed]
}

func (input DataFeedsConsensusInput) ToSteps() sdk.StepInputs {
	return sdk.StepInputs{
		Mapping: map[string]any{
			"observations": input.Observations.Ref(),
		},
	}
}