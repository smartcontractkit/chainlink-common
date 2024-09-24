package atlasdon

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type PlatformOpts struct {
	Platform     grafana.TypePlatform
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string
	Platform          grafana.TypePlatform
	MetricsDataSource *grafana.DataSource
	PlatformOpts      PlatformOpts
	OCRVersion        string
}

// PlatformPanelOpts generate different queries depending on params
func PlatformPanelOpts(platform grafana.TypePlatform, ocrVersion string) PlatformOpts {
	po := PlatformOpts{
		LabelFilters: map[string]string{
			"contract": `=~"${contract}"`,
		},
		Platform: platform,
	}

	variableFeedID := "feed_id"
	if ocrVersion == "ocr3" {
		variableFeedID = "feed_id_name"
	}

	switch ocrVersion {
	case "ocr2":
		po.LabelFilters[variableFeedID] = `=~"${` + variableFeedID + `}"`
	case "ocr3":
		po.LabelFilters[variableFeedID] = `=~"${` + variableFeedID + `}"`
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
