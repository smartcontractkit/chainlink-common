package pkg

type ProtocHelper interface {
	// FullGoPackageName specifies a go package name for the capability.
	FullGoPackageName(c *CapabilityConfig) string
	// SdkPgk returns where the base SDK and metadata protos are generated
	SdkPgk() string
}
