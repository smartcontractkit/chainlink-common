package otelhealth

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

func NewChecker(ver, sha string, meter metric.Meter) (*services.HealthChecker, error) {
	cfg, err := ConfigureHooks(services.HealthCheckerConfig{Ver: ver, Sha: sha}, meter)
	if err != nil {
		return nil, err
	}
	return cfg.New(), nil
}

func ConfigureHooks(orig services.HealthCheckerConfig, meter metric.Meter) (services.HealthCheckerConfig, error) {
	cfg := orig // copy
	healthStatus, err := meter.Int64Gauge("health", metric.WithDescription("Health status by service"))
	if err != nil {
		return services.HealthCheckerConfig{}, err
	}
	version, err := meter.Int64Counter("version", metric.WithDescription("Version and SHA of the service"))
	if err != nil {
		return services.HealthCheckerConfig{}, err
	}
	uptimeSeconds, err := meter.Float64Counter("uptime_seconds", metric.WithDescription("Uptime of the service in seconds"))
	if err != nil {
		return services.HealthCheckerConfig{}, err
	}
	cfg.AddUptime = func(ctx context.Context, d time.Duration) {
		if orig.AddUptime != nil {
			orig.AddUptime(ctx, d)
		}
		uptimeSeconds.Add(ctx, d.Seconds())
	}
	cfg.IncVersion = func(ctx context.Context, ver string, sha string) {
		if orig.IncVersion != nil {
			orig.IncVersion(ctx, ver, sha)
		}
		version.Add(ctx, 1, metric.WithAttributes(attribute.String("version", ver), attribute.String("commit", sha)))
	}
	cfg.SetStatus = func(ctx context.Context, name string, value int) {
		if orig.SetStatus != nil {
			orig.SetStatus(ctx, name, value)
		}
		healthStatus.Record(ctx, int64(value), metric.WithAttributes(attribute.String("service_id", name)))
	}
	return cfg, nil
}
