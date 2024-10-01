package sdk_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testdata/fixtures/capabilities/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-types --dir $GOFILE

// Note that the set of tests in this file cover the conversion from existing YAML -> code
// along with testing the structure of what is generated from the builders.
// This implicitly tests the code generators functionally, as the generated code is used in the tests.

type Config struct {
	Workflow    sdk.NewWorkflowParams
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

	workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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
	Workflow                sdk.NewWorkflowParams
	AllowedPartialStaleness string
	MaxFrequencyMs          uint64
	DefaultHeartbeat        uint64        `yaml:"default_heartbeat" json:"default_heartbeat"`
	DefaultDeviation        string        `yaml:"default_deviation" json:"default_deviation"`
	FeedInfo                []FeedInfo    `yaml:"feed_info" json:"feed_info"`
	ReportID                ocr3.ReportId `yaml:"report_id" json:"report_id"`
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

	workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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

	workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
	notStreamsTrigger := conf.NotStream.New(workflow)

	md := streams.NewSignersMetadataFromFields(
		sdk.ConstantDefinition(int64(1)), sdk.ListOf(notStreamsTrigger.Metadata().Signer()))

	payload := streams.NewFeedReportFromFields(
		notStreamsTrigger.Payload().BuyPrice(),
		sdk.ConstantDefinition[streams.FeedId](anyFakeFeedID),
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

		testutils.AssertWorkflowSpec(t, notStreamSepoliaWorkflowSpec, actual)
	})

	t.Run("duplicate names causes errors", func(t *testing.T) {
		conf, err := UnmarshalYaml[Config](sepoliaConfig)
		require.NoError(t, err)

		workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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

		workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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

		workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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

		workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
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
	Workflow    sdk.NewWorkflowParams
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
}

func UnmarshalYaml[T any](raw []byte) (*T, error) {
	var v T
	err := yaml.Unmarshal(raw, &v)
	return &v, err
}
