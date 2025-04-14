package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PluginConfig represents the structure of our plugin configuration files
type PluginConfig struct {
	Defaults DefaultsConfig         `yaml:"defaults"`
	Plugins  map[string][]PluginDef `yaml:"plugins"`
}

// DefaultsConfig holds the default configuration values
type DefaultsConfig struct {
	GoFlags   string `yaml:"goflags"`
	GoPrivate string `yaml:"goprivate"`
}

// PluginDef defines a single plugin instance
type PluginDef struct {
	Enabled     *bool    `yaml:"enabled,omitempty"`
	ModuleURI   string   `yaml:"moduleURI"`
	GitRef      string   `yaml:"gitRef"`
	InstallPath string   `yaml:"installPath"`
	Libs        []string `yaml:"libs"`
}

// ModDownloadResult represents the JSON directory (dir) output from 'go mod download -json'
type ModDownloadResult struct {
	Dir string `json:"Dir"`
}

// PluginInstallTask represents a plugin to be installed
type PluginInstallTask struct {
	PluginType string
	Plugin     PluginDef
	Defaults   DefaultsConfig
	ConfigFile string
}

// PluginInstallResult represents the result of a plugin installation
type PluginInstallResult struct {
	PluginType string
	Plugin     PluginDef
	Error      error
}

// DefaultsSource tracks which file has provided default values
type DefaultsSource struct {
	GoFlagsSource   string
	GoPrivateSource string
}

// BuildManifest represents the complete build information
type BuildManifest struct {
	BuildTime string                               `json:"buildTime"`
	Sources   map[string]map[string]PluginManifest `json:"sources"`
}

// PluginManifest represents a single plugin's build information
type PluginManifest struct {
	ModuleURI   string   `json:"moduleURI"`
	GitRef      string   `json:"gitRef"`
	InstallPath string   `json:"installPath"`
	Libs        []string `json:"libs,omitempty"`
}

// pluginKey creates a unique key for a plugin to detect duplicates
func pluginKey(pluginType string, plugin PluginDef) string {
	return fmt.Sprintf("%s:%s:%s", pluginType, plugin.ModuleURI, plugin.InstallPath)
}

// expandEnvVars replaces environment variables in a string
func expandEnvVars(s string) string {
	if !strings.Contains(s, "${") {
		return s
	}

	// Simple environment variable expansion
	result := s
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key, val := parts[0], parts[1]
			placeholder := "${" + key + "}"
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}
	return result
}

// isPluginEnabled checks if a plugin is enabled (defaults to true if not explicitly set to false)
func isPluginEnabled(plugin PluginDef) bool {
	return plugin.Enabled == nil || *plugin.Enabled
}

// validateModuleURI ensures the module URI follows Go module conventions
func validateModuleURI(uri string) error {
	// Check for valid Go module path format
	if !regexp.MustCompile(`^[\w./-]+(/[\w./-]+)*$`).MatchString(uri) {
		return fmt.Errorf("invalid module URI format: %s", uri)
	}
	return nil
}

// validateGitRef ensures the git reference is safe
func validateGitRef(ref string) error {
	// Git refs should only contain alphanumeric, dot, dash, slash, underscore
	if !regexp.MustCompile(`^[\w./-]+$`).MatchString(ref) {
		return fmt.Errorf("invalid git reference: %s", ref)
	}
	return nil
}

// validateInstallPath ensures the install path is safe
func validateInstallPath(path string) error {
	// Validate as Go import path
	if !regexp.MustCompile(`^[.\w/-]+(/[\w./-]+)*$`).MatchString(path) {
		return fmt.Errorf("invalid install path: %s", path)
	}
	return nil
}

// validateGoFlags ensures flags are safe and prevents command injection
func validateGoFlags(flags string) error {
	// Check for potentially dangerous characters that could enable command injection
	dangerousPatterns := []string{
		";", "&&", "||", "`", "$", "|", "<", ">", "#", "//",
		"shutdown", "reboot", "rm -", "format", "mkfs", "dd",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(flags, pattern) {
			return fmt.Errorf("potentially unsafe pattern in go flags: %s", pattern)
		}
	}

	return nil
}

