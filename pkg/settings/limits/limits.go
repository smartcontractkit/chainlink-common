// Package limits helps enforce request-scoped, multi-scope limits.
package limits

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/exp/constraints"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// Number includes all integer and float types, although metrics will be emitted either as int64 or float64.
type Number interface {
	constraints.Integer | constraints.Float
}

// A Limiter limits usage of a resource by tracking allocations and ensuring that the threshold is not exceeded.
// TODO Allocator.Alloc?
// TODO Resource.Take? Return?
type Limiter[N Number] interface {
	// Use reserves amount of resources, or returns an error. The free function must be
	// called to release resources, and should be deferred immediately when possible.
	// Execution is typically blocked until resources are available (or context expires), but may short-circuit.
	// ErrorLimitReached is returned when a limit is reached.
	Use(ctx context.Context, amount N) (free func(context.Context), err error)
}

// Config holds optional configuration fields for Limiters.
type Config[N Number] struct {
	// GetLimit should return the current limit, for the given settings.Scope scope, or an error if none is set.
	GetLimit settings.GetScoped[N]

	// GaugeFn is an optional way to emit limit and usage metrics.
	// - resource.*.limit
	// - resource.*.used
	GaugeFn func(string) (Gauge[N], error)

	// If ShortCircuit is true, Limiter.Use will return immediately when there are not enough resources available,
	// instead of waiting for them to become free.
	ShortCircuit bool
}

// UnscopedLimiter returns an unscoped Limiter with default options.
// See Config.UnscopedLimiter for dynamic limits, metering, and more.
func UnscopedLimiter[N Number](key string, defaultLimit N) (Limiter[N], error) {
	return Config[N]{}.UnscopedLimiter(key, defaultLimit)
}

// UnscopedLimiter returns an unscoped Limiter with the given Configuration.
func (c Config[N]) UnscopedLimiter(key string, defaultLimit N) (Limiter[N], error) {
	l := &unscopedLimiter[N]{
		limiter: limiter[N]{
			key:          key,
			defaultLimit: defaultLimit,
			shortCircuit: c.ShortCircuit,
			getLimitFn: func(ctx context.Context) (N, error) {
				return c.GetLimit(ctx, settings.ScopeGlobal, key)
			},
		},
	}
	return l, l.init(c.GaugeFn)
}

type limiter[N Number] struct {
	key          string // required
	defaultLimit N
	shortCircuit bool
	getLimitFn   func(context.Context) (N, error)

	recordUsage Gauge[N]
	recordLimit Gauge[N]

	mu   sync.RWMutex
	cond sync.Cond
}

func (l *limiter[N]) init(factoryFn func(string) (Gauge[N], error)) error {
	if factoryFn == nil {
		return nil
	}
	var err error
	l.recordLimit, err = factoryFn("resource." + l.key + ".limit")
	if err != nil {
		return err
	}
	l.recordUsage, err = factoryFn("resource." + l.key + ".used")
	if err != nil {
		return err
	}
	return nil
}

func (l *limiter[N]) getLimit(ctx context.Context) N {
	limit, err := l.getLimitFn(ctx)
	if err != nil {
		//TODO log about fallback to default?
		limit = l.defaultLimit
	}
	l.recordLimit(ctx, limit) //TODO can't from here w/o attributes....
	return limit
}

type unscopedLimiter[N Number] struct {
	limiter[N]

	used N
}

func (g *unscopedLimiter[N]) Use(ctx context.Context, amount N) (func(context.Context), error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	limit := g.getLimit(ctx)

	for g.used+amount > limit {
		if g.shortCircuit {
			return nil, ErrorLimitReached[N]{
				key:    g.key,
				used:   g.used,
				limit:  limit,
				amount: amount,
			}
		}
		g.cond.Wait()
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	g.used += amount
	return func(ctx context.Context) { g.free(ctx, amount) }, nil
}

func (g *unscopedLimiter[N]) free(ctx context.Context, amount N) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.used -= amount
	g.recordUsage(ctx, g.used)
	//TODO sanity check >0?

	//TODO cache limit instead? (avoids risk of wrongly scoped context)
	if g.used < g.getLimit(ctx) {
		g.cond.Broadcast()
	}
	return
}

