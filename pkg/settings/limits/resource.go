package limits

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// A ResourceLimiter limits usage by tracking allocations and ensuring that the threshold is not exceeded.
type ResourceLimiter[N Number] interface {
	Limiter

	// Use reserves amount of resources, or returns an error. The free function must be
	// called to release resources, and should be deferred immediately when possible.
	// Blocks until resources are available, or context has expired.
	Use(ctx context.Context, amount N) (free func(), err error)

	// TryUse is like Use, but returns immediately with ErrorResourceLimited when resources are not available.
	TryUse(ctx context.Context, amount N) (free func(), err error)
}

// GlobalResourceLimiter returns an unscoped ResourceLimiter with default options.
// See Resource.GlobalLimiter for dynamic limits, metering, and more.
func GlobalResourceLimiter[N Number](limit N) ResourceLimiter[N] {
	return newUnscopedLimiter(limit)
}

// GlobalLimiter returns an unscoped ResourceLimiter for Key with the given Configuration.
func newGlobalResourceLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourceLimiter[N], error) {
	l := newUnscopedLimiter(limit.DefaultValue)
	l.key = limit.Key

	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
		l.lggr = logger.Sugared(f.Logger).Named("ResourceLimiter").With("key", limit.Key)
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
		l.cre.Store(contexts.CRE{})
		go l.updateLoop(contexts.CRE{})
	}
	if f.Meter != nil {
		if err := l.createGauges(f.Meter); err != nil {
			return nil, err
		}
		l.resourceUsage = l.newLimitUsage(func(ctx context.Context, value N) {
			if l.resourceLimiter.recordUsage != nil {
				l.resourceLimiter.recordUsage(ctx, value)
			}
		}, func(ctx context.Context, value N) {
			if l.resourceLimiter.recordLimit != nil {
				l.resourceLimiter.recordLimit(ctx, value)
			}
		}, func(ctx context.Context, value time.Duration) {
			if l.resourceLimiter.recordBlockTime != nil {
				l.resourceLimiter.recordBlockTime(ctx, int64(value))
			}
		})
	}
	return l, nil
}

type resourceLimiter[N Number] struct {
	lggr logger.Logger

	key string // optional

	getLimitFn func(context.Context) (N, error)
	subFn      func(ctx context.Context) (<-chan settings.Update[N], func()) // optional

	recordUsage     func(ctx context.Context, value N, options ...metric.RecordOption)    // optional
	recordLimit     func(ctx context.Context, value N, options ...metric.RecordOption)    // optional
	recordBlockTime func(ctx context.Context, incr int64, options ...metric.RecordOption) // optional                                              // optional

}

func (l *resourceUsage[N]) Close() error {
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
	<-l.done
	return nil
}

func (l *resourceUsage[N]) updateCRE(cre contexts.CRE) {
	cur := l.cre.Load().(contexts.CRE)
	if cur == cre {
		return
	}
	l.cre.Store(cre)
	select {
	case l.creCh <- struct{}{}:
	default:
	}
}

func (l *resourceUsage[N]) updateLoop(cre contexts.CRE) {
	close(l.done)
	ctx, cancel := l.stopCh.NewCtx()
	defer cancel()

	var updates <-chan settings.Update[N]
	var cancelSub func()
	var c <-chan time.Time
	if l.subFn != nil {
		updates, cancelSub = l.subFn(contexts.WithCRE(ctx, cre))
		defer cancelSub()
	} else {
		t := time.NewTicker(pollPeriod)
		defer t.Stop()
		c = t.C
	}
	for {
		select {
		case <-ctx.Done():
			return

		case <-c:
			limit := l.getLimit(contexts.WithCRE(ctx, cre))
			l.recordLimit(ctx, limit)

		case update := <-updates:
			if update.Err != nil {
				l.lggr.Errorw("Failed to update resource limit. Using default value", "default", update.Value, "err", update.Err)
			}
			l.recordLimit(ctx, update.Value)

		case <-l.creCh:
			cre = l.cre.Load().(contexts.CRE)
			if l.subFn != nil {
				cancelSub()
				updates, cancelSub = l.subFn(contexts.WithCRE(ctx, cre))
			}
		}
	}
}

