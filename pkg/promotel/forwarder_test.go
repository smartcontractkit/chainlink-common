package promotel_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/promotel"
	internal "github.com/smartcontractkit/chainlink-common/pkg/promotel/internal"
)

func TestExample(t *testing.T) {
	var (
		g = prometheus.DefaultGatherer
		r = prometheus.DefaultRegisterer
		// todo: use logger.TestObserved
		lggr, observed = logger.TestObserved(t, zap.DebugLevel)
		testMetricName = "test_counter_metric"
		interval       = 10 * time.Millisecond
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go internal.ReportTestMetrics(ctx, r, testMetricName)

	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, l := range observed.All() {
					metricName, ok := l.ContextMap()["name"].(string)
					if ok && strings.Contains(metricName, testMetricName) {
						doneCh <- struct{}{}
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	srv := httptest.NewServer(nil)
	// TODO: add mocked GRPC endpoint for exporter
	forwarder, err := promotel.NewForwarder(g, r, lggr, promotel.ForwarderOptions{
		Endpoint:    srv.URL,
		TLSInsecure: true,
		Verbose:     true,
		Interval:    interval,
	})
	require.NoError(t, err)
	// Start the forwarder
	go func() {
		assert.NoError(t, forwarder.Start(ctx))
	}()
	// Gracefully shuts down the forwarder
	defer func() {
		assert.NoError(t, forwarder.Close())
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out. Expected metric not found")
	case <-doneCh:
		t.Log("Found metric.")
	}
}
