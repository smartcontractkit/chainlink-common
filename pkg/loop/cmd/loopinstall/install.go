package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	shellwords "github.com/mattn/go-shellwords"
)

// execCommand is a function variable that can be replaced in tests
var execCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// mergeOrReplaceEnvVars merges new environment variables into an existing slice,
// replacing any existing variables with the same key
func mergeOrReplaceEnvVars(existing []string, newVars []string) []string {
	result := make([]string, len(existing))
	copy(result, existing)

	for _, newVar := range newVars {
		parts := strings.SplitN(newVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]

		// Find and replace if it exists
		found := false
		for i, existingVar := range result {
			if strings.HasPrefix(existingVar, key+"=") {
				result[i] = newVar
				found = true
				break
			}
		}

		// Append if not found
		if !found {
			result = append(result, newVar)
		}
	}

	return result
}

func determineModuleDirectory(goPrivate, fullModulePath string) (string, error) {
	cmd := exec.Command("go", "mod", "download", "-json", fullModulePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if goPrivate != "" {
		// Inherit the current environment and override GOPRIVATE.
		// Note: Not really sure why this is needed - tried to simplify existing logic
		cmd.Env = append(os.Environ(), "GOPRIVATE="+goPrivate)
	}

	if err := execCommand(cmd); err != nil {
		return "", fmt.Errorf("failed to download module %s: %w", fullModulePath, err)
	}

	var result ModDownloadResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return "", fmt.Errorf("failed to parse go mod download output: %w", err)
	}

	if result.Dir == "" {
		return "", fmt.Errorf("empty module directory returned for %s", fullModulePath)
	}

	return result.Dir, nil
}

func determineGoFlags(defaultGoFlags, pluginGoFlags string) ([]string, error) {
	var flags []string
	parser := shellwords.NewParser()

	// Determine base flags
	if envGoFlags := os.Getenv("CL_PLUGIN_GOFLAGS"); envGoFlags != "" {
		log.Printf("Overriding config's default goflags with CL_PLUGIN_GOFLAGS env var: %s", envGoFlags)
		f, err := parser.Parse(envGoFlags)
		if err != nil {
			return nil, err
		}
		flags = f
	} else if defaultGoFlags != "" {
		f, err := parser.Parse(defaultGoFlags)
		if err != nil {
			return nil, err
		}
		flags = f
	}

	// Append plugin-specific flags
	if pluginGoFlags != "" {
		f, err := parser.Parse(pluginGoFlags)
		if err != nil {
			return nil, err
		}
		flags = append(flags, f...)
	}

	// Validate
	if len(flags) > 0 {
		if err := validateGoFlags(strings.Join(flags, " ")); err != nil {
			return nil, err
		}
	}

	return flags, nil
}

func determineInstallArg(installPath, moduleURI string) string {
	// Determine the actual argument for 'go install' based on installPath and moduleURI.
	// - installPath is the user-provided path from YAML (no environment variable expansion).
	// - moduleURI is the URI of the module being downloaded and installed (no environment variable expansion).
	// The 'go install' command will be run with cmd.Dir set to the root of the downloaded moduleURI.
	// Therefore, installArg must be "." or a path starting with "./" relative to the module root.

	// Case 1: installPath is the moduleURI itself. Install the module root.
	if installPath == moduleURI {
		return "."
	}

	// Case 2: installPath is a sub-package of moduleURI (e.g., "moduleURI/cmd/plugin").
	if after, ok := strings.CutPrefix(installPath, moduleURI+"/"); ok {
		// Extract the relative path and prefix with "./".
		relativePath := after
		cleanedRelativePath := strings.TrimLeft(relativePath, "/")   // Handles "moduleURI///subpath"
		if cleanedRelativePath == "" || cleanedRelativePath == "." { // Handles "moduleURI/" or "moduleURI/."
			return "."
		}

		// cleanedRelativePath is like "cmd/plugin" or "sub/../pkg". Prepend "./".
		return "./" + cleanedRelativePath
	}

	// Case 3: installPath is not moduleURI and not a sub-package of moduleURI.
	// Assumed to be:
	//  a) A path already relative to the module root (e.g., "cmd/plugin", "./cmd/plugin", ".").
	//  b) A full path to a different module (e.g., "github.com/other/mod").
	//     For (b), prefixing with "./" when cmd.Dir is set is problematic but replicates prior behavior if any.

	// Simple case
	if installPath == "." {
		return "."
	}

	// Already correctly formatted (e.g., "./cmd/plugin", "./sub/../pkg")
	if strings.HasPrefix(installPath, "./") {
		return installPath
	}

	// Needs "./" prefix. Handles "cmd/plugin", "/cmd/plugin", "github.com/other/mod".
	return "./" + strings.TrimLeft(installPath, "/")
}

