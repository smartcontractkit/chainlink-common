package sdk_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/anymapaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/capabilities/mapaction"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testdata/fixtures/capabilities/listtrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testdata/fixtures/capabilities/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-types --dir $GOFILE

// Note that the set of tests in this file cover the conversion from existing YAML -> code
// along with testing the structure of what is generated from the builders.
// This implicitly tests the code generators functionally, as the generated code is used in the tests.

type Config struct {
	Streams     *streams.TriggerConfig
	Ocr         *ocr3.DataFeedsConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

func NewWorkflowSpec(rawConfig []byte) (*sdk.WorkflowSpecFactory, error) {
	conf, err := UnmarshalYaml[Config](rawConfig)
	if err != nil {
		return nil, err
	}

	workflow := sdk.NewWorkflowSpecFactory()
	streamsTrigger := conf.Streams.New(workflow)
	consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
		Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
	)

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

	return workflow, nil
}

// ModifiedConfig, and the test it's used in, show how you can structure config to remove copy/paste issues when data
// needs to be repeated in multiple capability configurations.
type ModifiedConfig struct {
	AllowedPartialStaleness string
	MaxFrequencyMs          uint64
	DefaultHeartbeat        uint64        `yaml:"default_heartbeat" json:"default_heartbeat"`
	DefaultDeviation        string        `yaml:"default_deviation" json:"default_deviation"`
	FeedInfo                []FeedInfo    `yaml:"feed_info" json:"feed_info"`
	ReportID                ocr3.ReportId `yaml:"report_id" json:"report_id"`
	KeyID                   ocr3.KeyId    `yaml:"key_id" json:"key_id"`
	Encoder                 ocr3.Encoder
	EncoderConfig           ocr3.EncoderConfig `yaml:"encoder_config" json:"encoder_config"`
	ChainWriter             *chainwriter.TargetConfig
	TargetChain             string
}

type FeedInfo struct {
	FeedID     streams.FeedId
	Deviation  *string
	Heartbeat  *uint64
	RemappedID *string
}

func NewWorkflowRemapped(rawConfig []byte) (*sdk.WorkflowSpecFactory, error) {
	conf, err := UnmarshalYaml[ModifiedConfig](rawConfig)
	if err != nil {
		return nil, err
	}

	streamsConfig := streams.TriggerConfig{MaxFrequencyMs: conf.MaxFrequencyMs}
	ocr3Config := ocr3.DataFeedsConsensusConfig{
		AggregationMethod: "data_feeds",
		Encoder:           conf.Encoder,
		EncoderConfig:     conf.EncoderConfig,
		ReportId:          conf.ReportID,
		KeyId:             conf.KeyID,
		AggregationConfig: ocr3.DataFeedsConsensusConfigAggregationConfig{
			AllowedPartialStaleness: conf.AllowedPartialStaleness,
		},
	}

	feeds := ocr3.DataFeedsConsensusConfigAggregationConfigFeeds{}
	for _, elm := range conf.FeedInfo {
		streamsConfig.FeedIds = append(streamsConfig.FeedIds, elm.FeedID)
		feed := ocr3.FeedValue{
			Deviation:  conf.DefaultDeviation,
			Heartbeat:  conf.DefaultHeartbeat,
			RemappedID: elm.RemappedID,
		}
		if elm.Deviation != nil {
			feed.Deviation = *elm.Deviation
		}

		if elm.Heartbeat != nil {
			feed.Heartbeat = *elm.Heartbeat
		}

		feeds[string(elm.FeedID)] = feed
	}
	ocr3Config.AggregationConfig.Feeds = feeds

	workflow := sdk.NewWorkflowSpecFactory()
	streamsTrigger := streamsConfig.New(workflow)

	consensus := ocr3Config.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
		Observations: sdk.ListOf[streams.Feed](streamsTrigger),
	})

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

	return workflow, nil
}

