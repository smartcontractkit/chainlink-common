package prometheusreceiver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	promcfg "github.com/prometheus/prometheus/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	promreceiver "github.com/smartcontractkit/chainlink-common/pkg/promotel/prometheusreceiver"
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
		if len(metrics) > 0 {
			// If we don't have enough scrapes yet lets return false and wait for another tick
			return true
		}
		return false
	}, 30*time.Second, 500*time.Millisecond)

	// This begins the processing of the scrapes collected by the receiver
	metrics := cms.AllMetrics()
	// split and store results by target name
	pResults := splitMetricsByTarget(metrics)
	for name, scrapes := range pResults {
		// validate scrapes here
		fmt.Printf("name %s, \nscrapes %+v", name, scrapes)
		assert.NotEmpty(t, scrapes)
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