// execCommand is a function variable that can be replaced in tests
var execCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// downloadAndInstallPlugin downloads and installs a single plugin
func downloadAndInstallPlugin(pluginType string, pluginIdx int, plugin PluginDef, defaults DefaultsConfig) error {
	if !isPluginEnabled(plugin) {
		log.Printf("Skipping disabled plugin %s[%d]", pluginType, pluginIdx)
		return nil
	}

	// Handle GOPRIVATE environment variable
	origGoPrivate := os.Getenv("GOPRIVATE")
	envGoPrivate := origGoPrivate

	// Set GOPRIVATE based on the defaults and existing environment variable
	if defaults.GoPrivate != "" || envGoPrivate != "" {
		// Case 1: defaults.GoPrivate exists and envGoPrivate exists - append them
		// Case 2: defaults.GoPrivate exists but no envGoPrivate - use defaults
		// Case 3: defaults.GoPrivate empty but envGoPrivate exists - use environment value
		goPrivate := defaults.GoPrivate

		if goPrivate != "" && envGoPrivate != "" {
			// Append with comma separator
			goPrivate = goPrivate + "," + envGoPrivate
		} else if goPrivate == "" && envGoPrivate != "" {
			// Just use the environment variable
			goPrivate = envGoPrivate
		}

		os.Setenv("GOPRIVATE", goPrivate)
		defer os.Setenv("GOPRIVATE", origGoPrivate)
	}

	// Expand environment variables
	moduleURI := expandEnvVars(plugin.ModuleURI)
	gitRef := expandEnvVars(plugin.GitRef)
	installPath := expandEnvVars(plugin.InstallPath)

	// Validate inputs
	if err := validateModuleURI(moduleURI); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if gitRef != "" {
		if err := validateGitRef(gitRef); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	if err := validateInstallPath(installPath); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Clean the install path
	cleanInstallPath := filepath.Clean(installPath)

	// Full module path with git reference
	fullModulePath := moduleURI
	if gitRef != "" {
		fullModulePath = fmt.Sprintf("%s@%s", moduleURI, gitRef)
	}

	log.Printf("Installing plugin %s[%d] from %s", pluginType, pluginIdx, fullModulePath)

	// Download the module and get its directory
	var moduleDir string
	{
		cmd := exec.Command("go", "mod", "download", "-json", fullModulePath)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr

		if err := execCommand(cmd); err != nil {
			return fmt.Errorf("failed to download module %s: %w", fullModulePath, err)
		}

		var result ModDownloadResult
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			return fmt.Errorf("failed to parse go mod download output: %w", err)
		}

		moduleDir = result.Dir
		if moduleDir == "" {
			return fmt.Errorf("empty module directory returned for %s", fullModulePath)
		}
	}

	// Build goflags
	goflags := defaults.GoFlags
	if envGoFlags := os.Getenv("CL_PLUGIN_GOFLAGS"); envGoFlags != "" {
		goflags = envGoFlags
	}

	// Validate goflags
	if goflags != "" {
		if err := validateGoFlags(goflags); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Install the plugin
	{
		args := []string{"install"}
		if goflags != "" {
			args = append(args, strings.Fields(goflags)...)
		}

		// Add the install path
		args = append(args, filepath.Join(cleanInstallPath))

		cmd := exec.Command("go", args...)
		cmd.Dir = moduleDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("Running install command: go %s (in directory: %s)", strings.Join(args, " "), moduleDir)

		if err := execCommand(cmd); err != nil {
			return fmt.Errorf("failed to install plugin %s[%d]: %w", pluginType, pluginIdx, err)
		}
	}

	return nil
}

// writeBuildManifest writes installation artifacts to the specified file
func writeBuildManifest(tasks []PluginInstallTask, outputFile string) error {
	manifest := BuildManifest{
		BuildTime: time.Now().UTC().Format(time.RFC3339),
		Sources:   make(map[string]map[string]PluginManifest),
	}

	// Group tasks by source file
	for _, task := range tasks {
		// Get absolute path for the config file
		configPath := task.ConfigFile
		if !filepath.IsAbs(configPath) {
			// Get absolute path if it's not already absolute
			absPath, err := filepath.Abs(configPath)
			if err == nil {
				configPath = absPath
			}
		}

		// Initialize the map for this source file if it doesn't exist
		if _, ok := manifest.Sources[configPath]; !ok {
			manifest.Sources[configPath] = make(map[string]PluginManifest)
		}

		// Create the plugin manifest with the libs field
		pluginManifest := PluginManifest{
			ModuleURI:   task.Plugin.ModuleURI,
			GitRef:      task.Plugin.GitRef,
			InstallPath: task.Plugin.InstallPath,
		}

		// Only add libs if there are any
		if len(task.Plugin.Libs) > 0 {
			pluginManifest.Libs = task.Plugin.Libs
		}

		// Add the plugin to the appropriate source file's map
		manifest.Sources[configPath][task.PluginType] = pluginManifest
	}

	// Ensure directory exists for output file
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for output file: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal installation artifacts: %w", err)
	}

	// Change file permission from 0644 to 0600 to fix G306
	if err := os.WriteFile(outputFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write installation artifacts: %w", err)
	}

	log.Printf("Wrote installation artifacts to %s", outputFile)
	return nil
}

