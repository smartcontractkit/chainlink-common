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

type GateLimiter interface {
	Limiter[bool]
	AllowErr(context.Context) error
}

func NewGateLimiter(open bool) GateLimiter {
	return &simpleGateLimiter{open: open}
}

type simpleGateLimiter struct {
	open   bool
	closed atomic.Bool
}

func (s *simpleGateLimiter) Close() error { s.closed.Store(true); return nil }

func (s *simpleGateLimiter) Limit(ctx context.Context) (bool, error) {
	return s.open, nil
}

func (s *simpleGateLimiter) AllowErr(ctx context.Context) error {
	if ok, err := s.Limit(ctx); err != nil {
		return err
	} else if !ok {
		return ErrorNotAllowed{}
	}
	return nil
}

func newGateLimiter(f Factory, limit settings.SettingSpec[bool]) (GateLimiter, error) {
	g := &gateLimiter{
		updater: newUpdater[bool](nil, func(ctx context.Context) (bool, error) {
			return limit.GetOrDefault(ctx, f.Settings)
		}, nil),
		key:   limit.GetKey(),
		scope: limit.GetScope(),
	}
	g.updater.recordLimit = func(ctx context.Context, b bool) { g.recordStatus(ctx, b) }

	if f.Meter != nil {
		if g.key == "" {
			return nil, errors.New("metrics require Key to be set")
		}
		unit := limit.GetUnit()
		limitGauge, err := f.Meter.Int64Gauge("gate."+g.key+".limit", metric.WithUnit(unit))
		if err != nil {
			return nil, err
		}
		g.recordStatus = func(ctx context.Context, b bool, options ...metric.RecordOption) {
			var val int64
			if b {
				val = 1
			}
			limitGauge.Record(ctx, val, options...)
		}
		usageCounter, err := f.Meter.Int64Counter("gate."+g.key+".usage", metric.WithUnit(unit))
		if err != nil {
			return nil, err
		}
		g.recordUsage = func(ctx context.Context, options ...metric.AddOption) {
			usageCounter.Add(ctx, 1, options...)
		}
		deniedCounter, err := f.Meter.Int64Counter("gate."+g.key+".denied", metric.WithUnit(unit))
		if err != nil {
			return nil, err
		}
		g.recordDenied = func(ctx context.Context, options ...metric.AddOption) {
			deniedCounter.Add(ctx, 1, options...)
		}
	} else {
		g.recordStatus = func(ctx context.Context, value bool, options ...metric.RecordOption) {}
		g.recordUsage = func(ctx context.Context, options ...metric.AddOption) {}
		g.recordDenied = func(ctx context.Context, options ...metric.AddOption) {}
	}

	if f.Logger != nil {
		g.lggr = logger.Sugared(f.Logger).Named("GateLimiter").With("key", limit.GetKey())
	}

	// OPT: support settings.Registry subscriptions
	//if f.Settings != nil {
	//	if r, ok := f.Settings.(settings.Registry); ok {
	//		g.subFn = func(ctx context.Context) (<-chan settings.Update[bool], func()) {
	//			return limit.Subscribe(ctx, r)
	//		}
	//	}
	//}

	// OPT: restore with support for SettingMap
	//if limit.Default.Scope == settings.ScopeGlobal {
	//	g.updateCRE(contexts.CRE{})
	//	go g.updateLoop(contexts.CRE{})
	//}
	close(g.done)

	return g, nil
}

type gateLimiter struct {
	*updater[bool]

	key   string // optional
	scope settings.Scope

	recordStatus func(ctx context.Context, value bool, options ...metric.RecordOption)
	recordUsage  func(ctx context.Context, options ...metric.AddOption)
	recordDenied func(ctx context.Context, options ...metric.AddOption)

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[N]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (g *gateLimiter) Close() (err error) {
	g.wg.Wait()

	// cleanup
	if g.scope == settings.ScopeGlobal {
		return g.updater.Close()
	} else {
		g.updaters.Range(func(key, value any) bool {
			// opt: parallelize
			err = errors.Join(err, value.(*updater[bool]).Close())
			return true
		})
	}
	return
}

func (g *gateLimiter) Limit(ctx context.Context) (bool, error) {
	if err := g.wg.TryAdd(1); err != nil {
		return false, err
	}
	defer g.wg.Done()

	_, limit, err := g.get(ctx)
	if err != nil {
		return false, err
	}

	return limit, nil
}

func (g *gateLimiter) AllowErr(ctx context.Context) error {
	if err := g.wg.TryAdd(1); err != nil {
		return err
	}
	defer g.wg.Done()

	tenant, open, err := g.get(ctx)
	if err != nil {
		return err
	} else if !open {
		g.recordDenied(ctx, withScope(ctx, g.scope))
		return ErrorNotAllowed{Key: g.key, Scope: g.scope, Tenant: tenant}
	}
	g.recordUsage(ctx, withScope(ctx, g.scope))
	return nil
}

func (g *gateLimiter) get(ctx context.Context) (tenant string, open bool, err error) {
	if g.scope != settings.ScopeGlobal {
		tenant = g.scope.Value(ctx)
		if tenant == "" {
			if !g.scope.IsTenantRequired() {
				kvs := contexts.CREValue(ctx).LoggerKVs()
				g.lggr.Warnw("Unable to get scoped gate status due to missing tenant: failing open", append([]any{"scope", g.scope}, kvs...)...)
				return
			}
			err = fmt.Errorf("unable to get scoped gate status due to missing tenant for scope: %s", g.scope)
			return
		}

		u := newUpdater(g.lggr, g.getLimitFn, g.subFn)
		actual, loaded := g.updaters.LoadOrStore(tenant, u)
		cre := g.scope.RoundCRE(contexts.CREValue(ctx))
		if !loaded {
			// OPT: restore with support for SettingMap
			//u.cre.Store(cre)
			//go u.updateLoop(cre)
			close(u.done)
		} else {
			u = actual.(*updater[bool])
			u.updateCRE(cre)
		}
	}

	open, err = g.getLimitFn(ctx)
	if err != nil {
		g.lggr.Errorw("Failed to get status. Using default value", "default", open, "err", err)
	}
	// TODO: include map key in attributes
	g.recordStatus(ctx, open, withScope(ctx, g.scope))
	return
}
