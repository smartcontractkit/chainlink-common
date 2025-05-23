package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestExpandEnvVars tests the expansion of environment variables in strings
func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		env      map[string]string
		expected string
	}{
		{
			name:     "no variables",
			input:    "hello world",
			env:      nil,
			expected: "hello world",
		},
		{
			name:     "single variable",
			input:    "hello ${USER}",
			env:      map[string]string{"USER": "alice"},
			expected: "hello alice",
		},
		{
			name:     "multiple variables",
			input:    "${GREETING} ${USER}!",
			env:      map[string]string{"GREETING": "hello", "USER": "bob"},
			expected: "hello bob!",
		},
		{
			name:     "undefined variable",
			input:    "hello ${UNDEFINED}",
			env:      map[string]string{},
			expected: "hello ${UNDEFINED}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.env != nil {
				for k, v := range tt.env {
					os.Setenv(k, v)
					defer os.Unsetenv(k)
				}
			}

			got := expandEnvVars(tt.input)
			if got != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestIsPluginEnabled tests the enabled state of plugins
func TestIsPluginEnabled(t *testing.T) {
	trueBool := true
	falseBool := false

	tests := []struct {
		name     string
		plugin   PluginDef
		expected bool
	}{
		{
			name:     "nil enabled flag",
			plugin:   PluginDef{Enabled: nil},
			expected: true,
		},
		{
			name:     "explicitly enabled",
			plugin:   PluginDef{Enabled: &trueBool},
			expected: true,
		},
		{
			name:     "explicitly disabled",
			plugin:   PluginDef{Enabled: &falseBool},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPluginEnabled(tt.plugin); got != tt.expected {
				t.Errorf("isPluginEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestConfigParsing tests the parsing of plugin configuration files
func TestConfigParsing(t *testing.T) {
	data, err := os.ReadFile("testdata/plugins.test.yaml")
	if err != nil {
		t.Fatalf("Failed to read test config: %v", err)
	}

	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse test config: %v", err)
	}

	// Check defaults
	if config.Defaults.GoFlags != "-ldflags=-s" {
		t.Errorf("Expected GoFlags=%q, got %q", "-ldflags=-s", config.Defaults.GoFlags)
	}

	// Check plugins
	plugins, ok := config.Plugins["test-plugin"]
	if !ok {
		t.Fatal("Expected test-plugin in plugins")
	}
	if len(plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(plugins))
	}

	plugin := plugins[0]
	if plugin.ModuleURI != "github.com/example/module" {
		t.Errorf("Expected ModuleURI=%q, got %q", "github.com/example/module", plugin.ModuleURI)
	}
	if len(plugin.Libs) != 1 {
		t.Fatalf("Expected 1 lib path, got %d", len(plugin.Libs))
	}

	expectedLibPath := "/go/pkg/mod/github.com/example/module@v1.0.0/lib/libexample.so"
	if plugin.Libs[0] != expectedLibPath {
		t.Errorf("Expected lib path=%q, got %q", expectedLibPath, plugin.Libs[0])
	}
}

// TestWriteBuildManifest tests the writing of build manifests
func TestWriteBuildManifest(t *testing.T) {
	dir, err := os.MkdirTemp("", "build-manifest-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create tasks with and without libs
	tasks := []PluginInstallTask{
		{
			PluginType: "test-with-libs",
			Plugin: PluginDef{
				ModuleURI:   "github.com/example/test",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/test",
				Libs:        []string{"/go/pkg/mod/github.com/test@v1.0.0/lib/libtest.so"},
			},
			ConfigFile: "config1.yaml",
		},
		{
			PluginType: "test-no-libs",
			Plugin: PluginDef{
				ModuleURI:   "github.com/example/test2",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/test2",
				Libs:        []string{},
			},
			ConfigFile: "config1.yaml",
		},
	}

	outputFile := filepath.Join(dir, "build-manifest.json")
	if writeErr := writeBuildManifest(tasks, outputFile); writeErr != nil {
		t.Fatalf("writeBuildManifest failed: %v", writeErr)
	}

	// Read and verify manifest
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Verify sources map structure
	if len(manifest.Sources) == 0 {
		t.Fatal("Expected at least one source in manifest")
	}

	// Find the config1.yaml source
	source, exists := manifest.Sources["config1.yaml"]
	if !exists {
		// Try with absolute path if relative path failed
		absPath, err := filepath.Abs("config1.yaml")
		if err == nil {
			source, exists = manifest.Sources[absPath]
		}
	}

	if !exists {
		t.Fatal("Failed to find source config1.yaml in manifest")
	}

	// Check plugin with libs
	testWithLibs, exists := source["test-with-libs"]
	if !exists {
		t.Error("Plugin test-with-libs not found in manifest")
	} else {
		if testWithLibs.ModuleURI != "github.com/example/test" {
			t.Errorf("Expected ModuleURI=%q, got %q", "github.com/example/test", testWithLibs.ModuleURI)
		}

		if len(testWithLibs.Libs) != 1 {
			t.Fatalf("Expected 1 lib path, got %d", len(testWithLibs.Libs))
		}

		expectedLibPath := "/go/pkg/mod/github.com/test@v1.0.0/lib/libtest.so"
		if testWithLibs.Libs[0] != expectedLibPath {
			t.Errorf("Expected lib path=%q, got %q", expectedLibPath, testWithLibs.Libs[0])
		}
	}

	// Check plugin without libs
	testNoLibs, exists := source["test-no-libs"]
	if !exists {
		t.Error("Plugin test-no-libs not found in manifest")
	} else {
		if testNoLibs.ModuleURI != "github.com/example/test2" {
			t.Errorf("Expected ModuleURI=%q, got %q", "github.com/example/test2", testNoLibs.ModuleURI)
		}

		// For a plugin with empty libs, the libs field should be omitted or empty in the JSON
		if testNoLibs.Libs != nil && len(testNoLibs.Libs) > 0 {
			t.Errorf("Expected empty libs, got %v", testNoLibs.Libs)
		}
	}
}

// TestFileSpecificDefaults tests that defaults from each YAML file
// are applied only to plugins from that file
func TestFileSpecificDefaults(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "loopinstall-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create first config file with complex ldflags
	file1Content := `
defaults:
  goflags: "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0"
plugins:
  test1:
    - moduleURI: "github.com/test/module1"
      gitRef: "v1.0.0"
      installPath: "github.com/test/module1/cmd/test"
`
	file1Path := filepath.Join(tempDir, "config1.yaml")
	// Use writeErr instead of err to avoid shadowing
	if writeErr := os.WriteFile(file1Path, []byte(file1Content), 0600); writeErr != nil {
		t.Fatalf("Failed to write test file: %v", writeErr)
	}

	// Create second config file with different ldflags
	file2Content := `
defaults:
  goflags: "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Sha=abcdef"
plugins:
  test2:
    - moduleURI: "github.com/test/module2"
      gitRef: "v1.0.0"
      installPath: "github.com/test/module2/cmd/test"
`
	file2Path := filepath.Join(tempDir, "config2.yaml")
	// Use writeErr2 instead of err to avoid shadowing
	if writeErr2 := os.WriteFile(file2Path, []byte(file2Content), 0600); writeErr2 != nil {
		t.Fatalf("Failed to write test file: %v", writeErr2)
	}

	// Create third config file with local installPath and different ldflags
	file3Content := `
defaults:
  goflags: "-custom-flag-for-local-path -X github.com/smartcontractkit/chainlink/v2/core/static.Build=test3"
plugins:
  test3:
    - moduleURI: "github.com/test/module3"
      gitRef: "v1.0.0"
      installPath: "./cmd/localtest" # Explicitly local path
`
	file3Path := filepath.Join(tempDir, "config3.yaml")
	if writeErr3 := os.WriteFile(file3Path, []byte(file3Content), 0600); writeErr3 != nil {
		t.Fatalf("Failed to write test file: %v", writeErr3)
	}

	// Process all files
	tasks1, err := processConfigFile(file1Path, false)
	if err != nil {
		t.Fatalf("Failed to process first config file: %v", err)
	}

	tasks2, err := processConfigFile(file2Path, false)
	if err != nil {
		t.Fatalf("Failed to process second config file: %v", err)
	}

	tasks3, err := processConfigFile(file3Path, false)
	if err != nil {
		t.Fatalf("Failed to process third config file: %v", err)
	}

	// Combine tasks
	allTasks := append(tasks1, tasks2...)
	allTasks = append(allTasks, tasks3...)

	// Verify each task has the correct defaults from its source file
	for _, task := range allTasks {
		switch task.PluginType {
		case "test1":
			expectedFlags := "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0"
			if task.Defaults.GoFlags != expectedFlags {
				t.Errorf("test1 plugin has incorrect goflags: %s, expected: %s", task.Defaults.GoFlags, expectedFlags)
			}
		case "test2":
			expectedFlags := "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Sha=abcdef"
			if task.Defaults.GoFlags != expectedFlags {
				t.Errorf("test2 plugin has incorrect goflags: %s, expected: %s", task.Defaults.GoFlags, expectedFlags)
			}
		case "test3":
			expectedFlags := "-custom-flag-for-local-path -X github.com/smartcontractkit/chainlink/v2/core/static.Build=test3"
			if task.Defaults.GoFlags != expectedFlags {
				t.Errorf("test3 plugin has incorrect goflags: %s, expected: %s", task.Defaults.GoFlags, expectedFlags)
			}
		default:
			t.Errorf("Unexpected plugin type: %s", task.PluginType)
		}
	}

	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Define a struct to hold command call information
	type commandCallInfo struct {
		Args []string
		Dir  string
	}

	// Mock execCommand to avoid actual command execution and provide proper JSON output
	execCommandCalls := []commandCallInfo{}
	execCommand = func(cmd *exec.Cmd) error {
		// Store args and dir for later verification
		execCommandCalls = append(execCommandCalls, commandCallInfo{Args: cmd.Args, Dir: cmd.Dir})

		// For the go mod download -json commands, we need to provide valid JSON output
		// Reconstruct cmdStr here if needed for this specific check, or pass cmd.Args directly
		cmdStrForModDownload := strings.Join(cmd.Args, " ")
		if strings.Contains(cmdStrForModDownload, "go mod download -json") {
			// Extract module path to use in the mocked directory path
			parts := strings.Split(cmdStrForModDownload, " ")
			modulePath := parts[len(parts)-1]

			// Create a fake module directory based on the module name
			moduleDir := filepath.Join(tempDir, "mocked-modules", strings.ReplaceAll(modulePath, "@", "-"))

			// Create a mock response with proper JSON format
			jsonResponse := fmt.Sprintf(`{"Path":"%s","Version":"v1.0.0","Dir":"%s"}`,
				strings.Split(modulePath, "@")[0], moduleDir)

			// Access stdout field of the exec.Cmd struct to write our mocked JSON response
			if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
				stdout.WriteString(jsonResponse)
			}
		}

		return nil
	}

	// Set test output directory and file
	outputFile := filepath.Join(tempDir, "output", "build-manifest.json")

	// Skip actual module download and installation since we're mocking
	// by not executing the real commands, but still call installPlugins
	// to test the file-specific defaults behavior
	if err := installPlugins(allTasks, 1, true, outputFile); err != nil {
		t.Fatalf("Failed to install plugins: %v", err)
	}

	// Verify commands were called with the correct flags - looking for the complex flag values
	foundTest1 := false
	foundTest2 := false
	foundTest3 := false
	for _, call := range execCommandCalls {
		cmdStr := strings.Join(call.Args, " ")
		// Check for the install command specifically
		if strings.Contains(cmdStr, "go install") {
			if strings.Contains(call.Dir, "github.com/test/module1") && // Check against cmd.Dir
				strings.Contains(cmdStr, "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0") &&
				strings.Contains(cmdStr, " ./cmd/test") { // Corrected to expect relative path
				foundTest1 = true
			}
			if strings.Contains(call.Dir, "github.com/test/module2") && // Check against cmd.Dir
				strings.Contains(cmdStr, "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Sha=abcdef") &&
				strings.Contains(cmdStr, " ./cmd/test") { // Corrected to expect relative path
				foundTest2 = true
			}
			if strings.Contains(call.Dir, "github.com/test/module3") && // Check against cmd.Dir for module3
				strings.Contains(cmdStr, "-custom-flag-for-local-path -X github.com/smartcontractkit/chainlink/v2/core/static.Build=test3") && // test3's specific goflags
				strings.Contains(cmdStr, " ./cmd/localtest") { // test3's local installPath
				foundTest3 = true
			}
		}
	}

	if !foundTest1 {
		t.Error("Did not find command with correct flags for test1 module")
	}
	if !foundTest2 {
		t.Error("Did not find command with correct flags for test2 module")
	}
	if !foundTest3 {
		t.Error("Did not find command with correct flags for test3 module")
	}
}

// TestValidateGoFlags tests the validation of Go build flags
func TestValidateGoFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   string
		wantErr bool
	}{
		{
			name:    "simple flags",
			flags:   "-v -ldflags -tags",
			wantErr: false,
		},
		{
			name:    "flags with values",
			flags:   "-ldflags=-s -tags=netgo",
			wantErr: false,
		},
		{
			name:    "complex ldflags",
			flags:   "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0",
			wantErr: false,
		},
		{
			name:    "quoted values",
			flags:   `-ldflags="-s" -tags="netgo osusergo"`,
			wantErr: false,
		},
		{
			name:    "race flag",
			flags:   "-race -ldflags=-s",
			wantErr: false,
		},
		{
			name:    "bench flag",
			flags:   "-bench=. -ldflags=-s",
			wantErr: false,
		},
		{
			name:    "fuzz flag now allowed",
			flags:   "-fuzz=FuzzFunc -ldflags=-s",
			wantErr: false,
		},
		{
			name:    "dangerous character semicolon",
			flags:   "-ldflags=-s; rm -rf /",
			wantErr: true,
		},
		{
			name:    "dangerous command chaining",
			flags:   "-ldflags=-s && echo malicious",
			wantErr: true,
		},
		{
			name:    "dangerous rm command",
			flags:   "-ldflags=-s -X rm -rf /",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGoFlags(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGoFlags(%q) error = %v, wantErr %v", tt.flags, err, tt.wantErr)
			}
		})
	}
}

// TestValidationFunctions tests the input validation functions
func TestValidationFunctions(t *testing.T) {
	tests := []struct {
		name      string
		validator func(string) error
		valid     []string
		invalid   []string
	}{
		{
			name:      "validateModuleURI",
			validator: validateModuleURI,
			valid: []string{
				"github.com/example/module",
				"github.com/example-org/module-name",
				"github.com/example/module/v2",
				"example.com/org/repo",
			},
			invalid: []string{
				"github.com/example/module;rm -rf /",
				"github.com/example/module||echo hack",
				"github.com/example/module\nls",
			},
		},
		{
			name:      "validateGitRef",
			validator: validateGitRef,
			valid: []string{
				"v1.0.0",
				"main",
				"feature/new-feature",
				"123abc",
				"fix-bug.2",
			},
			invalid: []string{
				"v1.0.0;ls",
				"main||cat /etc/passwd",
				"feature/new-feature\nrm",
			},
		},
		{
			name:      "validateInstallPath",
			validator: validateInstallPath,
			valid: []string{
				"./cmd/example",
				"github.com/example/module/cmd/tool",
				"../relative/path",
			},
			invalid: []string{
				"/absolute/path;rm",
				"./cmd/example||echo hack",
				"github.com/example/module\nls",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test valid inputs
			for _, input := range tc.valid {
				if err := tc.validator(input); err != nil {
					t.Errorf("Expected %q to be valid, but got error: %v", input, err)
				}
			}

			// Test invalid inputs
			for _, input := range tc.invalid {
				if err := tc.validator(input); err == nil {
					t.Errorf("Expected %q to be invalid, but no error was returned", input)
				}
			}
		})
	}
}

// TestProcessConfigFile tests the config file parsing
func TestProcessConfigFile(t *testing.T) {
	// Create a temporary config file with a duplicate plugin
	tempDir, err := os.MkdirTemp("", "config-duplicate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Config with duplicate plugin definition
	duplicateContent := `
defaults:
  goflags: "-ldflags=-s"
plugins:
  test:
    - moduleURI: "github.com/example/test"
      gitRef: "v1.0.0"
      installPath: "./cmd/test"
    - moduleURI: "github.com/example/test"  # Same module URI
      gitRef: "v1.1.0"                      # Different git ref
      installPath: "./cmd/test"             # Same install path
`
	duplicateFile := filepath.Join(tempDir, "duplicate.yaml")
	if err := os.WriteFile(duplicateFile, []byte(duplicateContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test the duplicate detection
	_, err = processConfigFile(duplicateFile, false)
	if err == nil {
		t.Error("Expected error for duplicate plugin, but got nil")
	} else if !strings.Contains(err.Error(), "duplicate plugin found") {
		t.Errorf("Expected duplicate plugin error, got: %v", err)
	}

	// Test non-existent file
	_, err = processConfigFile("/nonexistent/file.yaml", false)
	if err == nil {
		t.Error("Expected error for non-existent file, but got nil")
	}

	// Test invalid YAML content
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte("invalid: yaml: content:"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = processConfigFile(invalidFile, false)
	if err == nil {
		t.Error("Expected error for invalid YAML, but got nil")
	}
}

// TestDownloadAndInstallPlugin tests the plugin download and installation
func TestDownloadAndInstallPlugin(t *testing.T) {
	// Save the original execCommand to restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Create a temp directory for test files
	tempDir, err := os.MkdirTemp("", "download-install-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name           string
		plugin         PluginDef
		defaults       DefaultsConfig
		mockDownload   func(*exec.Cmd) error
		mockInstall    func(*exec.Cmd) error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful installation",
			plugin: PluginDef{
				ModuleURI:   "github.com/example/test",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/test",
			},
			defaults: DefaultsConfig{
				GoFlags: "-ldflags=-s",
			},
			mockDownload: func(cmd *exec.Cmd) error {
				if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
					moduleDir := filepath.Join(tempDir, "modules", "github.com", "example", "test")
					stdout.WriteString(fmt.Sprintf(`{"Dir":"%s"}`, moduleDir))
				}
				return nil
			},
			mockInstall: func(cmd *exec.Cmd) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "download failure",
			plugin: PluginDef{
				ModuleURI:   "github.com/example/bad",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/bad",
			},
			defaults: DefaultsConfig{},
			mockDownload: func(cmd *exec.Cmd) error {
				return fmt.Errorf("failed to download module")
			},
			mockInstall: func(cmd *exec.Cmd) error {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to download module",
		},
		{
			name: "install failure",
			plugin: PluginDef{
				ModuleURI:   "github.com/example/test",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/test",
			},
			defaults: DefaultsConfig{},
			mockDownload: func(cmd *exec.Cmd) error {
				if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
					moduleDir := filepath.Join(tempDir, "modules", "github.com", "example", "test")
					stdout.WriteString(fmt.Sprintf(`{"Dir":"%s"}`, moduleDir))
				}
				return nil
			},
			mockInstall: func(cmd *exec.Cmd) error {
				return fmt.Errorf("failed to install plugin")
			},
			expectError:    true,
			expectedErrMsg: "failed to install plugin",
		},
		{
			name: "disabled plugin",
			plugin: PluginDef{
				Enabled:     func() *bool { b := false; return &b }(),
				ModuleURI:   "github.com/example/disabled",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/disabled",
			},
			defaults:    DefaultsConfig{},
			expectError: false,
		},
		{
			name: "with environment variable expansion",
			plugin: PluginDef{
				ModuleURI:   "github.com/${TEST_ORG}/test",
				GitRef:      "${TEST_REF}",
				InstallPath: "./cmd/${TEST_CMD}",
			},
			defaults: DefaultsConfig{},
			mockDownload: func(cmd *exec.Cmd) error {
				cmdLine := strings.Join(cmd.Args, " ")
				if !strings.Contains(cmdLine, "github.com/testorg/test@testref") {
					return fmt.Errorf("environment variables not expanded correctly: %s", cmdLine)
				}
				if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
					moduleDir := filepath.Join(tempDir, "modules", "github.com", "testorg", "test")
					stdout.WriteString(fmt.Sprintf(`{"Dir":"%s"}`, moduleDir))
				}
				return nil
			},
			mockInstall: func(cmd *exec.Cmd) error {
				cmdLine := strings.Join(cmd.Args, " ")
				if len(cmd.Args) < 2 {
					return fmt.Errorf("install command has too few arguments")
				}
				lastArg := cmd.Args[len(cmd.Args)-1]
				// Expect "./cmd/testcmd" after env var expansion and prefixing
				expectedInstallArg := "./cmd/testcmd"
				if lastArg != expectedInstallArg {
					return fmt.Errorf("environment variables not expanded correctly in install path: expected %q, got %q in %s", expectedInstallArg, lastArg, cmdLine)
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "full import path stripping",
			plugin: PluginDef{
				ModuleURI:   "github.com/example/full",
				GitRef:      "v1.0.0",
				InstallPath: "github.com/example/full/cmd/plugin", // Full path that needs stripping
			},
			defaults: DefaultsConfig{},
			mockDownload: func(cmd *exec.Cmd) error {
				if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
					// Derive moduleDir from ModuleURI for consistency
					parts := strings.Split("github.com/example/full", "/")
					moduleDir := filepath.Join(append([]string{tempDir, "modules"}, parts...)...)
					stdout.WriteString(fmt.Sprintf(`{"Dir":"%s"}`, moduleDir))
				}
				return nil
			},
			mockInstall: func(cmd *exec.Cmd) error {
				if len(cmd.Args) < 2 {
					return fmt.Errorf("install command has too few arguments: %v", cmd.Args)
				}
				packageArg := cmd.Args[len(cmd.Args)-1]
				expectedPackage := "./cmd/plugin" // Expected after stripping, cleaning, and prefixing with ./
				if packageArg != expectedPackage {
					return fmt.Errorf("expected install package %q, got %q", expectedPackage, packageArg)
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "install path is module URI",
			plugin: PluginDef{
				ModuleURI:   "github.com/example/rootinstall",
				GitRef:      "v1.0.0",
				InstallPath: "github.com/example/rootinstall", // Same as ModuleURI
			},
			defaults: DefaultsConfig{},
			mockDownload: func(cmd *exec.Cmd) error {
				if stdout, ok := cmd.Stdout.(*bytes.Buffer); ok {
					parts := strings.Split("github.com/example/rootinstall", "/")
					moduleDir := filepath.Join(append([]string{tempDir, "modules"}, parts...)...)
					stdout.WriteString(fmt.Sprintf(`{"Dir":"%s"}`, moduleDir))
				}
				return nil
			},
			mockInstall: func(cmd *exec.Cmd) error {
				if len(cmd.Args) < 2 {
					return fmt.Errorf("install command has too few arguments: %v", cmd.Args)
				}
				packageArg := cmd.Args[len(cmd.Args)-1]
				expectedPackage := "." // Expected when InstallPath is the same as ModuleURI
				if packageArg != expectedPackage {
					return fmt.Errorf("expected install package %q, got %q", expectedPackage, packageArg)
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables for the test with env var expansion
			if tc.name == "with environment variable expansion" {
				os.Setenv("TEST_ORG", "testorg")
				os.Setenv("TEST_REF", "testref")
				os.Setenv("TEST_CMD", "testcmd")
				defer func() {
					os.Unsetenv("TEST_ORG")
					os.Unsetenv("TEST_REF")
					os.Unsetenv("TEST_CMD")
				}()
			}

			// Mock the command execution
			var downloadCalled, installCalled bool
			execCommand = func(cmd *exec.Cmd) error {
				cmdLine := strings.Join(cmd.Args, " ")

				if strings.Contains(cmdLine, "go mod download") {
					downloadCalled = true
					if tc.mockDownload != nil {
						return tc.mockDownload(cmd)
					}
				} else if strings.Contains(cmdLine, "go install") {
					installCalled = true
					if tc.mockInstall != nil {
						// For the environment variable test case, check the install path differently
						if tc.name == "with environment variable expansion" {
							// The actual install command will have the full installPath, not just cmd/testcmd
							// Fix the check by examining the last argument instead
							if len(cmd.Args) < 2 {
								return fmt.Errorf("install command has too few arguments")
							}
							lastArg := cmd.Args[len(cmd.Args)-1]
							// Expect "./cmd/testcmd" after env var expansion and prefixing
							expectedInstallArg := "./cmd/testcmd"
							if lastArg != expectedInstallArg {
								return fmt.Errorf("environment variables not expanded correctly in install path: expected %q, got %q in %s", expectedInstallArg, lastArg, cmdLine)
							}
							return nil
						}
						return tc.mockInstall(cmd)
					}
				}
				return nil
			}

			// Call the function
			err := downloadAndInstallPlugin("test", 0, tc.plugin, tc.defaults)

			// Check results
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tc.expectedErrMsg != "" && !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got %q", tc.expectedErrMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Verify commands called (except for disabled plugins)
			if !tc.expectError && tc.plugin.Enabled == nil || (tc.plugin.Enabled != nil && *tc.plugin.Enabled) {
				if !downloadCalled && tc.mockDownload != nil {
					t.Error("Download command was not called")
				}
				if !installCalled && tc.mockInstall != nil {
					t.Error("Install command was not called")
				}
			}
		})
	}
}

// TestSetupOutputFile tests the setupOutputFile function
func TestSetupOutputFile(t *testing.T) {
	// Test with relative path
	relPath := "test-output.json"
	absPath, err := setupOutputFile(relPath)
	if err != nil {
		t.Errorf("setupOutputFile with relative path failed: %v", err)
	}
	if !filepath.IsAbs(absPath) {
		t.Errorf("Expected absolute path, got %s", absPath)
	}

	// Test with absolute path
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	origAbsPath := filepath.Join(wd, "test-abs-output.json")
	resultPath, err := setupOutputFile(origAbsPath)
	if err != nil {
		t.Errorf("setupOutputFile with absolute path failed: %v", err)
	}
	if resultPath != origAbsPath {
		t.Errorf("Expected path to be unchanged, got %s instead of %s", resultPath, origAbsPath)
	}
}
