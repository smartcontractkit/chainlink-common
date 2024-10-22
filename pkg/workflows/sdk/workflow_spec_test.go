package sdk_test

import (
	_ "embed"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

var notStreamSepoliaWorkflowSpec = sdk.WorkflowSpec{
	Name:  "notccipethsep",
	Owner: "0x00000000000000000000000000000000000000aa",
	Triggers: []sdk.StepDefinition{
		{
			ID:             "notstreams@1.0.0",
			Ref:            "trigger",
			Inputs:         sdk.StepInputs{},
			Config:         map[string]any{"maxFrequencyMs": 5000},
			CapabilityType: capabilities.CapabilityTypeTrigger,
		},
	},
	Actions: make([]sdk.StepDefinition, 0),
	Consensus: []sdk.StepDefinition{
		{
			ID:  "offchain_reporting@1.0.0",
			Ref: "data-feeds-report",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"observations": []map[string]any{
					{
						"Metadata": map[string]any{
							"MinRequiredSignatures": 1,
							"Signers":               []string{"$(trigger.outputs.Metadata.Signer)"},
						},
						"Payload": []map[string]any{
							{
								"BenchmarkPrice":       "$(trigger.outputs.Payload.BuyPrice)",
								"FeedID":               anyFakeFeedID,
								"FullReport":           "$(trigger.outputs.Payload.FullReport)",
								"ObservationTimestamp": "$(trigger.outputs.Payload.ObservationTimestamp)",
								"ReportContext":        "$(trigger.outputs.Payload.ReportContext)",
								"Signatures":           []string{"$(trigger.outputs.Payload.Signature)"},
							},
						},
						"Timestamp": "$(trigger.outputs.Timestamp)",
					},
				}},
			},
			Config: map[string]any{
				"aggregation_config": ocr3.DataFeedsConsensusConfigAggregationConfig{
					AllowedPartialStaleness: "0.5",
					Feeds: map[string]ocr3.FeedValue{
						anyFakeFeedID: {
							Deviation: "0.5",
							Heartbeat: 3600,
						},
					},
				},
				"aggregation_method": "data_feeds",
				"encoder":            "EVM",
				"encoder_config": ocr3.EncoderConfig{
					"Abi": "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
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
