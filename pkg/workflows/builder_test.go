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
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter/chainwritercap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams/streamscap"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testdata/fixtures/capabilities/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testdata/fixtures/capabilities/notstreams/notstreamscap"
)

type Config struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *streams.TriggerConfig
	Ocr         *ocr3.ConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

func NewWorkflowSpec(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := Config{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	streamsTrigger := streamscap.NewTrigger(workflow, *conf.Streams)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](streamsTrigger)}
	consensus := ocr3cap.NewConsensus(workflow, "ccip_feeds", ocrInput, *conf.Ocr)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, conf.TargetChain, input, *conf.ChainWriter)

	return workflow.Spec()
}

// What if there were hundreds of feeds?  Like feeds that aren't for CCIP?

type ModifiedConfig struct {
	Workflow                workflows.NewWorkflowParams
	AllowedPartialStaleness string
	MaxFrequencyMs          int
	DefaultHeartbeat        int        `yaml:"default_heartbeat" json:"default_heartbeat"`
	DefaultDeviation        string     `yaml:"default_deviation" json:"default_deviation"`
	FeedInfo                []FeedInfo `yaml:"feed_info" json:"feed_info"`
	ReportId                string     `yaml:"report_id" json:"report_id"`
	Encoder                 ocr3.ConsensusConfigEncoder
	EncoderConfig           ocr3.ConsensusConfigEncoderConfig `yaml:"encoder_config" json:"encoder_config"`
	ChainWriter             *chainwriter.TargetConfig
	TargetChain             string
}

type FeedInfo struct {
	FeedId     streams.FeedId
	Deviation  *string
	Heartbeat  *int
	RemappedId *string
}

func NewWorkflowRemapped(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := ModifiedConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	streamsConfig := streams.TriggerConfig{MaxFrequencyMs: conf.MaxFrequencyMs}
	ocr3Config := ocr3.ConsensusConfig{
		AggregationMethod: "data_feeds",
		Encoder:           conf.Encoder,
		EncoderConfig:     conf.EncoderConfig,
		ReportId:          conf.ReportId,
		AggregationConfig: ocr3.ConsensusConfigAggregationConfig{
			AllowedPartialStaleness: conf.AllowedPartialStaleness,
		},
	}

	feeds := ocr3.ConsensusConfigAggregationConfigFeeds{}
	for _, elm := range conf.FeedInfo {
		streamsConfig.FeedIds = append(streamsConfig.FeedIds, elm.FeedId)
		feed := ocr3.FeedValue{
			Deviation:  conf.DefaultDeviation,
			Heartbeat:  conf.DefaultHeartbeat,
			RemappedID: elm.RemappedId,
		}
		if elm.Deviation != nil {
			feed.Deviation = *elm.Deviation
		}

		if elm.Heartbeat != nil {
			feed.Heartbeat = *elm.Heartbeat
		}

		feeds[string(elm.FeedId)] = feed
	}
	ocr3Config.AggregationConfig.Feeds = feeds

	workflow := workflows.NewWorkflow(conf.Workflow)
	streamsTrigger := streamscap.NewTrigger(workflow, streamsConfig)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](streamsTrigger)}
	consensus := ocr3cap.NewConsensus(workflow, "ccip_feeds", ocrInput, ocr3Config)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, conf.TargetChain, input, *conf.ChainWriter)

	return workflow.Spec()
}

const anyFakeFeedId = "0x0000000000000000000000000000000000000000000000000000000000000000"

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := NotStreamsConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	notStreamsTrigger := notstreamscap.NewTrigger(workflow, *conf.NotStream)

	feedsInput := streamscap.NewTriggerFromFields(
		notStreamsTrigger.Price().PriceA(),
		workflows.ConstantDefinition[streams.FeedId](anyFakeFeedId),
		notStreamsTrigger.FullReport(),
		notStreamsTrigger.Timestamp(),
		notStreamsTrigger.ReportContext(),
		notStreamsTrigger.Signatures(),
	)
	ocrConfig := ocr3.ConsensusConfig{
		AggregationConfig: ocr3.ConsensusConfigAggregationConfig{
			AllowedPartialStaleness: conf.Ocr.AllowedPartialStaleness,
			Feeds: map[string]ocr3.FeedValue{
				anyFakeFeedId: {
					Deviation: conf.Ocr.Deviation,
					Heartbeat: conf.Ocr.Heartbeat,
				},
			},
		},
		AggregationMethod: conf.Ocr.AggregationMethod,
		Encoder:           conf.Ocr.Encoder,
		EncoderConfig:     conf.Ocr.EncoderConfig,
		ReportId:          conf.Ocr.ReportId,
	}

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](feedsInput)}
	consensus := ocr3cap.NewConsensus(workflow, "data-feeds-report", ocrInput, ocrConfig)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, conf.TargetChain, input, *conf.ChainWriter)

	return workflow.Spec()
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

	t.Run("maping different types without compute", func(t *testing.T) {
		actual, err := NewWorkflowSpecFromPrimitives(notStreamSepoliaConfig)
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
						Mapping: map[string]any{"observations": []map[string]any{
							{
								"benchmarkPrice":       "$(trigger.outputs.Price.PriceA)",
								"feedId":               anyFakeFeedId,
								"fullReport":           "$(trigger.outputs.FullReport)",
								"observationTimestamp": "$(trigger.outputs.Timestamp)",
								"reportContext":        "$(trigger.outputs.ReportContext)",
								"signatures":           "$(trigger.outputs.Signatures)",
							},
						}},
					},
					Config: map[string]any{
						"aggregation_config": ocr3.ConsensusConfigAggregationConfig{
							AllowedPartialStaleness: "0.5",
							Feeds: map[string]ocr3.FeedValue{
								anyFakeFeedId: {
									Deviation: "0.5",
									Heartbeat: 3600,
								},
							},
						},
						"aggregation_method": "data_feeds",
						"encoder":            "EVM",
						"encoder_config": ocr3.ConsensusConfigEncoderConfig{
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
}

func runSepoliaStagingTest(t *testing.T, config []byte, gen func([]byte) (workflows.WorkflowSpec, error)) {
	testWorkflowSpec, err := gen(config)
	require.NoError(t, err)

	expectedSpecYaml := workflows.WorkflowSpecYaml{}
	require.NoError(t, yaml.Unmarshal(expectedSepolia, &expectedSpecYaml))
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
	AllowedPartialStaleness string                                `json:"allowedPartialStaleness" yaml:"allowedPartialStaleness" mapstructure:"allowedPartialStaleness"`
	Deviation               string                                `json:"deviation" yaml:"deviation" mapstructure:"deviation"`
	Heartbeat               int                                   `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`
	AggregationMethod       ocr3.ConsensusConfigAggregationMethod `json:"aggregation_method" yaml:"aggregation_method" mapstructure:"aggregation_method"`
	Encoder                 ocr3.ConsensusConfigEncoder           `json:"encoder" yaml:"encoder" mapstructure:"encoder"`
	EncoderConfig           ocr3.ConsensusConfigEncoderConfig     `json:"encoder_config" yaml:"encoder_config" mapstructure:"encoder_config"`
	ReportId                string                                `json:"report_id" yaml:"report_id" mapstructure:"report_id"`
}
