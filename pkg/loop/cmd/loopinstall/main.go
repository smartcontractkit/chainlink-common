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
	fmt.Println("")
	fmt.Println("  plugins:")
	fmt.Println("    plugin-type:")
	fmt.Println("      - moduleURI: \"github.com/example/module\"")
	fmt.Println("        gitRef: \"v1.0.0\"")
	fmt.Println("        installPath: \"./cmd/example\"")
	fmt.Println("        libs: [\"/path/to/libs/*.so\"]  # Optional library paths (can include glob patterns)")
	fmt.Println("        # enabled: false  # Optional, defaults to true if omitted")
}
