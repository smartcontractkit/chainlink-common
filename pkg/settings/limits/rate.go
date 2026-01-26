package limits

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// A RateLimiter applies rate limits. These methods are a subset of [rate.Limiter], with context.Context based scoping,
// and some *Err variants. Methods with errors will return ErrorRateLimited when limits are encountered.
type RateLimiter interface {
	Limiter[config.Rate]

	// Allow reports whether an event may happen now.
	Allow(ctx context.Context) bool
	// AllowN reports whether n events may happen at time t.
	// Use this method if you intend to drop / skip events that exceed the rate limit.
	// Otherwise, use Reserve or Wait.
	AllowN(ctx context.Context, t time.Time, n int) bool
	// AllowErr is like Allow, but returns an error.
	AllowErr(ctx context.Context) error
	// AllowNErr is like AllowN, but returns an error.
	AllowNErr(ctx context.Context, t time.Time, n int) error

	// Reserve is shorthand for ReserveN(time.Now(), 1).
	Reserve(ctx context.Context) (Reservation, error)
	// ReserveN returns a Reservation that indicates how long the caller must wait before n events happen.
	// The Limiter takes this Reservation into account when allowing future events.
	// The returned Reservation’s OK() method returns false if n exceeds the Limiter's burst size.
	// Usage example:
	//
	//	r := lim.ReserveN(time.Now(), 1)
	//	if !r.OK() {
	//	  // Not allowed to act! Did you remember to set lim.burst to be > 0 ?
	//	  return
	//	}
	//	time.Sleep(r.Delay())
	//	Act()
	//
	// Use this method if you wish to wait and slow down in accordance with the rate limit without dropping events.
	// If you need to respect a deadline or cancel the delay, use Wait instead.
	// To drop or skip events exceeding rate limit, use Allow instead.
	ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error)

	// Wait is shorthand for WaitN(ctx, 1).
	Wait(ctx context.Context) (err error)
	// WaitN blocks until lim permits n events to happen.
	// It returns an error if n exceeds the Limiter's burst size, the Context is
	// canceled, or the expected wait time exceeds the Context's Deadline.
	WaitN(ctx context.Context, n int) (err error)
}

// Reservation extends the exported interface of [*rate.Reservation].
type Reservation interface {
	OK() bool
	Delay() time.Duration
	DelayFrom(time.Time) time.Duration
	Cancel()
	CancelAt(time.Time)

	// Allow returns true if no Delay is required.
	Allow() bool
	// AllowErr is like Allow but includes a detailed error.
	AllowErr() error
}

type reservation struct {
	*rate.Reservation
	key    string
	scope  settings.Scope
	tenant string
	n      int
}

func (r *reservation) Allow() bool { return r.Delay() <= 0 }
func (r *reservation) AllowErr() error {
	if r.Delay() > 0 {
		return ErrorRateLimited{
			Key:    r.key,
			Scope:  r.scope,
			Tenant: r.tenant,
			N:      r.n,
		}
	}
	return nil
}

// GlobalRateLimiter returns an unscoped RateLimiter for the given limit and burst.
func GlobalRateLimiter(limit rate.Limit, burst int) RateLimiter {
	return newRateLimiter(settings.ScopeGlobal, limit, burst)
}

// OrgRateLimiter returns a RateLimiter scoped by org for the given limit and burst.
func OrgRateLimiter(limit rate.Limit, burst int) RateLimiter {
	return newRateLimiter(settings.ScopeOrg, limit, burst)
}

// OwnerRateLimiter returns a RateLimiter scoped by owner for the given limit and burst.
func OwnerRateLimiter(limit rate.Limit, burst int) RateLimiter {
	return newRateLimiter(settings.ScopeOwner, limit, burst)
}

// WorkflowRateLimiter returns a RateLimiter scoped by workflow for the given limit and burst.
func WorkflowRateLimiter(limit rate.Limit, burst int) RateLimiter {
	return newRateLimiter(settings.ScopeWorkflow, limit, burst)
}

func newRateLimiter(scope settings.Scope, limit rate.Limit, burst int) RateLimiter {
	r := config.Rate{Limit: limit, Burst: burst}
	if scope == settings.ScopeGlobal {
		rl := &rateLimiter{
			updater: newUpdater[config.Rate](nil, func(ctx context.Context) (config.Rate, error) {
				return r, nil
			}, nil),

			limiter: rate.NewLimiter(limit, burst),

			recordLimit:  func(ctx context.Context, value float64) {},
			recordBurst:  func(ctx context.Context, value int64) {},
			addUsage:     func(ctx context.Context, incr int64) {},
			recordDenied: func(ctx context.Context, incr int) {},
		}
		close(rl.updater.done) // no background routine
		return rl
	}
	return &scopedRateLimiter{
		lggr:        logger.Nop(),
		scope:       scope,
		defaultRate: r,
		rateFn:      func(ctx context.Context) (config.Rate, error) { return r, nil },
	}
}

