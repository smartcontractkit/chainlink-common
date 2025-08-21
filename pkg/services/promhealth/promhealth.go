package promhealth

import (
	"context"
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
	return ConfigureHooks(services.HealthCheckerConfig{Ver: ver, Sha: sha}).New()
}

func ConfigureHooks(orig services.HealthCheckerConfig) services.HealthCheckerConfig {
	cfg := orig // copy
	cfg.AddUptime = func(ctx context.Context, d time.Duration) {
		if orig.AddUptime != nil {
			orig.AddUptime(ctx, d)
		}
		uptimeSeconds.Add(d.Seconds())
	}
	cfg.IncVersion = func(ctx context.Context, ver string, sha string) {
		if orig.IncVersion != nil {
			orig.IncVersion(ctx, ver, sha)
		}
		version.WithLabelValues(ver, sha).Inc()
	}
	cfg.SetStatus = func(ctx context.Context, name string, value int) {
		if orig.SetStatus != nil {
			orig.SetStatus(ctx, name, value)
		}
		healthStatus.WithLabelValues(name).Set(float64(value))
	}
	cfg.Delete = func(ctx context.Context, name string) {
		if orig.Delete != nil {
			orig.Delete(ctx, name)
		}
		healthStatus.DeleteLabelValues(name)
	}
	return cfg
}
