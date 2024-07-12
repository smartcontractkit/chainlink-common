package workflows_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	chainwriter "github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chain_writer"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func NewWorkflowSpec() (workflows.WorkflowSpec, error) {
	workflow := workflows.NewWorkflow(workflows.NewWorkflowParams{
		Owner: "00000000000000000000000000000000000000aa",
		Name:  "ccipethsep",
	})

	streamsConfig := streams.StreamsTriggerConfig{
		FeedIds: []streams.FeedId{
			"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd",
			"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d",
			"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12",
		},
		MaxFrequencyMs: 100,
	}
	streamsTrigger, err := streams.NewStreamsTriggerCapability(workflow, "streams", streamsConfig)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	ocr3Config := ocr3.Ocr3ConsensusConfig{
		AggregationMethod: "data_feeds",
		AggregationConfig: []ocr3.Ocr3ConsensusConfigAggregationConfigElem{
			{
				Deviation: 0.05,
				FeedId:    "0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd",
				Heartbeat: 3600,
			},
			{
				FeedId:    "0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d",
				Deviation: 0.05,
				Heartbeat: 3600,
			},
			{
				FeedId:    "0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12",
				Deviation: 0.05,
				Heartbeat: 3600,
			},
		},
		ReportId: "0001",
		Encoder:  "EVM",
		EncoderConfig: ocr3.Ocr3ConsensusConfigEncoderConfig{
			Abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
		},
	}

	ocrInput := ocr3.Ocr3ConsensusCapabilityInput{Observations: streamsTrigger}

	consensus, err := ocr3.NewOcr3ConsensusCapability(workflow, "data-feeds-report", ocrInput, ocr3Config)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	writerConfig := chainwriter.ChainwriterTargetConfig{Address: "0x1234567890123456789012345678901234567890"}
	// TODO ideally some of these will have references so the type can be some interface
	input := chainwriter.ChainwriterTargetCapabilityInput{
		Err: consensus.Err(),
		// I'm assuming this contains the signature...
		Value:               consensus.Value(),
		WorkflowExecutionID: consensus.WorkflowExecutionID(),
	}

	if err = chainwriter.NewChainwriterTargetCapability(workflow, "chain-writer", input, writerConfig); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	return workflow.Spec(), nil
}

func TestBuilder_ValidSpec(t *testing.T) {
	testWorkflowSpec, err := NewWorkflowSpec()
	require.NoError(t, err)

	expectedSpec := workflows.WorkflowSpec{
		Name:  "ccipethsep",
		Owner: "00000000000000000000000000000000000000aa",
		Triggers: []workflows.StepDefinition{
			{
				ID:  "streams-trigger@1.0.0",
				Ref: "streams",
				Config: map[string]interface{}{
					"feedIds": []string{
						"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd",
						"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d",
						"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12",
					},
					"maxFrequencyMs": 100,
				},
				CapabilityType: capabilities.CapabilityTypeTrigger,
			},
		},
		Consensus: []workflows.StepDefinition{
			{
				ID:  "offchain_reporting@1.0.0",
				Ref: "data-feeds-report",
				Inputs: workflows.StepInputs{
					Mapping: map[string]any{
						"observations": "$(streams.outputs)",
					},
				},
				Config: map[string]interface{}{
					"report_id":          "0001",
					"aggregation_method": "data_feeds",
					"aggregation_config": map[string]interface{}{
						"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
						"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
						"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
					},
					"encoder": "EVM",
					"encoder_config": map[string]interface{}{
						"abi": "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
					},
				},
				CapabilityType: capabilities.CapabilityTypeConsensus,
			},
		},
		Targets: []workflows.StepDefinition{
			{
				ID:  "",
				Ref: "chain-writer",
				Inputs: workflows.StepInputs{
					Mapping: map[string]any{"Err": "$(offchain_reporting.outputs.Err)", "Value": "$(offchain_reporting.outputs.Value)", "WorkflowExecutionID": "$(offchain_reporting.outputs.WorkflowExecutionID)"},
				},
				Config:         map[string]any{"Address": "0x1234567890123456789012345678901234567890"},
				CapabilityType: capabilities.CapabilityTypeTarget,
			},
		},
	}

	assert.Equal(t, expectedSpec, testWorkflowSpec)
}
