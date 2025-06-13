// Package limits helps enforce request-scoped, multi-scope limits.
package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// A ResourceLimiter limits usage by tracking allocations and ensuring that the threshold is not exceeded.
type ResourceLimiter[N Number] interface {
	// Use reserves amount of resources, or returns an error. The free function must be
	// called to release resources, and should be deferred immediately when possible.
	// Blocks until resources are available, or context has expired.
	Use(ctx context.Context, amount N) (free func(context.Context), err error)

	// TryUse is like Use, but returns immediately with ErrorLimited when resources are not available.
	//TODO consider removal
	TryUse(ctx context.Context, amount N) (free func(context.Context), err error)
}

// Resource configures a resource for limiting.
type Resource[N Number] struct {
	// GetLimit should return the current limit for the given settings.Scope scope, or an error if none is set.
	GetLimit settings.GetScoped[N]

	// GaugeFn is an optional factory for creating gauges to emit limit and usage metrics.
	// - resource.*.limit
	// - resource.*.used
	GaugeFn GaugeFactory[N]
}

// GlobalResourceLimiter returns an unscoped ResourceLimiter with default options.
// See Resource.GlobalLimiter for dynamic limits, metering, and more.
func GlobalResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newUnscopedLimiter(defaultLimit)
}

// TODO limits.ResourceLimiter{}.Global() ResourceLimiter
// TODO limits.Rate{}.Global() RateLimiter
// GlobalResourceLimiter returns an unscoped ResourceLimiter for Key with the given Configuration.
func (c Resource[N]) GlobalLimiter(key string, defaultLimit N) (ResourceLimiter[N], error) {
	l := newUnscopedLimiter(defaultLimit)
	l.key = key

	if c.GetLimit != nil {
		if key == "" {
			return nil, errors.New("key is required with GetLimit")
		}
		l.getLimitFn = func(ctx context.Context) (N, error) {
			return c.GetLimit(ctx, settings.ScopeGlobal, key)
		}
	}
	if c.GaugeFn != nil {
		if err := l.createGauges(c.GaugeFn); err != nil {
			return nil, err
		}
		l.usage = l.newLimitUsage(func(ctx context.Context, value N) {
			if l.limit.recordUsage != nil {
				l.limit.recordUsage(ctx, value)
			}
		}, func(ctx context.Context, value N) {
			if l.limit.recordLimit != nil {
				l.limit.recordLimit(ctx, value)
			}
		})
	}
	return l, nil
}

// limit holds a defaultLimit, and optional extensions.
type limit[N Number] struct {
	defaultLimit N

	key string // optional

	getLimitFn func(context.Context) (N, error) // optional

	recordUsage Gauge[N] // optional
	recordLimit Gauge[N] // optional
}

func (l *limit[N]) createGauges(gaugeFn func(string) (Gauge[N], error)) error {
	if l.key == "" {
		return errors.New("metrics require Key to be set")
	}
	var err error
	l.recordLimit, err = gaugeFn("resource." + l.key + ".limit")
	if err != nil {
		return err
	}
	l.recordUsage, err = gaugeFn("resource." + l.key + ".used")
	if err != nil {
		return err
	}
	return nil
}

func (l *limit[N]) getLimit(ctx context.Context) (limit N) {
	if l.getLimitFn == nil {
		return l.defaultLimit
	}
	limit, err := l.getLimitFn(ctx)
	if err != nil {
		//TODO log about fallback to default?
		limit = l.defaultLimit
	}
	return limit
}

type usage[N Number] struct {
	*limit[N]
	scope       settings.Scope // optional
	tenant      string         // optional
	mu          sync.Mutex
	cond        sync.Cond
	used        N
	recordUsage func(ctx context.Context, value N)
	recordLimit func(ctx context.Context, value N)
}

func (l *limit[N]) newLimitUsage(
	recordUsage func(ctx context.Context, value N),
	recordLimit func(ctx context.Context, value N),
) *usage[N] {
	u := usage[N]{
		limit:       l,
		recordUsage: recordUsage,
		recordLimit: recordLimit,
	}
	u.cond.L = &u.mu
	return &u
}