// printHelp prints usage and help information
func printHelp() {
	fmt.Println("Chainlink Plugin Installer")
	fmt.Println("\nA tool for installing Chainlink plugins from YAML configuration files.")
	fmt.Println("\nUsage:")
	fmt.Println("  loopinstall [options] <plugin-config-file> [<plugin-config-file>...]")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help                Show this help message")
	fmt.Println("  -v, --verbose             Enable verbose output")
	fmt.Println("  -c, --concurrency <num>   Set maximum number of concurrent installations (default: 5)")
	fmt.Println("  -s, --sequential          Install plugins sequentially (no concurrency)")
	fmt.Println("  -o, --output-installation-artifacts <file>  Path for installation artifacts JSON file")
	fmt.Println("                             (optional, no installation artifacts written if not specified)")
	fmt.Println("\nExamples:")
	fmt.Println("  # Install plugins from the default configuration")
	fmt.Println("  loopinstall plugins.default.yaml")
	fmt.Println("")
	fmt.Println("  # Install plugins with custom installation artifacts filename")
	fmt.Println("  loopinstall -o ./installation-artifacts.json plugins.default.yaml")
	fmt.Println("")
	fmt.Println("  # Install plugins sequentially")
	fmt.Println("  loopinstall -s plugins.default.yaml")
	fmt.Println("")
	fmt.Println("  # Install plugins with environment variable overrides")
	fmt.Println("  CL_PLUGIN_GOFLAGS=\"-ldflags='-s'\" loopinstall plugins.default.yaml")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  CL_PLUGIN_GOFLAGS  Override the goflags option from the configuration")
	fmt.Println("")
	fmt.Println("Plugin Configuration Format:")
	fmt.Println("  defaults:")
	fmt.Println("    goflags: \"\"     # Default Go build flags")
	fmt.Println("    goprivate: \"\"   # GOPRIVATE setting for private repos")
	fmt.Println("")
	fmt.Println("  plugins:")
	fmt.Println("    plugin-type:")
	fmt.Println("      - moduleURI: \"github.com/example/module\"")
	fmt.Println("        gitRef: \"v1.0.0\"")
	fmt.Println("        installPath: \"./cmd/example\"")
	fmt.Println("        libs: [\"/path/to/libs/*.so\"]  # Optional library paths (can include glob patterns)")
	fmt.Println("        # enabled: false  # Optional, defaults to true if omitted")
}

// processConfigFile parses a plugin configuration file and collects installation tasks
func processConfigFile(configFile string, verbose bool) ([]PluginInstallTask, error) {
	log.Printf("Processing plugin configuration file: %s", configFile)

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	var tasks []PluginInstallTask
	seenPlugins := make(map[string]bool)

	// Process each plugin type
	for pluginType, plugins := range config.Plugins {
		for _, plugin := range plugins {
			key := pluginKey(pluginType, plugin)

			// Check for duplicates within this file
			if seenPlugins[key] {
				return nil, fmt.Errorf("duplicate plugin found: %s in file %s", key, configFile)
			}
			seenPlugins[key] = true

			if verbose {
				log.Printf("Found plugin %s, enabled: %v", key, isPluginEnabled(plugin))
			}

			// Add to installation tasks if enabled
			if isPluginEnabled(plugin) {
				tasks = append(tasks, PluginInstallTask{
					PluginType: pluginType,
					Plugin:     plugin,
					Defaults:   config.Defaults,
					ConfigFile: configFile,
				})
			}
		}
	}

	return tasks, nil
}

