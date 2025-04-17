package main

// PluginConfig represents the structure of our plugin configuration files
type PluginConfig struct {
	Defaults DefaultsConfig         `yaml:"defaults"`
	Plugins  map[string][]PluginDef `yaml:"plugins"`
}

// DefaultsConfig holds the default configuration values
type DefaultsConfig struct {
	GoFlags string `yaml:"goflags"`
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
