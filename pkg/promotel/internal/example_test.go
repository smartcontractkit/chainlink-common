package internal_test

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

func TestExample(t *testing.T) {
	var (
		g              = prometheus.DefaultGatherer
		r              = prometheus.DefaultRegisterer
		logger         = logger.Test(t)
		timeout        = 10 * time.Second
		testMetricName = "test_counter_metric"
		doneCh         = make(chan struct{})
		interval       = 10 * time.Millisecond
	)
	if deadline, ok := t.Deadline(); !ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Report metrics
	go internal.ReportTestMetrics(ctx, r, testMetricName)

	// Create exporter
	expConfig, err := internal.NewDefaultExporterConfig()
	require.NoError(t, err)

	// Sends metrics data in OTLP format to otel-collector endpoint
	exporter, err := internal.NewMetricExporter(expConfig, logger)
	require.NoError(t, err)

	// Create receiver
	config, err := internal.NewReceiverConfig()
	require.NoError(t, err)
	nextFunc := func(ctx context.Context, md pmetric.Metrics) error {
		if internal.FindExpectedMetric(testMetricName, md) {
			doneCh <- struct{}{}
		}
		return exporter.Export(ctx, md)
	}
	receiver, err := internal.NewMetricReceiver(config, g, r, interval, logger, nextFunc)
	require.NoError(t, err)

	// Start exporter
	go func() {
		assert.NoError(t, exporter.Start(ctx))
	}()
	// Gracefully shuts down the exporter
	defer func() {
		assert.NoError(t, exporter.Close())
	}()

	// Start receiver
	go func() {
		assert.NoError(t, receiver.Start(ctx))
	}()
	// Gracefully shuts down the receiver
	defer func() {
		assert.NoError(t, receiver.Close())
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Timed out waiting for metric")
	case <-doneCh:
		t.Log("Found metric")
	}
}