// TODO ctx is only to record usage metric - can we drop it?
func (u *usage[N]) free(ctx context.Context, amount N) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.used -= amount
	u.recordUsage(ctx, u.used)

	if u.used == 0 {
		//TODO remove self from sync.Map?
	} //TODO sanity check >0?

	u.cond.Broadcast()

	return
}

func (u *usage[N]) newErrorLimitReached(limit, amount N) ErrorLimited[N] {
	return ErrorLimited[N]{
		Key:    u.key,
		Scope:  u.scope,
		Tenant: u.tenant,
		Used:   u.used,
		Limit:  limit,
		Amount: amount,
	}
}

func (u *usage[N]) use(ctx context.Context, amount N, block bool) (func(context.Context), error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	limit := u.getLimit(ctx)
	u.recordLimit(ctx, limit)

	if u.used+amount > limit {
		if !block {
			return nil, u.newErrorLimitReached(limit, amount)
		}
		// Ensure cond.Wait() yields to context expiration
		stop := context.AfterFunc(ctx, func() {
			u.mu.Lock()
			defer u.mu.Unlock()
			u.cond.Broadcast()
		})
		defer stop()
		start := time.Now()
		for u.used+amount > limit {
			u.cond.Wait()
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("context error (%w) after waiting %s for limit: %w", err, time.Since(start), u.newErrorLimitReached(limit, amount))
			}
		}
		//TODO metric for blocking time?
	}
	u.used += amount
	u.recordUsage(ctx, u.used)
	return func(ctx context.Context) { u.free(ctx, amount) }, nil
}

type unscopedLimiter[N Number] struct {
	limit[N]

	*usage[N]
}

func newUnscopedLimiter[N Number](defaultLimit N) *unscopedLimiter[N] {
	l := &unscopedLimiter[N]{
		limit: limit[N]{
			defaultLimit: defaultLimit,
		},
	}
	noop := func(ctx context.Context, value N) {}
	l.usage = l.newLimitUsage(noop, noop)
	return l
}

func (u *unscopedLimiter[N]) Use(ctx context.Context, amount N) (func(context.Context), error) {
	return u.usage.use(ctx, amount, true)
}

func (u *unscopedLimiter[N]) TryUse(ctx context.Context, amount N) (func(context.Context), error) {
	return u.usage.use(ctx, amount, false)
}

// OrgResourceLimiter creates a new ResourceLimiter scoped per organization.
func OrgResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeOrg, "", defaultLimit)
}

// OrgLimiter creates a new ResourceLimiter scoped per organization, with the given Key and Resource.
func (c Resource[N]) OrgLimiter(key string, defaultLimit N) (ResourceLimiter[N], error) {
	return c.newScopedResource(settings.ScopeOrg, key, defaultLimit)
}

// UserResourceLimiter creates a new ResourceLimiter scoped per user.
func UserResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeUser, "", defaultLimit)
}

// UserLimiter creates a new ResourceLimiter scoped per user, with the given Key and Resource.
func (c Resource[N]) UserLimiter(key string, defaultLimit N) (ResourceLimiter[N], error) {
	return c.newScopedResource(settings.ScopeUser, key, defaultLimit)
}

// WorkflowResourceLimiter creates a new ResourceLimiter scoped per workflow.
func WorkflowResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeWorkflow, "", defaultLimit)
}

// WorkflowLimiter creates a new ResourceLimiter scoped per workflow, with the given Key and Resource.
func (c Resource[N]) WorkflowLimiter(key string, defaultLimit N) (ResourceLimiter[N], error) {
	return c.newScopedResource(settings.ScopeWorkflow, key, defaultLimit)
}

func newScopedResource[N Number](scope settings.Scope, key string, defaultLimit N) *scopedResource[N] {
	l := &scopedResource[N]{
		limit: limit[N]{
			key:          key,
			defaultLimit: defaultLimit,
		},
		scope: scope,
	}
	return l
}

