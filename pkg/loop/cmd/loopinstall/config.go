package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

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

// pluginKey creates a unique key for a plugin to detect duplicates
func pluginKey(pluginType string, plugin PluginDef) string {
	return fmt.Sprintf("%s:%s:%s", pluginType, plugin.ModuleURI, plugin.InstallPath)
}

// isPluginEnabled checks if a plugin is enabled (defaults to true if not explicitly set to false)
func isPluginEnabled(plugin PluginDef) bool {
	return plugin.Enabled == nil || *plugin.Enabled
}
