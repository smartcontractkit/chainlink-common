package corenodecomponents

import "github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

type platformOpts struct {
	// Platform is infrastructure deployment platform: docker or k8s
	Platform     string
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string               // Name is the name of the dashboard
	Platform          grafana.TypePlatform // Platform is infrastructure deployment platform: docker or k8s
	MetricsDataSource *grafana.DataSource  // MetricsDataSource is the datasource for querying metrics
	LogsDataSource    *grafana.DataSource  // LogsDataSource is the datasource for querying logs
	platformOpts      platformOpts
}

// PlatformPanelOpts generate different queries for "docker" and "k8s" deployment platforms
func platformPanelOpts() platformOpts {
	po := platformOpts{
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
