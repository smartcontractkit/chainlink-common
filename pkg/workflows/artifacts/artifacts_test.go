package artifacts

import (
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/sha3"
)

type ArtifactsTestSuite struct {
	suite.Suite

	lggr *slog.Logger
}

func (s *ArtifactsTestSuite) SetupSuite() {
	s.lggr = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ArtifactsTestSuite) TearDownSuite() {
	// Cleanup code
}

func (s *ArtifactsTestSuite) TestArtifacts() {
	artifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config.yaml",
		BinaryPath:    "testdata/binary",
	}, s.lggr)
	err := artifacts.Compile()
	s.NoError(err, "failed to compile workflow")

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	b64EncodedBinaryData := artifacts.GetBinaryData()
	s.NotEmpty(b64EncodedBinaryData, "binary data should not be empty")

	s.lggr.Info("WorkflowCompiledBinary Size", "size", len(b64EncodedBinaryData))

	// Compare the keccak256 hash of the binary data with the keccak256 hash of the
	// base64 encoded binary from CRE-CLI
	expKeccak256Hash, err := hex.DecodeString("7385ae61173b2886a12250b508e2d361af0f6a40b6d0dda153fd4c20cb7e0c10")
	s.NoError(err, "failed to decode expected keccak256 hash")
	keccak256FromSha3Lib := sha3.NewLegacyKeccak256()
	keccak256FromSha3Lib.Write(b64EncodedBinaryData)
	actualKeccak256Hash := keccak256FromSha3Lib.Sum(nil)
	s.Equal(actualKeccak256Hash, expKeccak256Hash)

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	base64EncodedBinaryData := artifacts.GetBinaryData()
	// Compare if the compiled WASM binary is the same as the CRE-CLI output
	s.Equal(563960, len(base64EncodedBinaryData), "binary data size should be same as CRE-CLI output")
	s.Equal("m6kIGWtQvQYAALADgIPzAADQt3oerhGwcQAA8PV/uALY1AHwSGF6AACo9vMAqKqqqqqqqqqqqmsSAsIx",
		string(base64EncodedBinaryData[0:80]))
	s.Equal("iBgWTlBKb+iQflnikISk/ALWkBmXnTIkLiWJYdvkGOrbpWD+9ybhB5ISI+JHMO5tcviErMSoqk5p1CY=",
		string(base64EncodedBinaryData[len(base64EncodedBinaryData)-80:]))

	s.Equal("myContract: 0x44DD9D24349965E5e20E3D6118F560BCd64828E9\nchainID: 11155111", string(artifacts.GetConfigData()))

	s.Equal("008619548c29a2ed3eee5f904dc34305191e23e22559d788272b0d4587d776ef", artifacts.GetWorkflowID())
}

func (s *ArtifactsTestSuite) TestArtifactsSadPaths() {
	missingBinaryArtifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config.yaml",
		BinaryPath:    "testdata/binary",
	}, s.lggr)
	err := missingBinaryArtifacts.Prepare()
	s.ErrorContains(err, "no such file or directory", "failed to prepare artifacts")

	invalidBinaryArtifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config.yaml",
		BinaryPath:    "./testdata/main.go", // Not valid base64 encoded
	}, s.lggr)
	err = invalidBinaryArtifacts.Prepare()
	s.ErrorContains(err, "failed to decode base64 binary data", "failed to prepare artifacts")

	invalidConfigArtifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config2.yaml", // Not valid yaml
		BinaryPath:    "testdata/main.go",
	}, s.lggr)
	err = invalidConfigArtifacts.Prepare()
	s.ErrorContains(err, "no such file or directory", "failed to prepare artifacts")

	invalidWorkflowPathArtifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main2.go",
		ConfigPath:    "./testdata/config.yaml",
	}, s.lggr)
	err = invalidWorkflowPathArtifacts.Compile()
	s.ErrorContains(err, "workflow file not found", "failed to prepare artifacts")

	invalidBinaryFolderArtifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config.yaml",
		BinaryPath:    "testdata2/binary",
	}, s.lggr)
	err = invalidBinaryFolderArtifacts.Compile()
	s.ErrorContains(err, "failed to base64 encode the WASM binary: open", "failed to compile workflow")
}

