package corenode

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type platformOpts struct {
	// Platform is infrastructure deployment platform: docker or k8s
	Platform     grafana.TypePlatform
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string               // Name is the name of the dashboard
	Platform          grafana.TypePlatform // Platform is infrastructure deployment platform: docker or k8s
	MetricsDataSource *grafana.DataSource  // MetricsDataSource is the datasource for querying metrics
	SlackChannel      string               // SlackChannel is the channel to send alerts to
	SlackWebhookURL   string               // SlackWebhookURL is the URL to send alerts to
	AlertsTags        map[string]string    // AlertsTags is the tags to map with notification policy
	AlertsFilters     string               // AlertsFilters is the filters to apply to alerts
	platformOpts      platformOpts
}

// PlatformPanelOpts generate different queries for "docker" and "k8s" deployment platforms
func platformPanelOpts(platform grafana.TypePlatform) platformOpts {
	po := platformOpts{
		LabelFilters: map[string]string{},
		Platform:     platform,
	}
	switch platform {
	case grafana.TypePlatformKubernetes:
		po.LabelFilters["namespace"] = `=~"${namespace}"`
		po.LabelFilters["job"] = `=~"${job}"`
		po.LabelFilters["pod"] = `=~"${pod}"`
		po.LabelFilter = "job"
		po.LegendString = "pod"
	case grafana.TypePlatformDocker:
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
