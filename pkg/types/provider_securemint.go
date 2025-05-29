package types

// SecureMintProvider provides all components needed for a secure mint OCR2 plugin.
type SecureMintProvider interface {
	PluginProvider
	// TODO(gg): more fields needed?

	// ReportCodec() median.ReportCodec
	// MedianContract() median.MedianContract
	// OnchainConfigCodec() median.OnchainConfigCodec
}
