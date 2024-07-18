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
	chainwriter "github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chain_writer"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/notstreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

type Config struct {
	Workflow    workflows.NewWorkflowParams
	Streams     *streams.StreamsTriggerConfig
	Ocr         *ocr3.Ocr3ConsensusConfig
	ChainWriter *chainwriter.ChainwriterTargetConfig
}

func NewWorkflowSpec(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := Config{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	streamsTrigger, err := streams.NewStreamsTriggerCapability(workflow, "streams", *conf.Streams)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	streamsList := workflows.ListOf[streams.Feed](streamsTrigger)
	ocrInput := ocr3.Ocr3ConsensusCapabilityInput{Observations: streamsList}

	consensus, err := ocr3.NewOcr3ConsensusCapability(workflow, "data-feeds-report", ocrInput, *conf.Ocr)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	input := chainwriter.ChainwriterTargetCapabilityInput{SignedReport: consensus}

	if err = chainwriter.NewChainwriterTargetCapability(workflow, "chain-writer", input, *conf.ChainWriter); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	return workflow.Spec(), nil
}

// What if there were hundreds of feeds?  Like feeds that aren't for CCIP?

type ModifiedConfig struct {
	Workflow         workflows.NewWorkflowParams
	MaxFrequencyMs   int
	DefaultHeartbeat int
	DefaultDeviation float64
	FeedInfos        []FeedInfo
	ReportId         string
	Encoder          ocr3.Ocr3ConsensusConfigEncoder
	EncoderConfig    ocr3.Ocr3ConsensusConfigEncoderConfig
	ChainWriter      *chainwriter.ChainwriterTargetConfig
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

	streamsConfig := streams.StreamsTriggerConfig{MaxFrequencyMs: conf.MaxFrequencyMs}
	ocr3Config := ocr3.Ocr3ConsensusConfig{
		AggregationMethod: "data_feeds",
		Encoder:           conf.Encoder,
		EncoderConfig:     conf.EncoderConfig,
		ReportId:          conf.ReportId,
	}
	for _, elm := range conf.FeedInfos {
		streamsConfig.FeedIds = append(streamsConfig.FeedIds, elm.FeedId)
		aggConfig := ocr3.Ocr3ConsensusConfigAggregationConfigElem{
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
	streamsTrigger, err := streams.NewStreamsTriggerCapability(workflow, "streams", streamsConfig)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	streamsList := workflows.ListOf[streams.Feed](streamsTrigger)
	ocrInput := ocr3.Ocr3ConsensusCapabilityInput{Observations: streamsList}

	consensus, err := ocr3.NewOcr3ConsensusCapability(workflow, "data-feeds-report", ocrInput, ocr3Config)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	input := chainwriter.ChainwriterTargetCapabilityInput{SignedReport: consensus}

	if err = chainwriter.NewChainwriterTargetCapability(workflow, "chain-writer", input, *conf.ChainWriter); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	return workflow.Spec(), nil
}

// What if inputs and outputs don't match exactly?

func NewWorkflowSpecFromPrimitives(rawConfig []byte) (workflows.WorkflowSpec, error) {
	conf := NotStreamsConfig{}
	if err := yaml.Unmarshal(rawConfig, &conf); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	workflow := workflows.NewWorkflow(conf.Workflow)
	notStreamsTrigger, err := notstreams.NewNotstreamsTriggerCapability(workflow, "notstreams", *conf.Streams)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	feedsInput := streams.NewStreamsTriggerCapabilityFromComponents(
		notStreamsTrigger.Price().PriceA(),
		workflows.ConstantDefinition[streams.FeedId]("0x0000000000000000000000000000000000000000000000000000000000000000"),
		notStreamsTrigger.FullReport(),
		notStreamsTrigger.Timestamp(),
		notStreamsTrigger.ReportContext(),
		notStreamsTrigger.Signatures(),
	)
	notStreamsList := workflows.ListOf[streams.Feed](feedsInput)
	ocrInput := ocr3.Ocr3ConsensusCapabilityInput{Observations: notStreamsList}

	consensus, err := ocr3.NewOcr3ConsensusCapability(workflow, "data-feeds-report", ocrInput, *conf.Ocr)
	if err != nil {
		return workflows.WorkflowSpec{}, err
	}

	input := chainwriter.ChainwriterTargetCapabilityInput{SignedReport: consensus}

	if err = chainwriter.NewChainwriterTargetCapability(workflow, "chain-writer", input, *conf.ChainWriter); err != nil {
		return workflows.WorkflowSpec{}, err
	}

	return workflow.Spec(), nil
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
	Streams     *notstreams.NotstreamsTriggerConfig
	Ocr         *ocr3.Ocr3ConsensusConfig
	ChainWriter *chainwriter.ChainwriterTargetConfig
}
