package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// BoundLimiter is a limiter for simple bounds checks.
type RangeLimiter[N Number] interface {
	Limiter[settings.Range[N]]
	// Check returns ErrorBoundLimited if the value is above the limit.
	Check(context.Context, N) error
}

// NewRangeLimiter returns a RangeLimiter with the given lower bounds.
func NewRangeLimiter[N Number](bounds settings.Range[N]) RangeLimiter[N] {
	return &simpleRangeLimiter[N]{bounds: bounds}
}

var _ RangeLimiter[int64] = &simpleRangeLimiter[int64]{}

type simpleRangeLimiter[N Number] struct {
	bounds settings.Range[N]
	closed atomic.Bool
}

func (s *simpleRangeLimiter[N]) Close() error { s.closed.Store(true); return nil }

func (s *simpleRangeLimiter[N]) Check(ctx context.Context, n N) error {
	if s.closed.Load() {
		return errors.New("closed")
	}
	if !s.bounds.Contains(n) {
		return ErrorRangeLimited[N]{Limit: s.bounds, Amount: n}
	}
	return nil
}
func (s *simpleRangeLimiter[N]) Limit(ctx context.Context) (settings.Range[N], error) {
	return s.bounds, nil
}

func newRangeLimiter[N Number](f Factory, bound settings.SettingSpec[settings.Range[N]]) (RangeLimiter[N], error) {
	b := &rangeLimiter[N]{
		updater: newUpdater[settings.Range[N]](nil, func(ctx context.Context) (settings.Range[N], error) {
			return bound.GetOrDefault(ctx, f.Settings)
		}, nil),
		key:   bound.GetKey(),
		scope: bound.GetScope(),
	}
	b.recordLimit = func(ctx context.Context, n settings.Range[N]) { b.recordBound(ctx, n) }

	if f.Meter != nil {
		if b.key == "" {
			return nil, errors.New("metrics require Key to be set")
		}
		newGauge, newHist := metricConstructors[N](f.Meter, bound.GetUnit())

		key := bound.GetKey()
		lowerLimitGauge, err := newGauge("range." + key + ".lower.limit")
		if err != nil {
			return nil, err
		}
		upperLimitGauge, err := newGauge("range." + key + ".upper.limit")
		if err != nil {
			return nil, err
		}
		b.recordBound = func(ctx context.Context, value settings.Range[N], options ...metric.RecordOption) {
			lowerLimitGauge.Record(ctx, value.Lower, options...)
			upperLimitGauge.Record(ctx, value.Upper, options...)
		}
		usageHist, err := newHist("range." + key + ".usage")
		if err != nil {
			return nil, err
		}
		b.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {
			usageHist.Record(ctx, value, options...)
		}
		deniedHist, err := newHist("range." + key + ".denied")
		if err != nil {
			return nil, err
		}
		b.recordDenied = func(ctx context.Context, value N, options ...metric.RecordOption) {
			deniedHist.Record(ctx, value, options...)
		}
	} else {
		b.recordBound = func(ctx context.Context, value settings.Range[N], options ...metric.RecordOption) {}
		b.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {}
		b.recordDenied = func(ctx context.Context, value N, options ...metric.RecordOption) {}
	}

	if f.Logger != nil {
		b.lggr = logger.Sugared(f.Logger).Named("RangeLimiter").With("key", bound.GetKey())
	}

	if f.Settings != nil {
		if r, ok := f.Settings.(settings.Registry); ok {
			b.subFn = func(ctx context.Context) (<-chan settings.Update[settings.Range[N]], func()) {
				return bound.Subscribe(ctx, r)
			}
		}
	}

	if bound.GetScope() == settings.ScopeGlobal {
		go b.updateLoop(context.Background())
	}

	return b, nil
}

