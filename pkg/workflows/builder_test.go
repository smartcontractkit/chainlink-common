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
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams/streamscap"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

type Config struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *streams.TriggerConfig
	Ocr         *ocr3.ConsensusConfig
	ChainWriter *chainwriter.TargetConfig
}

func NewWorkflowSpec(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := Config{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	streamsTrigger := streamscap.NewTrigger(workflow, "streams", *conf.Streams)

	ocrInput := ocr3cap.ConsensusInput{Observations: workflows.ListOf[streams.Feed](streamsTrigger)}
	consensus := ocr3cap.NewConsensus(workflow, "data-feeds-report", ocrInput, *conf.Ocr)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, "chain-writer", input, *conf.ChainWriter)

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
}

type FeedInfo struct {
	FeedId    streams.FeedId
	Deviation *float64
	Heartbeat *int
}

func NewModifiedWorkflowSpec(rawConfig []byte) (workflows.WorkflowSpec, error) {
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
		aggConfig := ocr3.ConsensusConfigAggregationConfigElem{
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
	consensus := ocr3cap.NewConsensus(workflow, "data-feeds-report", ocrInput, ocr3Config)

	input := chainwritercap.TargetInput{SignedReport: consensus}
	chainwritercap.NewTarget(workflow, "chain-writer", input, *conf.ChainWriter)

	return workflow.Spec()
}

// What if inputs and outputs don't match exactly?

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := NotStreamsConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	notStreamsTrigger := notstreams.NewNotstreamsTriggerCapability(workflow, "notstreams", *conf.Streams)

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
	chainwritercap.NewTarget(workflow, "chain-writer", input, *conf.ChainWriter)

	return workflow.Spec()
}

//go:embed testdata/fixtures/workflows/sepolia.yaml
var sepolia string

func TestBuilder_ValidSpec(t *testing.T) {
	testWorkflowSpec, err := NewWorkflowSpec([]byte(sepolia))
	require.NoError(t, err)

	expectedSpec := workflows.WorkflowSpec{
		Name:  "ccipethsep",
		Owner: "00000000000000000000000000000000000000aa",
		Triggers: []workflows.StepDefinition{
			{
				ID:  "streams-trigger@1.0.0",
				Ref: "streams",
				Inputs: workflows.StepInputs{
					Mapping:   map[string]any{},
					OutputRef: "",
				},
				Config: map[string]interface{}{
					"feedIds": []string{
						"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd",
						"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d",
						"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12",
					},
					"maxFrequencyMs": 100,
				},
				CapabilityType: capabilities.CapabilityTypeTrigger,
			},
		},
		Consensus: []workflows.StepDefinition{
			{
				ID:  "offchain_reporting@1.0.0",
				Ref: "data-feeds-report",
				Inputs: workflows.StepInputs{
					Mapping: map[string]any{
						"observations": "$(streams.outputs)",
					},
				},
				Config: map[string]interface{}{
					"report_id":          "0001",
					"aggregation_method": "data_feeds",
					"aggregation_config": map[string]interface{}{
						"0x0003fbba4fce42f65d6032b18aee53efdf526cc734ad296cb57565979d883bdd": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
						"0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
						"0x0003da6ab44ea9296674d80fe2b041738189103d6b4ea9a4d34e2f891fa93d12": map[string]interface{}{
							"deviation": "0.05",
							"heartbeat": 3600,
						},
					},
					"encoder": "EVM",
					"encoder_config": map[string]interface{}{
						"abi": "(bytes32 FeedID, uint224 Price, uint32 Timestamp)[] Reports",
					},
				},
				CapabilityType: capabilities.CapabilityTypeConsensus,
			},
		},
		Targets: []workflows.StepDefinition{
			{
				ID:  "",
				Ref: "chain-writer",
				Inputs: workflows.StepInputs{
					Mapping: map[string]any{"Err": "$(offchain_reporting.outputs.Err)", "Value": "$(offchain_reporting.outputs.Value)", "WorkflowExecutionID": "$(offchain_reporting.outputs.WorkflowExecutionID)"},
				},
				Config: map[string]any{
					"Address":    "0x1234567890123456789012345678901234567890",
					"DeltaStage": "5s",
					"Schedule":   "oneAtATime",
				},
				CapabilityType: capabilities.CapabilityTypeTarget,
			},
		},
	}

	expected, err := json.Marshal(expectedSpec)
	require.NoError(t, err)

	actual, err := json.Marshal(testWorkflowSpec)
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}

type NotStreamsConfig struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *notstreams.TriggerConfig
	Ocr         *ocr3.ConsensusConfig
	ChainWriter *chainwriter.TargetConfig
}
