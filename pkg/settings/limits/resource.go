package limits

import (
	"context"
	"errors"
	"fmt"
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
	Limiter[N]
	// Use increases the resource count by amount, or returns an error if the limit is reached.
	// It does not block. Use a ResourcePoolLimiter for blocking semantics.
	Use(ctx context.Context, amount N) error
	// Free is the counterpart to Use and releases amount of resources from use.
	Free(ctx context.Context, amount N) error
	// Available returns the available remaining capacity.
	Available(ctx context.Context) (N, error)
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
// See [MakeResourcePoolLimiter] for dynamic limits, metering, and more.
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
		l.resourcePoolUsage = l.newLimitUsage()
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

	go l.updateLoop(context.Background())

	return l, nil
}

type resourcePoolLimiter[N Number] struct {
	*updater[N]

	key string // optional

	recordUsage     func(context.Context, N, ...metric.RecordOption)       // optional
	recordLimit     func(context.Context, N, ...metric.RecordOption)       // optional
	recordBlockTime func(context.Context, float64, ...metric.RecordOption) // optional
	recordAmount    func(context.Context, N, ...metric.RecordOption)       // optional
	recordDenied    func(context.Context, N, ...metric.RecordOption)       // optional
}

func (l *resourcePoolLimiter[N]) setOnLimitUpdate(fn func(ctx context.Context)) {
	l.updater.onLimitUpdate = fn
}

