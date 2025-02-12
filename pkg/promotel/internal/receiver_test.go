package internal_test

import (
	"context"
	"testing"
	"time"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

// TestPrometheusReceiver verifies the initialization, startup, and shutdown
// processes of the Prometheus receiver. It ensures that no errors occur when
// creating a metrics receiver from a loaded configuration, starting it, and
// gracefully stopping it.
func TestPrometheusReceiver(t *testing.T) {
	// Load configuration from a YAML file
	testConfig, err := internal.NewReceiverConfig()
	require.NoError(t, err)
	// Creates a new Prometheus receiver factory
	factory := prometheusreceiver.NewFactory()
	// Creates a metrics receiver with the context, settings, config, and consumer
	receiver, err := factory.CreateMetrics(context.Background(), receivertest.NewNopSettings(), testConfig, consumertest.NewNop())
	// Verifies the receiver was created without error
	require.NoError(t, err)
	// Creates a no-operation host for the receiver
	host := componenttest.NewNopHost()
	// Ensures no error occurred before continuing
	require.NoError(t, err)
	// Starts the receiver with the provided host
	require.NoError(t, receiver.Start(context.Background(), host))
	// Gracefully shuts down the receiver
	require.NoError(t, receiver.Shutdown(context.Background()))
}

// TestMetricReceiverStartClose verifies the initialization, startup, and shutdown
func TestMetricReceiverStartClose(t *testing.T) {
	testConfig, err := internal.NewReceiverConfig()
	require.NoError(t, err)
	noopConsumerFunc := func(context.Context, pmetric.Metrics) error { return nil }
	receiver, err := internal.NewMetricReceiver(testConfig, prometheus.DefaultGatherer, prometheus.DefaultRegisterer, nil, noopConsumerFunc)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background()))
	require.NoError(t, receiver.Close())
}

func TestMetricReceiver(t *testing.T) {
	var (
		g = prometheus.DefaultGatherer
		r = prometheus.DefaultRegisterer
		// todo: use logger.TestObserved
		logger                = logger.Test(t)
		timeout               = 10 * time.Second
		testCounterMetricName = "test_counter_metric"
		receiver              internal.MetricReceiver
	)
	if deadline, ok := t.Deadline(); !ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start reporting metrics
	go func() {
		testCounter := promauto.With(r).NewCounter(prometheus.CounterOpts{
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
	}()

	doneCh := make(chan struct{})
	nextFunc := func(ctx context.Context, md pmetric.Metrics) error {
		if findExpectedMetric(testCounterMetricName, md) {
			doneCh <- struct{}{}
		}
		return nil
	}

	// Start the metric receiver
	go func() {
		config, err := internal.NewReceiverConfig()
		assert.NoError(t, err)
		// MetricReceiver fetches metrics from prometheus.Gatherer
		receiver, err = internal.NewMetricReceiver(config, g, r, logger, nextFunc)
		assert.NoError(t, err)
		// Starts the promotel
		assert.NoError(t, receiver.Start(context.Background()))
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out, expected metric not found")
	case <-doneCh:
		t.Log("Found metric")
	}
	receiver.Close()
}

func findExpectedMetric(name string, md pmetric.Metrics) bool {
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
					v := metric.Sum().DataPoints().At(0).DoubleValue()
					if v > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}
