package atlasdon

import (
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type platformOpts struct {
	LabelFilters map[string]string
	LabelFilter  string
	LegendString string
	LabelQuery   string
}

type Props struct {
	Name              string              // Name is the name of the dashboard
	MetricsDataSource *grafana.DataSource // MetricsDataSource is the datasource for querying metrics
	OCRVersion        string              // OCRVersion is the version of the OCR (ocr, ocr2, ocr3)
	platformOpts      platformOpts
	Tested            bool
}

// PlatformPanelOpts generate different queries depending on params
func platformPanelOpts(ocrVersion string) platformOpts {
	po := platformOpts{
		LabelFilters: map[string]string{
			"contract": `=~"${contract}"`,
		},
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
	namespace := "otpe"
	if ocrVersion == "ocr2" {
		namespace = "otpe2"
	} else if ocrVersion == "ocr3" {
		namespace = "otpe3"
	}

	po.LabelFilters["namespace"] = `="` + namespace + `"`
	po.LabelFilters["job"] = `=~"${job}"`
	po.LabelFilter = "job"
	po.LegendString = "job"

	for key, value := range po.LabelFilters {
		po.LabelQuery += key + value + ", "
	}
	return po
}
