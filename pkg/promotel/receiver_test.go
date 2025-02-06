package promotel_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/smartcontractkit/chainlink-common/pkg/promotel"
)

// TestPrometheusReceiver verifies the initialization, startup, and shutdown
// processes of the Prometheus receiver. It ensures that no errors occur when
// creating a metrics receiver from a loaded configuration, starting it, and
// gracefully stopping it.
func TestPrometheusReceiver(t *testing.T) {
	// Load configuration from a YAML file
	configFile := filepath.Join("testdata", "promconfig.yaml")
	testConfig, err := promotel.LoadTestConfig(configFile, "withOnlyScrape")
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

func TestMetricReceiver(t *testing.T) {
	configFile := filepath.Join("testdata", "promconfig.yaml")
	testConfig, err := promotel.LoadTestConfig(configFile, "withOnlyScrape")
	require.NoError(t, err)
	noopConsumerFunc := func(context.Context, pmetric.Metrics) error { return nil }
	receiver, err := promotel.NewMetricReceiver(testConfig, prometheus.DefaultGatherer, prometheus.DefaultRegisterer, noopConsumerFunc, nil)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background()))
	require.NoError(t, receiver.Close())
}