func (s *ArtifactsTestSuite) TestScanFilesForContent() {
	// Create temporary testworkflow directory
	testDir := "testworkflow"
	err := os.MkdirAll(testDir, 0755)
	s.NoError(err, "failed to create testworkflow directory")

	// Ensure cleanup at the end
	defer func() {
		err := os.RemoveAll(testDir)
		s.NoError(err, "failed to remove testworkflow directory")
	}()

	// Create workflow.go file
	workflowGoContent := `//go:build wasip1

package main

import (
	"log/slog"

	"github.com/smartcontractkit/cre-sdk-go/cre"
	"github.com/smartcontractkit/cre-sdk-go/cre/wasm"

	wfcommon "github.com/smartcontractkit/cre-workflow-utils"
)

func main() {
	r := wasm.NewRunner(wfcommon.ParseWorkflowConfig)
	r.Run(func(cfg *wfcommon.Config, _ *slog.Logger, _ cre.SecretsProvider) (cre.Workflow[*wfcommon.Config], error) {
		// Reuse common initializer with the operation-status specific handler.
		return wfcommon.InitEventListenerWorkflow(cfg, wf.OnLog)
	})
}
`
	workflowGoPath := filepath.Join(testDir, "workflow.go")
	err = os.WriteFile(workflowGoPath, []byte(workflowGoContent), 0644)
	s.NoError(err, "failed to create workflow.go file")

	// Create utils.go file
	utilsGoContent := `package workflow

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	gethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	httpcap "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http"
	"github.com/smartcontractkit/cre-sdk-go/cre"

	wfcommon "github.com/smartcontractkit/cre-workflow-utils"
)

// OnLog reads log events
func OnLog(cfg *wfcommon.Config, rt cre.Runtime, payload *evm.Log) (string, error) {
	// Log receipt of trigger similar to operation-writer, but minimal
	rt.Logger().Info("STATUS: received log trigger",
		"blockNumber", payload.BlockNumber,
		// avoid logging raw bytes or full content unless necessary debug
		"txHash_short", func() string {
			if len(payload.TxHash) >= 4 {
				return hex.EncodeToString(payload.TxHash[:4]) + "..."
			}
			return "invalid"
		}(),
	)
}
`
	utilsGoPath := filepath.Join(testDir, "utils.go")
	err = os.WriteFile(utilsGoPath, []byte(utilsGoContent), 0644)
	s.NoError(err, "failed to create utils.go file")

	// Test ScanFilesForContent
	artifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  testDir,
		ConfigPath:    "",
	}, s.lggr)
	foundPath, err := artifacts.GetWorkflowMainFile(testDir)
	s.NoError(err, "ScanFilesForContent should not return an error")
	s.NotEmpty(foundPath, "ScanFilesForContent should return a file path")

	// Verify that the found path is the workflow.go file
	expectedPath, err := filepath.Abs(workflowGoPath)
	s.NoError(err, "failed to get absolute path for workflow.go")
	actualPath, err := filepath.Abs(foundPath)
	s.NoError(err, "failed to get absolute path for found file")
	s.Equal(expectedPath, actualPath, "ScanFilesForContent should identify workflow.go correctly")
}

func (s *ArtifactsTestSuite) TestIsBinaryFile() {
	// Test binary file extensions - should return true, nil
	testCases := []struct {
		name     string
		fileName string
		expected bool
		hasError bool
	}{
		// Binary files - should return true, nil
		{
			name:     "wasm.br file",
			fileName: "binary.wasm.br",
			expected: true,
			hasError: false,
		},
		{
			name:     "wasm file",
			fileName: "binary.wasm",
			expected: true,
			hasError: false,
		},
		{
			name:     "wasm.br with path",
			fileName: "/path/to/binary.wasm.br",
			expected: true,
			hasError: false,
		},
		{
			name:     "wasm with path",
			fileName: "./output/binary.wasm",
			expected: true,
			hasError: false,
		},
		// Non-binary files - should return false, nil
		{
			name:     "yaml file",
			fileName: "config.yaml",
			expected: false,
			hasError: false,
		},
		{
			name:     "yml file",
			fileName: "config.yml",
			expected: false,
			hasError: false,
		},
		{
			name:     "json file",
			fileName: "config.json",
			expected: false,
			hasError: false,
		},
		{
			name:     "yaml with path",
			fileName: "/path/to/config.yaml",
			expected: false,
			hasError: false,
		},
		{
			name:     "yml with path",
			fileName: "./testdata/config.yml",
			expected: false,
			hasError: false,
		},
		{
			name:     "json with path",
			fileName: "secrets/config.json",
			expected: false,
			hasError: false,
		},
		// Unsupported extensions - should return false, error
		{
			name:     "go file",
			fileName: "main.go",
			expected: false,
			hasError: true,
		},
		{
			name:     "txt file",
			fileName: "readme.txt",
			expected: false,
			hasError: true,
		},
		{
			name:     "no extension",
			fileName: "binary",
			expected: false,
			hasError: true,
		},
		{
			name:     "unsupported extension",
			fileName: "file.xml",
			expected: false,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			isBinary, err := IsBinaryFile(tc.fileName)
			s.Equal(tc.expected, isBinary, "IsBinaryFile should return correct boolean value for %s", tc.fileName)
			if tc.hasError {
				s.Error(err, "IsBinaryFile should return an error for unsupported extension: %s", tc.fileName)
				s.Contains(err.Error(), "file extension not supported", "error message should mention unsupported extension")
			} else {
				s.NoError(err, "IsBinaryFile should not return an error for supported extension: %s", tc.fileName)
			}
		})
	}
}

func TestArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactsTestSuite))
}
