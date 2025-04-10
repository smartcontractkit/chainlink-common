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
	if config.Defaults.GoFlags != "-ldflags=-s -w" {
		t.Errorf("Expected GoFlags=%q, got %q", "-ldflags=-s -w", config.Defaults.GoFlags)
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
  goflags: "-ldflags=-s -w -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0"
  goprivate: "github.com/test/private1"
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
  goprivate: "github.com/test/private2"
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

	// Process both files
	tasks1, err := processConfigFile(file1Path, false)
	if err != nil {
		t.Fatalf("Failed to process first config file: %v", err)
	}

	tasks2, err := processConfigFile(file2Path, false)
	if err != nil {
		t.Fatalf("Failed to process second config file: %v", err)
	}

	// Combine tasks
	allTasks := append(tasks1, tasks2...)

	// Verify each task has the correct defaults from its source file
	for _, task := range allTasks {
		switch task.PluginType {
		case "test1":
			expectedFlags := "-ldflags=-s -w -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0"
			if task.Defaults.GoFlags != expectedFlags {
				t.Errorf("test1 plugin has incorrect goflags: %s, expected: %s", task.Defaults.GoFlags, expectedFlags)
			}
			if task.Defaults.GoPrivate != "github.com/test/private1" {
				t.Errorf("test1 plugin has incorrect goprivate: %s", task.Defaults.GoPrivate)
			}
		case "test2":
			expectedFlags := "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Sha=abcdef"
			if task.Defaults.GoFlags != expectedFlags {
				t.Errorf("test2 plugin has incorrect goflags: %s, expected: %s", task.Defaults.GoFlags, expectedFlags)
			}
			if task.Defaults.GoPrivate != "github.com/test/private2" {
				t.Errorf("test2 plugin has incorrect goprivate: %s", task.Defaults.GoPrivate)
			}
		default:
			t.Errorf("Unexpected plugin type: %s", task.PluginType)
		}
	}

	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Mock execCommand to avoid actual command execution and provide proper JSON output
	execCommandCalls := []string{}
	execCommand = func(cmd *exec.Cmd) error {
		cmdStr := strings.Join(cmd.Args, " ")
		execCommandCalls = append(execCommandCalls, cmdStr)

		// For the go mod download -json commands, we need to provide valid JSON output
		if strings.Contains(cmdStr, "go mod download -json") {
			// Extract module path to use in the mocked directory path
			parts := strings.Split(cmdStr, " ")
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
	for _, cmdStr := range execCommandCalls {
		if strings.Contains(cmdStr, "github.com/test/module1") &&
			strings.Contains(cmdStr, "-ldflags=-s -w -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0") {
			foundTest1 = true
		}
		if strings.Contains(cmdStr, "github.com/test/module2") &&
			strings.Contains(cmdStr, "-ldflags=-s -X github.com/smartcontractkit/chainlink/v2/core/static.Sha=abcdef") {
			foundTest2 = true
		}
	}

	if !foundTest1 {
		t.Error("Did not find command with correct flags for test1 module")
	}
	if !foundTest2 {
		t.Error("Did not find command with correct flags for test2 module")
	}
}

// Replace TestExtractFlagNames with TestValidateGoFlags
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
			flags:   "-ldflags=-s -w -tags=netgo",
			wantErr: false,
		},
		{
			name:    "complex ldflags",
			flags:   "-ldflags=-s -w -X github.com/smartcontractkit/chainlink/v2/core/static.Version=1.0.0",
			wantErr: false,
		},
		{
			name:    "quoted values",
			flags:   `-ldflags="-s -w" -tags="netgo osusergo"`,
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
