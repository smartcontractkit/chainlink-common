package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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
	Limiter

	Allow(ctx context.Context) bool
	AllowN(ctx context.Context, t time.Time, n int) bool
	AllowErr(ctx context.Context) error
	AllowNErr(ctx context.Context, t time.Time, n int) error

	Reserve(ctx context.Context) (Reservation, error)
	ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error)

	Wait(ctx context.Context) (err error)
	WaitN(ctx context.Context, n int) (err error)
}

var _ Reservation = (*rate.Reservation)(nil)

// Reservation is the exported interface of [*rate.Reservation].
type Reservation interface {
	OK() bool
	Delay() time.Duration
	DelayFrom(time.Time) time.Duration
	Cancel()
	CancelAt(time.Time)
}

// NewRateLimiter returns a RateLimiter for the given limit and burst.
func NewRateLimiter(limit rate.Limit, burst int) RateLimiter {
	return &rateLimiter{
		lggr: logger.Nop(),
		rateFn: func(ctx context.Context) (config.Rate, error) {
			return config.Rate{Limit: limit, Burst: burst}, nil
		},
		limiter: rate.NewLimiter(limit, burst),
		creCh:   make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (f Factory) globalRateLimiter(limit settings.Setting[config.Rate]) (RateLimiter, error) {
	l := &rateLimiter{
		recordLimit: func(ctx context.Context, value float64) {},
		recordBurst: func(ctx context.Context, value int64) {},
		addUsage:    func(ctx context.Context, incr int64) {},

		limiter: rate.NewLimiter(limit.DefaultValue.Limit, limit.DefaultValue.Burst),
		creCh:   make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
		done:    make(chan struct{}),
	}
	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
		l.lggr = logger.Sugared(f.Logger).Named("RateLimiter").With("key", limit.Key)
	}
	if f.Settings != nil {
		l.rateFn = func(ctx context.Context) (config.Rate, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}
		if registry, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[config.Rate], func()) {
				return limit.Subscribe(ctx, registry)
			}
		}
		l.cre.Store(contexts.CRE{})
		go l.updateLoop(contexts.CRE{})
	}
	if f.Meter != nil {
		if limitGauge, err := f.Meter.Float64Gauge("rate." + limit.Key + ".limit"); err != nil {
			return nil, err
		} else {
			l.recordLimit = func(ctx context.Context, value float64) { limitGauge.Record(ctx, value) }
		}
		if burstGauge, err := f.Meter.Int64Gauge("rate." + limit.Key + ".burst"); err != nil {
			return nil, err
		} else {
			l.recordBurst = func(ctx context.Context, value int64) { burstGauge.Record(ctx, value) }
		}
		if usageCounter, err := f.Meter.Int64Counter("rate." + limit.Key + ".usage"); err != nil {
			return nil, err
		} else {
			l.addUsage = func(ctx context.Context, value int64) { usageCounter.Add(ctx, value) }
		}
	}

	return l, nil
}

type rateLimiter struct {
	lggr logger.Logger

	key string // optional

	rateFn func(context.Context) (config.Rate, error)
	subFn  func(context.Context) (<-chan settings.Update[config.Rate], func()) // optional

	recordLimit func(ctx context.Context, value float64)
	recordBurst func(ctx context.Context, value int64)
	addUsage    func(ctx context.Context, incr int64)

	limiter *rate.Limiter

	creCh chan struct{}
	cre   atomic.Value

	stopOnce sync.Once
	stopCh   services.StopChan
	done     chan struct{}
}

func (l *rateLimiter) Close() error {
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
	<-l.done
	return nil

}

