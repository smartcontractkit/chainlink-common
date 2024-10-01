package sdk_test

import (
	"context"
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/mathutil"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

var update = flag.Bool("update", true, "update golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

//go:generate go test -run=TestWorkflowSpecFormatChart . -update
//go:embed testdata/fixtures/charts
var charts embed.FS

// requireEqualChart compares the formatted workflow chart against the golden file testdata/fixtures/charts/<name>.md,
// or updates it when update is true.
func requireEqualChart(t *testing.T, name string, workflow sdk.WorkflowSpec) {
	t.Helper()
	path := filepath.Join("testdata/fixtures/charts", name+".md")

	s, err := workflow.FormatChart()
	require.NoError(t, err)

	got := fmt.Sprintf("```mermaid\n%s\n```", s)

	if *update {
		require.NoError(t, os.WriteFile(path, []byte(got), 0600))
		return
	}

	b, err := charts.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, string(b), got)
}

// TestWorkflowSpecFormatChart depends on charts golden files, and will regenerate them
// when the -update flag is used.
func TestWorkflowSpecFormatChart(t *testing.T) {
	for _, tt := range []struct {
		name     string
		workflow sdk.WorkflowSpec
	}{
		{"notstreamssepolia", notStreamSepoliaWorkflowSpec},
		{"serial", serialWorkflowSpec},
		{"parallel", parallelWorkflowSpec},
		{"builder_parallel", buildSimpleWorkflowSpec(
			sdk.NewWorkflowSpecFactory(sdk.NewWorkflowParams{Owner: "test", Name: "parallel"}),
		).MustSpec(t)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requireEqualChart(t, tt.name, tt.workflow)
		})
	}
}

func buildSimpleWorkflowSpec(w *sdk.WorkflowSpecFactory) *sdk.WorkflowSpecFactory {
	trigger := basictrigger.TriggerConfig{
		Name:   "trigger",
		Number: 100,
	}.New(w)

	foo := sdk.Compute1(w, "get-foo", sdk.Compute1Inputs[string]{
		Arg0: trigger.CoolOutput(),
	}, func(runtime sdk.Runtime, s string) (int64, error) {
		ctx := context.Background()
		req, err := http.NewRequest("GET", "https://foo.com/"+s, nil)
		if err != nil {
			return -1, fmt.Errorf("failed to create request for foo.com: %w", err)
		}
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return -1, fmt.Errorf("failed to get data from foo.com: %w", err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("failed to read response from foo.com: %w", err)
		}
		return strconv.ParseInt(string(b), 10, 64)
	})
	bar := sdk.Compute1(w, "get-bar", sdk.Compute1Inputs[string]{
		Arg0: trigger.CoolOutput(),
	}, func(runtime sdk.Runtime, s string) (int64, error) {
		ctx := context.Background()
		req, err := http.NewRequest("GET", "https://bar.io/api/"+s, nil)
		if err != nil {
			return -1, fmt.Errorf("failed to create request for bar.io: %w", err)
		}
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return -1, fmt.Errorf("failed to get data from bar.io: %w", err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("failed to read response from bar.io: %w", err)
		}
		return strconv.ParseInt(string(b), 10, 64)
	})
	baz := sdk.Compute1(w, "get-baz", sdk.Compute1Inputs[string]{
		Arg0: trigger.CoolOutput(),
	}, func(runtime sdk.Runtime, s string) (int64, error) {
		ctx := context.Background()
		query := url.Values{"id": []string{s}}.Encode()
		req, err := http.NewRequest("GET", "https://baz.com/v2/path/to/thing?"+query, nil)
		if err != nil {
			return -1, fmt.Errorf("failed to create request for baz.com/v2: %w", err)
		}
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return -1, fmt.Errorf("failed to get data from baz.com/v2: %w", err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return -1, fmt.Errorf("failed to read response from baz.com/v2: %w", err)
		}
		return strconv.ParseInt(string(b), 10, 64)
	})

	compute := sdk.Compute3(w, "compute", sdk.Compute3Inputs[int64, int64, int64]{
		Arg0: foo.Value(),
		Arg1: bar.Value(),
		Arg2: baz.Value(),
	}, func(runtime sdk.Runtime, foo, bar, baz int64) ([]streams.Feed, error) {
		val, err := mathutil.Median(foo, bar, baz)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate median: %w", err)
		}
		return []streams.Feed{{
			Metadata: streams.SignersMetadata{},
			Payload: []streams.FeedReport{
				{FullReport: []byte(strconv.FormatInt(val, 10))},
			},
			Timestamp: 0,
		}}, nil
	})

	consensus := ocr3.DataFeedsConsensusConfig{}.New(w, "consensus", ocr3.DataFeedsConsensusInput{
		Observations: compute.Value(),
	})

	chainwriter.TargetConfig{
		Address: "0xfakeaddr",
	}.New(w, "id", chainwriter.TargetInput{
		SignedReport: consensus,
	})
	return w
}

