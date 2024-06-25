package corenode

import "fmt"

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
}

// PlatformPanelOpts generate different queries for "docker" and "k8s" deployment platforms
func PlatformPanelOpts(platform string) PlatformOpts {
	po := PlatformOpts{
		LabelFilters: map[string]string{},
		Platform:     platform,
	}
	switch platform {
	case "kubernetes":
		po.LabelFilters["namespace"] = `=~"${namespace}"`
		po.LabelFilters["job"] = `=~"${job}"`
		po.LabelFilter = "job"
		po.LegendString = "pod"
	case "docker":
		po.LabelFilters["instance"] = `=~"${instance}"`
		po.LabelFilter = "instance"
		po.LegendString = "instance"
	default:
		panic(fmt.Sprintf("failed to generate Platform dependent queries, unknown platform: %s", platform))
	}
	for key, value := range po.LabelFilters {
		po.LabelQuery += key + value + ", "
	}
	return po
}
