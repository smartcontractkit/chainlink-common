package prometheusreceiver_test

import (
	"context"
	"testing"
	"time"

	promcfg "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	promreceiver "github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheusreceiver"
)

func TestReceiverEndToEnd(t *testing.T) {
	//cfg, err := setupTestConfig("127.0.0.1:8888", "/metrics")
	//assert.NoError(t, err)
	ctx := context.Background()
	config := &promreceiver.Config{
		PrometheusConfig:     (*promreceiver.PromConfig)(&promcfg.Config{}),
		StartTimeMetricRegex: "",
	}

	cms := new(consumertest.MetricsSink)
	receiver := promreceiver.NewPrometheusReceiver(receivertest.NewNopSettings(), config, cms)

	require.NoError(t, receiver.Start(ctx, componenttest.NewNopHost()))
	// verify state after shutdown is called
	t.Cleanup(func() {
		// verify state after shutdown is called
		require.NoError(t, receiver.Shutdown(context.Background()))
		// assert.Empty(t, flattenTargets(receiver.scrapeManager.TargetsAll()), "expected scrape manager to have no targets")
	})
	// Wait for some scrape results to be collected
	assert.Eventually(t, func() bool {
		// This is the receiver's pov as to what should have been collected from the server
		metrics := cms.AllMetrics()
		return len(metrics) > 0
	}, 30*time.Second, 500*time.Millisecond)

	// This begins the processing of the scrapes collected by the receiver
	metrics := cms.AllMetrics()
	// split and store results by target name
	pResults := splitMetricsByTarget(metrics)
	for _, scrapes := range pResults {
		assert.NotEmpty(t, scrapes)
		for _, scrape := range scrapes {
			// Verify that each scrape contains expected metrics
			ilms := scrape.ScopeMetrics()
			for j := 0; j < ilms.Len(); j++ {
				metrics := ilms.At(j).Metrics()
				assert.NotEmpty(t, metrics, "expected non-empty metrics")
				for k := 0; k < metrics.Len(); k++ {
					metric := metrics.At(k)
					assert.NotEmpty(t, metric.Name(), "expected metric to have a name")
				}
			}
		}
	}
}

func splitMetricsByTarget(metrics []pmetric.Metrics) map[string][]pmetric.ResourceMetrics {
	pResults := make(map[string][]pmetric.ResourceMetrics)
	for _, md := range metrics {
		rms := md.ResourceMetrics()
		for i := 0; i < rms.Len(); i++ {
			name, _ := rms.At(i).Resource().Attributes().Get("service.name")
			pResults[name.AsString()] = append(pResults[name.AsString()], rms.At(i))
		}
	}
	return pResults
}
