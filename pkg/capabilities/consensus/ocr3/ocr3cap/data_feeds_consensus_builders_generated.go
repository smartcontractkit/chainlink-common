// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package ocr3cap

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
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
			"key_id":             cfg.KeyId,
			"report_id":          cfg.ReportId,
		},
		CapabilityType: capabilities.CapabilityTypeConsensus,
	}

	step := sdk.Step[SignedReport]{Definition: def}
	raw := step.AddTo(w)
	return SignedReportWrapper(raw)
}

// FeedValueWrapper allows access to field from an sdk.CapDefinition[FeedValue]
func FeedValueWrapper(raw sdk.CapDefinition[FeedValue]) FeedValueCap {
	wrapped, ok := raw.(FeedValueCap)
	if ok {
		return wrapped
	}
	return &feedValueCap{CapDefinition: raw}
}

type FeedValueCap interface {
	sdk.CapDefinition[FeedValue]
	Deviation() sdk.CapDefinition[string]
	Heartbeat() sdk.CapDefinition[uint64]
	RemappedID() sdk.CapDefinition[string]
	private()
}

type feedValueCap struct {
	sdk.CapDefinition[FeedValue]
}

func (*feedValueCap) private() {}
func (c *feedValueCap) Deviation() sdk.CapDefinition[string] {
	return sdk.AccessField[FeedValue, string](c.CapDefinition, "deviation")
}
func (c *feedValueCap) Heartbeat() sdk.CapDefinition[uint64] {
	return sdk.AccessField[FeedValue, uint64](c.CapDefinition, "heartbeat")
}
func (c *feedValueCap) RemappedID() sdk.CapDefinition[string] {
	return sdk.AccessField[FeedValue, string](c.CapDefinition, "remappedID")
}

func ConstantFeedValue(value FeedValue) FeedValueCap {
	return &feedValueCap{CapDefinition: sdk.ConstantDefinition(value)}
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
