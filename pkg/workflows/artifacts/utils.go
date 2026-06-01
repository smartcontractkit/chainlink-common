package artifacts

import (
	"bufio"
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

// ensureToolFn is the function used by EnsureTool; replaced in tests to mock.
var ensureToolFn = func(bin string) error {
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%q not found in PATH: %w", bin, err)
	}
	return nil
}

// EnsureTool checks that the binary exists on PATH
func EnsureTool(bin string) error {
	return ensureToolFn(bin)
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
		env := append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
		// Pin GOTOOLCHAIN so the compiled WASM is reproducible. Prefer the
		// configured GOTOOLCHAIN; when it's unset, fall back to the version
		// declared in the module's go.mod (the local Go version wouldn't pin).
		if toolchain := goToolchain(rootFolder); toolchain != "" {
			env = append(env, "GOTOOLCHAIN="+toolchain)
		}
		buildCmd.Env = env
	}

	buildCmd.Dir = rootFolder

	return buildCmd
}

// goToolchain returns a GOTOOLCHAIN value (e.g. "go1.26.2") to pin the build
// to. It prefers the configured `go env GOTOOLCHAIN`; when that is unset (e.g.
// "auto") it falls back to the go version declared in the module's go.mod. The
// local Go version is not used as a fallback because it would not pin a
// reproducible toolchain. Returns "" when nothing can be determined.
func goToolchain(dir string) string {
	if v := goEnv(dir, "GOTOOLCHAIN"); v != "" && v != "auto" {
		return v
	}
	return goToolchainFromMod(dir)
}

// goEnv runs `go env <name>` in dir and returns the trimmed value, or "" on error.
func goEnv(dir, name string) string {
	cmd := exec.Command("go", "env", name)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// goToolchainFromMod returns a GOTOOLCHAIN value derived from the go directive
// in the module's go.mod, located via `go env GOMOD`. Returns "" when no go.mod
// or go version can be determined.
func goToolchainFromMod(dir string) string {
	goModPath := goEnv(dir, "GOMOD")
	// `go env GOMOD` returns "" outside a module and os.DevNull when modules
	// are disabled (GO111MODULE=off); neither is a real go.mod file.
	if goModPath == "" || goModPath == os.DevNull {
		return ""
	}
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "go" {
			return "go" + fields[1]
		}
	}
	return ""
}
