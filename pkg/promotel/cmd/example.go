package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel"
)

const testCounterMetricName = "test_counter_metric"

func reportMetrics(reg prometheus.Registerer, logger *zap.Logger) {
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

func gatherMetricsDirectly(reg prometheus.Gatherer, logger *zap.Logger) {
	for {
		mf, err := reg.Gather()
		if err != nil {
			fmt.Printf("Error gathering metrics: %v\n", err)
		}
		for _, metricFamily := range mf {
			if *metricFamily.Name == testCounterMetricName {
				for _, metric := range metricFamily.Metric {
					logger.Info("Received Prometheus metric ", zap.Any("name", testCounterMetricName), zap.Any("value", metric.Counter.GetValue()))
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func startExporter(ctx context.Context, logger *zap.Logger) promotel.MetricExporter {
	expConfig, err := promotel.NewExporterConfig(map[string]any{
		"endpoint": "localhost:4317",
		"tls": map[string]any{
			"insecure": true,
		},
	})
	if err != nil {
		logger.Fatal("Failed to create exporter config", zap.Error(err))
	}
	// Sends metrics data in OTLP format to otel-collector endpoint
	exporter, err := promotel.NewMetricExporter(expConfig, logger)
	if err != nil {
		logger.Fatal("Failed to create metric exporter", zap.Error(err))
	}
	err = exporter.Start(ctx)
	if err != nil {
		logger.Fatal("Failed to start exporter", zap.Error(err))
	}
	return exporter
}

func startMetricReceiver(reg prometheus.Gatherer, logger *zap.Logger, next consumer.ConsumeMetricsFunc) promotel.Runnable {
	logger.Info("Starting promotel metric receiver")
	config, err := promotel.NewDefaultReceiverConfig()
	if err != nil {
		logger.Fatal("Failed to create config", zap.Error(err))
	}

	// Gather metrics via promotel
	// MetricReceiver fetches metrics from prometheus.Gatherer, then converts it to OTel format and writes formatted metrics to stdout
	receiver, err := promotel.NewMetricReceiver(config, reg, next, logger)

	if err != nil {
		logger.Fatal("Failed to create debug metric receiver", zap.Error(err))
	}
	// Starts the promotel
	if err := receiver.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start metric receiver", zap.Error(err))
	}
	return receiver
}

func main() {
	logger, _ := zap.NewDevelopment()

	go reportMetrics(prometheus.DefaultRegisterer, logger)
	// Gather metrics directly from DefaultGatherer to verify that the metrics are being reported
	go gatherMetricsDirectly(prometheus.DefaultGatherer, logger)

	exporter := startExporter(context.Background(), logger)
	// Fetches metrics from in memory prometheus.Gatherer and converts to OTel format
	receiver := startMetricReceiver(prometheus.DefaultGatherer, logger, func(ctx context.Context, md pmetric.Metrics) error {
		// Logs the converted OTel metric
		logOtelMetric(md, testCounterMetricName, logger)
		// Exports the converted OTel metric
		return exporter.Consumer().ConsumeMetrics(ctx, md)
	})

	// Wait for a signal to exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-signalChan
	logger.Info("Exiting promotel")
	// Gracefully shuts down promotel
	if err := receiver.Close(); err != nil {
		logger.Fatal("Failed to close scraper", zap.Error(err))
	}
	if err := exporter.Close(); err != nil {
		logger.Fatal("Failed to close exporter", zap.Error(err))
	}
}

func logOtelMetric(md pmetric.Metrics, name string, logger *zap.Logger) {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				if metric.Name() == name {
					logger.Info("Exporting OTel metric ", zap.Any("name", metric.Name()), zap.Any("value", metric.Sum().DataPoints().At(0).DoubleValue()))
				}
			}
		}
	}
}
