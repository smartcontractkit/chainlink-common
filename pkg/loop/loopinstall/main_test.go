package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

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
	if len(plugin.AdditionalFiles) != 1 {
		t.Fatalf("Expected 1 additional file, got %d", len(plugin.AdditionalFiles))
	}
}

func TestWriteBuildManifest(t *testing.T) {
	dir, err := os.MkdirTemp("", "build-manifest-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	tasks := []PluginInstallTask{
		{
			PluginType: "test",
			Plugin: PluginDef{
				ModuleURI:   "github.com/example/test",
				GitRef:      "v1.0.0",
				InstallPath: "./cmd/test",
				AdditionalFiles: []AdditionalFile{
					{
						Src:  "/go/pkg/mod/github.com/test@v1.0.0/lib/libtest.so",
						Dest: "/usr/lib/",
					},
				},
			},
		},
	}

	if writeErr := writeBuildManifest(tasks, dir); writeErr != nil {
		t.Fatalf("writeBuildManifest failed: %v", writeErr)
	}

	// Read and verify manifest
	data, err := os.ReadFile(filepath.Join(dir, "build-manifest.json"))
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Verify contents
	if _, ok := manifest.Plugins["test"]; !ok {
		t.Error("Expected test plugin in manifest")
	}

	// Check test plugin
	test := manifest.Plugins["test"]
	if test.ModuleURI != "github.com/example/test" {
		t.Errorf("Expected ModuleURI=%q, got %q", "github.com/example/test", test.ModuleURI)
	}
	if len(test.AdditionalFiles) != 1 {
		t.Fatalf("Expected 1 additional file, got %d", len(test.AdditionalFiles))
	}
	if test.AdditionalFiles[0].Src != "/go/pkg/mod/github.com/test@v1.0.0/lib/libtest.so" {
		t.Errorf("Unexpected additional file source: %s", test.AdditionalFiles[0].Src)
	}
}