func (c Resource[N]) newScopedResource(scope settings.Scope, key string, defaultLimit N) (*scopedResource[N], error) {
	l := newScopedResource(scope, key, defaultLimit)

	if c.GetLimit != nil {
		if key == "" {
			return nil, errors.New("key is required with GetLimit")
		}
		l.getLimitFn = func(ctx context.Context) (N, error) {
			return c.GetLimit(ctx, scope, key)
		}
	}
	if c.GaugeFn != nil {
		if err := l.createGauges(c.GaugeFn); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// scopedResource extends limit with a settings.Scope and enforces limits for each tenant separately.
type scopedResource[N Number] struct {
	limit[N]
	scope settings.Scope

	used sync.Map // map[string]usage[N]
}

func (s *scopedResource[N]) Use(ctx context.Context, amount N) (func(context.Context), error) {
	return s.use(ctx, amount, true)
}
func (s *scopedResource[N]) TryUse(ctx context.Context, amount N) (func(context.Context), error) {
	return s.use(ctx, amount, false)
}

func (s *scopedResource[N]) use(ctx context.Context, amount N, block bool) (func(context.Context), error) {
	tenant := s.scope.Value(ctx)
	if tenant == "" {
		//TODO fallback to global?
		return nil, fmt.Errorf("missing scope: %s", s.scope)
	}

	loaded, _ := s.used.LoadOrStore(tenant, s.newLimitUsage(tenant))
	usage := loaded.(*usage[N])
	return usage.use(ctx, amount, block)
}

func (s *scopedResource[N]) newLimitUsage(tenant string) *usage[N] {
	u := s.limit.newLimitUsage(s.recordScoped(tenant))
	u.scope = s.scope
	u.tenant = tenant
	return u
}

func (s *scopedResource[N]) recordScoped(tenant string) (usage, limit func(context.Context, N)) {
	if s.recordUsage == nil {
		usage = func(ctx context.Context, n N) {}
	} else {
		usage = func(ctx context.Context, value N) {
			s.recordUsage(ctx, value, metric.WithAttributes(attribute.String(s.scope.String(), tenant)))
		}
	}
	if s.recordLimit == nil {
		limit = func(ctx context.Context, n N) {}
	} else {
		limit = func(ctx context.Context, value N) {
			s.recordLimit(ctx, value, metric.WithAttributes(attribute.String(s.scope.String(), tenant)))
		}
	}
	return
}

type ErrorLimited[N Number] struct {
	Key                 string
	Scope               settings.Scope
	Tenant              string
	Used, Limit, Amount N
}

func (e ErrorLimited[N]) GRPCStatus() *status.Status {
	return status.New(codes.ResourceExhausted, e.Error())
}

func (e ErrorLimited[N]) Is(err error) bool {
	_, ok := err.(ErrorLimited[N]) //nolint:errcheck // implementing errors.Is
	//TODO also match grpc over the wire
	return ok
}

func (e ErrorLimited[N]) Error() string {
	scope := e.Scope.String()
	if e.Tenant != "" {
		scope += "[" + e.Tenant + "]"
	}
	var forKey string
	if e.Key != "" {
		forKey = fmt.Sprintf(" for key %s", e.Key)
	}
	return fmt.Sprintf("%s resource limited%s: cannot use %v, already using %v/%v", scope, forKey, e.Amount, e.Used, e.Limit)
}

// MultiResourceLimiter is a ResourceLimiter backed by other limiters, which are each called in order.
type MultiResourceLimiter[N Number] []ResourceLimiter[N]

func (m MultiResourceLimiter[N]) Use(ctx context.Context, amount N) (func(context.Context), error) {
	var frees freeFns
	for _, l := range m {
		free, err := l.Use(ctx, amount)
		if err != nil {
			frees.freeAll(ctx)
			return nil, err
		}
		frees = append(frees, free)
	}
	return frees.freeAll, nil
}

func (m MultiResourceLimiter[N]) TryUse(ctx context.Context, amount N) (func(context.Context), error) {
	var frees freeFns
	for _, l := range m {
		free, err := l.TryUse(ctx, amount)
		if err != nil {
			frees.freeAll(ctx)
			return nil, err
		}
		frees = append(frees, free)
	}
	return frees.freeAll, nil
}

type freeFns []func(context.Context)

func (f freeFns) freeAll(ctx context.Context) {
	for i := len(f) - 1; i >= 0; i-- {
		f[i](ctx)
	}
}
