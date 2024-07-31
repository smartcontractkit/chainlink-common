package nopocr

type PlatformOpts struct {
	// Platform is infrastructure deployment platform: docker or k8s
	Platform     string
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	MetricsDataSource string
	PlatformOpts      PlatformOpts
	OcrVersion        string
}

// PlatformPanelOpts generate different queries for "docker" and "k8s" deployment platforms
func PlatformPanelOpts(platform string, ocrVersion string) PlatformOpts {
	po := PlatformOpts{
		LabelFilters: map[string]string{
			"env":      `=~"${env}"`,
			"contract": `=~"${contract}"`,
			"oracle":   `=~"${oracle}"`,
		},
		Platform: platform,
	}

	for key, value := range po.LabelFilters {
		po.LabelQuery += key + value + ", "
	}
	return po
}
