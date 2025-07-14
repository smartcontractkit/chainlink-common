// Package limits helps enforce request-scoped, multi-tenant limits with three kinds of [Limiter]:
//  - [RateLimiter]: for throttling usage
//  - [ResourceLimiter]/[ResourcePoolLimiter]: for allocating resources
//  - [TimeLimiter]: for enforcing timeouts
//
// Every limit requires a default value. Additional features like Otel metrics and dynamic updates are available by
// using the [settings.Setting] variants.
//
// Limiter errors are GRPC [codes.ResourceExhausted] and [codes.DeadlineExceeded].
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
var pollPeriod = 5 * time.Second // reduced for tests