func (l *resourcePoolLimiter[N]) createGauges(meter metric.Meter, unit string) error {
	if l.key == "" {
		return errors.New("metrics require Key to be set")
	}
	newGauge, newHist := metricConstructors[N](meter, unit)

	limitGauge, err := newGauge("resource." + l.key + ".limit")
	if err != nil {
		return err
	}
	l.recordLimit = func(ctx context.Context, value N, options ...metric.RecordOption) {
		limitGauge.Record(ctx, value, options...)
	}
	usageGauge, err := newGauge("resource." + l.key + ".usage")
	if err != nil {
		return err
	}
	l.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {
		usageGauge.Record(ctx, value, options...)
	}
	blockTimeHistogram, err := meter.Float64Histogram("resource."+l.key+".block_time", metric.WithUnit(unit))
	if err != nil {
		return err
	}
	l.recordBlockTime = func(ctx context.Context, value float64, options ...metric.RecordOption) {
		blockTimeHistogram.Record(ctx, value, options...)
	}
	amountHistogram, err := newHist("resource." + l.key + ".amount")
	if err != nil {
		return err
	}
	l.recordAmount = func(ctx context.Context, value N, options ...metric.RecordOption) {
		amountHistogram.Record(ctx, value, options...)
	}
	deniedHistogram, err := newHist("resource." + l.key + ".denied")
	if err != nil {
		return err
	}
	l.recordDenied = func(ctx context.Context, value N, options ...metric.RecordOption) {
		deniedHistogram.Record(ctx, value, options...)
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

// waiter represents a goroutine waiting for resources in the FIFO queue.
type waiter[N Number] struct {
	amount N
	ready  chan struct{} // closed when resources are granted
}

type resourcePoolUsage[N Number] struct {
	resourcePoolLimiter[N]
	scope  settings.Scope // optional
	tenant string         // optional
	mu     sync.Mutex
	used   N
	// queue holds waiters in FIFO order; head of slice is first to be serviced
	queue []*waiter[N]
	// onEnqueue is an optional callback invoked (under lock) when a waiter is added to the queue.
	// Used for testing to synchronize without sleeps.
	onEnqueue func()

	recordUsage     func(context.Context, N)
	recordLimit     func(context.Context, N)
	recordBlockTime func(context.Context, float64)
	recordAmount    func(context.Context, N)
	recordDenied    func(context.Context, N)

	stopOnce  sync.Once
	stopCh    services.StopChan
	done      chan struct{}
	cancelSub func() // optional
}

// onLimitUpdate is invoked when the configured limit changes. It attempts to
// wake queued waiters using the new limit.
func (u *resourcePoolUsage[N]) onLimitUpdate() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tryWakeWaiters()
}

func (l *resourcePoolLimiter[N]) newLimitUsage(opts ...metric.RecordOption) *resourcePoolUsage[N] {
	u := resourcePoolUsage[N]{
		queue:  make([]*waiter[N], 0),
		stopCh: make(services.StopChan),
		done:   make(chan struct{}),
		recordUsage: func(ctx context.Context, n N) {
			if l.recordUsage != nil {
				l.recordUsage(ctx, n, opts...)
			}
		},
		recordLimit: func(ctx context.Context, n N) {
			if l.recordLimit != nil {
				l.recordLimit(ctx, n, opts...)
			}
		},
		recordBlockTime: func(ctx context.Context, n float64) {
			if l.recordBlockTime != nil {
				l.recordBlockTime(ctx, n, opts...)
			}
		},
		recordAmount: func(ctx context.Context, n N) {
			if l.recordAmount != nil {
				l.recordAmount(ctx, n, opts...)
			}
		},
		recordDenied: func(ctx context.Context, n N) {
			if l.recordDenied != nil {
				l.recordDenied(ctx, n, opts...)
			}
		},
	}
	// copy but replace updater
	u.resourcePoolLimiter = *l
	u.resourcePoolLimiter.updater = newUpdater(l.lggr, l.getLimitFn, l.subFn)
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

	u.tryWakeWaiters()
}

// tryWakeWaiters attempts to wake waiters at the head of the queue
// whose resource requests can now be satisfied.
// Must be called with u.mu held.
func (u *resourcePoolUsage[N]) tryWakeWaiters() {
	for len(u.queue) > 0 {
		head := u.queue[0]
		limit := u.getLimit(context.Background())
		if u.used+head.amount > limit {
			// Not enough resources for the head waiter; stop here to preserve FIFO
			break
		}
		// Grant resources to head waiter
		u.used += head.amount
		close(head.ready)
		u.queue = u.queue[1:]
	}
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

func (u *resourcePoolUsage[N]) get(ctx context.Context) (N, error) {
	limit := u.getLimit(ctx)
	u.recordLimit(ctx, limit)
	return limit, nil
}

func (u *resourcePoolUsage[N]) available(ctx context.Context) (N, error) {
	limit, err := u.get(ctx)
	if err != nil {
		var zero N
		return zero, err
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	return limit - u.used, nil
}

func (u *resourcePoolUsage[N]) use(ctx context.Context, amount N, block bool) error {
	limit, err := u.get(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	u.mu.Lock()

	// Fast path: resources available immediately and no one else waiting
	if len(u.queue) == 0 && u.used+amount <= limit {
		u.used += amount
		u.recordUsage(ctx, u.used)
		u.recordAmount(ctx, amount)
		u.recordBlockTime(ctx, time.Since(start).Seconds())
		u.mu.Unlock()
		return nil
	}

	// Not enough resources
	if !block {
		u.recordDenied(ctx, amount)
		err := u.newErrorLimitReached(limit, amount)
		u.mu.Unlock()
		return err
	}

	// Slow path: need to queue up and wait (FIFO ordering)
	w := &waiter[N]{
		amount: amount,
		ready:  make(chan struct{}),
	}
	u.queue = append(u.queue, w)
	if u.onEnqueue != nil {
		u.onEnqueue()
	}
	u.mu.Unlock()

	// Wait for our turn or context cancellation
	select {
	case <-w.ready:
		// Resources have been granted to us
		u.mu.Lock()
		u.recordUsage(ctx, u.used)
		u.recordAmount(ctx, amount)
		u.recordBlockTime(ctx, time.Since(start).Seconds())
		u.mu.Unlock()
		return nil
	case <-ctx.Done():
		// Context cancelled - remove ourselves from queue
		u.mu.Lock()
		defer u.mu.Unlock()

		// Check if we were already granted resources while acquiring the lock
		select {
		case <-w.ready:
			// We got resources just as we were cancelling; return them
			u.used -= amount
			u.tryWakeWaiters()
			u.recordDenied(ctx, amount)
			return fmt.Errorf("context error (%w) after waiting %s for limit: %w", ctx.Err(), time.Since(start), u.newErrorLimitReached(limit, amount))
		default:
		}

		// Remove from queue. Only needed when the context was cancelled before the element got to the head of the queue.
		// Otherwise it is already removed by tryWakeWaiters().
		for i, waiter := range u.queue {
			if waiter == w {
				u.queue = append(u.queue[:i], u.queue[i+1:]...)
				break
			}
		}
		u.recordDenied(ctx, amount)
		return fmt.Errorf("context error (%w) after waiting %s for limit: %w", ctx.Err(), time.Since(start), u.newErrorLimitReached(limit, amount))
	}
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
	l.resourcePoolUsage = l.newLimitUsage()
	l.setOnLimitUpdate(func(context.Context) {
		l.resourcePoolUsage.onLimitUpdate()
	})
	return l
}

func (u *unscopedResourcePoolLimiter[N]) Limit(ctx context.Context) (N, error) {
	return u.get(ctx)
}

func (u *unscopedResourcePoolLimiter[N]) Available(ctx context.Context) (N, error) {
	return u.available(ctx)
}

func (u *unscopedResourcePoolLimiter[N]) Wait(ctx context.Context, amount N) (func(), error) {
	return u.wait(ctx, amount)
}

func (u *unscopedResourcePoolLimiter[N]) Use(ctx context.Context, amount N) error {
	return u.use(ctx, amount, false)
}

func (u *unscopedResourcePoolLimiter[N]) Free(_ context.Context, amount N) error {
	u.free(amount)
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

func (m MultiResourcePoolLimiter[N]) Limit(ctx context.Context) (N, error) {
	if len(m) == 0 {
		var zero N
		return zero, fmt.Errorf("no limiters")
	}
	return m[0].Limit(ctx)
}

func (m MultiResourcePoolLimiter[N]) Available(ctx context.Context) (N, error) {
	if len(m) == 0 {
		var zero N
		return zero, fmt.Errorf("no limiters")
	}
	return m[0].Available(ctx)
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
	l.setOnLimitUpdate(func(ctx context.Context) {
		tenant := l.scope.Value(ctx)
		if tenant == "" {
			return
		}
		if usage, ok := l.used.Load(tenant); ok {
			usage.(*resourcePoolUsage[N]).onLimitUpdate()
		}
	})
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
	s.used.Range(func(tenant, value any) bool {
		// opt: parallelize
		err = errors.Join(err, value.(*resourcePoolUsage[N]).Close())
		return true
	})
	return
}

func (s *scopedResourcePoolLimiter[N]) Limit(ctx context.Context) (N, error) {
	usage, done, err := s.getOrCreate(ctx)
	if err != nil {
		var zero N
		return zero, err
	}
	defer done()
	return usage.get(ctx)
}

func (s *scopedResourcePoolLimiter[N]) Available(ctx context.Context) (N, error) {
	usage, done, err := s.getOrCreate(ctx)
	if err != nil {
		var zero N
		return zero, err
	}
	defer done()
	return usage.available(ctx)
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

type resourcePool[N Number] interface {
	use(ctx context.Context, amount N, block bool) error
	free(N)
	get(context.Context) (N, error)
	available(context.Context) (N, error)
}

// unlimitedResourcePool is a no-op resourcePool.
type unlimitedResourcePool[N Number] struct{}

func (u unlimitedResourcePool[N]) use(ctx context.Context, amount N, block bool) error { return nil }

func (u unlimitedResourcePool[N]) get(ctx context.Context) (N, error) {
	return maxVal[N]()
}

func (u unlimitedResourcePool[N]) available(ctx context.Context) (N, error) {
	return maxVal[N]()
}

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
	creCtx := contexts.WithCRE(ctx, s.scope.RoundCRE(contexts.CREValue(ctx)))
	if !loaded {
		go usage.updateLoop(creCtx)
	} else {
		usage = actual.(*resourcePoolUsage[N])
		usage.updateCtx(creCtx)
	}

	return usage, s.wg.Done, nil
}

func (s *scopedResourcePoolLimiter[N]) newLimitUsage(tenant string) *resourcePoolUsage[N] {
	u := s.resourcePoolLimiter.newLimitUsage(metric.WithAttributes(attribute.String(s.scope.String(), tenant)))
	u.scope = s.scope
	u.tenant = tenant
	return u
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

func (u unlimitedResourceLimiter[N]) Limit(ctx context.Context) (n N, err error) {
	return maxVal[N]()
}
func (u unlimitedResourceLimiter[N]) Available(ctx context.Context) (n N, err error) {
	return maxVal[N]()
}

func (u unlimitedResourceLimiter[N]) Use(ctx context.Context, amount N) error { return nil }

func (u unlimitedResourceLimiter[N]) Free(ctx context.Context, amount N) error { return nil }
