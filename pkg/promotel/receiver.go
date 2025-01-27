package promotel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheus/scrape"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheusreceiver"
)

type Runnable interface {
	Start(context.Context) error
	Close() error
}

type MetricReceiver interface {
	Runnable
}

type metricReceiver struct {
	factory  receiver.Factory
	host     component.Host
	receiver receiver.Metrics
}

func (p *metricReceiver) Start(ctx context.Context) error {
	return p.receiver.Start(ctx, p.host)
}

func (p *metricReceiver) Close() error {
	return p.receiver.Shutdown(context.Background())
}

func NewMetricReceiver(config ReceiverConfig, g prometheus.Gatherer, consumerFunc consumer.ConsumeMetricsFunc, logger *zap.Logger) (Runnable, error) {
	// Scrape from the provided gatherer
	scrape.SetDefaultGatherer(g)

	factory := prometheusreceiver.NewFactory()
	// Creates a metrics receiver with the context, settings, config, and consumer
	receiver, err := factory.CreateMetrics(
		context.Background(),
		internal.NewReceiverSettings(logger),
		config,
		internal.NewConsumer(consumerFunc))
	if err != nil {
		return nil, err
	}
	// Creates a no-operation host for the receiver
	host := internal.NewNopHost()
	return &metricReceiver{factory, host, receiver}, nil
}

func NewDebugMetricReceiver(config ReceiverConfig, g prometheus.Gatherer, logger *zap.Logger) (MetricReceiver, error) {
	debugExporter := internal.NewDebugExporter(logger)
	// Creates a no-operation consumer
	return NewMetricReceiver(config, g, func(ctx context.Context, md pmetric.Metrics) error {
		// Writes metrics data to stdout
		return debugExporter.Export(md)
	}, logger)
}
