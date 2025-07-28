package limits

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// ResourceLimiter is a limiter for resources, where each interaction is typically single-action.
type ResourceLimiter[N Number] interface {
	Limiter
	// Use increases the resource count by amount, or returns an error if the limit is reached.
	// It does not block. Use a ResourcePoolLimiter for blocking semantics.
	Use(ctx context.Context, amount N) error
	// Free is the counterpart to Use and releases amount of resources from use.
	Free(ctx context.Context, amount N) error
}

// ResourcePoolLimiter is a limiter for a pool of resources, with concurrent active use, and extends the ResourceLimiter
// API with a Wait method to simplify the typical two-step interaction via a free func() to return resources to the pool.
type ResourcePoolLimiter[N Number] interface {
	ResourceLimiter[N]
	// Wait is like Use, but blocks until resources are available, or context has expired. The free func must be
	// called and should be deferred immediately when possible. It effectively calls Free to release N resources.
	Wait(context.Context, N) (free func(), err error)
}

// GlobalResourcePoolLimiter returns an unscoped ResourcePoolLimiter with default options.
// See [NewResourcePoolLimiter] for dynamic limits, metering, and more.
func GlobalResourcePoolLimiter[N Number](limit N) ResourcePoolLimiter[N] {
	return newUnscopedResourcePoolLimiter(limit)
}

// newGlobalResourcePoolLimiter returns an unscoped ResourcePoolLimiter for Key with the given Configuration.
func newGlobalResourcePoolLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourcePoolLimiter[N], error) {
	l := newUnscopedResourcePoolLimiter(limit.DefaultValue)
	l.key = limit.Key

	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
		l.lggr = logger.Sugared(f.Logger).Named("ResourcePoolLimiter").With("key", limit.Key)
	}

	if f.Meter != nil {
		if err := l.createGauges(f.Meter, limit.Unit); err != nil {
			return nil, err
		}
		l.resourcePoolUsage = l.newLimitUsage(func(ctx context.Context, value N) {
			if l.resourcePoolLimiter.recordUsage != nil {
				l.resourcePoolLimiter.recordUsage(ctx, value)
			}
		}, func(ctx context.Context, value N) {
			if l.resourcePoolLimiter.recordLimit != nil {
				l.resourcePoolLimiter.recordLimit(ctx, value)
			}
		}, func(ctx context.Context, value time.Duration) {
			if l.resourcePoolLimiter.recordBlockTime != nil {
				l.resourcePoolLimiter.recordBlockTime(ctx, int64(value))
			}
		})
	}

	if f.Settings != nil {
		l.getLimitFn = func(ctx context.Context) (N, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if registry, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (updates <-chan settings.Update[N], cancelSub func()) {
				return limit.Subscribe(ctx, registry)
			}

		}
	}

	l.cre.Store(contexts.CRE{})
	go l.updateLoop(contexts.CRE{})

	return l, nil
}

type resourcePoolLimiter[N Number] struct {
	*updater[N]

	key string // optional

	recordUsage     func(ctx context.Context, value N, options ...metric.RecordOption)    // optional
	recordLimit     func(ctx context.Context, value N, options ...metric.RecordOption)    // optional
	recordBlockTime func(ctx context.Context, incr int64, options ...metric.RecordOption) // optional
	recordAmount    func(ctx context.Context, value N, options ...metric.RecordOption)    // optional
}

func (l *resourcePoolLimiter[N]) createGauges(meter metric.Meter, unit string) error {
	if l.key == "" {
		return errors.New("metrics require Key to be set")
	}
	var gaugeFn func(key string) (gauge[N], error)
	var histogramFn func(key string) (histogram[N], error)
	var n N
	if k := reflect.TypeOf(n).Kind(); k == reflect.Float64 || k == reflect.Float32 {
		gaugeFn = func(key string) (gauge[N], error) {
			g, err := meter.Float64Gauge(key, metric.WithUnit(unit))
			return &floatRecorder[N]{g}, err
		}
		histogramFn = func(key string) (histogram[N], error) {
			g, err := meter.Float64Histogram(key, metric.WithUnit(unit))
			return &floatRecorder[N]{g}, err
		}
	} else {
		gaugeFn = func(key string) (gauge[N], error) {
			g, err := meter.Int64Gauge(key, metric.WithUnit(unit))
			return &intRecorder[N]{g}, err
		}
		histogramFn = func(key string) (histogram[N], error) {
			g, err := meter.Int64Histogram(key, metric.WithUnit(unit))
			return &intRecorder[N]{g}, err
		}
	}

	limitGauge, err := gaugeFn("resource." + l.key + ".limit")
	if err != nil {
		return err
	}
	l.recordLimit = func(ctx context.Context, value N, options ...metric.RecordOption) {
		limitGauge.Record(ctx, value, options...)
	}
	usageGauge, err := gaugeFn("resource." + l.key + ".usage")
	if err != nil {
		return err
	}
	l.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {
		usageGauge.Record(ctx, value, options...)
	}
	amountHistogram, err := histogramFn("resource." + l.key + ".amount")
	if err != nil {
		return err
	}
	l.recordAmount = func(ctx context.Context, value N, options ...metric.RecordOption) {
		amountHistogram.Record(ctx, value, options...)
	}
	return nil
}

