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
)

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
		goPrivate := defaults.GoPrivate

		if goPrivate != "" && envGoPrivate != "" {
			goPrivate = goPrivate + "," + envGoPrivate
		} else if goPrivate == "" && envGoPrivate != "" {
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