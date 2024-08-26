package corenodecomponents

import "github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

type PlatformOpts struct {
	// Platform is infrastructure deployment platform: docker or k8s
	Platform     string
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string
	MetricsDataSource *grafana.DataSource
	LogsDataSource    *grafana.DataSource
	FolderUID         string
	PlatformOpts      PlatformOpts
}

// PlatformPanelOpts generate different queries for "docker" and "k8s" deployment platforms
func PlatformPanelOpts() PlatformOpts {
	po := PlatformOpts{
		LabelFilters: map[string]string{
			"env":          `=~"${env}"`,
			"cluster":      `=~"${cluster}"`,
			"blockchain":   `=~"${blockchain}"`,
			"product":      `=~"${product}"`,
			"network_type": `=~"${network_type}"`,
			"component":    `=~"${component}"`,
			"service":      `=~"${service}"`,
		},
	}
	for key, value := range po.LabelFilters {
		po.LabelQuery += key + value + ", "
	}
	return po
}