func (f Factory) globalRateLimiter(limit settings.Setting[config.Rate]) (RateLimiter, error) {
	l := &rateLimiter{
		updater: newUpdater[config.Rate](nil, func(ctx context.Context) (config.Rate, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}, nil),

		recordLimit:  func(ctx context.Context, value float64) {},
		recordBurst:  func(ctx context.Context, value int64) {},
		addUsage:     func(ctx context.Context, incr int64) {},
		recordDenied: func(ctx context.Context, incr int) {},

		limiter: rate.NewLimiter(limit.DefaultValue.Limit, limit.DefaultValue.Burst),
	}

	if f.Logger != nil {
		l.lggr = logger.Sugared(f.Logger).Named("RateLimiter").With("key", limit.Key)
	}

	if f.Meter != nil {
		if limitGauge, err := f.Meter.Float64Gauge("rate."+limit.Key+".limit", metric.WithUnit("rps")); err != nil {
			return nil, err
		} else {
			l.recordLimit = func(ctx context.Context, value float64) { limitGauge.Record(ctx, value) }
		}
		if burstGauge, err := f.Meter.Int64Gauge("rate."+limit.Key+".burst", metric.WithUnit(limit.Unit)); err != nil {
			return nil, err
		} else {
			l.recordBurst = func(ctx context.Context, value int64) { burstGauge.Record(ctx, value) }
		}
		if usageCounter, err := f.Meter.Int64Counter("rate."+limit.Key+".usage", metric.WithUnit(limit.Unit)); err != nil {
			return nil, err
		} else {
			l.addUsage = func(ctx context.Context, value int64) { usageCounter.Add(ctx, value) }
		}
		if deniedCounter, err := f.Meter.Int64Histogram("rate."+limit.Key+".denied", metric.WithUnit(limit.Unit)); err != nil {
			return nil, err
		} else {
			l.recordDenied = func(ctx context.Context, value int) { deniedCounter.Record(ctx, int64(value)) }
		}
	} else {
		l.recordLimit = func(ctx context.Context, value float64) {}
		l.recordBurst = func(ctx context.Context, value int64) {}
		l.addUsage = func(ctx context.Context, value int64) {}
		l.recordDenied = func(ctx context.Context, value int) {}
	}

	if f.Settings != nil {
		if registry, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[config.Rate], func()) {
				return limit.Subscribe(ctx, registry)
			}
		}
	}
	l.updater.recordLimit = func(ctx context.Context, r config.Rate) {
		l.limiter.SetLimit(r.Limit)
		l.limiter.SetBurst(r.Burst)
		l.recordLimit(ctx, float64(r.Limit))
		l.recordBurst(ctx, int64(r.Burst))
	}
	l.cre.Store(contexts.CRE{})
	go l.updateLoop(contexts.CRE{})

	return l, nil
}

type rateLimiter struct {
	*updater[config.Rate]

	key    string         // optional
	scope  settings.Scope // optional
	tenant string         // optional

	recordLimit  func(ctx context.Context, value float64)
	recordBurst  func(ctx context.Context, value int64)
	addUsage     func(ctx context.Context, incr int64)
	recordDenied func(ctx context.Context, value int)

	limiter *rate.Limiter
}

func (l *rateLimiter) getRate(ctx context.Context) config.Rate {
	r, err := l.getLimitFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to get limit. Using default value", "default", r, "err", err)
	}
	return r
}

func (l *rateLimiter) Limit(ctx context.Context) (config.Rate, error) {
	return l.getRate(ctx), nil
}

func (l *rateLimiter) Allow(ctx context.Context) bool {
	if l.limiter.Allow() {
		l.addUsage(ctx, 1)
		return true
	}
	l.recordDenied(ctx, 1)
	return false
}

func (l *rateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	if l.limiter.AllowN(t, n) {
		l.addUsage(ctx, int64(n))
		return true
	}
	l.recordDenied(ctx, n)
	return false
}

func (l *rateLimiter) AllowErr(ctx context.Context) error {
	if !l.Allow(ctx) {
		return ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: 1}
	}
	l.recordDenied(ctx, 1)
	return nil
}