func OrgLimiter[N Number](key string, defaultValue N) (Limiter[N], error) {
	return Config[N]{}.newScopedLimiter(settings.ScopeOrg, key, defaultValue)
}

func (c Config[N]) OrgLimiter(key string, defaultValue N) (Limiter[N], error) {
	return c.newScopedLimiter(settings.ScopeOrg, key, defaultValue)
}

func UserLimiter[N Number](key string, defaultValue N) (Limiter[N], error) {
	return Config[N]{}.newScopedLimiter(settings.ScopeUser, key, defaultValue)
}

func (c Config[N]) UserLimiter(key string, defaultValue N) (Limiter[N], error) {
	return c.newScopedLimiter(settings.ScopeUser, key, defaultValue)
}

func WorkflowLimiter[N Number](key string, defaultValue N) (Limiter[N], error) {
	return Config[N]{}.newScopedLimiter(settings.ScopeWorkflow, key, defaultValue)
}

func (c Config[N]) WorkflowLimiter(key string, defaultValue N) (Limiter[N], error) {
	return c.newScopedLimiter(settings.ScopeWorkflow, key, defaultValue)
}

// TODO export?
func (c Config[N]) newScopedLimiter(scope settings.Scope, key string, defaultValue N) (*scopedLimiter[N], error) {
	l := &scopedLimiter[N]{
		limiter: limiter[N]{
			key:          key,
			defaultLimit: defaultValue,
			shortCircuit: c.ShortCircuit,
			getLimitFn: func(ctx context.Context) (N, error) {
				return c.GetLimit(ctx, scope, key)
			},
		},
		scope: scope,
	}
	return l, l.init(c.GaugeFn)
}

// scopedLimiter extends limiter with a settings.Scope and enforces limits for each tenant separately.
type scopedLimiter[N Number] struct {
	limiter[N]
	scope settings.Scope

	used map[string]N
}

func (l *scopedLimiter[N]) Use(ctx context.Context, amount N) (func(context.Context), error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	limit := l.getLimit(ctx)

	tenant := l.scope.Value(ctx)
	if tenant == "" {
		//TODO fallback to global?
		return nil, fmt.Errorf("missing scope: %s", l.scope)
	}

	used := l.used[tenant]
	for used+amount > limit {
		if l.shortCircuit {
			return nil, ErrorLimitReached[N]{
				key:    l.key,
				used:   used,
				limit:  limit,
				amount: amount,
			}
		}
		l.cond.Wait()
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		used = l.used[tenant] // try again
	}
	l.used[tenant] = used + amount
	return func(ctx context.Context) { l.free(ctx, tenant, amount) }, nil
}

func (l *scopedLimiter[N]) free(ctx context.Context, tenant string, amount N) {
	l.mu.Lock()
	defer l.mu.Unlock()

	used := l.used[tenant]
	used -= amount
	if used == 0 {
		delete(l.used, tenant)
	} else {
		l.used[tenant] = used
		l.recordUsage(ctx, used, metric.WithAttributes(attribute.String(l.scope.String(), tenant)))
	} //TODO sanity check >0?

	//TODO cache limit instead? (avoids risk of wrongly scoped context)
	if used < l.getLimit(ctx) {
		l.cond.Broadcast()
	}
	return
}

type ErrorLimitReached[N Number] struct {
	key                 string
	used, limit, amount N
}

func (e ErrorLimitReached[N]) Error() string {
	return fmt.Sprintf("limit reached for key %s: cannot use %v, already using %v/%v", e.key, e.amount, e.used, e.limit)
}

//TODO layered cases: "MultiLimit" of Workflow, then User, then Org, then Global