func (l *resourceLimiter[N]) createGauges(meter metric.Meter) error {
	if l.key == "" {
		return errors.New("metrics require Key to be set")
	}
	var gaugeFn func(key string) (gauge[N], error)
	var n N
	if k := reflect.TypeOf(n).Kind(); k == reflect.Float64 || k == reflect.Float32 {
		gaugeFn = func(key string) (gauge[N], error) {
			g, err := meter.Float64Gauge(key)
			if err != nil {
				return nil, err
			}
			return &floatGauge[N]{g}, nil
		}
	} else {
		gaugeFn = func(key string) (gauge[N], error) {
			g, err := meter.Int64Gauge(key)
			if err != nil {
				return nil, err
			}
			return &intGauge[N]{g}, nil
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
	return nil
}

func (l *resourceLimiter[N]) getLimit(ctx context.Context) (limit N) {
	limit, err := l.getLimitFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to get limit. Using default value", "default", limit, "err", err)
	}
	return limit
}

type resourceUsage[N Number] struct {
	*resourceLimiter[N]
	scope  settings.Scope // optional
	tenant string         // optional
	mu     sync.Mutex
	cond   sync.Cond
	used   N

	recordLimit   func(ctx context.Context, value N)
	recordUsage   func(ctx context.Context, value N)
	recordBlocked func(ctx context.Context, value time.Duration)

	creCh chan struct{}
	cre   atomic.Value

	stopOnce sync.Once
	stopCh   services.StopChan
	done     chan struct{}
}

func (l *resourceLimiter[N]) newLimitUsage(
	recordUsage func(ctx context.Context, value N),
	recordLimit func(ctx context.Context, value N),
	recordBlocked func(ctx context.Context, value time.Duration),
) *resourceUsage[N] {
	u := resourceUsage[N]{
		resourceLimiter: l,
		recordUsage:     recordUsage,
		recordLimit:     recordLimit,
		recordBlocked:   recordBlocked,
		creCh:           make(chan struct{}, 1),
		stopCh:          make(services.StopChan),
		done:            make(chan struct{}),
	}
	u.cond.L = &u.mu
	return &u
}

func (u *resourceUsage[N]) free(amount N) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.used -= amount
	ctx, cancel := u.stopCh.NewCtx()
	defer cancel()
	u.recordUsage(ctx, u.used)

	//TODO sanity check >0?

	u.cond.Broadcast()

	return
}

func (u *resourceUsage[N]) newErrorLimitReached(limit, amount N) ErrorResourceLimited[N] {
	return ErrorResourceLimited[N]{
		Key:    u.key,
		Scope:  u.scope,
		Tenant: u.tenant,
		Used:   u.used,
		Limit:  limit,
		Amount: amount,
	}
}

func (u *resourceUsage[N]) use(ctx context.Context, amount N, block bool) (func(), error) {
	start := time.Now()
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
	}
	u.used += amount
	u.recordUsage(ctx, u.used)
	u.recordBlocked(ctx, time.Since(start))
	var once sync.Once
	return func() { once.Do(func() { u.free(amount) }) }, nil
}

type unscopedResourceLimiter[N Number] struct {
	resourceLimiter[N]

	*resourceUsage[N]
}

func newUnscopedLimiter[N Number](defaultLimit N) *unscopedResourceLimiter[N] {
	l := &unscopedResourceLimiter[N]{
		resourceLimiter: resourceLimiter[N]{
			getLimitFn: func(ctx context.Context) (N, error) { return defaultLimit, nil },
		},
	}
	l.resourceUsage = l.newLimitUsage(recordNoop[N], recordNoop[N], recordNoop[time.Duration])
	return l
}

func recordNoop[T any](ctx context.Context, value T) {}

func (u *unscopedResourceLimiter[N]) Use(ctx context.Context, amount N) (func(), error) {
	return u.resourceUsage.use(ctx, amount, true)
}

func (u *unscopedResourceLimiter[N]) TryUse(ctx context.Context, amount N) (func(), error) {
	return u.resourceUsage.use(ctx, amount, false)
}