const anyFakeFeedID = "0x0000000000000000000000000000000000000000000000000000000000000000"

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (*sdk.WorkflowSpecFactory, error) {
	conf, err := UnmarshalYaml[NotStreamsConfig](rawConfig)
	if err != nil {
		return nil, err
	}

	workflow := sdk.NewWorkflowSpecFactory()
	notStreamsTrigger := conf.NotStream.New(workflow)

	md := streams.NewSignersMetadataFromFields(
		sdk.ConstantDefinition(int64(1)), sdk.ListOf(notStreamsTrigger.Metadata().Signer()))

	payload := streams.NewFeedReportFromFields(
		notStreamsTrigger.Payload().BuyPrice(),
		streams.FeedIdWrapper(sdk.ConstantDefinition[streams.FeedId](anyFakeFeedID)),
		notStreamsTrigger.Payload().FullReport(),
		notStreamsTrigger.Payload().ObservationTimestamp(),
		notStreamsTrigger.Payload().ReportContext(),
		sdk.ListOf(notStreamsTrigger.Payload().Signature()),
	)

	feedsInput := streams.NewFeedFromFields(
		md,
		sdk.ListOf[streams.FeedReport](payload),
		notStreamsTrigger.Timestamp(),
	)

	ocrConfig := ocr3.DataFeedsConsensusConfig{
		AggregationConfig: ocr3.DataFeedsConsensusConfigAggregationConfig{
			AllowedPartialStaleness: conf.Ocr.AllowedPartialStaleness,
			Feeds: map[string]ocr3.FeedValue{
				anyFakeFeedID: {
					Deviation: conf.Ocr.Deviation,
					Heartbeat: conf.Ocr.Heartbeat,
				},
			},
		},
		AggregationMethod: conf.Ocr.AggregationMethod,
		Encoder:           conf.Ocr.Encoder,
		EncoderConfig:     conf.Ocr.EncoderConfig,
		ReportId:          conf.Ocr.ReportID,
		KeyId:             conf.Ocr.KeyID,
	}

	consensus := ocrConfig.New(workflow, "data-feeds-report", ocr3.DataFeedsConsensusInput{
		Observations: sdk.ListOf[streams.Feed](feedsInput),
	})

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

	return workflow, nil
}

//go:embed testdata/fixtures/workflows/sepolia.yaml
var sepoliaConfig []byte

//go:embed testdata/fixtures/workflows/sepolia_defaults.yaml
var sepoliaDefaultConfig []byte

//go:embed testdata/fixtures/workflows/expected_sepolia.yaml
var expectedSepolia []byte

//go:embed testdata/fixtures/workflows/notstreamssepolia.yaml
var notStreamSepoliaConfig []byte

