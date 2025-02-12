package internal_test

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

func TestExample(t *testing.T) {
	var (
		g = prometheus.DefaultGatherer
		r = prometheus.DefaultRegisterer
		// todo: use logger.TestObserved
		logger = logger.Test(t)
	)

	go reportMetrics(r, logger)

	// Fetches metrics from in memory prometheus.Gatherer and converts to OTel format

	foundCh := make(chan struct{})
	// TODO: add mocked GRPC endpoint for exporter
	exporter := startExporter(context.Background(), logger)
	nextFunc := func(ctx context.Context, md pmetric.Metrics) error {
		if findExpectedMetric(testCounterMetricName, md) {
			foundCh <- struct{}{}
		}
		return exporter.Export(ctx, md)
	}
	receiver := startMetricReceiver(g, r, logger, nextFunc)

	defer receiver.Close()

	timeout := 10 * time.Second
	if deadline, ok := t.Deadline(); !ok {
		timeout = time.Until(deadline)
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("Timed out waiting for metric")
	case <-foundCh:
		t.Log("Found metric")
	}
}

const testCounterMetricName = "test_counter_metric"

func reportMetrics(reg prometheus.Registerer, logger logger.Logger) {
	testCounter := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: testCounterMetricName,
		ConstLabels: prometheus.Labels{
			"app": "promotel-demo",
		},
	})
	for {
		testCounter.Inc()
		m := &dto.Metric{}
		_ = testCounter.Write(m)
		logger.Info("Reported Prometheus metric ", zap.Any("name", testCounterMetricName), zap.Any("value", m.GetCounter().GetValue()))
		time.Sleep(1 * time.Second)
	}
}

func startExporter(ctx context.Context, logger logger.Logger) internal.MetricExporter {
	expConfig, err := internal.TestExporterConfig(map[string]any{
		"endpoint": "localhost:4317",
		"tls": map[string]any{
			"insecure": true,
		},
	})
	if err != nil {
		logger.Fatal("Failed to create exporter config", zap.Error(err))
	}
	// Sends metrics data in OTLP format to otel-collector endpoint
	exporter, err := internal.NewMetricExporter(expConfig, logger)
	if err != nil {
		logger.Fatal("Failed to create metric exporter", zap.Error(err))
	}
	err = exporter.Start(ctx)
	if err != nil {
		logger.Fatal("Failed to start exporter", zap.Error(err))
	}
	return exporter
}

func startMetricReceiver(g prometheus.Gatherer, r prometheus.Registerer, logger logger.Logger, next internal.NextFunc) internal.Runnable {
	logger.Info("Starting promotel metric receiver")
	config, err := internal.NewReceiverConfig()
	if err != nil {
		logger.Fatal("Failed to create config", zap.Error(err))
	}

	// Gather metrics via promotel
	// MetricReceiver fetches metrics from prometheus.Gatherer, then converts it to OTel format and writes formatted metrics to stdout
	receiver, err := internal.NewMetricReceiver(config, g, r, logger, next)
	if err != nil {
		logger.Fatal("Failed to create debug metric receiver", zap.Error(err))
	}
	// Starts the promotel
	if err := receiver.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start metric receiver", zap.Error(err))
	}
	return receiver
}
