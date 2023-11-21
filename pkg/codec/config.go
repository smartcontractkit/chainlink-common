package codec

type ModificationConfig struct {
	ElementExtractions map[string]ElementExtractorLocation
	Renames            map[string]string
}
