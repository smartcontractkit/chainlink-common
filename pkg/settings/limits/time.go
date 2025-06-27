package limits

import (
	"context"
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
	Limiter
	// WithTimeout is like context.WithTimeout, but automatically applies the timeout
	// from this TimeLimiter. The CancelFunc must be called to signal completion.
	WithTimeout(context.Context) (context.Context, context.CancelFunc)
}

// NewTimeLimiter returns a simple TimeLimiter with the given time out.
func NewTimeLimiter(timeout time.Duration) (TimeLimiter, error) {
	return &timeLimiter{
		defaultTimeout: timeout,
		getTimeout: func(ctx context.Context) (time.Duration, error) {
			return timeout, nil
		},
		creCh:  make(chan struct{}, 1),
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}, nil
}

func (f Factory) newTimeLimiter(timeout settings.Setting[time.Duration]) (TimeLimiter, error) {
	l := &timeLimiter{
		lggr:           f.Logger,
		key:            timeout.Key,
		defaultTimeout: timeout.DefaultValue,
		getTimeout: func(ctx context.Context) (time.Duration, error) {
			return timeout.GetOrDefault(ctx, f.Settings)
		},
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
	if f.Logger == nil {
		l.lggr = logger.Nop()
	} else {
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
	lggr           logger.Logger
	defaultTimeout time.Duration

	getTimeout func(context.Context) (time.Duration, error)
	subFn      func(ctx context.Context) (<-chan settings.Update[time.Duration], func()) // optional

	key   string // optional
	scope settings.Scope

	timeoutGauge   metric.Int64Gauge
	runtimeGauge   metric.Int64Histogram
	timeoutCounter metric.Int64Counter
	successCounter metric.Int64Counter

	creCh chan struct{}
	cre   atomic.Value

	closeOnce sync.Once
	stopCh    services.StopChan
	done      chan struct{}
	cancelSub func() // optional
}

func (l *timeLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.stopCh)
		if l.cancelSub != nil {
			l.cancelSub()
		}
	})
	<-l.done
	return nil
}

func (l *timeLimiter) updateCRE(cre contexts.CRE) {
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

func (l *timeLimiter) updateLoop(cre contexts.CRE) {
	ctx, cancel := l.stopCh.NewCtx()
	defer cancel()

	var updates <-chan settings.Update[time.Duration]
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
			to, err := l.getTimeout(contexts.WithCRE(ctx, cre))
			if err != nil {
				l.lggr.Errorw("Failed to update timeout. Using default value", "default", to, "err", err)
			}
			l.timeoutGauge.Record(ctx, int64(to), withScope(ctx, l.scope))

		case to := <-updates:
			if to.Err != nil {
				l.lggr.Errorw("Failed to update timeout. Using default value", "default", to, "err", to.Err)
			}
			l.timeoutGauge.Record(ctx, int64(to.Value), withScope(ctx, l.scope))

		case <-l.creCh:
			cre = l.cre.Load().(contexts.CRE)
			if l.subFn != nil {
				cancelSub()
				updates, cancelSub = l.subFn(contexts.WithCRE(ctx, cre))
			}
		}
	}
}

//TODO with each use, go updateLoop if not present for this tenant
func (l *timeLimiter) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	to, err := l.getTimeout(ctx)
	if err != nil {
		l.lggr.Errorw("Failed to  update timeout. Using default value", "default", to, "err", err)
	}
	opts := withScope(ctx, l.scope)
	l.timeoutGauge.Record(ctx, int64(to), opts)

	ctx, cancel := context.WithTimeoutCause(ctx, to, ErrorTimeLimited{Key: l.key, Timeout: to})
	stop := context.AfterFunc(ctx, func() {
		l.timeoutCounter.Add(ctx, 1, opts)
	})

	start := time.Now()
	return ctx, func() {
		elapsed := time.Since(start)
		l.runtimeGauge.Record(ctx, int64(elapsed), opts)
		if stop() {
			l.successCounter.Add(ctx, 1, opts)
		}
		cancel()
	}
}
