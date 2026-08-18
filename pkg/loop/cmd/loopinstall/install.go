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
// replacing any existing variables with the same key.
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

// determineModuleDirectory locates the directory to build from.
// - Local path (absolute or "./relative"): resolve and return the directory (no download).
func determineModuleDirectoryLocal(pluginKey, moduleURI string) (string, error) {
	log.Printf("%s - resolving local module path %q", pluginKey, moduleURI)
	abs, err := filepath.Abs(moduleURI)
	if err != nil {
		return "", fmt.Errorf("failed to resolve local module path %q: %w", moduleURI, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("local module path %q not accessible: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("local module path %q is not a directory", abs)
	}
	return abs, nil
}

// determineModuleDirectory locates the directory to build from.
// - Remote module path (e.g., "github.com/org/repo@ref"): use `go mod download -json` to get a module cache dir.
func determineModuleDirectoryRemote(pluginKey, moduleURI, gitRef, goPrivate string) (string, error) {
	fullModulePath := moduleURI
	if gitRef != "" {
		fullModulePath = fmt.Sprintf("%s@%s", moduleURI, gitRef)
	}
	log.Printf("%s - downloading remote module %s", pluginKey, fullModulePath)

	cmd := exec.Command("go", "mod", "download", "-json", fullModulePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if goPrivate != "" {
		// Inherit the current environment and override GOPRIVATE.
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

// determineGoFlags resolves go build flags in priority order:
// 1) CL_PLUGIN_GOFLAGS env var (overrides config)
// 2) defaults.GoFlags
// 3) plugin-specific goflags appended
// It validates flags if any are present.
func determineGoFlags(pluginKey, defaultGoFlags, pluginGoFlags string) ([]string, error) {
	var flags []string
	parser := shellwords.NewParser()

	// Determine base flags
	if envGoFlags := os.Getenv("CL_PLUGIN_GOFLAGS"); envGoFlags != "" {
		log.Printf("%s - overriding config's default goflags with CL_PLUGIN_GOFLAGS env var: %s", pluginKey, envGoFlags)
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

// determineInstallArg computes the argument passed to `go build` given we're changing cmd.Dir.
// For remote modules, we keep the legacy behavior.
// For local moduleURIs, we compute a relative path from the module root to the installPath
// so the resulting arg is "." or "./sub/package".
func determineInstallArg(installPath, moduleURI string, isLocal bool) string {
	cleanInstallPath := filepath.Clean(installPath)
	cleanModuleURI := filepath.Clean(moduleURI)

	// Local modules
	if isLocal {
		// 1 - If building the module root
		if cleanInstallPath == cleanModuleURI || cleanInstallPath == "." {
			return "."
		}
		// 2 - If installPath is inside the module root, return "./<rel>"
		if rel, err := filepath.Rel(cleanModuleURI, cleanInstallPath); err == nil && rel != "" && !strings.HasPrefix(rel, "..") {
			rel = filepath.ToSlash(rel)
			if rel == "." {
				return "."
			}
			return "./" + rel
		}

		// 3 - If installPath is already relative to the module root, normalize "./" prefix
		if !filepath.IsAbs(cleanInstallPath) {
			cleanInstallPath = filepath.ToSlash(cleanInstallPath)
			if cleanInstallPath == "." || strings.HasPrefix(cleanInstallPath, "./") {
				return cleanInstallPath
			}
			return "./" + strings.TrimLeft(cleanInstallPath, "/")
		}

		// Absolute path outside module root: still give a relative-looking arg;
		// cmd.Dir will be set to module root so Go expects package paths like "./x/y".
		return "./" + filepath.ToSlash(strings.TrimLeft(cleanInstallPath, string(filepath.Separator)))
	}

	// Remote modules
	// 1 - installPath is the module root itself.
	if installPath == moduleURI {
		return "."
	}

	// 2 - installPath is a sub-package of moduleURI.
	if after, ok := strings.CutPrefix(installPath, moduleURI+"/"); ok {
		cleanedRelativePath := strings.TrimLeft(after, "/")
		if cleanedRelativePath == "" || cleanedRelativePath == "." {
			return "."
		}
		return "./" + cleanedRelativePath
	}

	// 3 - other inputs; normalize to a "./" path.
	if installPath == "." {
		return "."
	}
	if strings.HasPrefix(installPath, "./") {
		return installPath
	}
	return "./" + strings.TrimLeft(installPath, "/")
}

// downloadAndInstallPlugin downloads (if remote) and builds the plugin.
// For local moduleURIs (absolute or "./relative"), we skip network download,
// ignore gitRef (with a log message), and build directly from the local dir.
func downloadAndInstallPlugin(pluginType string, pluginIdx int, plugin PluginDef, defaults DefaultsConfig) error {
	pluginKey := fmt.Sprintf("%s[%d]", pluginType, pluginIdx)
	if !isPluginEnabled(plugin) {
		log.Printf("%s - skipping disabled plugin", pluginKey)
		return nil
	}

	// Validate inputs
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("%s - plugin input validation failed: %w", pluginKey, err)
	}

	moduleURI := plugin.ModuleURI
	gitRef := plugin.GitRef
	installPath := plugin.InstallPath

	goPrivate := os.Getenv("GOPRIVATE")

	// Determine the directory to run `go build` in.
	isLocal := filepath.IsAbs(moduleURI) || strings.HasPrefix(moduleURI, "."+string(filepath.Separator))
	moduleDir, err := func() (string, error) {
		if isLocal {
			return determineModuleDirectoryLocal(pluginKey, moduleURI)
		}
		return determineModuleDirectoryRemote(pluginKey, moduleURI, gitRef, goPrivate)
	}()
	if err != nil {
		return fmt.Errorf("%s - failed to determine module directory: %w", pluginKey, err)
	}
	if moduleDir == "" {
		return fmt.Errorf("%s - empty module directory resolved", pluginKey)
	}

	log.Printf("%s - installing plugin from %s", pluginKey, moduleDir)

	// Build env vars from defaults, environment variable, and plugin-specific settings.
	envVars := defaults.EnvVars
	if envEnvVars := os.Getenv("CL_PLUGIN_ENVVARS"); envEnvVars != "" {
		envVars = mergeOrReplaceEnvVars(envVars, strings.Fields(envEnvVars))
	}
	if len(plugin.EnvVars) != 0 {
		envVars = mergeOrReplaceEnvVars(envVars, plugin.EnvVars)
	}

	// Compute build target relative to module root ('.' or './subpkg').
	installArg := determineInstallArg(installPath, moduleURI, isLocal)

	// Derive output binary name. When arg is ".", use the module/repo (or local dir) name.
	binaryName := filepath.Base(installArg)
	if binaryName == "." {
		binaryName = filepath.Base(filepath.Clean(moduleURI))
	}

	// Determine output directory (GOBIN, or GOPATH/bin, or $HOME/go/bin).
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
	goflags, err := determineGoFlags(pluginKey, defaults.GoFlags, plugin.Flags)
	if err != nil {
		return fmt.Errorf("%s - goflags validation failed: %w", pluginKey, err)
	}

	// Assemble `go build` command.
	args := []string{"build", "-o", outputPath}
	if len(goflags) != 0 {
		args = append(args, goflags...)
	}
	args = append(args, installArg)

	cmd := exec.Command("go", args...)
	cmd.Dir = moduleDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start with all current environment variables.
	cmd.Env = os.Environ()
	if goPrivate != "" {
		cmd.Env = mergeOrReplaceEnvVars(cmd.Env, []string{"GOPRIVATE=" + goPrivate})
	}
	cmd.Env = mergeOrReplaceEnvVars(cmd.Env, envVars)

	log.Printf("%s - running install command: go %s (in directory: %s)", pluginKey, strings.Join(args, " "), moduleDir)

	if err := execCommand(cmd); err != nil {
		return fmt.Errorf("%s - failed to install plugin: %w", pluginKey, err)
	}

	return nil
}

// writeBuildManifest writes installation artifacts to the specified file.
func writeBuildManifest(tasks []PluginInstallTask, outputFile string) error {
	manifest := BuildManifest{
		BuildTime: time.Now().UTC().Format(time.RFC3339),
		Sources:   make(map[string]map[string]PluginManifest),
	}

	// Group tasks by source file
	for _, task := range tasks {
		configPath := task.ConfigFile
		if !filepath.IsAbs(configPath) {
			if absPath, err := filepath.Abs(configPath); err == nil {
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

// installPlugins installs plugins concurrently using a worker pool.
func installPlugins(tasks []PluginInstallTask, concurrency int, verbose bool, outputFile string) error {
	if len(tasks) == 0 {
		log.Println("No enabled plugins found to install")
		return nil
	}

	log.Printf("Installing %d plugins with concurrency %d", len(tasks), concurrency)

	// Optionally write the manifest first (so artifacts exist even if a build fails).
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

// setupOutputFile ensures the output path is absolute (and its directory exists is handled elsewhere).
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
