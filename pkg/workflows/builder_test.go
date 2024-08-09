package workflows_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

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

// What if inputs and outputs don't match exactly?

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := NotStreamsConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	notStreamsTrigger := notstreamscap.NewTrigger(workflow, *conf.Streams)

	feedsInput := streamscap.NewTriggerFromFields(
		notStreamsTrigger.Price().PriceA(),
		workflows.ConstantDefinition[streams.FeedId]("0x0000000000000000000000000000000000000000000000000000000000000000"),
		notStreamsTrigger.FullReport(),
		notStreamsTrigger.Timestamp(),
		notStreamsTrigger.ReportContext(),
		notStreamsTrigger.Signatures(),
	)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](feedsInput)}
	consensus := ocr3cap.NewConsensus(workflow, "data-feeds-report", ocrInput, *conf.Ocr)

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

func TestBuilder_ValidSpec(t *testing.T) {
	t.Run("basic config", func(t *testing.T) {
		runSepoliaStagingTest(t, sepoliaConfig, NewWorkflowSpec)
	})

	t.Run("remapping config", func(t *testing.T) {
		runSepoliaStagingTest(t, sepoliaDefaultConfig, NewWorkflowRemapped)
	})
}

func runSepoliaStagingTest(t *testing.T, config []byte, gen func([]byte) (workflows.WorkflowSpec, error)) {
	testWorkflowSpec, err := gen(config)
	require.NoError(t, err)

	expectedSpecYaml := workflows.WorkflowSpecYaml{}
	require.NoError(t, yaml.Unmarshal(expectedSepolia, &expectedSpecYaml))
	expectedSpec := expectedSpecYaml.ToWorkflowSpec()

	expected, err := json.Marshal(expectedSpec)
	require.NoError(t, err)

	actual, err := json.Marshal(testWorkflowSpec)
	require.NoError(t, err)

	// TODO rtinianov REMOVE BEFORE COMMIT
	if string(expected) != string(actual) {
		os.WriteFile("/Volumes/RAM/expected.json", expected, 0644)
		os.WriteFile("/Volumes/RAM/actual.json", actual, 0644)
	}

	assert.Equal(t, string(expected), string(actual))
}

type NotStreamsConfig struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *notstreams.TriggerConfig
	Ocr         *ocr3.ConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}
