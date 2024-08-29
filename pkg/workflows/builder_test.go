package workflows_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testdata/fixtures/capabilities/notstreams"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-types --dir $GOFILE

// Note that the set of tests in this file cover the conversion from existing YAML -> code
// along with testing the structure of what is generated from the builders.
// This implicitly tests the code generators functionally, as the generated code is used in the tests.

type Config struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *streams.TriggerConfig
	Ocr         *ocr3.DataFeedsConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

func NewWorkflowSpec(rawConfig []byte) (*workflows.WorkflowSpecFactory, error) {
	conf, err := UnmarshalYaml[Config](rawConfig)
	if err != nil {
		return nil, err
	}

	workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
	streamsTrigger := conf.Streams.New(workflow)
	consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
		Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
	)

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

	return workflow, nil
}

// ModifiedConfig, and the test it's used in, show how you can structure config to remove copy/paste issues when data
// needs to be repeated in multiple capability configurations.
type ModifiedConfig struct {
	Workflow                workflows.NewWorkflowParams
	AllowedPartialStaleness string
	MaxFrequencyMs          int
	DefaultHeartbeat        int        `yaml:"default_heartbeat" json:"default_heartbeat"`
	DefaultDeviation        string     `yaml:"default_deviation" json:"default_deviation"`
	FeedInfo                []FeedInfo `yaml:"feed_info" json:"feed_info"`
	ReportID                string     `yaml:"report_id" json:"report_id"`
	Encoder                 ocr3.DataFeedsConsensusConfigEncoder
	EncoderConfig           ocr3.DataFeedsConsensusConfigEncoderConfig `yaml:"encoder_config" json:"encoder_config"`
	ChainWriter             *chainwriter.TargetConfig
	TargetChain             string
}

type FeedInfo struct {
	FeedID     streams.FeedId
	Deviation  *string
	Heartbeat  *int
	RemappedID *string
}

func NewWorkflowRemapped(rawConfig []byte) (*workflows.WorkflowSpecFactory, error) {
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

	workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
	streamsTrigger := streamsConfig.New(workflow)

	consensus := ocr3Config.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
		Observations: workflows.ListOf[[]streams.Feed](streamsTrigger),
	})

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

	return workflow, nil
}

