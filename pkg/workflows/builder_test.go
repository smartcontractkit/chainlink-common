package workflows_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

func NewWorkflowSpec() (workflows.WorkflowSpec, error) {
	workflow := workflows.NewWorkflow(workflows.NewWorkflowParams{
		Owner: "00000000000000000000000000000000000000aa",
		Name:  "ccipethsep",
	})

	mercuryTriggerOutput := workflows.AddTrigger(workflow, triggers.NewMercuryTrigger(
		triggers.NewMercuryTriggerParams{
			Config: triggers.Config{
				FeedIDs: []string{
					"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd",
					"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d",
					"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12",
				},
				MaxFrequencyMs: 100,
			},
		},
	))

	fmt.Print("mercuryTriggerOutput", mercuryTriggerOutput)

	// TODO: Input shows up as an interpolated value in the spec
	// QQ: How is interpolated value represented today?
	// A: workflows.StepInputs
	// QQ: How does workflows.StepInputs look like?
	// A: It has OutputRef and Mapping
	// QQ: How would one represent inputs from multiple nodes to the consensus
	// capability?
	// A: Consensus capability can multiple the inputs for testing purposes.

	consensusOutput := workflows.AddConsensus(workflow, ocr3.NewOCR3Consensus())

	fmt.Print("consensusOutput", consensusOutput)

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
				Ref: "trigger-0",
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
		Actions: []workflows.StepDefinition{
			{
				ID:  "offchain_reporting@1.0.0",
				Ref: "consensus-0",
				Inputs: workflows.StepInputs{
					OutputRef: "trigger-0",
					Mapping: map[string]any{
						"observations": "$(trigger-0.outputs)",
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
			},
			//    inputs:
			//      observations:
			//        - "$(trigger.outputs)"
			//    config:
			//      report_id: "0001"
			//      aggregation_method: "data_feeds"
			//      aggregation_config:
			//        "0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd":
			//          deviation: "0.05"
			//          heartbeat: 3600
			//        "0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d":
			//          deviation: "0.05"
			//          heartbeat: 3600
			//        "0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12":
			//          deviation: "0.05"
			//          heartbeat: 3600
			//      encoder: "EVM"
			//      encoder_config:
			//        abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports"

		},
		Consensus: []workflows.StepDefinition{},
		Targets:   []workflows.StepDefinition{},
	}

	assert.Equal(t, expectedSpec, testWorkflowSpec)
}
