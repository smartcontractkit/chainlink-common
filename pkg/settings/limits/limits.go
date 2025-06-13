// Package limits helps enforce request-scoped, multi-tenant limits with three kinds of [Limiter]:
//  - [RateLimiter]: for throttling access
//  - [ResourceLimiter]: for allocating resources
//  - [TimeLimiter]: for enforcing timeouts
//
// Every limit requires a default value. Additional features like Otel metrics and dynamic updates are available
// via [settings.Setting]s.
//
// Limiter errors are GRPC [codes.ResourceExhausted].
package limits

import (
	"io"
	"time"

	"golang.org/x/exp/constraints"
)

// Number includes all integer and float types, although metrics will be emitted either as int64 or float64.
type Number interface {
	constraints.Integer | constraints.Float
}

type Limiter interface {
	io.Closer // Limiters spawn background goroutines and must be closed.
}

// pollPeriod is how often settings are refreshed via [settings.Getter.GetScoped]
const pollPeriod = 5 * time.Second
