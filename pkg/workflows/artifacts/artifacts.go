package artifacts

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"

	workflowUtils "github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

type Input struct {
	WorkflowOwner string
	WorkflowName  string
	WorkflowPath  string
	ConfigPath    string
	BinaryPath    string
}

type Artifacts struct {
	input *Input

	binaryData []byte
	configData []byte
	workflowID string
	log        *slog.Logger
}

// Constructor for WorkflowArtifacts
func NewWorkflowArtifacts(
	input *Input,
	lggr *slog.Logger,
) *Artifacts {
	return &Artifacts{
		input: input,
		log:   lggr,
	}
}

// Reads a file and returns the data
func (a *Artifacts) readFile(fileName string, artifactName string) ([]byte, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	a.log.Info("Read file successfully", "file", fileName, "artifact", artifactName)
	return data, nil
}

// Prepares the workflow artifacts for workflow deployment
func (a *Artifacts) Prepare() error {
	var err error
	a.binaryData, err = a.readFile(a.input.BinaryPath, "binary wasm")
	if err != nil {
		return err
	}
	a.configData, err = a.readFile(a.input.ConfigPath, "config")
	if err != nil {
		return err
	}
	// Note: the binary data read from file is base64 encoded, so we need to decode it before generating the workflow ID.
	binaryDataDecoded, err := base64.StdEncoding.DecodeString(string(a.binaryData))
	if err != nil {
		return fmt.Errorf("failed to decode base64 binary data: %w", err)
	}
	a.workflowID, err = workflowUtils.GenerateWorkflowIDFromStrings(a.input.WorkflowOwner,
		a.input.WorkflowName, binaryDataDecoded, a.configData, "")
	if err != nil {
		return fmt.Errorf("failed to generate workflow ID: %w", err)
	}
	return nil
}

// Returns the generated workflow ID after preparing the artifacts
func (a *Artifacts) GetWorkflowID() string {
	return a.workflowID
}