// MultiResourceLimiter is a ResourceLimiter backed by other limiters, which are each called in order.
type MultiResourceLimiter[N Number] []ResourceLimiter[N]

func (m MultiResourceLimiter[N]) Use(ctx context.Context, amount N) (func(), error) {
	var frees freeFns
	for _, l := range m {
		free, err := l.Use(ctx, amount)
		if err != nil {
			frees.freeAll()
			return nil, err
		}
		frees = append(frees, free)
	}
	return frees.freeAll, nil
}

func (m MultiResourceLimiter[N]) TryUse(ctx context.Context, amount N) (func(), error) {
	var frees freeFns
	for _, l := range m {
		free, err := l.TryUse(ctx, amount)
		if err != nil {
			frees.freeAll()
			return nil, err
		}
		frees = append(frees, free)
	}
	return frees.freeAll, nil
}

type freeFns []func()

func (f freeFns) freeAll() {
	for i := len(f) - 1; i >= 0; i-- {
		f[i]()
	}
}

// OrgResourceLimiter creates a new ResourceLimiter scoped per organization.
func OrgResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeOrg, "", defaultLimit)
}

// OwnerResourceLimiter creates a new ResourceLimiter scoped per user.
func OwnerResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeOwner, "", defaultLimit)
}

// WorkflowResourceLimiter creates a new ResourceLimiter scoped per workflow.
func WorkflowResourceLimiter[N Number](defaultLimit N) ResourceLimiter[N] {
	return newScopedResource(settings.ScopeWorkflow, "", defaultLimit)
}

func newScopedResource[N Number](scope settings.Scope, key string, defaultLimit N) *scopedResourceLimiter[N] {
	l := &scopedResourceLimiter[N]{
		resourceLimiter: resourceLimiter[N]{
			key: key,
			getLimitFn: func(ctx context.Context) (N, error) {
				return defaultLimit, nil
			},
		},
		scope: scope,
	}
	return l
}

func newScopedResourceLimiter[N Number](f Factory, limit settings.Setting[N]) (ResourceLimiter[N], error) {
	l := newScopedResource(limit.Scope, limit.Key, limit.DefaultValue)

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
	if f.Meter != nil {
		if err := l.createGauges(f.Meter); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// scopedResourceLimiter extends resourceLimiter with a settings.Scope and enforces limits for each tenant separately.
type scopedResourceLimiter[N Number] struct {
	resourceLimiter[N]
	scope settings.Scope

	//TODO when to reap?
	used sync.Map // map[string]*resourceUsage[N]
}

func (s *scopedResourceLimiter[N]) Close() (err error) {
	s.used.Range(func(tenant, value interface{}) bool {
		err = errors.Join(err, value.(*resourceUsage[N]).Close())
		return true
	})
	return
}

func (s *scopedResourceLimiter[N]) Use(ctx context.Context, amount N) (func(), error) {
	return s.use(ctx, amount, true)
}
func (s *scopedResourceLimiter[N]) TryUse(ctx context.Context, amount N) (func(), error) {
	return s.use(ctx, amount, false)
}

func (s *scopedResourceLimiter[N]) use(ctx context.Context, amount N, block bool) (func(), error) {
	tenant := s.scope.Value(ctx)
	if tenant == "" {
		return nil, fmt.Errorf("missing tenant for scope: %s", s.scope)
	}

	usage := s.newLimitUsage(tenant)
	actual, loaded := s.used.LoadOrStore(tenant, usage)
	cre := s.scope.RoundCRE(contexts.CREValue(ctx))
	if !loaded {
		usage.cre.Store(cre)
		go usage.updateLoop(cre)
	} else {
		usage = actual.(*resourceUsage[N])
		usage.updateCRE(cre)
	}
	return usage.use(ctx, amount, block)
}

func (s *scopedResourceLimiter[N]) newLimitUsage(tenant string) *resourceUsage[N] {
	u := s.resourceLimiter.newLimitUsage(s.recordScoped(tenant))
	u.scope = s.scope
	u.tenant = tenant
	return u
}

func (s *scopedResourceLimiter[N]) recordScoped(tenant string) (usage, limit func(context.Context, N), blocked func(ctx context.Context, value time.Duration)) {
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
