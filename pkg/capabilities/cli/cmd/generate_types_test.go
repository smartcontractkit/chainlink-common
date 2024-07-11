package cmd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

const TriggerSchemaFilePath = "testdata/streams_trigger.json"
const ActionSchemaFilePath = "pkg/capabilities/cli/cmd/testdata/read_chain_action.json"
const ConsensusSchemaFilePath = "pkg/capabilities/cli/cmd/testdata/ocr3_consensus.json"
const TargetSchemaFilePath = "pkg/capabilities/cli/cmd/testdata/write_chain_target.json"

func Test_CapabilitySchemaFilePattern(t *testing.T) {
	matches := cmd.CapabilitySchemaFilePattern.FindStringSubmatch(TriggerSchemaFilePath)
	assert.Equal(t, "streams", matches[1])
	assert.Equal(t, "trigger", matches[2])

	matches = cmd.CapabilitySchemaFilePattern.FindStringSubmatch(ActionSchemaFilePath)
	assert.Equal(t, "read_chain", matches[1])
	assert.Equal(t, "action", matches[2])

	matches = cmd.CapabilitySchemaFilePattern.FindStringSubmatch(ConsensusSchemaFilePath)
	assert.Equal(t, "ocr3", matches[1])
	assert.Equal(t, "consensus", matches[2])

	matches = cmd.CapabilitySchemaFilePattern.FindStringSubmatch(TargetSchemaFilePath)
	assert.Equal(t, "write_chain", matches[1])
	assert.Equal(t, "target", matches[2])
}

func Test_TypesFromJSONSchema(t *testing.T) {
	schemaFilePath := "testdata/streams_trigger.json"
	expectedOutputFilePath := "testdata/streams_trigger_generated.go"

	expectedOutputFileContents, err := os.ReadFile(expectedOutputFilePath)
	assert.NoError(t, err)

	generatedFilepath, generatedContents, rootType, capabilityType, err := cmd.TypesFromJSONSchema(schemaFilePath)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutputFilePath, generatedFilepath, "Generated file path does not match expected file path")
	assert.Equal(t, string(expectedOutputFileContents), generatedContents, "Generated file contents do not match expected contents")
	assert.Equal(t, capabilities.CapabilityTypeTrigger.String(), capabilityType, "Wrong type for capability")
	assert.Equal(t, "StreamsTrigger", rootType, "Root type does not match expected root type")
	assert.Equal(t, "trigger", capabilityType, "Capability type does not match expected capability type")
}
