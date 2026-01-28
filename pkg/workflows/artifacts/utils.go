package artifacts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	WorkflowLanguageGolang     = "golang"
	WorkflowLanguageTypeScript = "typescript"
)

// There is only a small group of acceptable file extensions by this tool and only few of them are considered to be binary files
func IsBinaryFile(fileName string) (bool, error) {
	// this is binary wasm file (additional .br extension if it's compressed by Brotli)
	if strings.HasSuffix(fileName, ".wasm.br") ||
		strings.HasSuffix(fileName, ".wasm") {
		return true, nil
		// this is a configuration or secrets file
	} else if strings.HasSuffix(fileName, ".yaml") ||
		strings.HasSuffix(fileName, ".yml") ||
		strings.HasSuffix(fileName, ".json") {
		return false, nil
	}
	return false, fmt.Errorf("file extension not supported by the tool: %s, supported extensions: .wasm.br, .json, .yaml, .yml", fileName)
}

// GetWorkflowLanguage determines the workflow language based on the file extension
// Note: inputFile can be a file path (e.g., "main.ts" or "main.go") or a directory (for Go workflows, e.g., ".")
// Returns constants.WorkflowLanguageTypeScript for .ts or .tsx files, constants.WorkflowLanguageGolang otherwise
func GetWorkflowLanguage(inputFile string) string {
	if strings.HasSuffix(inputFile, ".ts") || strings.HasSuffix(inputFile, ".tsx") {
		return WorkflowLanguageTypeScript
	}
	return WorkflowLanguageGolang
}

// EnsureTool checks that the binary exists on PATH
func EnsureTool(bin string) error {
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%q not found in PATH: %w", bin, err)
	}
	return nil
}

// Gets a build command for either Golang or Typescript based on the filename
func GetBuildCmd(inputFile string, outputFile string, rootFolder string) *exec.Cmd {
	isTypescriptWorkflow := strings.HasSuffix(inputFile, ".ts") || strings.HasSuffix(inputFile, ".tsx")

	var buildCmd *exec.Cmd
	if isTypescriptWorkflow {
		buildCmd = exec.Command(
			"bun",
			"cre-compile",
			inputFile,
			outputFile,
		)
	} else {
		// The build command for reproducible and trimmed binaries.
		// -trimpath removes all file system paths from the compiled binary.
		// -ldflags="-buildid= -w -s" further reduces the binary size:
		//   -buildid= removes the build ID, ensuring reproducibility.
		//   -w disables DWARF debugging information.
		//   -s removes the symbol table.
		buildCmd = exec.Command(
			"go",
			"build",
			"-o", outputFile,
			"-trimpath",
			"-ldflags=-buildid= -w -s",
			"-buildvcs=false",
			inputFile,
		)
		buildCmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	}

	buildCmd.Dir = rootFolder

	return buildCmd
}