func (l *rateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	if !l.AllowN(ctx, t, n) {
		return ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: 1}
	}
	l.recordDenied(ctx, n)
	return nil
}

func (l *rateLimiter) Reserve(ctx context.Context) (Reservation, error) {
	r := l.limiter.Reserve()
	if r.OK() {
		l.addUsage(ctx, 1)
		return &reservation{
			Reservation: r,
			key:         l.key,
			scope:       l.scope,
			tenant:      l.tenant,
			n:           1,
		}, nil
	}
	l.recordDenied(ctx, 1)
	return nil, ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: 1}
}

func (l *rateLimiter) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	r := l.limiter.ReserveN(t, n)
	if r.OK() {
		l.addUsage(ctx, int64(n))
		return &reservation{
			Reservation: r,
			key:         l.key,
			scope:       l.scope,
			tenant:      l.tenant,
			n:           n,
		}, nil
	}
	l.recordDenied(ctx, n)
	return nil, ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: n}
}

func (l *rateLimiter) Wait(ctx context.Context) error {
	if err := l.limiter.Wait(ctx); err != nil {
		l.recordDenied(ctx, 1)
		return ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: 1, Err: err}
	}
	l.addUsage(ctx, 1)
	return nil
}

func (l *rateLimiter) WaitN(ctx context.Context, n int) error {
	if err := l.limiter.WaitN(ctx, n); err != nil {
		l.recordDenied(ctx, n)
		return ErrorRateLimited{Key: l.key, Scope: l.scope, Tenant: l.tenant, N: n, Err: err}
	}
	l.addUsage(ctx, int64(n))
	return nil
}

func (f Factory) newScopedRateLimiter(limit settings.Setting[config.Rate]) (RateLimiter, error) {
	l := &scopedRateLimiter{
		key:         limit.Key,
		scope:       limit.Scope,
		defaultRate: limit.DefaultValue,
	}

	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
		l.lggr = logger.Sugared(f.Logger).Named("RateLimiter").With("key", limit.Key, "scope", limit.Scope)
	}

	if f.Meter != nil {
		var err error
		l.limitGauge, err = f.Meter.Float64Gauge("rate."+limit.Key+".limit", metric.WithUnit("rps"))
		if err != nil {
			return nil, err
		}

		l.burstGauge, err = f.Meter.Int64Gauge("rate."+limit.Key+".burst", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}

		l.usageCounter, err = f.Meter.Int64Counter("rate."+limit.Key+".usage", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
		l.deniedHist, err = f.Meter.Int64Histogram("rate."+limit.Key+".denied", metric.WithUnit(limit.Unit))
		if err != nil {
			return nil, err
		}
	}

	if f.Settings != nil {
		l.rateFn = func(ctx context.Context) (config.Rate, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if r, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[config.Rate], func()) {
				return limit.Subscribe(ctx, r)
			}
		}
	} else {
		l.rateFn = func(ctx context.Context) (config.Rate, error) {
			return limit.DefaultValue, nil
		}
	}

	return l, nil
}

type scopedRateLimiter struct {
	lggr        logger.Logger
	key         string
	scope       settings.Scope
	defaultRate config.Rate

	rateFn func(context.Context) (config.Rate, error)
	subFn  func(ctx context.Context) (<-chan settings.Update[config.Rate], func()) // optional

	limitGauge   metric.Float64Gauge   // optional
	burstGauge   metric.Int64Gauge     // optional
	usageCounter metric.Int64Counter   // optional
	deniedHist   metric.Int64Histogram // optional

	// opt: reap after period of non-use
	limiters sync.Map           // map[string]*rateLimiter
	wg       services.WaitGroup // tracks and blocks limiters background routines
}

func (s *scopedRateLimiter) Close() (err error) {
	s.wg.Wait()

	// cleanup
	s.limiters.Range(func(tenant, value any) bool {
		// opt: parallelize
		err = errors.Join(err, value.(*rateLimiter).Close())
		return true
	})
	return
}

