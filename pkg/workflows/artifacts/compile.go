package artifacts

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
)

var defaultBinaryPath = "./binary.wasm.br.b64"

// Function to compile CRE workflow into a WASM binary, brotli compress it and base64 encode it
// Input.WorkflowPath is the path to the workflow directory or file. If it's a directory, it
// will be scanned for a main workflow file either in .go or .ts/.tsx files for CRE Workflow
// Runner execution.
func (a *Artifacts) Compile() error {
	a.log.Info("Compiling workflow", "workflow path", a.input.WorkflowPath)

	workflowAbsFile, err := filepath.Abs(a.input.WorkflowPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for the workflow file: %w", err)
	}

	if _, err := os.Stat(workflowAbsFile); os.IsNotExist(err) {
		return fmt.Errorf("workflow file not found: %s", workflowAbsFile)
	}

	workflowMainFilePath, err := a.GetWorkflowMainFile(a.input.WorkflowPath)
	if err != nil {
		return fmt.Errorf("failed to get workflow main file: %w", err)
	}
	workflowRootFolder := filepath.Dir(workflowMainFilePath)
	workflowMainFile := filepath.Base(workflowMainFilePath)
	tmpWasmFileName := "tmp.wasm"
	if a.input.BinaryPath == "" {
		a.input.BinaryPath = filepath.Join(workflowRootFolder, defaultBinaryPath)
	}
	if !strings.HasSuffix(a.input.BinaryPath, ".b64") {
		if !strings.HasSuffix(a.input.BinaryPath, ".br") {
			if !strings.HasSuffix(a.input.BinaryPath, ".wasm") {
				a.input.BinaryPath += ".wasm" // Append ".wasm" if it doesn't already end with ".wasm"
			}
			a.input.BinaryPath += ".br" // Append ".br" if it doesn't already end with ".br"
		}
		a.input.BinaryPath += ".b64" // Append ".b64" if it doesn't already end with ".b64"
	}
	a.log.Info("Workflow details", "main_file", workflowMainFilePath,
		"parent_folder", workflowRootFolder,
		"binary_path", a.input.BinaryPath)

	// Set language based on workflow file extension
	workflowLanguage := GetWorkflowLanguage(workflowMainFile)
	switch workflowLanguage {
	case WorkflowLanguageTypeScript:
		if err := EnsureTool("bun"); err != nil {
			return errors.New("bun is required for TypeScript workflows but was not found in PATH" +
				"; install from https://bun.com/docs/installation")
		}
	case WorkflowLanguageGolang:
		if err := EnsureTool("go"); err != nil {
			return errors.New("go toolchain is required for Go workflows but was not found in PATH" +
				"; install from https://go.dev/dl")
		}
	default:
		return fmt.Errorf("unsupported workflow language for file %s", workflowMainFile)
	}

	buildCmd := GetBuildCmd(".", tmpWasmFileName, workflowRootFolder)
	a.log.Info("Executing workflow build command", "workflow directory", buildCmd.Dir, "command", buildCmd.String())

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		a.log.Error("Failed to compile workflow", "error", err, "build output", string(buildOutput))
		out := strings.TrimSpace(string(buildOutput))
		return fmt.Errorf("failed to compile workflow: %w\nbuild output:\n%s", err, out)
	}
	a.log.Info("Workflow compiled successfully", "build_output", buildOutput)

	tmpWasmFilePath := filepath.Join(workflowRootFolder, tmpWasmFileName)
	compressedWasmBytes, err := brotliCompressFile(tmpWasmFilePath)
	if err != nil {
		return fmt.Errorf("failed to compress WASM binary: %w", err)
	}
	a.log.Info("WASM binary compressed", "compressed_file_size", len(compressedWasmBytes))

	if err = a.b64EncodeAndWriteFile(compressedWasmBytes, a.input.BinaryPath); err != nil {
		return fmt.Errorf("failed to base64 encode the WASM binary: %w", err)
	}

	if err = os.Remove(tmpWasmFilePath); err != nil {
		return fmt.Errorf("failed to remove the temporary file:  %w", err)
	}

	return nil
}

// Read WASM file content from temporary workflow binary file and compress it using
// Brotli compression algorithm.
func brotliCompressFile(tmpWasmFilePath string) ([]byte, error) {
	wasmFileBytes, err := os.ReadFile(tmpWasmFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow binary: %w", err)
	}
	var buffer bytes.Buffer

	// Compress using Brotli with default options
	writer := brotli.NewWriter(&buffer)

	_, err = writer.Write(wasmFileBytes)
	if err != nil {
		return nil, err
	}

	// must close it to flush the writer and ensure all data is stored to the buffer
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Encode the input bytes to base64 and write them to the output file.
func (a *Artifacts) b64EncodeAndWriteFile(input []byte, outputFile string) error {
	encoded := base64.StdEncoding.EncodeToString(input)
	a.log.Debug("WASM binary encoded", "encoded_file_size", len(encoded))
	err := os.WriteFile(outputFile, []byte(encoded), 0666) //nolint:gosec
	if err != nil {
		return err
	}
	return nil
}

// Get the path to main workflow file for a Go workflow
func (a *Artifacts) GetWorkflowMainFile(workflowPath string) (string, error) {
	f, err := os.Stat(workflowPath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info for the workflow path: %w", err)
	}
	// Direct file passed in workflow path, just return the path
	if !f.IsDir() {
		return workflowPath, nil
	}

	// Directory passed in workflow path, find the main Go/TSX file
	a.log.Debug("Scanning directory for main workflow file", "directory", workflowPath)
	mainWorkflowFile, err := ScanFilesForContent(workflowPath)
	if err != nil {
		return "", err
	}
	a.log.Debug("Found main workflow file", "file", mainWorkflowFile)
	return mainWorkflowFile, nil
}

// ScanFilesForContent scans all files in the specified directory (non-recursive)
// and returns the name of the first file whose content contains all the specified substrings.
// Returns an error if no file matches all substrings or if there's an error reading the directory.
func ScanFilesForContent(dirPath string) (string, error) {
	goSubstrings := []string{
		"github.com/smartcontractkit/cre-sdk-go/cre/wasm",
		"wasm.NewRunner",
	}
	tsSubstrings := []string{
		"@chainlink/cre-sdk",
		"Runner.newRunner<",
	}
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, file := range files {
		allowedSourceFile := strings.HasSuffix(file.Name(), ".go") ||
			strings.HasSuffix(file.Name(), ".ts") ||
			strings.HasSuffix(file.Name(), ".tsx")
		if file.IsDir() || !allowedSourceFile {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read, continue to next file
			continue
		}

		substrings := goSubstrings
		if strings.HasSuffix(file.Name(), ".ts") || strings.HasSuffix(file.Name(), ".tsx") {
			substrings = tsSubstrings
		}

		contentStr := string(content)
		matchesAll := true
		for _, substr := range substrings {
			if !strings.Contains(contentStr, substr) {
				matchesAll = false
				break
			}
		}

		if matchesAll {
			return filepath.Join(dirPath, file.Name()), nil
		}
	}

	return "", fmt.Errorf("no ts or go workflow file found in directory %s", dirPath)
}
