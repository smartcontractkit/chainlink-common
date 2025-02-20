package internal

import (
	"context"
	"time"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
	logger   logger.Logger
}

func (p *metricReceiver) Start(ctx context.Context) error {
	return p.receiver.Start(ctx, p.host)
}

func (p *metricReceiver) Close() error {
	return p.receiver.Shutdown(context.Background())
}

func NewMetricReceiver(config *ReceiverConfig, g prometheus.Gatherer, r prometheus.Registerer, interval time.Duration, logger logger.Logger, next NextFunc) (Runnable, error) {
	factory := prometheusreceiver.NewFactoryWithOptions(
		prometheusreceiver.WithGatherer(g),
		prometheusreceiver.WithRegisterer(r),
		prometheusreceiver.WithGathererInterval(interval),
	)
	// Creates a metrics receiver with the context, settings, config, and consumer
	receiver, err := factory.CreateMetrics(
		context.Background(),
		NewReceiverSettings(logger),
		config,
		NewConsumer(consumer.ConsumeMetricsFunc(next)))
	if err != nil {
		return nil, err
	}
	// Creates a no-operation host for the receiver
	host := NewNopHost()
	return &metricReceiver{factory, host, receiver, logger}, nil
}

func NewDebugMetricReceiver(config *ReceiverConfig, g prometheus.Gatherer, r prometheus.Registerer, interval time.Duration, logger logger.Logger) (MetricReceiver, error) {
	debugExporter := NewDebugExporter(logger)
	// Creates a no-operation consumer
	return NewMetricReceiver(config, g, r, interval, logger, func(_ context.Context, md pmetric.Metrics) error {
		// Writes metrics data to stdout
		return debugExporter.Export(md)
	})
}
