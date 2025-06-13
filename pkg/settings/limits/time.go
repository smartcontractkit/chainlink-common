package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// TimeLimiter is a Limiter that enforces timeouts.
type TimeLimiter interface {
	Limiter
	// WithTimeout is like context.WithTimeout, but automatically applies the timeout
	// from this TimeLimiter, and returns a done func() that must be called to signal completion.
	WithTimeout(context.Context) (ctx context.Context, done func(), err error)
}

// NewTimeLimiter returns a simple TimeLimiter with the given time out.
func NewTimeLimiter(timeout time.Duration) TimeLimiter {
	return &timeLimiter{
		updater: newUpdater[time.Duration](nil, func(ctx context.Context) (time.Duration, error) {
			return timeout, nil
		}, nil),
		defaultTimeout: timeout,
	}
}

func (f Factory) newTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	l := &timeLimiter{
		updater: newUpdater[time.Duration](nil, func(ctx context.Context) (time.Duration, error) {
			return timeout.GetOrDefault(ctx, f.Settings)
		}, nil),
		key:            timeout.Key,
		defaultTimeout: timeout.DefaultValue,
	}
	l.updater.recordLimit = l.recordTimeout
	if f.Logger != nil {
		l.lggr = logger.Sugared(f.Logger).Named("TimeLimiter").With("key", timeout.Key)
	}
	if f.Settings != nil {
		if r, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[time.Duration], func()) {
				return timeout.Subscribe(ctx, r)
			}
		}
		if timeout.Scope == settings.ScopeGlobal {
			l.updateCRE(contexts.CRE{})
			go l.updateLoop(contexts.CRE{})
		}
	}
	if f.Meter != nil {
		var err error
		l.timeoutGauge, err = f.Meter.Int64Gauge("time." + timeout.Key + ".t")
		if err != nil {
			return nil, err
		}
		l.runtimeGauge, err = f.Meter.Int64Histogram("time." + timeout.Key + ".usage")
		if err != nil {
			return nil, err
		}
		l.timeoutCounter, err = f.Meter.Int64Counter("time." + timeout.Key + ".timeout")
		if err != nil {
			return nil, err
		}
		l.successCounter, err = f.Meter.Int64Counter("time." + timeout.Key + ".success")
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

type timeLimiter struct {
	*updater[time.Duration]
	defaultTimeout time.Duration

	key   string // optional
	scope settings.Scope

	timeoutGauge   metric.Int64Gauge     // optional
	runtimeGauge   metric.Int64Histogram // optional
	timeoutCounter metric.Int64Counter   // optional
	successCounter metric.Int64Counter   // optional

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[time.Duration]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (l *timeLimiter) Close() (err error) {
	close(l.stopCh)
	l.wg.Wait()

	// cleanup
	if l.scope == settings.ScopeGlobal {
		return l.updater.Close()
	}
	l.updaters.Range(func(key, value interface{}) bool {
		// opt: parallelize
		err = errors.Join(err, value.(*updater[time.Duration]).Close())
		return true
	})
	return
}

func (l *timeLimiter) recordTimeout(ctx context.Context, to time.Duration) {
	if l.timeoutGauge == nil {
		return
	}
	l.timeoutGauge.Record(ctx, int64(to), withScope(ctx, l.scope))
}

func (l *timeLimiter) recordRuntime(ctx context.Context, elapsed time.Duration) {
	if l.runtimeGauge == nil {
		return
	}
	l.runtimeGauge.Record(ctx, int64(elapsed), withScope(ctx, l.scope))
}

func (l *timeLimiter) countTimeout(ctx context.Context) {
	if l.timeoutCounter == nil {
		return
	}
	l.timeoutCounter.Add(ctx, 1, withScope(ctx, l.scope))
}

func (l *timeLimiter) countSuccess(ctx context.Context) {
	if l.successCounter == nil {
		return
	}
	l.successCounter.Add(ctx, 1, withScope(ctx, l.scope))
}

func (l *timeLimiter) WithTimeout(ctx context.Context) (context.Context, func(), error) {
	if err := l.wg.TryAdd(1); err != nil {
		return nil, nil, err
	}
	defer l.wg.Done()

	var tenant string
	if l.scope != settings.ScopeGlobal {
		tenant = l.scope.Value(ctx)
		if tenant == "" {
			if !l.scope.IsTenantRequired() {
				kvs := contexts.CREValue(ctx).LoggerKVs()
				l.lggr.Warnw("Unable to apply scoped time limit due to missing tenant: failing open", append([]any{"scope", l.scope}, kvs...)...)
				return ctx, nil, nil
			}
			return nil, nil, fmt.Errorf("unable to apply scoped time limit due to missing tenant for scope: %s", l.scope)
		}

		u := newUpdater(l.lggr, l.getLimitFn, l.subFn)
		actual, loaded := l.updaters.LoadOrStore(tenant, u)
		cre := l.scope.RoundCRE(contexts.CREValue(ctx))
		if !loaded {
			u.cre.Store(cre)
			go u.updateLoop(cre)
		} else {
			u = actual.(*updater[time.Duration])
			u.updateCRE(cre)
		}
	}

	to, err := l.getLimitFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to get limit. Using default value", "default", to, "err", err)
	}
	l.recordTimeout(ctx, to)

	countTimeout := func() { l.countTimeout(ctx) } // constructing this first to reference the original ctx
	ctx, cancel := context.WithTimeoutCause(ctx, to, ErrorTimeLimited{Key: l.key, Scope: l.scope, Tenant: tenant, Timeout: to})
	stop := context.AfterFunc(ctx, countTimeout)

	start := time.Now()
	return ctx, func() {
		elapsed := time.Since(start)

		l.recordRuntime(ctx, elapsed)
		if stop() {
			l.countSuccess(ctx)
		}
		cancel()
	}, nil
}
