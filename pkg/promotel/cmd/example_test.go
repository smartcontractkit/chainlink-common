package main

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestExample(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	go reportMetrics(prometheus.DefaultRegisterer, logger)

	// Fetches metrics from in memory prometheus.Gatherer and converts to OTel format
	foundCh := make(chan struct{})
	receiver := startMetricReceiver(prometheus.DefaultGatherer, logger, func(_ context.Context, md pmetric.Metrics) error {
		// Logs the converted OTel metric
		rms := md.ResourceMetrics()
		for i := 0; i < rms.Len(); i++ {
			rm := rms.At(i)
			ilms := rm.ScopeMetrics()
			for j := 0; j < ilms.Len(); j++ {
				ilm := ilms.At(j)
				metrics := ilm.Metrics()
				for k := 0; k < metrics.Len(); k++ {
					metric := metrics.At(k)
					if metric.Name() == testCounterMetricName {
						v := metric.Sum().DataPoints().At(0).DoubleValue()
						logger.Info("Exporting OTel metric ", zap.Any("name", metric.Name()), zap.Any("value", v))
						if v > 0 {
							foundCh <- struct{}{}
							return nil
						}
					}
				}
			}
		}
		return nil
	})
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
