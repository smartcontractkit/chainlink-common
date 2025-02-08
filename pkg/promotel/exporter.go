package promotel

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

type MetricExporter interface {
	Runnable
	Consumer() consumer.Metrics
	Export(ctx context.Context, md pmetric.Metrics) error
}

type metricExporter struct {
	factory  exporter.Factory
	host     component.Host
	exporter exporter.Metrics
}

type NextFunc consumer.ConsumeMetricsFunc

func (me *metricExporter) Start(ctx context.Context) error {
	return me.exporter.Start(ctx, me.host)
}

func (me *metricExporter) Close() error {
	return me.exporter.Shutdown(context.Background())
}

func (me *metricExporter) Consumer() consumer.Metrics {
	return me.exporter
}

func (me *metricExporter) Export(ctx context.Context, md pmetric.Metrics) error {
	return me.exporter.ConsumeMetrics(ctx, md)
}

func NewMetricExporter(config ExporterConfig, logger logger.Logger) (MetricExporter, error) {
	factory := otlpexporter.NewFactory()
	// Creates a metrics receiver with the context, settings, config, and consumer
	exporter, err := factory.CreateMetrics(
		context.Background(),
		internal.NewExporterSettings(logger),
		config)
	if err != nil {
		return nil, err
	}
	// Creates a no-operation host for the receiver
	host := internal.NewNopHost()
	return &metricExporter{factory, host, exporter}, nil
}
