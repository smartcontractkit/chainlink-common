package main

import (
	"encoding/hex"
	"flag"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/artifacts"
)

// Command line interface for workflow artifacts package
// Workflow folder should contain a main.go file and a config.yaml file.
/* go run . --workflow-folder /path/to/my/workflow/folder \
       --workflow-owner 0xABc1234567890defABc1234567890DEFaBc12345 \
	   --workflow-name wf-test-1
*/

// Parse CLI args
func parseArgs() *artifacts.Input {
	workflowOwner := flag.String("workflow-owner", "", "the owner of the workflow")
	workflowName := flag.String("workflow-name", "", "the name of the workflow")
	workflowFolderPath := flag.String("workflow-folder", "", "the path to the workflow folder")
	flag.Parse()

	// Validate input args
	if *workflowOwner == "" || *workflowName == "" || *workflowFolderPath == "" {
		log.Fatalf("workflow-owner, workflow-name, and workflow-folder-path are required")
	}
	if fi, err := os.Stat(*workflowFolderPath); os.IsNotExist(err) || !fi.IsDir() {
		log.Fatalf("workflow folder path does not exist: %s", *workflowFolderPath)
	}
	workflowMainFile := filepath.Join(*workflowFolderPath, "")
	workflowConfigFile := filepath.Join(*workflowFolderPath, "config.yaml")
	binaryPath := filepath.Join(*workflowFolderPath, "binary.wasm.br.b64")
	if _, err := os.Stat(workflowMainFile); os.IsNotExist(err) {
		log.Fatalf("workflow main file does not exist: %s", workflowMainFile)
	}
	if _, err := os.Stat(workflowConfigFile); os.IsNotExist(err) {
		log.Fatalf("workflow config file does not exist: %s", workflowConfigFile)
	}

	if !isValidEVMAddr(*workflowOwner) {
		log.Fatalf("workflow owner must be a valid Ethereum address")
	}

	return &artifacts.Input{
		WorkflowOwner: *workflowOwner,
		WorkflowName:  *workflowName,
		WorkflowPath:  workflowMainFile,
		ConfigPath:    workflowConfigFile,
		BinaryPath:    binaryPath,
	}
}

func isValidEVMAddr(addr string) bool {
	if len(addr) != 42 || !strings.HasPrefix(addr, "0x") {
		return false
	}
	addr = strings.TrimPrefix(addr, "0x")
	if _, err := hex.DecodeString(addr); err != nil {
		return false
	}
	return true
}

// Sample CLI to test out workflow artifacts pkg functionality
func main() {
	input := parseArgs()
	lggr := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lggr.Info("Parsing input arguments", "input", input)
	artifacts := artifacts.NewWorkflowArtifacts(input, lggr)
	if err := artifacts.Compile(); err != nil {
		log.Fatalf("failed to compile workflow: %s", err)
	}
	if err := artifacts.Prepare(); err != nil {
		log.Fatalf("failed to prepare artifacts: %s", err)
	}
	lggr.Info("Artifacts prepared successfully", "workflow ID", artifacts.GetWorkflowID())
}