func TestBuilder_ValidSpec(t *testing.T) {
	t.Run("basic config", func(t *testing.T) {
		runSepoliaStagingTest(t, sepoliaConfig, NewWorkflowSpec)
	})

	t.Run("remapping config", func(t *testing.T) {
		runSepoliaStagingTest(t, sepoliaDefaultConfig, NewWorkflowRemapped)
	})

	// This test intentionally uses a similar complex type to the real steams trigger
	// this helps assure that mapping works correctly under many circumstances, including hard-coding
	// and wrapping values into arrays, while still remaining somewhat realistic
	t.Run("mapping different types without compute", func(t *testing.T) {
		factory, err := NewWorkflowSpecFromPrimitives(notStreamSepoliaConfig)
		require.NoError(t, err)

		actual, err := factory.Spec()
		require.NoError(t, err)

		expected := sdk.WorkflowSpec{
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
						"key_id":    "evm",
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
						"address":          "0xE0082363396985ae2FdcC3a9F816A586Eed88416",
						"deltaStage":       "45s",
						"schedule":         "oneAtATime",
						"cre_step_timeout": 0,
					},
					CapabilityType: capabilities.CapabilityTypeTarget,
				},
			},
		}

		testutils.AssertWorkflowSpec(t, expected, actual)
	})

	t.Run("maps work correctly", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "1", Number: 1}.New(workflow)
		mapaction.ActionConfig{}.New(workflow, "ref", mapaction.ActionInput{Payload: sdk.Map[string, mapaction.ActionInputsPayload](map[string]sdk.CapDefinition[string]{"Foo": trigger.CoolOutput()})})
		spec, err := workflow.Spec()
		require.NoError(t, err)
		testutils.AssertWorkflowSpec(t, sdk.WorkflowSpec{
			Triggers: []sdk.StepDefinition{
				{
					ID:     "basic-test-trigger@1.0.0",
					Ref:    "trigger",
					Inputs: sdk.StepInputs{},
					Config: map[string]any{
						"name":   "1",
						"number": uint64(1),
					},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "mapaction@1.0.0",
					Ref: "ref",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"payload": map[string]string{"Foo": "$(trigger.outputs.cool_output)"}},
					},
					Config:         map[string]any{},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{},
			Targets:   []sdk.StepDefinition{},
		}, spec)
	})

	t.Run("any maps work correctly", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "1", Number: 1}.New(workflow)
		anymapaction.MapActionConfig{}.New(workflow, "ref", anymapaction.MapActionInput{Payload: sdk.AnyMap[anymapaction.MapActionInputsPayload](sdk.CapMap{"Foo": trigger.CoolOutput()})})
		spec, err := workflow.Spec()
		require.NoError(t, err)
		testutils.AssertWorkflowSpec(t, sdk.WorkflowSpec{
			Triggers: []sdk.StepDefinition{
				{
					ID:     "basic-test-trigger@1.0.0",
					Ref:    "trigger",
					Inputs: sdk.StepInputs{},
					Config: map[string]any{
						"name":   "1",
						"number": uint64(1),
					},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "anymapaction@1.0.0",
					Ref: "ref",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"payload": map[string]string{"Foo": "$(trigger.outputs.cool_output)"}},
					},
					Config:         map[string]any{},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{},
			Targets:   []sdk.StepDefinition{},
		}, spec)
	})

	t.Run("ToListDefinition works correctly for list elements", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := listtrigger.TriggerConfig{Name: "1"}.New(workflow)
		asList := sdk.ToListDefinition[string](trigger.CoolOutput())
		sdk.Compute1(workflow, "compute", sdk.Compute1Inputs[[]string]{Arg0: asList}, func(_ sdk.Runtime, inputs []string) (string, error) {
			return inputs[0], nil
		})
		sdk.Compute1(workflow, "compute again", sdk.Compute1Inputs[string]{Arg0: asList.Index(0)}, func(runtime sdk.Runtime, input string) (string, error) {
			return input, nil
		})

		spec, err := workflow.Spec()
		require.NoError(t, err)

		testutils.AssertWorkflowSpec(t, sdk.WorkflowSpec{
			Triggers: []sdk.StepDefinition{
				{
					ID:             "list@1.0.0",
					Ref:            "trigger",
					Inputs:         sdk.StepInputs{},
					Config:         map[string]any{"name": "1"},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"Arg0": "$(trigger.outputs.cool_output)"},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute again",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"Arg0": "$(trigger.outputs.cool_output.0)"},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{},
			Targets:   []sdk.StepDefinition{},
		}, spec)
	})

	t.Run("ToListDefinition works correctly for built up lists", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "1"}.New(workflow)
		asList := sdk.ToListDefinition(sdk.ListOf(trigger.CoolOutput()))
		sdk.Compute1(workflow, "compute", sdk.Compute1Inputs[[]string]{Arg0: asList}, func(_ sdk.Runtime, inputs []string) (string, error) {
			return inputs[0], nil
		})
		sdk.Compute1(workflow, "compute again", sdk.Compute1Inputs[string]{Arg0: asList.Index(0)}, func(runtime sdk.Runtime, input string) (string, error) {
			return input, nil
		})

		spec, err := workflow.Spec()
		require.NoError(t, err)

		testutils.AssertWorkflowSpec(t, sdk.WorkflowSpec{
			Triggers: []sdk.StepDefinition{
				{
					ID:             "basic-test-trigger@1.0.0",
					Ref:            "trigger",
					Inputs:         sdk.StepInputs{},
					Config:         map[string]any{"name": "1", "number": uint64(0)},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"Arg0": []any{"$(trigger.outputs.cool_output)"}},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute again",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{"Arg0": "$(trigger.outputs.cool_output)"},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{},
			Targets:   []sdk.StepDefinition{},
		}, spec)
	})

	t.Run("ToListDefinition works correctly for hard-coded lists", func(t *testing.T) {
		workflow := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "1"}.New(workflow)
		list := sdk.ToListDefinition(sdk.ConstantDefinition([]string{"1", "2"}))
		sdk.Compute2(workflow, "compute", sdk.Compute2Inputs[string, []string]{Arg0: trigger.CoolOutput(), Arg1: list}, func(_ sdk.Runtime, t string, l []string) (string, error) {
			return "", nil
		})
		sdk.Compute2(workflow, "compute again", sdk.Compute2Inputs[string, string]{Arg0: trigger.CoolOutput(), Arg1: list.Index(0)}, func(_ sdk.Runtime, t string, l string) (string, error) {
			return "", nil
		})

		spec, err := workflow.Spec()
		require.NoError(t, err)

		testutils.AssertWorkflowSpec(t, sdk.WorkflowSpec{
			Triggers: []sdk.StepDefinition{
				{
					ID:             "basic-test-trigger@1.0.0",
					Ref:            "trigger",
					Inputs:         sdk.StepInputs{},
					Config:         map[string]any{"name": "1", "number": uint64(0)},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: []sdk.StepDefinition{
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{
							"Arg0": "$(trigger.outputs.cool_output)",
							"Arg1": []string{"1", "2"},
						},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
				{
					ID:  "custom-compute@1.0.0",
					Ref: "compute again",
					Inputs: sdk.StepInputs{
						Mapping: map[string]any{
							"Arg0": "$(trigger.outputs.cool_output)",
							"Arg1": "1",
						},
					},
					Config: map[string]any{
						"config": "$(ENV.config)",
						"binary": "$(ENV.binary)",
					},
					CapabilityType: capabilities.CapabilityTypeAction,
				},
			},
			Consensus: []sdk.StepDefinition{},
			Targets:   []sdk.StepDefinition{},
		}, spec)
	})

	t.Run("AnyListOf works like list of but returns a type any", func(t *testing.T) {
		workflow1 := sdk.NewWorkflowSpecFactory()
		trigger := basictrigger.TriggerConfig{Name: "foo", Number: 0}
		list := sdk.ListOf(trigger.New(workflow1).CoolOutput())
		sdk.Compute1(workflow1, "compute", sdk.Compute1Inputs[[]string]{Arg0: list}, func(_ sdk.Runtime, inputs []string) (string, error) {
			return inputs[0], nil
		})

		workflow2 := sdk.NewWorkflowSpecFactory()
		anyList := sdk.AnyListOf(trigger.New(workflow2).CoolOutput())
		sdk.Compute1(workflow2, "compute", sdk.Compute1Inputs[[]any]{Arg0: anyList}, func(_ sdk.Runtime, inputs []any) (any, error) {
			return inputs[0], nil
		})

		spec1, err := workflow1.Spec()
		require.NoError(t, err)
		spec2, err := workflow2.Spec()
		require.NoError(t, err)

		testutils.AssertWorkflowSpec(t, spec1, spec2)
	})

	t.Run("duplicate names causes errors", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := sdk.NewWorkflowSpecFactory()
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
		)

		consensus2 := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus2})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("empty ref causes an error", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := sdk.NewWorkflowSpecFactory()
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("bad capability type causes an error", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := sdk.NewWorkflowSpecFactory()
		badStep := sdk.Step[streams.Feed]{
			Definition: sdk.StepDefinition{
				ID:             "streams-trigger@1.0.0",
				Ref:            "Trigger",
				Inputs:         sdk.StepInputs{},
				Config:         map[string]any{},
				CapabilityType: "fake",
			},
		}

		badCap := badStep.AddTo(workflow)

		consensus := conf.Ocr.New(workflow, "", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](badCap)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("Capabilities can be used multiple times with different references", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := sdk.NewWorkflowSpecFactory()
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
		)

		consensus2 := conf.Ocr.New(workflow, "ccip_feeds_different", ocr3.DataFeedsConsensusInput{
			Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus2})

		_, err = workflow.Spec()
		require.NoError(t, err)
	})
}

