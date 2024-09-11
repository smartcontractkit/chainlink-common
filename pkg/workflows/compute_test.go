package workflows_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testdata/fixtures/capabilities/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testutils"
)

func TestCompute(t *testing.T) {
	anyNotStreamsInput := notstreams.Feed{
		ID:       "id",
		Metadata: notstreams.SignerMetadata{Signer: "signer1"},
		Payload: notstreams.FeedReport{
			BuyPrice:             "123",
			FullReport:           "report",
			ObservationTimestamp: 2,
			ReportContext:        "context",
			SellPrice:            "124",
			Signature:            "sig",
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
		expectedSpec := workflows.WorkflowSpec{
			Name:  "name",
			Owner: "owner",
			Triggers: []workflows.StepDefinition{
				{
					ID:             "notstreams@1.0.0",
					Ref:            "trigger",
					Inputs:         workflows.StepInputs{},
					Config:         map[string]any{"maxFrequencyMs": 5000},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []workflows.StepDefinition{
				{
					ID:  "__internal__custom_compute@1.0.0",
					Ref: "Compute",
					Inputs: workflows.StepInputs{
						Mapping: map[string]any{"Arg0": "$(trigger.outputs)"},
					},
					Config:         map[string]any{},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []workflows.StepDefinition{
				{
					ID:  "offchain_reporting@1.0.0",
					Ref: "data-feeds-report",
					Inputs: workflows.StepInputs{
						Mapping: map[string]any{"observations": "$(Compute.outputs.Value)"},
					},
					Config: map[string]any{
						"aggregation_config": ocr3.DataFeedsConsensusConfigAggregationConfig{
							AllowedPartialStaleness: "false",
							Feeds: map[string]ocr3.FeedValue{
								anyFakeFeedID: {
									Deviation: "0.5",
									Heartbeat: 3600,
								},
							},
						},
						"aggregation_method": "data_feeds",
						"encoder":            ocr3.EncoderEVM,
						"encoder_config":     ocr3.EncoderConfig{},
						"report_id":          "0001",
					},
					CapabilityType: capabilities.CapabilityTypeConsensus,
				},
			},
			Targets: []workflows.StepDefinition{
				{
					ID: "write_ethereum-testnet-sepolia@1.0.0",
					Inputs: workflows.StepInputs{
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
		actual, err := fn(struct{}{}, req)
		require.NoError(t, err)

		expected, err := convertFeed(nil, anyNotStreamsInput)
		require.NoError(t, err)

		computed := &workflows.ComputeOutput[[]streams.Feed]{}
		err = actual.Value.UnwrapTo(computed)
		require.NoError(t, err)

		assert.Equal(t, expected, computed.Value)
	})

	t.Run("compute returns errors correctly", func(t *testing.T) {
		anyErr := errors.New("nope")
		workflow := createWorkflow(func(_ workflows.SDK, inputFeed notstreams.Feed) ([]streams.Feed, error) {
			return nil, anyErr
		})

		fn := workflow.GetFn("Compute")
		require.NotNil(t, fn)

		req := capabilities.CapabilityRequest{Inputs: nsf}
		_, err := fn(struct{}{}, req)
		require.Equal(t, anyErr, err)
	})
}

func createWorkflow(fn func(_ workflows.SDK, inputFeed notstreams.Feed) ([]streams.Feed, error)) *workflows.WorkflowSpecFactory {
	workflow := workflows.NewWorkflowSpecFactory(workflows.NewWorkflowParams{
		Owner: "owner",
		Name:  "name",
	})

	trigger := notstreams.TriggerConfig{MaxFrequencyMs: 5000}.New(workflow)
	computed := workflows.Compute1(workflow, "Compute", workflows.Compute1Inputs[notstreams.Feed]{Arg0: trigger}, fn)

	consensus := ocr3.DataFeedsConsensusConfig{
		AggregationConfig: ocr3.DataFeedsConsensusConfigAggregationConfig{
			AllowedPartialStaleness: "false",
			Feeds: map[string]ocr3.FeedValue{
				anyFakeFeedID: {
					Deviation: "0.5",
					Heartbeat: 3600,
				},
			},
		},
		AggregationMethod: "data_feeds",
		Encoder:           ocr3.EncoderEVM,
		EncoderConfig:     ocr3.EncoderConfig{},
		ReportId:          "0001",
	}.New(workflow, "data-feeds-report", ocr3.DataFeedsConsensusInput{
		Observations: computed.Value(),
	})

	chainwriter.TargetConfig{
		Address:    "0xE0082363396985ae2FdcC3a9F816A586Eed88416",
		DeltaStage: "45s",
		Schedule:   "oneAtATime",
	}.New(workflow, "write_ethereum-testnet-sepolia@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

	return workflow
}

func convertFeed(_ workflows.SDK, inputFeed notstreams.Feed) ([]streams.Feed, error) {
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
					Signatures:           []string{inputFeed.Payload.Signature},
				},
			},
			Timestamp:   inputFeed.Timestamp,
			TriggerType: inputFeed.TriggerType,
		},
	}, nil
}