func (l *rateLimiter) updateCRE(cre contexts.CRE) {
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

func (l *rateLimiter) updateLoop(cre contexts.CRE) {
	defer close(l.done)
	ctx, cancel := l.stopCh.NewCtx()
	defer cancel()

	var updates <-chan settings.Update[config.Rate]
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
			r := l.getRate(contexts.WithCRE(ctx, cre))
			l.limiter.SetLimit(r.Limit)
			l.limiter.SetBurst(r.Burst)
			l.recordLimit(ctx, float64(r.Limit))
			l.recordBurst(ctx, int64(r.Burst))

		case update := <-updates:
			if update.Err != nil {
				l.lggr.Errorw("Failed to update rate. Using default value", "default", update.Value, "err", update.Err)
			}
			l.limiter.SetLimit(update.Value.Limit)
			l.limiter.SetBurst(update.Value.Burst)
			l.recordLimit(ctx, float64(update.Value.Limit))
			l.recordBurst(ctx, int64(update.Value.Burst))

		case <-l.creCh:
			cre = l.cre.Load().(contexts.CRE)
			if l.subFn != nil {
				cancelSub()
				updates, cancelSub = l.subFn(contexts.WithCRE(ctx, cre))
			}
		}
	}
}

func (l *rateLimiter) getRate(ctx context.Context) config.Rate {
	r, err := l.rateFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to update rate. Using default value", "default", r, "err", err)
	}
	return r
}

func (l *rateLimiter) Allow(ctx context.Context) bool {
	if l.limiter.Allow() {
		l.addUsage(ctx, 1)
		return true
	}
	return false
}

func (l *rateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	if l.limiter.AllowN(t, n) {
		l.addUsage(ctx, int64(n))
		return true
	}
	return false
}

func (l *rateLimiter) AllowErr(ctx context.Context) error {
	_, err := l.Reserve(ctx)
	return err
}

func (l *rateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	_, err := l.ReserveN(ctx, t, n)
	return err
}

func (l *rateLimiter) Reserve(ctx context.Context) (Reservation, error) {
	r := l.limiter.Reserve()
	if r.OK() {
		l.addUsage(ctx, 1)
		return r, nil
	}
	return nil, ErrorRateLimited{Key: l.key, N: 1}
}

func (l *rateLimiter) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	r := l.limiter.ReserveN(t, n)
	if r.OK() {
		l.addUsage(ctx, int64(n))
		return r, nil
	}
	return nil, ErrorRateLimited{Key: l.key, N: n}
}

func (l *rateLimiter) Wait(ctx context.Context) error {
	if err := l.limiter.Wait(ctx); err != nil {
		return ErrorRateLimited{Key: l.key, N: 1, Err: err}
	}
	l.addUsage(ctx, 1)
	return nil
}

func (l *rateLimiter) WaitN(ctx context.Context, n int) error {
	if err := l.limiter.WaitN(ctx, n); err != nil {
		return ErrorRateLimited{Key: l.key, N: n, Err: err}
	}
	l.addUsage(ctx, int64(n))
	return nil
}