// downloadAndInstallPlugin downloads and installs a single plugin
func downloadAndInstallPlugin(pluginType string, pluginIdx int, plugin PluginDef, defaults DefaultsConfig) error {
	if !isPluginEnabled(plugin) {
		log.Printf("Skipping disabled plugin %s[%d]", pluginType, pluginIdx)
		return nil
	}

	// Validate inputs
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("plugin input validation failed: %w", err)
	}

	moduleURI := plugin.ModuleURI
	gitRef := plugin.GitRef
	installPath := plugin.InstallPath

	// Full module path with git reference
	fullModulePath := moduleURI
	if gitRef != "" {
		fullModulePath = fmt.Sprintf("%s@%s", moduleURI, gitRef)
	}

	log.Printf("Installing plugin %s[%d] from %s", pluginType, pluginIdx, fullModulePath)

	// Get GOPRIVATE environment variable
	goPrivate := os.Getenv("GOPRIVATE")

	// Download the module and get its directory
	moduleDir, err := determineModuleDirectory(goPrivate, fullModulePath)
	if err != nil {
		return fmt.Errorf("failed to determine module directory: %w", err)
	}

	// Build env vars from defaults, environment variable, and plugin-specific settings
	envVars := defaults.EnvVars
	if envEnvVars := os.Getenv("CL_PLUGIN_ENVVARS"); envEnvVars != "" {
		envVars = mergeOrReplaceEnvVars(envVars, strings.Fields(envEnvVars))
	}

	// Merge plugin-specific env vars
	if len(plugin.EnvVars) != 0 {
		envVars = mergeOrReplaceEnvVars(envVars, plugin.EnvVars)
	}

	// Install the plugin
	{
		installArg := determineInstallArg(installPath, moduleURI)

		binaryName := filepath.Base(installArg)
		if binaryName == "." {
			binaryName = filepath.Base(moduleURI)
		}

		// Determine output directory
		outputDir := os.Getenv("GOBIN")
		if outputDir == "" {
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				gopath = filepath.Join(os.Getenv("HOME"), "go")
			}
			outputDir = filepath.Join(gopath, "bin")
		}

		outputPath := filepath.Join(outputDir, binaryName)

		// Build goflags
		goflags, err := determineGoFlags(defaults.GoFlags, plugin.Flags)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		args := []string{"build", "-o", outputPath}
		if len(goflags) != 0 {
			args = append(args, goflags...)
		}
		args = append(args, installArg)

		cmd := exec.Command("go", args...)
		cmd.Dir = moduleDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Start with all current environment variables
		cmd.Env = os.Environ()

		// Set GOPRIVATE environment variable if provided
		if goPrivate != "" {
			cmd.Env = mergeOrReplaceEnvVars(cmd.Env, []string{"GOPRIVATE=" + goPrivate})
		}

		// Add/replace custom environment variables (e.g., GOOS, GOARCH, CGO_ENABLED)
		cmd.Env = mergeOrReplaceEnvVars(cmd.Env, envVars)

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
		configPath := task.ConfigFile
		if !filepath.IsAbs(configPath) {
			absPath, err := filepath.Abs(configPath)
			if err == nil {
				configPath = absPath
			}
		}

		if _, ok := manifest.Sources[configPath]; !ok {
			manifest.Sources[configPath] = make(map[string]PluginManifest)
		}

		pluginManifest := PluginManifest{
			ModuleURI:   task.Plugin.ModuleURI,
			GitRef:      task.Plugin.GitRef,
			InstallPath: task.Plugin.InstallPath,
		}

		if len(task.Plugin.Libs) > 0 {
			pluginManifest.Libs = task.Plugin.Libs
		}

		manifest.Sources[configPath][task.PluginType] = pluginManifest
	}

	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for output file: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal installation artifacts: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write installation artifacts: %w", err)
	}

	log.Printf("Wrote installation artifacts to %s", outputFile)
	return nil
}

// installPlugins installs plugins concurrently using worker pool pattern
func installPlugins(tasks []PluginInstallTask, concurrency int, verbose bool, outputFile string) error {
	if len(tasks) == 0 {
		log.Println("No enabled plugins found to install")
		return nil
	}

	log.Printf("Installing %d plugins with concurrency %d", len(tasks), concurrency)

	if outputFile != "" {
		if err := writeBuildManifest(tasks, outputFile); err != nil {
			return fmt.Errorf("failed to write installation artifacts: %w", err)
		}
	}

	taskCh := make(chan PluginInstallTask, len(tasks))
	resultCh := make(chan PluginInstallResult, len(tasks))

	var wg sync.WaitGroup
	for i := range concurrency {
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

	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var hasErrors bool
	resultMap := make(map[string]error)

	for result := range resultCh {
		key := pluginKey(result.PluginType, result.Plugin)
		resultMap[key] = result.Error
		if result.Error != nil {
			hasErrors = true
		}
	}

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
	if !filepath.IsAbs(outputFile) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		outputFile = filepath.Join(wd, outputFile)
	}

	return outputFile, nil
}
