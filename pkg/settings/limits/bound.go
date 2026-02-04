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
type BoundLimiter[N Number] interface {
	Limiter[N]
	// Check returns ErrorBoundLimited if the value is above the limit.
	Check(context.Context, N) error
}

// NewBoundLimiter returns a BoundLimiter with the given bound.
func NewBoundLimiter[N Number](bound N) BoundLimiter[N] {
	return &simpleBoundLimiter[N]{bound: bound}
}

var _ BoundLimiter[int64] = &simpleBoundLimiter[int64]{}

type simpleBoundLimiter[N Number] struct {
	bound  N
	closed atomic.Bool
}

func (s *simpleBoundLimiter[N]) Close() error { s.closed.Store(true); return nil }

func (s *simpleBoundLimiter[N]) Check(ctx context.Context, n N) error {
	if s.closed.Load() {
		return errors.New("closed")
	}
	if n > s.bound {
		return ErrorBoundLimited[N]{Limit: s.bound, Amount: n}
	}
	return nil
}
func (s *simpleBoundLimiter[N]) Limit(ctx context.Context) (N, error) {
	return s.bound, nil
}

func newBoundLimiter[N Number](f Factory, bound settings.SettingSpec[N]) (BoundLimiter[N], error) {
	b := &boundLimiter[N]{
		updater: newUpdater[N](nil, func(ctx context.Context) (N, error) {
			return bound.GetOrDefault(ctx, f.Settings)
		}, nil),
		key:   bound.GetKey(),
		scope: bound.GetScope(),
	}
	b.updater.recordLimit = func(ctx context.Context, n N) { b.recordBound(ctx, n) }

	if f.Meter != nil {
		if b.key == "" {
			return nil, errors.New("metrics require Key to be set")
		}
		newGauge, newHist := metricConstructors[N](f.Meter, bound.GetUnit())

		key := bound.GetKey()
		limitGauge, err := newGauge("bound." + key + ".limit")
		if err != nil {
			return nil, err
		}
		b.recordBound = func(ctx context.Context, value N, options ...metric.RecordOption) {
			limitGauge.Record(ctx, value, options...)
		}
		usageHist, err := newHist("bound." + key + ".usage")
		if err != nil {
			return nil, err
		}
		b.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {
			usageHist.Record(ctx, value, options...)
		}
		deniedHist, err := newHist("bound." + key + ".denied")
		if err != nil {
			return nil, err
		}
		b.recordDenied = func(ctx context.Context, value N, options ...metric.RecordOption) {
			deniedHist.Record(ctx, value, options...)
		}
	} else {
		b.recordBound = func(ctx context.Context, value N, options ...metric.RecordOption) {}
		b.recordUsage = func(ctx context.Context, value N, options ...metric.RecordOption) {}
		b.recordDenied = func(ctx context.Context, value N, options ...metric.RecordOption) {}
	}

	if f.Logger != nil {
		b.lggr = logger.Sugared(f.Logger).Named("BoundLimiter").With("key", bound.GetKey())
	}

	if f.Settings != nil {
		if r, ok := f.Settings.(settings.Registry); ok {
			b.subFn = func(ctx context.Context) (<-chan settings.Update[N], func()) {
				return bound.Subscribe(ctx, r)
			}
		}
	}

	if bound.GetScope() == settings.ScopeGlobal {
		go b.updateLoop(context.Background())
	}

	return b, nil
}

type boundLimiter[N Number] struct {
	*updater[N]

	key   string // optional
	scope settings.Scope

	recordBound  func(ctx context.Context, value N, options ...metric.RecordOption)
	recordUsage  func(ctx context.Context, value N, options ...metric.RecordOption)
	recordDenied func(ctx context.Context, value N, options ...metric.RecordOption)

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[N]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (b *boundLimiter[N]) Close() (err error) {
	b.wg.Wait()

	// cleanup
	if b.scope == settings.ScopeGlobal {
		return b.updater.Close()
	} else {
		b.updaters.Range(func(key, value any) bool {
			// opt: parallelize
			err = errors.Join(err, value.(*updater[N]).Close())
			return true
		})
	}
	return
}

func (b *boundLimiter[N]) Check(ctx context.Context, amount N) error {
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

	if amount > bound {
		b.recordDenied(ctx, amount, withScope(ctx, b.scope))
		return ErrorBoundLimited[N]{Key: b.key, Scope: b.scope, Tenant: tenant, Limit: bound, Amount: amount}
	}

	b.recordUsage(ctx, amount, withScope(ctx, b.scope))
	return nil
}

func (b *boundLimiter[N]) Limit(ctx context.Context) (N, error) {
	var zero N
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

func (b *boundLimiter[N]) get(ctx context.Context) (tenant string, bound N, err error) {
	if b.scope != settings.ScopeGlobal {
		tenant = b.scope.Value(ctx)
		if tenant == "" {
			if !b.scope.IsTenantRequired() {
				kvs := contexts.CREValue(ctx).LoggerKVs()
				b.lggr.Warnw("Unable to get scoped bound limit due to missing tenant: failing open", append([]any{"scope", b.scope}, kvs...)...)
				return
			}
			err = fmt.Errorf("unable to get scoped bound limit due to missing tenant for scope: %s", b.scope)
			return
		}

		u := newUpdater(b.lggr, b.getLimitFn, b.subFn)
		actual, loaded := b.updaters.LoadOrStore(tenant, u)
		creCtx := contexts.WithCRE(ctx, b.scope.RoundCRE(contexts.CREValue(ctx)))
		if !loaded {
			go u.updateLoop(creCtx)
		} else {
			u = actual.(*updater[N])
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