func (s *scopedRateLimiter) getOrCreate(ctx context.Context) (RateLimiter, func(), error) {
	if err := s.wg.TryAdd(1); err != nil {
		return nil, nil, err
	}

	tenant := s.scope.Value(ctx)
	if tenant == "" {
		if !s.scope.IsTenantRequired() {
			kvs := contexts.CREValue(ctx).LoggerKVs()
			s.lggr.Warnw("Unable to apply scoped rate limit due to missing tenant: failing open", append([]any{"scope", s.scope}, kvs...)...)
			return UnlimitedRateLimiter(), s.wg.Done, nil
		}
		s.wg.Done()
		return nil, nil, fmt.Errorf("failed to get rate limiter: missing tenant for scope: %s", s.scope)
	}

	limiter := s.newRateLimiter(tenant)
	actual, loaded := s.limiters.LoadOrStore(tenant, limiter)
	cre := s.scope.RoundCRE(contexts.CREValue(ctx))
	if !loaded {
		limiter.cre.Store(cre)
		go limiter.updateLoop(cre)
	} else {
		limiter = actual.(*rateLimiter)
		limiter.updateCRE(cre)
	}
	return limiter, s.wg.Done, nil
}

func (s *scopedRateLimiter) newRateLimiter(tenant string) *rateLimiter {
	l := &rateLimiter{
		key:     s.key,
		scope:   s.scope,
		tenant:  tenant,
		updater: newUpdater[config.Rate](logger.With(s.lggr, s.scope.String(), tenant), s.rateFn, s.subFn),
		limiter: rate.NewLimiter(s.defaultRate.Limit, s.defaultRate.Burst),
		recordLimit: func(ctx context.Context, value float64) {
			if s.limitGauge != nil {
				s.limitGauge.Record(ctx, value, withScope(ctx, s.scope))
			}
		},
		recordBurst: func(ctx context.Context, value int64) {
			if s.burstGauge != nil {
				s.burstGauge.Record(ctx, value, withScope(ctx, s.scope))
			}
		},
		addUsage: func(ctx context.Context, incr int64) {
			if s.usageCounter != nil {
				s.usageCounter.Add(ctx, incr, withScope(ctx, s.scope))
			}
		},
		recordDenied: func(ctx context.Context, value int) {
			if s.deniedHist != nil {
				s.deniedHist.Record(ctx, int64(value), withScope(ctx, s.scope))
			}
		},
	}
	l.updater.recordLimit = func(ctx context.Context, r config.Rate) {
		l.limiter.SetLimit(r.Limit)
		l.limiter.SetBurst(r.Burst)
		l.recordLimit(ctx, float64(r.Limit))
		l.recordBurst(ctx, int64(r.Burst))
	}
	return l
}

func (s *scopedRateLimiter) Limit(ctx context.Context) (config.Rate, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return config.Rate{}, err
	}
	defer done()
	return l.Limit(ctx)
}

func (s *scopedRateLimiter) Allow(ctx context.Context) bool {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return false
	}
	defer done()
	return l.Allow(ctx)
}

func (s *scopedRateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return false
	}
	defer done()
	return l.AllowN(ctx, t, n)
}

func (s *scopedRateLimiter) AllowErr(ctx context.Context) error {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return l.AllowErr(ctx)
}

func (s *scopedRateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return l.AllowNErr(ctx, t, n)
}

func (s *scopedRateLimiter) Reserve(ctx context.Context) (Reservation, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return l.Reserve(ctx)
}

func (s *scopedRateLimiter) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return l.ReserveN(ctx, t, n)
}

func (s *scopedRateLimiter) Wait(ctx context.Context) (err error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return l.Wait(ctx)
}

func (s *scopedRateLimiter) WaitN(ctx context.Context, n int) (err error) {
	l, done, err := s.getOrCreate(ctx)
	if err != nil {
		return err
	}
	defer done()
	return l.WaitN(ctx, n)
}

var _ RateLimiter = MultiRateLimiter{}

// MultiRateLimiter is a RateLimiter composed of other RateLimiters which are applied in order.
type MultiRateLimiter []RateLimiter

func (m MultiRateLimiter) Close() (err error) {
	for _, l := range m {
		err = errors.Join(err, l.Close())
	}
	return
}

func (m MultiRateLimiter) Limit(ctx context.Context) (config.Rate, error) {
	if len(m) == 0 {
		return config.Rate{}, fmt.Errorf("empty")
	}
	return m[0].Limit(ctx)
}

func (m MultiRateLimiter) Allow(ctx context.Context) bool {
	return m.AllowN(ctx, time.Now(), 1)
}

func (m MultiRateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	var mr multiReservation
	for _, l := range m {
		r, err := l.ReserveN(ctx, t, n)
		if err != nil || !r.Allow() {
			mr.Cancel()
			return false
		}
		mr = append(mr, r)
	}
	return true
}

func (m MultiRateLimiter) AllowErr(ctx context.Context) error {
	return m.AllowNErr(ctx, time.Now(), 1)
}

