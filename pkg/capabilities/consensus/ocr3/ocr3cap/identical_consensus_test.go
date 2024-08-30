package ocr3cap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

func TestIdenticalConsensus(t *testing.T) {
	t.Parallel()
	workflow := sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{
		Owner: "0x1234",
		Name:  "Test",
	})

	trigger := basictrigger.TriggerConfig{Name: "1234", Number: 1}.New(workflow)

	consensus := ocr3cap.IdenticalConsensusConfig[basictrigger.TriggerOutputs]{
		Encoder:       "EVM",
		EncoderConfig: ocr3cap.EncoderConfig{Abi: "(string CoolOutput)"},
	}.New(workflow, "consensus", ocr3cap.IdenticalConsensusInput[basictrigger.TriggerOutputs]{Observations: trigger})

	chainwriter.TargetConfig{
		Address:    "0x1235",
		DeltaStage: "45s",
		Schedule:   "oneAtATime",
	}.New(workflow, "chainwriter@1.0.0", chainwriter.TargetInput{SignedReport: consensus})

	actual, err := workflow.Spec()
	require.NoError(t, err)

	expected := sdk.WorkflowSpec{
		Name:  "Test",
		Owner: "0x1234",
		Triggers: []sdk.StepDefinition{
			{
				ID:     "basic-test-trigger@1.0.0",
				Ref:    "trigger",
				Inputs: sdk.StepInputs{},
				Config: map[string]any{
					"name":   "1234",
					"number": 1,
				},
				CapabilityType: capabilities.CapabilityTypeTrigger,
			},
		},
		Actions: []sdk.StepDefinition{},
		Consensus: []sdk.StepDefinition{
			{
				ID:             "offchain_reporting@1.0.0",
				Ref:            "consensus",
				Inputs:         sdk.StepInputs{Mapping: map[string]any{"observations": "$(trigger.outputs)"}},
				Config:         map[string]any{"encoder": "EVM", "encoder_config": map[string]any{"abi": "(string CoolOutput)"}},
				CapabilityType: capabilities.CapabilityTypeConsensus,
			},
		},
		Targets: []sdk.StepDefinition{
			{
				ID: "chainwriter@1.0.0",
				Inputs: sdk.StepInputs{
					Mapping: map[string]any{"signed_report": "$(consensus.outputs)"},
				},
				Config: map[string]any{
					"address":    "0x1235",
					"deltaStage": "45s",
					"schedule":   "oneAtATime",
				},
				CapabilityType: capabilities.CapabilityTypeTarget,
			},
		},
	}

	testutils.AssertWorkflowSpec(t, expected, actual)
}