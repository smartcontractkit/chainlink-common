package cmd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

// 1. Input filepath
// 2. Output filepath
// Test that the contents produced by the function matches the contents of 2.

func Test_TypesFromJSONSchema(t *testing.T) {
	schemaFilePath := "testdata/streams.capability.json"
	expectedOutputFilePath := "testdata/streams.generated.go"

	expectedOutputFileContents, err := os.ReadFile(expectedOutputFilePath)
	assert.NoError(t, err)

	generatedFilepath, generatedContents, err := cmd.TypesFromJSONSchema(schemaFilePath)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutputFilePath, generatedFilepath, "Generated file path does not match expected file path")
	assert.Equal(t, expectedOutputFileContents, generatedContents, "Generated file contents do not match expected contents")

}
