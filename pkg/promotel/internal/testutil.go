package internal

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func FindExpectedMetric(name string, md pmetric.Metrics) bool {
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

func ReportTestMetrics(ctx context.Context, reg prometheus.Registerer, metricName string) {
	m := promauto.With(reg).NewCounter(prometheus.CounterOpts{Name: metricName})
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m.Inc()
			time.Sleep(1 * time.Second)
		}
	}

}
