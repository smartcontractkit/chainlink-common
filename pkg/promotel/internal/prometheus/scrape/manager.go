package scrape

import (
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Options are the configuration parameters to the scrape manager.
type Options struct {
	ExtraMetrics  bool
	NoDefaultPort bool
	// Option used by downstream scraper users like OpenTelemetry Collector
	// to help lookup metric metadata. Should be false for Prometheus.
	PassMetadataInContext bool
	// Option to enable appending of scraped Metadata to the TSDB/other appenders. Individual appenders
	// can decide what to do with metadata, but for practical purposes this flag exists so that metadata
	// can be written to the WAL and thus read for remote write.
	// TODO: implement some form of metadata storage
	AppendMetadata bool
	// Option to increase the interval used by scrape manager to throttle target groups updates.
	DiscoveryReloadInterval model.Duration
	// Option to enable the ingestion of the created timestamp as a synthetic zero sample.
	// See: https://github.com/prometheus/proposals/blob/main/proposals/2023-06-13_created-timestamp.md
	EnableCreatedTimestampZeroIngestion bool
	// Option to enable the ingestion of native histograms.
	EnableNativeHistogramsIngestion bool

	// Optional HTTP client options to use when scraping.
	HTTPClientOptions []config_util.HTTPClientOption
}
