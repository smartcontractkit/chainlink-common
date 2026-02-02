package limits

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// TimeLimiter is a Limiter that enforces timeouts.
type TimeLimiter interface {
	Limiter[time.Duration]
	// WithTimeout is like context.WithTimeout, but automatically applies the timeout
	// from this TimeLimiter, and returns a done func() that must be called to signal completion.
	WithTimeout(context.Context) (ctx context.Context, done func(), err error)
}

// NewTimeLimiter returns a simple TimeLimiter with the given time out.
func NewTimeLimiter(timeout time.Duration) TimeLimiter {
	return &simpleTimeLimiter{timeout: timeout}
}

type simpleTimeLimiter struct {
	timeout time.Duration
	closed  atomic.Bool
}

func (s *simpleTimeLimiter) Limit(ctx context.Context) (time.Duration, error) {
	return s.timeout, nil
}

func (s *simpleTimeLimiter) Close() error { s.closed.Store(true); return nil }

func (s *simpleTimeLimiter) WithTimeout(ctx context.Context) (context.Context, func(), error) {
	if s.closed.Load() {
		return nil, nil, errors.New("closed")
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, s.timeout)
	return ctx, cancel, nil
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
		go l.updateLoop(context.Background())
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
		l.updaters.Range(func(key, value any) bool {
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

	tenant, timeout, err := l.get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if tenant == "" && l.scope != settings.ScopeGlobal {
		return ctx, func() {}, nil // fail open
	}

	countTimeout := func() { l.countTimeout(ctx) } // constructing this first to reference the original ctx
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, ErrorTimeLimited{Key: l.key, Scope: l.scope, Tenant: tenant, Timeout: timeout})
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

func (l *timeLimiter) Limit(ctx context.Context) (time.Duration, error) {
	if err := l.wg.TryAdd(1); err != nil {
		return -1, err
	}
	defer l.wg.Done()

	tenant, timeout, err := l.get(ctx)
	if err != nil {
		return -1, err
	}
	if tenant == "" && l.scope != settings.ScopeGlobal {
		return -1, nil // fail open
	}

	return timeout, nil
}

func (l *timeLimiter) get(ctx context.Context) (tenant string, timeout time.Duration, err error) {
	if l.scope != settings.ScopeGlobal {
		tenant = l.scope.Value(ctx)
		if tenant == "" {
			if !l.scope.IsTenantRequired() {
				kvs := contexts.CREValue(ctx).LoggerKVs()
				l.lggr.Warnw("Unable to get scoped time limit due to missing tenant: failing open", append([]any{"scope", l.scope}, kvs...)...)
				return
			}
			err = fmt.Errorf("unable to get scoped time limit due to missing tenant for scope: %s", l.scope)
			return
		}

		u := newUpdater(l.lggr, l.getLimitFn, l.subFn)
		actual, loaded := l.updaters.LoadOrStore(tenant, u)
		creCtx := contexts.WithCRE(ctx, l.scope.RoundCRE(contexts.CREValue(ctx)))
		if !loaded {
			go u.updateLoop(creCtx)
		} else {
			u = actual.(*updater[time.Duration])
			u.updateCtx(creCtx)
		}
	}

	timeout, err = l.getLimitFn(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to get limit. Using default value", "default", timeout, "err", err)
	}
	l.recordTimeout(ctx, timeout)

	return
}