const anyFakeFeedID = "0x0000000000000000000000000000000000000000000000000000000000000000"

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (*workflows.WorkflowSpecFactory, error) {
	conf, err := UnmarshalYaml[NotStreamsConfig](rawConfig)
	if err != nil {
		return nil, err
	}

	workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
	notStreamsTrigger := conf.NotStream.New(workflow)

	feedsInput := streams.NewFeedFromFields(
		notStreamsTrigger.Price().PriceA(),
		workflows.ConstantDefinition[streams.FeedId](anyFakeFeedID),
		notStreamsTrigger.FullReport(),
		notStreamsTrigger.Timestamp(),
		notStreamsTrigger.ReportContext(),
		notStreamsTrigger.Signatures(),
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
	}

	consensus := ocrConfig.New(workflow, "data-feeds-report", ocr3.DataFeedsConsensusInput{
		Observations: workflows.ListOf[[]streams.Feed](workflows.ListOf[streams.Feed](feedsInput)),
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

	t.Run("mapping different types without compute", func(t *testing.T) {
		factory, err := NewWorkflowSpecFromPrimitives(notStreamSepoliaConfig)
		require.NoError(t, err)

		actual, err := factory.Spec()
		require.NoError(t, err)

		expected := workflows.WorkflowSpec{
			Name:  "notccipethsep",
			Owner: "0x00000000000000000000000000000000000000aa",
			Triggers: []workflows.StepDefinition{
				{
					ID:             "notstreams@1.0.0",
					Ref:            "trigger",
					Inputs:         workflows.StepInputs{},
					Config:         map[string]any{"maxFrequencyMs": 5000},
					CapabilityType: capabilities.CapabilityTypeTrigger,
				},
			},
			Actions: make([]workflows.StepDefinition, 0),
			Consensus: []workflows.StepDefinition{
				{
					ID:  "offchain_reporting@1.0.0",
					Ref: "data-feeds-report",
					Inputs: workflows.StepInputs{
						Mapping: map[string]any{"observations": [][]map[string]any{
							{
								{
									"benchmarkPrice":       "$(trigger.outputs.Price.PriceA)",
									"feedId":               anyFakeFeedID,
									"fullReport":           "$(trigger.outputs.fullReport)",
									"observationTimestamp": "$(trigger.outputs.Timestamp)",
									"reportContext":        "$(trigger.outputs.reportContext)",
									"signatures":           "$(trigger.outputs.signatures)",
								},
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
						"encoder_config": ocr3.DataFeedsConsensusConfigEncoderConfig{
							Abi: "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
						},
						"report_id": "0001",
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

		assertWorkflowSpec(t, expected, actual)
	})

	t.Run("duplicate names causes errors", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
		)

		consensus2 := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus2})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("empty ref causes an error", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("bad capability type causes an error", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
		badStep := workflows.Step[[]streams.Feed]{
			Definition: workflows.StepDefinition{
				ID:             "streams-trigger@1.0.0",
				Ref:            "Trigger",
				Inputs:         workflows.StepInputs{},
				Config:         map[string]any{},
				CapabilityType: "fake",
			},
		}

		badCap := badStep.AddTo(workflow)

		consensus := conf.Ocr.New(workflow, "", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](badCap)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		_, err = workflow.Spec()
		require.Error(t, err)
	})

	t.Run("Capabilities can be used multiple times with different references", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := workflows.NewWorkflowSpecFactory(conf.Workflow)
		streamsTrigger := conf.Streams.New(workflow)
		consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
		)

		consensus2 := conf.Ocr.New(workflow, "ccip_feeds_different", ocr3.DataFeedsConsensusInput{
			Observations: workflows.ListOf[[]streams.Feed](streamsTrigger)},
		)

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})

		conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus2})

		_, err = workflow.Spec()
		require.NoError(t, err)
	})
}

func runSepoliaStagingTest(t *testing.T, config []byte, gen func([]byte) (*workflows.WorkflowSpecFactory, error)) {
	testFactory, err := gen(config)
	require.NoError(t, err)

	testWorkflowSpec, err := testFactory.Spec()
	require.NoError(t, err)

	expectedSpecYaml, err := UnmarshalYaml[workflows.WorkflowSpecYaml](expectedSepolia)
	require.NoError(t, err)
	expectedSpec := expectedSpecYaml.ToWorkflowSpec()
	assertWorkflowSpec(t, expectedSpec, testWorkflowSpec)
}

func assertWorkflowSpec(t *testing.T, expectedSpec, testWorkflowSpec workflows.WorkflowSpec) {
	expected, err := json.Marshal(expectedSpec)
	require.NoError(t, err)

	actual, err := json.Marshal(testWorkflowSpec)
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

type NotStreamsConfig struct {
	Workflow    workflows.NewWorkflowParams
	NotStream   *notstreams.TriggerConfig `yaml:"not_stream" json:"not_stream"`
	Ocr         *ModifiedConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

type ModifiedConsensusConfig struct {
	AllowedPartialStaleness string                                         `json:"allowedPartialStaleness" yaml:"allowedPartialStaleness" mapstructure:"allowedPartialStaleness"`
	Deviation               string                                         `json:"deviation" yaml:"deviation" mapstructure:"deviation"`
	Heartbeat               int                                            `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`
	AggregationMethod       ocr3.DataFeedsConsensusConfigAggregationMethod `json:"aggregation_method" yaml:"aggregation_method" mapstructure:"aggregation_method"`
	Encoder                 ocr3.DataFeedsConsensusConfigEncoder           `json:"encoder" yaml:"encoder" mapstructure:"encoder"`
	EncoderConfig           ocr3.DataFeedsConsensusConfigEncoderConfig     `json:"encoder_config" yaml:"encoder_config" mapstructure:"encoder_config"`
	ReportID                string                                         `json:"report_id" yaml:"report_id" mapstructure:"report_id"`
}

func UnmarshalYaml[T any](raw []byte) (*T, error) {
	var v T
	err := yaml.Unmarshal(raw, &v)
	return &v, err
}