func (f Factory) newScopedRateLimiter(limit settings.Setting[config.Rate]) (RateLimiter, error) {
	l := &scopedRateLimiter{
		scope:       limit.Scope,
		defaultRate: limit.DefaultValue,
	}

	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
		l.lggr = logger.Sugared(f.Logger).Named("RateLimiter").With("key", limit.Key, "scope", limit.Scope)
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

	if f.Meter != nil {
		var err error
		l.limitGauge, err = f.Meter.Float64Gauge("rate." + limit.Key + ".limit")
		if err != nil {
			return nil, err
		}

		l.burstGauge, err = f.Meter.Int64Gauge("rate." + limit.Key + ".burst")
		if err != nil {
			return nil, err
		}

		l.usageCounter, err = f.Meter.Int64Counter("rate." + limit.Key + ".usage")
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

type scopedRateLimiter struct {
	lggr        logger.Logger
	scope       settings.Scope
	defaultRate config.Rate

	rateFn func(context.Context) (config.Rate, error)
	subFn  func(ctx context.Context) (<-chan settings.Update[config.Rate], func()) // optional

	limitGauge   metric.Float64Gauge // optional
	burstGauge   metric.Int64Gauge   // optional
	usageCounter metric.Int64Counter // optional

	//TODO when to reap?
	limiters sync.Map // map[string]*rateLimiter
}

func (s *scopedRateLimiter) Close() (err error) {
	s.limiters.Range(func(tenant, value interface{}) bool {
		err = errors.Join(err, value.(*rateLimiter).Close())
		return true
	})
	return
}

func (s *scopedRateLimiter) get(ctx context.Context) (string, *rateLimiter, error) {
	tenant := s.scope.Value(ctx)
	if tenant == "" {
		return "", nil, fmt.Errorf("failed to get rate limiter: missing tenant for scope: %s", s.scope.String())
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
	return tenant, limiter, nil
}

func (s *scopedRateLimiter) newRateLimiter(tenant string) *rateLimiter {
	return &rateLimiter{
		lggr:    logger.With(s.lggr, s.scope.String(), tenant),
		stopCh:  make(chan struct{}),
		rateFn:  s.rateFn,
		subFn:   s.subFn,
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
	}
}

func (s *scopedRateLimiter) Allow(ctx context.Context) bool {
	_, l, err := s.get(ctx)
	if err != nil {
		return false
	}
	return l.Allow(ctx)
}

func (s *scopedRateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	_, l, err := s.get(ctx)
	if err != nil {
		return false
	}
	return l.AllowN(ctx, t, n)
}

func (s *scopedRateLimiter) AllowErr(ctx context.Context) error {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return err
	}
	err = l.AllowErr(ctx)
	if err != nil {
		return augmentErrorRateLimited(err, s.scope, tenant)
	}
	return nil
}

func augmentErrorRateLimited(err error, scope settings.Scope, tenant string) error {
	var erl ErrorRateLimited
	if err != nil && errors.As(err, &erl) {
		erl.Scope = scope
		erl.Tenant = tenant
		return erl
	}
	return err
}

func (s *scopedRateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return err
	}
	err = l.AllowNErr(ctx, t, n)
	if err != nil {
		return augmentErrorRateLimited(err, s.scope, tenant)
	}
	return nil
}

func (s *scopedRateLimiter) Reserve(ctx context.Context) (Reservation, error) {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.Reserve(ctx)
	if err != nil {
		return nil, augmentErrorRateLimited(err, s.scope, tenant)
	}
	return r, nil
}

func (s *scopedRateLimiter) ReserveN(ctx context.Context, t time.Time, n int) (Reservation, error) {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return nil, err
	}
	r, err := l.ReserveN(ctx, t, n)
	if err != nil {
		return nil, augmentErrorRateLimited(err, s.scope, tenant)
	}
	return r, nil
}

func (s *scopedRateLimiter) Wait(ctx context.Context) (err error) {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return err
	}
	err = l.Wait(ctx)
	if err != nil {
		return augmentErrorRateLimited(err, s.scope, tenant)
	}
	return nil
}

func (s *scopedRateLimiter) WaitN(ctx context.Context, n int) (err error) {
	tenant, l, err := s.get(ctx)
	if err != nil {
		return err
	}
	err = l.WaitN(ctx, n)
	if err != nil {
		return augmentErrorRateLimited(err, s.scope, tenant)
	}
	return nil
}

var _ RateLimiter = MultiRateLimiter{}

// MultiRateLimiter is a RateLimiter composed of other RateLimiters which are applied in order.
// All RateLimiters must agree. If any blocks
type MultiRateLimiter []RateLimiter

func (m MultiRateLimiter) Close() (err error) {
	for _, l := range m {
		err = errors.Join(err, l.Close())
	}
	return
}

func (m MultiRateLimiter) Allow(ctx context.Context) bool {
	var cancels []func()
	for _, l := range m {
		r, err := l.Reserve(ctx)
		if err != nil || !r.OK() {
			for _, cancel := range cancels {
				cancel()
			}
			return false
		}
		cancels = append(cancels, r.Cancel)
	}
	return true
}

func (m MultiRateLimiter) AllowN(ctx context.Context, t time.Time, n int) bool {
	var cancels []func()
	for _, l := range m {
		r, err := l.ReserveN(ctx, t, n)
		if err != nil {
			for _, cancel := range cancels {
				cancel()
			}
			return false
		}
		cancels = append(cancels, r.Cancel)
	}
	return true
}

func (m MultiRateLimiter) AllowErr(ctx context.Context) error {
	_, err := m.Reserve(ctx)
	return err
}

func (m MultiRateLimiter) AllowNErr(ctx context.Context, t time.Time, n int) error {
	_, err := m.ReserveN(ctx, t, n)
	return err
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
