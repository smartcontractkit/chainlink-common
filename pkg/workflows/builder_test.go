package workflows_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

// 1. Capability defines JSON schema for inputs and outputs of a capability.
// Trigger: triggerOutputType := workflowBuilder.addTrigger(DataStreamsTrigger.Config{})
// Adds metadata to the builder. Returns output type.

func NewWorkflowSpec() (workflows.WorkflowSpec, error) {
	workflow := workflows.NewWorkflow(workflows.NewWorkflowParams{
		Owner: "test",
		Name:  "test",
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

	return workflow.Spec(), nil
}

func TestBuilder_ValidSpec(t *testing.T) {
	testWorkflowSpec, err := NewWorkflowSpec()
	require.NoError(t, err)

	expectedSpec := workflows.WorkflowSpec{
		Name:  "test",
		Owner: "test",
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
		Actions:   []workflows.StepDefinition{},
		Consensus: []workflows.StepDefinition{},
		Targets:   []workflows.StepDefinition{},
	}

	assert.Equal(t, expectedSpec, testWorkflowSpec)
}
