package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

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

// printHelp prints usage and help information
func printHelp() {
	fmt.Print(`Chainlink Plugin Installer

A tool for installing Chainlink plugins from YAML configuration files.

Usage:
  loopinstall [options] <plugin-config-file> [<plugin-config-file>...]

Options:
  -h, --help                Show this help message
  -v, --verbose             Enable verbose output
  -c, --concurrency <num>   Set maximum number of concurrent installations (default: 5)
  -s, --sequential          Install plugins sequentially (no concurrency)
  -o, --output-installation-artifacts <file>  Path for installation artifacts JSON file
                             (optional, no installation artifacts written if not specified)

Examples:
  # Install plugins from the default configuration
  loopinstall plugins.default.yaml

  # Install plugins with custom installation artifacts filename
  loopinstall -o ./installation-artifacts.json plugins.default.yaml

  # Install plugins sequentially
  loopinstall -s plugins.default.yaml

  # Install plugins with go flags variable overrides
  CL_PLUGIN_GOFLAGS="-ldflags='-s'" loopinstall plugins.default.yaml

  # Install plugins with custom environment variables
  CL_PLUGIN_ENVVARS="GOOS=linux GOARCH=amd64 CGO_ENABLED=0" loopinstall plugins.default.yaml

Environment Variables:
  CL_PLUGIN_GOFLAGS  Override the goflags option from the configuration
  CL_PLUGIN_ENVVARS  Space-separated list of environment variables to set during installation (for example for cross-compilation)

Plugin Configuration Format:
  defaults:
    goflags: ""     # Default Go build flags
    envvars: []     # Default environment variables

  plugins:
    plugin-type:
      - moduleURI: "github.com/example/module"
        gitRef: "v1.0.0"
        installPath: "./cmd/example"
        libs: ["/path/to/libs/*.so"]  # Optional library paths (can include glob patterns)
        # enabled: false  # Optional, defaults to true if omitted
`)
}
