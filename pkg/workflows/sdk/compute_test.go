package sdk_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testdata/fixtures/capabilities/notstreams"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func TestCompute(t *testing.T) {
	anyNotStreamsInput := notstreams.Feed{
		ID:       "id",
		Metadata: notstreams.SignerMetadata{Signer: "signer1"},
		Payload: notstreams.FeedReport{
			BuyPrice:             []byte{1, 2, 3},
			FullReport:           []byte("report"),
			ObservationTimestamp: 2,
			ReportContext:        []byte("context"),
			SellPrice:            []byte{1, 2, 4},
			Signature:            []byte("sig"),
		},
		Timestamp:   "2022-01-05",
		TriggerType: "Type",
	}
	nsf, err := values.CreateMapFromStruct(map[string]any{"Arg0": anyNotStreamsInput})
	require.NoError(t, err)

	t.Run("creates correct workflow spec", func(t *testing.T) {
		workflow := createWorkflow(convertFeed)

		spec, err2 := workflow.Spec()
		require.NoError(t, err2)
		expectedSpec := sdk.WorkflowSpec{
			Name:  "name",
			Owner: "owner",
			Triggers: []sdk.StepDefinition{
				{
					ID:             "notstreams@1.0.0",
					Ref:            "trigger",
					Inputs:         sdk.StepInputs{},
					Config:         map[string]any{"maxFrequencyMs": 5000},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "__internal__custom_compute@1.0.0",
					Ref: "Compute",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"Arg0": "$(trigger.outputs)"},
					},
					Config:         map[string]any{},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{
				{
					ID:  "offchain_reporting@1.0.0",
					Ref: "data-feeds-report",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"observations": "$(Compute.outputs.Value)"},
					},
					Config: map[string]any{
						"aggregation_config": ocr3cap.DataFeedsConsensusConfigAggregationConfig{
							AllowedPartialStaleness: "false",
							Feeds: map[string]ocr3cap.FeedValue{
								anyFakeFeedID: {
									Deviation: "0.5",
									Heartbeat: 3600,
								},
							},
						},
						"aggregation_method": "data_feeds",
						"encoder":            "EVM",
						"encoder_config": ocr3cap.EncoderConfig{
							Abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
						},
						"report_id": "0001",
					},
					CapabilityType: capabilities.CapabilityTypeConsensus,
				},
			},
			Targets: []sdk.StepDefinition{
				{
					ID: "write_ethereum-testnet-sepolia@1.0.0",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"signed_report": "$(data-feeds-report.outputs)"},
					},
					Config: map[string]any{
						"address":    "0xE0082363396985ae2FdcC3a9F816A586Eed88416",
						"deltaStage": "45s",
						"schedule":   "oneAtATime",
					},
					CapabilityType: capabilities.CapabilityTypeTarget,
				},
			},
		}

		testutils.AssertWorkflowSpec(t, expectedSpec, spec)
	})

	t.Run("compute runs the function and returns the value", func(t *testing.T) {
		workflow := createWorkflow(convertFeed)

		fn := workflow.GetFn("Compute")
		require.NotNil(t, fn)

		req := capabilities.CapabilityRequest{Inputs: nsf}
		actual := fn(struct{}{}, req)
		require.NoError(t, actual.Err)

		expected, err := convertFeed(nil, anyNotStreamsInput)
		require.NoError(t, err)

		computed := &sdk.ComputeOutput[[]streams.Feed]{}
		err = actual.Value.UnwrapTo(computed)
		require.NoError(t, err)

		assert.Equal(t, expected, computed.Value)
	})

	t.Run("compute returns errors correctly", func(t *testing.T) {
		anyErr := errors.New("nope")
		workflow := createWorkflow(func(_ sdk.Runtime, inputFeed notstreams.Feed) ([]streams.Feed, error) {
			return nil, anyErr
		})

		fn := workflow.GetFn("Compute")
		require.NotNil(t, fn)

		req := capabilities.CapabilityRequest{Inputs: nsf}
		actual := fn(struct{}{}, req)
		require.Equal(t, anyErr, actual.Err)
	})
}

func createWorkflow(fn func(_ sdk.Runtime, inputFeed notstreams.Feed) ([]streams.Feed, error)) *sdk.WorkflowSpecFactory {
	workflow := sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{
		Owner: "owner",
		Name:  "name",
	})

	trigger := notstreams.TriggerConfig{MaxFrequencyMs: 5000}.New(workflow)
	computed := sdk.Compute1(workflow, "Compute", sdk.Compute1Inputs[notstreams.Feed]{Arg0: trigger}, fn)

	consensus := ocr3cap.DataFeedsConsensusConfig{
		AggregationConfig: ocr3cap.DataFeedsConsensusConfigAggregationConfig{
			AllowedPartialStaleness: "false",
			Feeds: map[string]ocr3cap.FeedValue{
				anyFakeFeedID: {
					Deviation: "0.5",
					Heartbeat: 3600,
				},
			},
		},
		AggregationMethod: "data_feeds",
		Encoder:           "EVM",
		EncoderConfig: ocr3cap.EncoderConfig{
			Abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
		},
		ReportId: "0001",
	}.New(workflow, "data-feeds-report", ocr3cap.DataFeedsConsensusInput{
		Observations: computed.Value(),
	})

	chainwriter.TargetConfig{
		Address:    "0xE0082363396985ae2FdcC3a9F816A586Eed88416",
		DeltaStage: "45s",
		Schedule:   "oneAtATime",
	}.New(workflow, "write_ethereum-testnet-sepolia@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

	return workflow
}

func convertFeed(_ sdk.Runtime, inputFeed notstreams.Feed) ([]streams.Feed, error) {
	return []streams.Feed{
		{
			ID:       inputFeed.ID,
			Metadata: streams.SignersMetadata{Signers: []string{inputFeed.Metadata.Signer}},
			Payload: []streams.FeedReport{
				{
					BenchmarkPrice:       inputFeed.Payload.BuyPrice,
					FeedID:               anyFakeFeedID,
					FullReport:           inputFeed.Payload.FullReport,
					ObservationTimestamp: inputFeed.Payload.ObservationTimestamp,
					ReportContext:        inputFeed.Payload.ReportContext,
					Signatures:           [][]byte{inputFeed.Payload.Signature},
				},
			},
			Timestamp:   inputFeed.Timestamp,
			TriggerType: inputFeed.TriggerType,
		},
	}, nil
}
