package internal

import (
	"github.com/prometheus/common/model"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/prometheusreceiver/scrape"
)

type dataPoint struct {
	value    float64
	boundary float64
}

// internalMetricMetadata allows looking up metadata for internal scrape metrics
var internalMetricMetadata = map[string]*scrape.MetricMetadata{
	scrapeUpMetricName: {
		Metric: scrapeUpMetricName,
		Type:   model.MetricTypeGauge,
		Help:   "The scraping was successful",
	},
	"scrape_duration_seconds": {
		Metric: "scrape_duration_seconds",
		Unit:   "seconds",
		Type:   model.MetricTypeGauge,
		Help:   "Duration of the scrape",
	},
	"scrape_samples_scraped": {
		Metric: "scrape_samples_scraped",
		Type:   model.MetricTypeGauge,
		Help:   "The number of samples the target exposed",
	},
	"scrape_series_added": {
		Metric: "scrape_series_added",
		Type:   model.MetricTypeGauge,
		Help:   "The approximate number of new series in this scrape",
	},
	"scrape_samples_post_metric_relabeling": {
		Metric: "scrape_samples_post_metric_relabeling",
		Type:   model.MetricTypeGauge,
		Help:   "The number of samples remaining after metric relabeling was applied",
	},
}

func metadataForMetric(metricName string, mc scrape.MetricMetadataStore) (*scrape.MetricMetadata, string) {
	if metadata, ok := internalMetricMetadata[metricName]; ok {
		return metadata, metricName
	}
	if metadata, ok := mc.GetMetadata(metricName); ok {
		return &metadata, metricName
	}
	// If we didn't find metadata with the original name,
	// try with suffixes trimmed, in-case it is a "merged" metric type.
	normalizedName := normalizeMetricName(metricName)
	if metadata, ok := mc.GetMetadata(normalizedName); ok {
		if metadata.Type == model.MetricTypeCounter {
			return &metadata, metricName
		}
		return &metadata, normalizedName
	}
	// Otherwise, the metric is unknown
	return &scrape.MetricMetadata{
		Metric: metricName,
		Type:   model.MetricTypeUnknown,
	}, metricName
}
