package pkg

// GeneratorHelper defines parameters for generating capability code with a protoc plugin.
type GeneratorHelper interface {
	// FullGoPackageName specifies a go package name for the capability.
	FullGoPackageName(c *CapabilityConfig) string
	// SdkPgk returns where the base SDK and metadata protos are generated
	SdkPgk() string
	// PluginName returns the fully quantified name of the protoc plugin to install.
	// It will be verified that the package generating uses the same version of github.com/smartcontractkit/chainlink-common/pkg/values
	// as the plugin.
	PluginName() string
	// HelperName returns the fully quantified name of the helper package for the generator.
	HelperName() string
}
