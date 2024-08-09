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
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/notstreams/notstreamscap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams/streamscap"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
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
	streamsTrigger := streamscap.NewTrigger(workflow, "trigger", *conf.Streams)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](streamsTrigger)}
	consensus := ocr3cap.NewConsensus(workflow, "ccip_feeds", ocrInput, *conf.Ocr)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, conf.TargetChain, input, *conf.ChainWriter)

	return workflow.Spec()
}

// What if there were hundreds of feeds?  Like feeds that aren't for CCIP?

type ModifiedConfig struct {
	Workflow         workflows.NewWorkflowParams
	MaxFrequencyMs   int
	DefaultHeartbeat int
	DefaultDeviation float64
	FeedInfos        []FeedInfo
	ReportId         string
	Encoder          ocr3.ConsensusConfigEncoder
	EncoderConfig    ocr3.ConsensusConfigEncoderConfig
	ChainWriter      *chainwriter.TargetConfig
	TargetChain      string
}

type FeedInfo struct {
	FeedId    streams.FeedId
	Deviation *float64
	Heartbeat *int
}

/*func NewModifiedWorkflowSpec(rawConfig []byte) (workflows.WorkflowSpec, error) {
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
	}
	for _, elm := range conf.FeedInfos {
		streamsConfig.FeedIds = append(streamsConfig.FeedIds, elm.FeedId)
		aggConfig := ocr3.ConsenssConfigAggregationConfigElem{
			FeedId:    elm.FeedId,
			Deviation: conf.DefaultDeviation,
			Heartbeat: conf.DefaultHeartbeat,
		}
		if elm.Deviation != nil {
			aggConfig.Deviation = *elm.Deviation
		}

		if elm.Heartbeat != nil {
			aggConfig.Heartbeat = *elm.Heartbeat
		}

		ocr3Config.AggregationConfig = append(ocr3Config.AggregationConfig, aggConfig)
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	streamsTrigger := streamscap.NewTrigger(workflow, "streams", streamsConfig)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](streamsTrigger)}
	consensus := ocr3cap.NewConsensus(workflow, "consensus", ocrInput, ocr3Config)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, conf.TargetChain, "write to chain", input, *conf.ChainWriter)

	return workflow.Spec()
}*/

// What if inputs and outputs don't match exactly?

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := NotStreamsConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	notStreamsTrigger := notstreamscap.NewTrigger(workflow, "notstreams", *conf.Streams)

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

//go:embed testdata/fixtures/workflows/expected_sepolia.yaml
var expectedSepolia []byte

func TestBuilder_ValidSpec(t *testing.T) {
	testWorkflowSpec, err := NewWorkflowSpec(sepoliaConfig)
	require.NoError(t, err)

	expectedSpecYaml := workflows.WorkflowSpecYaml{}
	require.NoError(t, yaml.Unmarshal(expectedSepolia, &expectedSpecYaml))
	expectedSpec := expectedSpecYaml.ToWorkflowSpec()

	expected, err := json.Marshal(expectedSpec)
	require.NoError(t, err)

	actual, err := json.Marshal(testWorkflowSpec)
	require.NoError(t, err)

	// TODO rtinianov NOW remove this
	if string(expected) != string(actual) {
		os.WriteFile("/Volumes/RAM/expected.txt", expected, 0644)
		os.WriteFile("/Volumes/RAM/actual.txt", actual, 0644)
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
