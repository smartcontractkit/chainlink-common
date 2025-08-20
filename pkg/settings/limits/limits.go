// Package limits helps enforce request-scoped, multi-tenant limits with three kinds of [Limiter]:
//  - [RateLimiter]: for throttling usage
//  - [ResourceLimiter]/[ResourcePoolLimiter]: for allocating resources
//  - [TimeLimiter]: for enforcing timeouts
//  - [BoundLimiter]: for enforcing bounds
//  - [QueueLimiter]: for limited capacity queues
//
// Every limit requires a default value. Additional features like Otel metrics and dynamic updates are available by
// using the [settings.Setting] variants.
//
// Limiter errors are GRPC [codes.ResourceExhausted] and [codes.DeadlineExceeded].
package limits

import (
	"context"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

// Number includes all integer and float types, although metrics will be emitted either as int64 or float64.
type Number interface {
	constraints.Integer | constraints.Float
}

type Limiter[N any] interface {
	io.Closer // Limiters spawn background goroutines and must be closed.
	// Limit returns the current limit.
	Limit(context.Context) (N, error)
}

// pollPeriod is how often settings are refreshed via [settings.Getter.GetScoped]
var pollPeriod = 5 * time.Second // reduced for tests

func maxVal[N Number]() (n N, err error) {
	val := reflect.ValueOf(n)
	switch val.Kind() {
	case reflect.Float64:
		val.SetFloat(math.MaxFloat64)
	case reflect.Float32:
		val.SetFloat(math.MaxFloat32)
	case reflect.Int64:
		val.SetInt(math.MaxInt64)
	case reflect.Int32:
		val.SetInt(math.MaxInt32)
	case reflect.Int16:
		val.SetInt(math.MaxInt16)
	case reflect.Int8:
		val.SetInt(math.MaxInt8)
	case reflect.Int:
		val.SetInt(math.MaxInt)
	case reflect.Uint64:
		val.SetUint(math.MaxUint64)
	case reflect.Uint32:
		val.SetUint(math.MaxUint32)
	case reflect.Uint16:
		val.SetUint(math.MaxUint16)
	case reflect.Uint8:
		val.SetUint(math.MaxUint8)
	case reflect.Uint:
		val.SetUint(math.MaxUint)
	default:
		return 0, fmt.Errorf("unsupported kind %s for type %T", val.Kind(), n)
	}
	return
}
