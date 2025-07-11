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
	tl := &timeLimiter{
		updater: newUpdater[time.Duration](nil, func(ctx context.Context) (time.Duration, error) {
			return timeout, nil
		}, nil),
		defaultTimeout: timeout,
	}
	close(tl.done) // no update routine
	return tl
}

func (f Factory) newTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	l := &timeLimiter{
		updater: newUpdater[time.Duration](nil, func(ctx context.Context) (time.Duration, error) {
			return timeout.GetOrDefault(ctx, f.Settings)
		}, nil),
		key:            timeout.Key,
		defaultTimeout: timeout.DefaultValue,
		scope:          timeout.Scope,
	}
	l.updater.recordLimit = l.recordTimeout

	if f.Logger != nil {
		l.lggr = logger.Sugared(f.Logger).Named("TimeLimiter").With("key", timeout.Key)
	}

	if f.Meter != nil {
		var err error
		l.limitGauge, err = f.Meter.Float64Gauge("time."+timeout.Key+".limit", metric.WithUnit("s"))
		if err != nil {
			return nil, err
		}
		l.runtimeHist, err = f.Meter.Float64Histogram("time."+timeout.Key+".runtime", metric.WithUnit("s"))
		if err != nil {
			return nil, err
		}
		l.timeoutCounter, err = f.Meter.Int64Counter("time."+timeout.Key+".timeout", metric.WithUnit(timeout.Unit))
		if err != nil {
			return nil, err
		}
		l.successCounter, err = f.Meter.Int64Counter("time."+timeout.Key+".success", metric.WithUnit(timeout.Unit))
		if err != nil {
			return nil, err
		}
	}

	if f.Settings != nil {
		if r, ok := f.Settings.(settings.Registry); ok {
			l.subFn = func(ctx context.Context) (<-chan settings.Update[time.Duration], func()) {
				return timeout.Subscribe(ctx, r)
			}
		}
	}

	if timeout.Scope == settings.ScopeGlobal {
		l.updateCRE(contexts.CRE{})
		go l.updateLoop(contexts.CRE{})
	}

	return l, nil
}

type timeLimiter struct {
	*updater[time.Duration]
	defaultTimeout time.Duration

	key   string // optional
	scope settings.Scope

	limitGauge     metric.Float64Gauge     // optional
	runtimeHist    metric.Float64Histogram // optional
	timeoutCounter metric.Int64Counter     // optional
	successCounter metric.Int64Counter     // optional

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[time.Duration]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (l *timeLimiter) Close() (err error) {
	l.wg.Wait()

	// cleanup
	if l.scope == settings.ScopeGlobal {
		return l.updater.Close()
	} else {
		l.updaters.Range(func(key, value interface{}) bool {
			// opt: parallelize
			err = errors.Join(err, value.(*updater[time.Duration]).Close())
			return true
		})
	}
	return
}

func (l *timeLimiter) recordTimeout(ctx context.Context, to time.Duration) {
	if l.limitGauge == nil {
		return
	}
	l.limitGauge.Record(ctx, to.Seconds(), withScope(ctx, l.scope))
}

func (l *timeLimiter) recordRuntime(ctx context.Context, elapsed time.Duration) {
	if l.runtimeHist == nil {
		return
	}
	l.runtimeHist.Record(ctx, elapsed.Seconds(), withScope(ctx, l.scope))
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