// installPlugins installs plugins concurrently using worker pool pattern
func installPlugins(tasks []PluginInstallTask, concurrency int, verbose bool, outputFile string) error {
	if len(tasks) == 0 {
		log.Println("No enabled plugins found to install")
		return nil
	}

	log.Printf("Installing %d plugins with concurrency %d", len(tasks), concurrency)

	// Write installation artifacts if outputFile is specified
	if outputFile != "" {
		if err := writeBuildManifest(tasks, outputFile); err != nil {
			return fmt.Errorf("failed to write installation artifacts: %w", err)
		}
	}

	// Create a channel for tasks and results
	taskCh := make(chan PluginInstallTask, len(tasks))
	resultCh := make(chan PluginInstallResult, len(tasks))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				if verbose {
					log.Printf("Worker %d: Installing plugin %s", workerID, task.PluginType)
				}

				start := time.Now()
				err := downloadAndInstallPlugin(task.PluginType, 0, task.Plugin, task.Defaults)
				duration := time.Since(start)

				if err != nil {
					log.Printf("Worker %d: Failed to install %s in %v: %v",
						workerID, task.PluginType, duration, err)
				} else if verbose {
					log.Printf("Worker %d: Successfully installed %s in %v",
						workerID, task.PluginType, duration)
				}

				resultCh <- PluginInstallResult{
					PluginType: task.PluginType,
					Plugin:     task.Plugin,
					Error:      err,
				}
			}
		}(i)
	}

	// Feed tasks to workers
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var hasErrors bool
	resultMap := make(map[string]error)

	for result := range resultCh {
		key := pluginKey(result.PluginType, result.Plugin)
		resultMap[key] = result.Error
		if result.Error != nil {
			hasErrors = true
		}
	}

	// Report results
	if hasErrors {
		log.Println("Some plugin installations failed:")
		for key, err := range resultMap {
			if err != nil {
				log.Printf("- %s: %v", key, err)
			}
		}
		return errors.New("some plugin installations failed")
	}

	log.Println("All plugins installed successfully")
	if outputFile != "" {
		log.Printf("installation artifacts saved to: %s", outputFile)
	}
	return nil
}

// setupOutputFile ensures the output directory exists
func setupOutputFile(outputFile string) (string, error) {
	// Ensure output path is absolute
	if !filepath.IsAbs(outputFile) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		outputFile = filepath.Join(wd, outputFile)
	}

	return outputFile, nil
}

func main() {
	var showHelp bool
	var verbose bool
	var concurrency int
	var sequential bool
	var outputFile string

	// Define flags
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help (shorthand)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.IntVar(&concurrency, "concurrency", 5, "Maximum number of concurrent installations")
	flag.IntVar(&concurrency, "c", 5, "Maximum number of concurrent installations (shorthand)")
	flag.BoolVar(&sequential, "sequential", false, "Install plugins sequentially")
	flag.BoolVar(&sequential, "s", false, "Install plugins sequentially (shorthand)")
	flag.StringVar(&outputFile, "output-installation-artifacts", "", "Path for installation artifacts JSON file (optional)")
	flag.StringVar(&outputFile, "o", "", "Path for installation artifacts JSON file (optional, shorthand)")

	// Parse flags
	flag.Parse()

	// Show help if requested or no arguments provided
	if showHelp || flag.NArg() == 0 {
		printHelp()
		if !showHelp {
			os.Exit(1)
		}
		return
	}

	// Validate concurrency
	if concurrency < 1 {
		log.Fatal("Concurrency must be at least 1")
	}

	// If sequential mode is enabled, set concurrency to 1
	if sequential {
		concurrency = 1
		if verbose {
			log.Println("Running in sequential mode, concurrency set to 1")
		}
	}

	// Setup output file only if it's specified
	if outputFile != "" {
		var err error
		outputFile, err = setupOutputFile(outputFile)
		if err != nil {
			log.Fatalf("Failed to setup output file: %v", err)
		}
	}

	// Track plugins to detect duplicates across files
	seenPlugins := make(map[string]string)
	var allTasks []PluginInstallTask

	// Process each config file
	for _, configFile := range flag.Args() {
		tasks, err := processConfigFile(configFile, verbose)
		if err != nil {
			log.Fatalf("Failed to process config file %s: %v", configFile, err)
		}

		// Check for duplicates across files
		for _, task := range tasks {
			key := pluginKey(task.PluginType, task.Plugin)
			if prevFile, exists := seenPlugins[key]; exists {
				log.Fatalf("Duplicate plugin found: %s in files %s and %s",
					key, prevFile, configFile)
			}
			seenPlugins[key] = configFile
		}

		allTasks = append(allTasks, tasks...)
	}

	// Install all plugins
	if err := installPlugins(allTasks, concurrency, verbose, outputFile); err != nil {
		os.Exit(1)
	}
}
