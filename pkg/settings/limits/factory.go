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

// Deprecated: use MakeRateLimiter
func (f Factory) NewRateLimiter(rate settings.Setting[config.Rate]) (RateLimiter, error) {
	return f.MakeRateLimiter(rate)
}

// MakeRateLimiter creates a RateLimiter for the given rate and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//   - rate.*.limit - float gauge
//   - rate.*.burst - int gauge
//   - rate.*.usage - int counter
//   - rate.*.denied - int histogram
func (f Factory) MakeRateLimiter(rate settings.Setting[config.Rate]) (RateLimiter, error) {
	if rate.Scope == settings.ScopeGlobal {
		return f.globalRateLimiter(rate)
	}
	return f.newScopedRateLimiter(rate)
}

// Deprecated: use MakeTimeLimiter
func (f Factory) NewTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	return f.newTimeLimiter(timeout)
}

// MakeTimeLimiter returns a TimeLimiter for given timeout, and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//   - time.*.limit - float gauge
//   - time.*.runtime - float gauge
//   - time.*.success - int counter
//   - time.*.timeout - int counter
//
// Note: Unit will be ignored. All TimeLimiters emit seconds as "s".
func (f Factory) MakeTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	return f.newTimeLimiter(timeout)
}

// Deprecated: use MakeResourcePoolLimiter
func NewResourcePoolLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourcePoolLimiter[N], error) {
	return MakeResourcePoolLimiter(f, limit)
}

// MakeResourcePoolLimiter returns a ResourcePoolLimiter for the given limit, and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//   - resource.*.limit - gauge
//   - resource.*.usage - gauge
//   - resource.*.amount - histogram
//   - resource.*.denied - histogram
func MakeResourcePoolLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourcePoolLimiter[N], error) {
	if limit.Scope == settings.ScopeGlobal {
		return newGlobalResourcePoolLimiter(f, limit)
	}
	return newScopedResourcePoolLimiterFromFactory(f, limit)
}

// MakeBoundLimiter returns a BoundLimiter for the given bound and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//   - bound.*.limit - gauge
//   - bound.*.usage - histogram
//   - bound.*.denied - histogram
func MakeBoundLimiter[N Number](f Factory, bound settings.IsSetting[N]) (BoundLimiter[N], error) {
	return newBoundLimiter(f, bound.GetSpec())
}

// MakeQueueLimiter returns a QueueLimiter for the given limit and configured by the Factory.
// If Meter is set, the following metrics will be emitted
//   - queue.*.limit - int gauge
//   - queue.*.usage - int gauge
//   - queue.*.denied - int histogram
func MakeQueueLimiter[T any](f Factory, limit settings.Setting[int]) (QueueLimiter[T], error) {
	if limit.Scope == settings.ScopeGlobal {
		return newUnscopedQueue[T](f, limit)
	}
	return newScopedQueue[T](f, limit)
}

// MakeGateLimiter returns a GateLimiter for the given limit and configured by the factory.
// If Meter is set, the following metrics will be emitted
//   - gate.*.limit - int gauge
//   - gate.*.usage - int counter
//   - gate.*.denied - int counter
func MakeGateLimiter(f Factory, limit settings.IsSetting[bool]) (GateLimiter, error) {
	return newGateLimiter(f, limit.GetSpec())
}