func (l *resourcePoolLimiter[N]) getLimit(ctx context.Context) (limit N) {
	limit, err := l.getLimitFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to get limit. Using default value", "default", limit, "err", err)
	}
	return limit
}

type resourcePoolUsage[N Number] struct {
	*resourcePoolLimiter[N]
	scope  settings.Scope // optional
	tenant string         // optional
	mu     sync.Mutex
	cond   sync.Cond
	used   N

	recordLimit   func(ctx context.Context, value N)
	recordUsage   func(ctx context.Context, value N)
	recordBlocked func(ctx context.Context, value time.Duration)

	stopOnce  sync.Once
	stopCh    services.StopChan
	done      chan struct{}
	cancelSub func() // optional
}

func (l *resourcePoolLimiter[N]) newLimitUsage(
	recordUsage func(ctx context.Context, value N),
	recordLimit func(ctx context.Context, value N),
	recordBlocked func(ctx context.Context, value time.Duration),
) *resourcePoolUsage[N] {
	u := resourcePoolUsage[N]{
		resourcePoolLimiter: l,
		recordUsage:         recordUsage,
		recordLimit:         recordLimit,
		recordBlocked:       recordBlocked,
		stopCh:              make(services.StopChan),
		done:                make(chan struct{}),
	}
	u.cond.L = &u.mu
	return &u
}

func (u *resourcePoolUsage[N]) free(amount N) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.used -= amount
	// opt: sanity check >0?
	ctx, cancel := u.stopCh.NewCtx()
	defer cancel()
	u.recordUsage(ctx, u.used)

	u.cond.Broadcast() // notify others blocked on cond.Wait

	return
}

func (u *resourcePoolUsage[N]) newErrorLimitReached(limit, amount N) ErrorResourceLimited[N] {
	return ErrorResourceLimited[N]{
		Key:    u.key,
		Scope:  u.scope,
		Tenant: u.tenant,
		Used:   u.used,
		Limit:  limit,
		Amount: amount,
	}
}

//opt: queue instead of racing for the [sync.Mutex] & [sync.Cond]
func (u *resourcePoolUsage[N]) use(ctx context.Context, amount N, block bool) error {
	start := time.Now()
	u.mu.Lock()
	defer u.mu.Unlock()

	limit := u.getLimit(ctx)
	u.recordLimit(ctx, limit)

	if u.used+amount > limit {
		if !block {
			return u.newErrorLimitReached(limit, amount)
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
			u.cond.Wait() // wait until some resources are freed, or context expiration
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context error (%w) after waiting %s for limit: %w", err, time.Since(start), u.newErrorLimitReached(limit, amount))
			}
		}
	}
	u.used += amount
	u.recordUsage(ctx, u.used)
	u.recordBlocked(ctx, time.Since(start))
	return nil
}

func (u *resourcePoolUsage[N]) wait(ctx context.Context, amount N) (free func(), err error) {
	err = u.use(ctx, amount, true)
	if err != nil {
		return
	}
	var once sync.Once
	free = func() { once.Do(func() { u.free(amount) }) }
	return
}

type unscopedResourcePoolLimiter[N Number] struct {
	resourcePoolLimiter[N]

	*resourcePoolUsage[N]
}

func newUnscopedResourcePoolLimiter[N Number](defaultLimit N) *unscopedResourcePoolLimiter[N] {
	l := &unscopedResourcePoolLimiter[N]{
		resourcePoolLimiter: resourcePoolLimiter[N]{
			updater: newUpdater[N](nil, func(ctx context.Context) (N, error) { return defaultLimit, nil }, nil),
		},
	}
	l.resourcePoolUsage = l.newLimitUsage(recordNoop[N], recordNoop[N], recordNoop[time.Duration])
	return l
}

