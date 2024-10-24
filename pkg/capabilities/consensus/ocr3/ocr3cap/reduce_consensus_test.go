package ocr3cap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

func TestReduceConsensus(t *testing.T) {
	t.Parallel()
	workflow := sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{
		Owner: "0x1234",
		Name:  "Test",
	})

	trigger := basictrigger.TriggerConfig{Name: "1234", Number: 1}.New(workflow)

	consensus := ocr3.ReduceConsensusConfig[basictrigger.TriggerOutputs]{
		Encoder:       ocr3.EncoderEVM,
		EncoderConfig: ocr3.EncoderConfig{},
		ReportID:      "0001",
		AggregationConfig: aggregators.ReduceAggConfig{
			Fields: []aggregators.AggregationField{
				{
					Method: "median",
				},
			},
		},
	}.New(workflow, "consensus", ocr3.ReduceConsensusInput[basictrigger.TriggerOutputs]{
		Observation:   trigger,
		Encoder:       "evm",
		EncoderConfig: ocr3.EncoderConfig(map[string]any{"foo": "bar"}),
	})

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
				ID:  "offchain_reporting@1.0.0",
				Ref: "consensus",
				Inputs: sdk.StepInputs{Mapping: map[string]any{
					"observations":  []any{"$(trigger.outputs)"},
					"encoder":       "evm",
					"encoderConfig": map[string]any{"foo": "bar"},
				}},
				Config: map[string]any{
					"encoder":            "EVM",
					"encoder_config":     map[string]any{},
					"aggregation_method": "reduce",
					"report_id":          "0001",
				},
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
