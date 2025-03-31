package promhealth

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

var (
	healthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health",
			Help: "Health status by service",
		},
		[]string{"service_id"},
	)
	uptimeSeconds = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "uptime_seconds",
			Help: "Uptime of the application measured in seconds",
		},
	)
	version = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "version",
			Help: "Application version information",
		},
		[]string{"version", "commit"},
	)
)

// NewChecker returns a *services.HealthChecker with hooks for prometheus metrics.
func NewChecker(ver, sha string) *services.HealthChecker {
	return services.HealthCheckerConfig{
		Ver: ver,
		Sha: sha,
		AddUptime: func(d time.Duration) {
			uptimeSeconds.Add(d.Seconds())
		},
		IncVersion: func(ver string, sha string) {
			version.WithLabelValues(ver, sha).Inc()
		},
		SetStatus: func(name string, value int) {
			healthStatus.WithLabelValues(name).Set(float64(value))
		},
		Delete: func(name string) {
			healthStatus.DeleteLabelValues(name)
		},
	}.New()
}