func (m MultiRateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	var mr multiReservation
	for _, l := range m {
		r, err := l.ReserveN(ctx, t, n)
		if err != nil {
			mr.Cancel()
			return err
		} else if err = r.AllowErr(); err != nil {
			mr.Cancel()
			return err
		}
		mr = append(mr, r)
	}
	return nil
}

func (m MultiRateLimiter) Reserve(ctx context.Context) (Reservation, error) {
	var mr multiReservation
	for _, l := range m {
		r, err := l.Reserve(ctx)
		if err != nil {
			mr.Cancel()
			return nil, err
		}
		mr = append(mr, r)
	}
	return &mr, nil
}

func (m MultiRateLimiter) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	var mr multiReservation
	for _, l := range m {
		r, err := l.ReserveN(ctx, t, n)
		if err != nil {
			mr.Cancel()
			return nil, err
		}
		mr = append(mr, r)
	}
	return &mr, nil
}

func (m MultiRateLimiter) Wait(ctx context.Context) error {
	var mr multiReservation
	for _, l := range m {
		r, err := l.Reserve(ctx)
		if err != nil {
			mr.Cancel()
			return err
		}
		mr = append(mr, r)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(mr.Delay()):
	}
	return nil
}

func (m MultiRateLimiter) WaitN(ctx context.Context, n int) (err error) {
	var mr multiReservation
	for _, l := range m {
		r, err := l.ReserveN(ctx, time.Now(), n)
		if err != nil {
			mr.Cancel()
			return err
		}
		mr = append(mr, r)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(mr.Delay()):
	}
	return nil
}

var _ Reservation = (*multiReservation)(nil)

type multiReservation []Reservation

func (m *multiReservation) OK() bool {
	for _, r := range *m {
		if !r.OK() {
			return false
		}
	}
	return true
}

func (m *multiReservation) Delay() time.Duration {
	var latest time.Time
	for _, r := range *m {
		if t := time.Now().Add(r.Delay()); t.After(latest) {
			latest = t
		}
	}
	return time.Until(latest)
}

func (m *multiReservation) DelayFrom(t time.Time) time.Duration {
	var latest time.Time
	for _, r := range *m {
		if t := time.Now().Add(r.DelayFrom(t)); t.After(latest) {
			latest = t
		}
	}
	return time.Until(latest)
}

func (m *multiReservation) Cancel() {
	for _, r := range *m {
		r.Cancel()
	}
}

func (m *multiReservation) CancelAt(t time.Time) {
	for _, r := range *m {
		r.CancelAt(t)
	}
}

func (m *multiReservation) Allow() bool {
	for _, r := range *m {
		if !r.Allow() {
			return false
		}
	}
	return true
}

func (m *multiReservation) AllowErr() error {
	for _, r := range *m {
		if err := r.AllowErr(); err != nil {
			return err
		}
	}
	return nil
}

type unlimitedRate struct{}

func (r unlimitedRate) Limit(context.Context) (config.Rate, error) {
	return config.Rate{Limit: rate.Inf, Burst: math.MaxInt}, nil
}

// UnlimitedRateLimiter returns a RateLimiter without any limit. Every call is allowed, all reservations are accepted
// without delay, and no calls have to wait.
func UnlimitedRateLimiter() RateLimiter { return unlimitedRate{} }

func (r unlimitedRate) Close() error { return nil }

func (r unlimitedRate) Allow(ctx context.Context) bool { return true }

func (r unlimitedRate) AllowN(ctx context.Context, t time.Time, n int) bool { return true }

func (r unlimitedRate) AllowErr(ctx context.Context) error { return nil }

func (r unlimitedRate) AllowNErr(ctx context.Context, t time.Time, n int) error { return nil }

func (r unlimitedRate) Reserve(ctx context.Context) (Reservation, error) {
	return unlimitedReservation{}, nil
}

func (r unlimitedRate) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	return unlimitedReservation{}, nil
}

func (r unlimitedRate) Wait(ctx context.Context) (err error) { return nil }

func (r unlimitedRate) WaitN(ctx context.Context, n int) (err error) { return nil }

// unlimitedReservation is a Reservation for any amount with no delay.
type unlimitedReservation struct{}

func (r unlimitedReservation) OK() bool { return true }

func (r unlimitedReservation) Delay() time.Duration { return 0 }

func (r unlimitedReservation) DelayFrom(t time.Time) time.Duration { return 0 }

func (r unlimitedReservation) Cancel() {}

func (r unlimitedReservation) CancelAt(t time.Time) {}

func (r unlimitedReservation) Allow() bool { return true }

func (r unlimitedReservation) AllowErr() error { return nil }
