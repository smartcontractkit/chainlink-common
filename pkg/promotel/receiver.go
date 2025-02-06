package promotel

import (
	"context"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
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

func NewMetricReceiver(config ReceiverConfig, g prometheus.Gatherer, r prometheus.Registerer, consumerFunc consumer.ConsumeMetricsFunc, logger logger.Logger) (Runnable, error) {
	factory := prometheusreceiver.NewFactoryWithOptions(
		prometheusreceiver.WithGatherer(g),
		prometheusreceiver.WithRegisterer(r),
	)
	settings, err := internal.NewReceiverSettings()
	if err != nil {
		return nil, err
	}
	// Creates a metrics receiver with the context, settings, config, and consumer
	receiver, err := factory.CreateMetrics(
		context.Background(),
		settings,
		config,
		internal.NewConsumer(consumerFunc))
	if err != nil {
		return nil, err
	}
	// Creates a no-operation host for the receiver
	host := internal.NewNopHost()
	return &metricReceiver{factory, host, receiver}, nil
}

func NewDebugMetricReceiver(config ReceiverConfig, g prometheus.Gatherer, r prometheus.Registerer, logger logger.Logger) (MetricReceiver, error) {
	debugExporter := internal.NewDebugExporter(logger)
	// Creates a no-operation consumer
	return NewMetricReceiver(config, g, r, func(_ context.Context, md pmetric.Metrics) error {
		// Writes metrics data to stdout
		return debugExporter.Export(md)
	}, logger)
}