func runSepoliaStagingTest(t *testing.T, config []byte, gen func([]byte) (*sdk.WorkflowSpecFactory, error)) {
	testFactory, err := gen(config)
	require.NoError(t, err)

	testWorkflowSpec, err := testFactory.Spec()
	require.NoError(t, err)

	expectedSpecYaml, err := UnmarshalYaml[workflows.WorkflowSpecYaml](expectedSepolia)
	require.NoError(t, err)
	expectedSpec := expectedSpecYaml.ToWorkflowSpec()
	testutils.AssertWorkflowSpec(t, expectedSpec, testWorkflowSpec)
}

type NotStreamsConfig struct {
	NotStream   *notstreams.TriggerConfig `yaml:"not_stream" json:"not_stream"`
	Ocr         *ModifiedConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

type ModifiedConsensusConfig struct {
	AllowedPartialStaleness string                                         `json:"allowedPartialStaleness" yaml:"allowedPartialStaleness" mapstructure:"allowedPartialStaleness"`
	Deviation               string                                         `json:"deviation" yaml:"deviation" mapstructure:"deviation"`
	Heartbeat               uint64                                         `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`
	AggregationMethod       ocr3.DataFeedsConsensusConfigAggregationMethod `json:"aggregation_method" yaml:"aggregation_method" mapstructure:"aggregation_method"`
	Encoder                 ocr3.Encoder                                   `json:"encoder" yaml:"encoder" mapstructure:"encoder"`
	EncoderConfig           ocr3.EncoderConfig                             `json:"encoder_config" yaml:"encoder_config" mapstructure:"encoder_config"`
	ReportID                ocr3.ReportId                                  `json:"report_id" yaml:"report_id" mapstructure:"report_id"`
	KeyID                   ocr3.KeyId                                     `json:"key_id" yaml:"key_id" mapstructure:"key_id"`
}

func UnmarshalYaml[T any](raw []byte) (*T, error) {
	var v T
	err := yaml.Unmarshal(raw, &v)
	return &v, err
}
