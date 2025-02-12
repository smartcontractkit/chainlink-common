package promotel_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
		testMetricName = t.Name() + "_test_counter_metric"
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

	forwarder, err := promotel.NewForwarder(g, r, lggr, promotel.ForwarderOptions{
		Endpoint:    "localhost:4317",
		TLSInsecure: true,
		Interval:    interval,
		Verbose:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = forwarder.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer forwarder.Close()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out. Expected metric not found")
	case <-doneCh:
		t.Log("Found metric.")
	}
}