type rangeLimiter[N Number] struct {
	*updater[settings.Range[N]]

	key   string // optional
	scope settings.Scope

	recordBound  func(ctx context.Context, value settings.Range[N], options ...metric.RecordOption)
	recordUsage  func(ctx context.Context, value N, options ...metric.RecordOption)
	recordDenied func(ctx context.Context, value N, options ...metric.RecordOption)

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[Range[N]]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (b *rangeLimiter[N]) Close() (err error) {
	b.wg.Wait()

	// cleanup
	if b.scope == settings.ScopeGlobal {
		return b.updater.Close()
	} else {
		b.updaters.Range(func(key, value any) bool {
			// opt: parallelize
			err = errors.Join(err, value.(*updater[settings.Range[N]]).Close())
			return true
		})
	}
	return
}

// Deprecated: use TryCleanup
func (b *rangeLimiter[N]) EvictTenant(tenant string) error {
	v, loaded := b.updaters.LoadAndDelete(tenant)
	if !loaded {
		return nil
	}
	return v.(*updater[settings.Range[N]]).Close()
}

func (b *rangeLimiter[N]) cleanup(ctx context.Context) {
	tenant := b.scope.Value(ctx)
	if tenant == "" {
		b.lggr.Warnw("Unable to cleanup scoped bounds limiter due to missing tenant", "scope", b.scope)
		return
	}
	v, loaded := b.updaters.LoadAndDelete(tenant)
	if !loaded {
		return
	}
	if err := v.(*updater[settings.Range[N]]).Close(); err != nil {
		b.lggr.Errorw("Failed to close bounds limiter", "tenant", tenant, "err", err)
	}
}

func (b *rangeLimiter[N]) Check(ctx context.Context, amount N) error {
	if err := b.wg.TryAdd(1); err != nil {
		return err
	}
	defer b.wg.Done()

	tenant, bound, err := b.get(ctx)
	if err != nil {
		return err
	}
	if tenant == "" && b.scope != settings.ScopeGlobal {
		return nil // fail open
	}
	if !bound.Contains(amount) {
		b.recordDenied(ctx, amount, withScope(ctx, b.scope))
		return ErrorRangeLimited[N]{Key: b.key, Scope: b.scope, Tenant: tenant, Limit: bound, Amount: amount}
	}

	b.recordUsage(ctx, amount, withScope(ctx, b.scope))
	return nil
}

func (b *rangeLimiter[N]) Limit(ctx context.Context) (settings.Range[N], error) {
	var zero settings.Range[N]
	if err := b.wg.TryAdd(1); err != nil {
		return zero, err
	}
	defer b.wg.Done()

	tenant, bound, err := b.get(ctx)
	if err != nil {
		return zero, err
	}
	if tenant == "" && b.scope != settings.ScopeGlobal {
		return zero, nil // fail open
	}

	return bound, nil
}

func (b *rangeLimiter[N]) get(ctx context.Context) (tenant string, bound settings.Range[N], err error) {
	if b.scope != settings.ScopeGlobal {
		tenant = b.scope.Value(ctx)
		if tenant == "" {
			if !b.scope.IsTenantRequired() {
				kvs := contexts.CREValue(ctx).LoggerKVs()
				b.lggr.Warnw("Unable to get scoped bounds limit due to missing tenant: failing open", append([]any{"scope", b.scope}, kvs...)...)
				return
			}
			err = fmt.Errorf("unable to get scoped bounds limit due to missing tenant for scope: %s", b.scope)
			return
		}

		u := newUpdater(b.lggr, b.getLimitFn, b.subFn)
		actual, loaded := b.updaters.LoadOrStore(tenant, u)
		creCtx := contexts.WithCRE(ctx, b.scope.RoundCRE(contexts.CREValue(ctx)))
		if !loaded {
			go u.updateLoop(creCtx)
		} else {
			u = actual.(*updater[settings.Range[N]])
			u.updateCtx(creCtx)
		}
	}

	bound, err = b.getLimitFn(ctx)
	if err != nil {
		b.lggr.Errorw("Failed to get limit. Using default value", "default", bound, "err", err)
	}
	b.recordBound(ctx, bound, withScope(ctx, b.scope))
	return
}
