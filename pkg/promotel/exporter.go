package promotel

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

type MetricExporter interface {
	Runnable
	Consumer() consumer.Metrics
}

type metricExporter struct {
	factory  exporter.Factory
	host     component.Host
	exporter exporter.Metrics
}

func (me *metricExporter) Start(ctx context.Context) error {
	return me.exporter.Start(ctx, me.host)
}

func (me *metricExporter) Close() error {
	return me.exporter.Shutdown(context.Background())
}

func (me *metricExporter) Consumer() consumer.Metrics {
	// Writes metrics data to stdout
	return me.exporter
}

func NewMetricExporter(config ExporterConfig, logger logger.Logger) (MetricExporter, error) {
	factory := otlpexporter.NewFactory()
	// Creates a metrics receiver with the context, settings, config, and consumer
	settings, err := internal.NewExporterSettings()
	if err != nil {
		return nil, err
	}
	exporter, err := factory.CreateMetrics(
		context.Background(),
		settings,
		config)
	if err != nil {
		return nil, err
	}
	// Creates a no-operation host for the receiver
	host := internal.NewNopHost()
	return &metricExporter{factory, host, exporter}, nil
}
