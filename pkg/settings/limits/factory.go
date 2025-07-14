package limits

import (
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// Factory holds optional configuration for constructing [Limit]s.
type Factory struct {
	// Settings is a source of dynamic limit and burst updates.
	// [settings.Getter.GetScoped] will be polled for updates, unless Settings is also a settings.Registry, in which case
	// the channel based [settings.Registry.SubscribeScoped] will be used instead.
	Settings settings.Getter // optional

	// Meter is an optional way to emit Open Telemetry metrics.
	Meter metric.Meter // optional

	// Logger is used when parsing fails and a limit falls back to the default value.
	Logger logger.Logger // optional
}

// NewRateLimiter creates a RateLimiter for the given settings.Setting and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//  - rate.*.limit - float gauge
//  - rate.*.burst - int gauge
//  - rate.*.usage - int counter
func (f Factory) NewRateLimiter(rate settings.Setting[config.Rate]) (RateLimiter, error) {
	if rate.Scope == settings.ScopeGlobal {
		return f.globalRateLimiter(rate)
	}
	return f.newScopedRateLimiter(rate)
}

// NewTimeLimiter returns a TimeLimiter for given timeout, and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//  - time.*.limit - float gauge
//  - time.*.usage - float gauge
//  - time.*.timeout - int counter
//  - time.*.success - int counter
// Note: Unit will be ignored. All TimeLimiters emit seconds as "s".
func (f Factory) NewTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	return f.newTimeLimiter(timeout)
}

// NewResourcePoolLimiter returns a ResourcePoolLimiter for the given limit, and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//  - resource.*.limit - gauge
//  - resource.*.usage - gauge
func NewResourcePoolLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourcePoolLimiter[N], error) {
	if limit.Scope == settings.ScopeGlobal {
		return newGlobalResourcePoolLimiter(f, limit)
	}
	return newScopedResourcePoolLimiterFromFactory(f, limit)
}
