package internal_test

import (
	"context"
	"testing"
	"time"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

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
	receiver, err := internal.NewMetricReceiver(testConfig, prometheus.DefaultGatherer, prometheus.DefaultRegisterer, 10*time.Second, nil, noopConsumerFunc)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background()))
	require.NoError(t, receiver.Close())
}

func TestMetricReceiver(t *testing.T) {
	var (
		g = prometheus.DefaultGatherer
		r = prometheus.DefaultRegisterer
		// todo: use logger.TestObserved
		logger         = logger.Test(t)
		timeout        = 10 * time.Second
		testMetricName = t.Name() + "_test_counter_metric"
		receiver       internal.MetricReceiver
		interval       = 10 * time.Millisecond
	)
	if deadline, ok := t.Deadline(); !ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start reporting metrics
	go internal.ReportTestMetrics(ctx, r, testMetricName)

	doneCh := make(chan struct{})
	nextFunc := func(ctx context.Context, md pmetric.Metrics) error {
		if internal.FindExpectedMetric(testMetricName, md) {
			doneCh <- struct{}{}
		}
		return nil
	}
	// Create new metric receiver
	config, err := internal.NewReceiverConfig()
	assert.NoError(t, err)
	receiver, err = internal.NewMetricReceiver(config, g, r, interval, logger, nextFunc)
	assert.NoError(t, err)
	// Start the metric receiver
	go func() {
		assert.NoError(t, receiver.Start(context.Background()))
	}()
	// Gracefully shuts down the receiver
	defer func() {
		assert.NoError(t, receiver.Close())
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out, expected metric not found")
	case <-doneCh:
		t.Log("Found metric")
	}
	receiver.Close()
}