func TestStepInputsOutput(t *testing.T) {
	os := notStreamSepoliaWorkflowSpec.Consensus[0].Inputs.Outputs()
	require.Equal(t, []sdk.Output{
		{Ref: "trigger",
			Name: "Metadata.Signer<br>Payload.BuyPrice<br>Payload.FullReport<br>Payload.ObservationTimestamp<br>Payload.ReportContext<br>Payload.Signature<br>Timestamp"},
	}, os)
}

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

var serialWorkflowSpec = sdk.WorkflowSpec{
	Name:  "serial",
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
			ID:  "custom_compute@1.0.0",
			Ref: "Compute",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(trigger.outputs)"},
			},
			Config: map[string]any{
				"binary": "$(ENV.binary)",
				"config": "$(ENV.config)",
			},
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

var parallelWorkflowSpec = sdk.WorkflowSpec{
	Name:  "parallel",
	Owner: "owner",
	Triggers: []sdk.StepDefinition{
		{
			ID:             "chain_reader@1.0.0",
			Ref:            "trigger-chain-event",
			Inputs:         sdk.StepInputs{},
			Config:         map[string]any{"maxFrequencyMs": 5000},
			CapabilityType: capabilities.CapabilityTypeTrigger,
		},
	},
	Actions: []sdk.StepDefinition{
		{
			ID:  "http@1.0.0",
			Ref: "get-foo",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(trigger-chain-event.outputs)"},
			},
			CapabilityType: capabilities.CapabilityTypeAction,
		},
		{
			ID:  "custom_compute@1.0.0",
			Ref: "compute-foo",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(get-foo.outputs)"},
			},
			CapabilityType: capabilities.CapabilityTypeAction,
		},
		{
			ID:  "http@1.0.0",
			Ref: "get-bar",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(trigger-chain-event.outputs)"},
			},
			CapabilityType: capabilities.CapabilityTypeAction,
		},
		{
			ID:  "custom_compute@1.0.0",
			Ref: "compute-bar",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(get-bar.outputs)"},
			},
			CapabilityType: capabilities.CapabilityTypeAction,
		},
		{
			ID:  "chain_reader@1.0.0",
			Ref: "read-token-price",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{"Arg0": "$(trigger-chain-event.outputs)"},
			},
			CapabilityType: capabilities.CapabilityTypeAction,
		},
	},
	Consensus: []sdk.StepDefinition{
		{
			ID:  "offchain_reporting@1.0.0",
			Ref: "data-feeds-report",
			Inputs: sdk.StepInputs{
				Mapping: map[string]any{
					"observations": []string{
						"$(compute-foo.outputs.Value)",
						"$(compute-bar.outputs.Value)",
					},
					"token_price": "$(read-token-price.outputs.Value)",
				},
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
			CapabilityType: capabilities.CapabilityTypeTarget,
		},
	},
}