func recordNoop[T any](ctx context.Context, value T) {}

func (u *unscopedResourcePoolLimiter[N]) Wait(ctx context.Context, amount N) (func(), error) {
	return u.resourcePoolUsage.wait(ctx, amount)
}

func (u *unscopedResourcePoolLimiter[N]) Use(ctx context.Context, amount N) error {
	return u.resourcePoolUsage.use(ctx, amount, false)
}

func (u *unscopedResourcePoolLimiter[N]) Free(_ context.Context, amount N) error {
	u.resourcePoolUsage.free(amount)
	return nil
}

var _ ResourcePoolLimiter[int] = MultiResourcePoolLimiter[int]{}

// MultiResourcePoolLimiter is a ResourcePoolLimiter backed by other limiters, which are each called in order.
type MultiResourcePoolLimiter[N Number] []ResourcePoolLimiter[N]

func (m MultiResourcePoolLimiter[N]) Close() (errs error) {
	for _, l := range m {
		if err := l.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return
}

func (m MultiResourcePoolLimiter[N]) Wait(ctx context.Context, amount N) (func(), error) {
	var frees freeFns
	for _, l := range m {
		free, err := l.Wait(ctx, amount)
		if err != nil {
			frees.freeAll()
			return nil, err
		}
		frees = append(frees, free)
	}
	return frees.freeAll, nil
}

func (m MultiResourcePoolLimiter[N]) Use(ctx context.Context, amount N) error {
	var frees freeFns
	for _, l := range m {
		err := l.Use(ctx, amount)
		if err != nil {
			frees.freeAll()
			return err
		}
		frees = append(frees, func() { l.Free(ctx, amount) })
	}
	return nil
}

func (m MultiResourcePoolLimiter[N]) Free(ctx context.Context, amount N) (errs error) {
	for _, l := range m {
		if err := l.Free(ctx, amount); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return
}

type freeFns []func()

func (f freeFns) freeAll() {
	for i := len(f) - 1; i >= 0; i-- {
		f[i]()
	}
}

// OrgResourcePoolLimiter creates a new ResourcePoolLimiter scoped per organization.
func OrgResourcePoolLimiter[N Number](defaultLimit N) ResourcePoolLimiter[N] {
	return newScopedResourcePoolLimiter(settings.ScopeOrg, "", defaultLimit)
}

// OwnerResourcePoolLimiter creates a new ResourcePoolLimiter scoped per user.
func OwnerResourcePoolLimiter[N Number](defaultLimit N) ResourcePoolLimiter[N] {
	return newScopedResourcePoolLimiter(settings.ScopeOwner, "", defaultLimit)
}

// WorkflowResourcePoolLimiter creates a new ResourcePoolLimiter scoped per workflow.
func WorkflowResourcePoolLimiter[N Number](defaultLimit N) ResourcePoolLimiter[N] {
	return newScopedResourcePoolLimiter(settings.ScopeWorkflow, "", defaultLimit)
}

func newScopedResourcePoolLimiter[N Number](scope settings.Scope, key string, defaultLimit N) *scopedResourcePoolLimiter[N] {
	l := &scopedResourcePoolLimiter[N]{
		resourcePoolLimiter: resourcePoolLimiter[N]{
			key:     key,
			updater: newUpdater[N](nil, func(ctx context.Context) (N, error) { return defaultLimit, nil }, nil),
		},
		scope: scope,
	}
	return l
}

func newScopedResourcePoolLimiterFromFactory[N Number](f Factory, limit settings.Setting[N]) (ResourcePoolLimiter[N], error) {
	l := newScopedResourcePoolLimiter(limit.Scope, limit.Key, limit.DefaultValue)

	if f.Meter != nil {
		if err := l.createGauges(f.Meter, limit.Unit); err != nil {
			return nil, err
		}
	}

	if f.Settings != nil {
		if limit.Key == "" {
			return nil, errors.New("key is required for dynamic Settings updates")
		}
		l.getLimitFn = func(ctx context.Context) (N, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if registry, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[N], func()) {
				return limit.Subscribe(ctx, registry)
			}
		}
	}

	return l, nil
}

// scopedResourcePoolLimiter extends resourcePoolLimiter with a settings.Scope and enforces limits for each tenant separately.
type scopedResourcePoolLimiter[N Number] struct {
	resourcePoolLimiter[N]
	scope settings.Scope

	// opt: reap after period of non-use
	used sync.Map           // map[string]*resourcePoolUsage[N]
	wg   services.WaitGroup // tracks and blocks used background routines
}

func (s *scopedResourcePoolLimiter[N]) Close() (err error) {
	s.wg.Wait()

	// cleanup
	s.used.Range(func(tenant, value interface{}) bool {
		// opt: parallelize
		err = errors.Join(err, value.(*resourcePoolUsage[N]).Close())
		return true
	})
	return
}

func (s *scopedResourcePoolLimiter[N]) Wait(ctx context.Context, amount N) (func(), error) {
	usage, done, err := s.getOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	err = usage.use(ctx, amount, true)
	if err != nil {
		return nil, err
	}
	return func() { usage.free(amount) }, nil
}

func (s *scopedResourcePoolLimiter[N]) Use(ctx context.Context, amount N) error {
	usage, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return usage.use(ctx, amount, false)
}

func (s *scopedResourcePoolLimiter[N]) Free(ctx context.Context, amount N) error {
	usage, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	usage.free(amount)
	return nil
}

type resourcePool[N any] interface {
	use(ctx context.Context, amount N, block bool) error
	free(N)
}

// unlimitedResourcePool is a no-op resourcePool.
type unlimitedResourcePool[N any] struct{}

func (u unlimitedResourcePool[N]) use(ctx context.Context, amount N, block bool) error { return nil }

func (u unlimitedResourcePool[N]) free(n N) {}

func (s *scopedResourcePoolLimiter[N]) getOrCreate(ctx context.Context) (resourcePool[N], func(), error) {
	if err := s.wg.TryAdd(1); err != nil {
		return nil, nil, err
	}

	tenant := s.scope.Value(ctx)
	if tenant == "" {
		if !s.scope.IsTenantRequired() {
			kvs := contexts.CREValue(ctx).LoggerKVs()
			s.lggr.Warnw("Unable to apply scoped resource pool limit due to missing tenant: failing open", append([]any{"scope", s.scope}, kvs...)...)
			return unlimitedResourcePool[N]{}, s.wg.Done, nil
		}
		s.wg.Done()
		return nil, nil, fmt.Errorf("failed to get resource pool: missing tenant for scope: %s", s.scope)
	}

	usage := s.newLimitUsage(tenant)
	actual, loaded := s.used.LoadOrStore(tenant, usage)
	cre := s.scope.RoundCRE(contexts.CREValue(ctx))
	if !loaded {
		usage.cre.Store(cre)
		go usage.updateLoop(cre)
	} else {
		usage = actual.(*resourcePoolUsage[N])
		usage.updateCRE(cre)
	}

	return usage, s.wg.Done, nil
}

func (s *scopedResourcePoolLimiter[N]) newLimitUsage(tenant string) *resourcePoolUsage[N] {
	u := s.resourcePoolLimiter.newLimitUsage(s.recordScoped(tenant))
	u.scope = s.scope
	u.tenant = tenant
	return u
}

func (s *scopedResourcePoolLimiter[N]) recordScoped(tenant string) (usage, limit func(context.Context, N), blocked func(ctx context.Context, value time.Duration)) {
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
	if s.recordBlockTime == nil {
		blocked = func(ctx context.Context, value time.Duration) {}
	} else {
		blocked = func(ctx context.Context, value time.Duration) {
			s.recordBlockTime(ctx, int64(value), metric.WithAttributes(attribute.String(s.scope.String(), tenant)))
		}
	}
	return
}

type unlimitedResourcePoolLimiter[N Number] struct {
	unlimitedResourceLimiter[N]
}

func UnlimitedResourcePoolLimiter[N Number]() ResourcePoolLimiter[N] {
	return unlimitedResourcePoolLimiter[N]{}
}

func (u unlimitedResourcePoolLimiter[N]) Wait(ctx context.Context, n N) (free func(), err error) {
	return func() {}, nil
}

type unlimitedResourceLimiter[N Number] struct{}

func UnlimitedResourceLimiter[N Number]() ResourceLimiter[N] {
	return unlimitedResourceLimiter[N]{}
}

func (u unlimitedResourceLimiter[N]) Close() error { return nil }

func (u unlimitedResourceLimiter[N]) Use(ctx context.Context, amount N) error { return nil }

func (u unlimitedResourceLimiter[N]) Free(ctx context.Context, amount N) error { return nil }
