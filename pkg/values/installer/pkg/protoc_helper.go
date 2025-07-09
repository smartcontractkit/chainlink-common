package pkg

type ProtocHelper interface {
	// FullGoPackageName specifies a go package name for the capability.
	FullGoPackageName(c *CapabilityConfig) string
}
