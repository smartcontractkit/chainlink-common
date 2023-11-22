package codec

type ModificationConfig struct {
	ElementExtractions      map[string]ElementExtractorLocation
	OnChainHardCodedValues  map[string]any
	OffChainHardCodedValues map[string]any
	Renames                 map[string]string
}
