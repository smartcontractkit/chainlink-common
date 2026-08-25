package limits

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// updater monitors limit updates via subscriptions or polling and reports them via recordLimit.
// If an updateLoop goroutine is spawned, then Close must be called.
type updater[N any] struct {
	lggr          logger.Logger
	getLimitFn    func(context.Context) (N, error)
	subFn         func(ctx context.Context) (<-chan settings.Update[N], func()) // optional
	recordLimit   func(context.Context, N)
	onLimitUpdate func(context.Context)

	ctxCh chan context.Context // to receive updates

	stopOnce  sync.Once
	stopCh    services.StopChan
	done      chan struct{}
	cancelSub func() // optional

	// lastGoodValue is the fallback used on a read failure, instead of Setting.DefaultValue.
	lastGoodValue atomic.Pointer[N]
}

// newUpdater returns a new updater. lggr and subFn are optional, but getLimitFn is required.
func newUpdater[N any](lggr logger.Logger, getLimitFn func(context.Context) (N, error), subFn func(ctx context.Context) (<-chan settings.Update[N], func())) *updater[N] {
	if lggr == nil {
		lggr = logger.Nop()
	}
	return &updater[N]{
		lggr:        lggr,
		getLimitFn:  getLimitFn,
		subFn:       subFn,
		recordLimit: func(ctx context.Context, n N) {}, // no-op
		ctxCh:       make(chan context.Context, 1),
		stopCh:      make(chan struct{}),
		done:        make(chan struct{}),
	}
}

func (u *updater[N]) Close() error {
	u.stopOnce.Do(func() {
		close(u.stopCh)
		<-u.done
		if u.cancelSub != nil {
			u.cancelSub()
		}
	})
	<-u.done
	return nil
}

func (u *updater[N]) updateCtx(ctx context.Context) {
	select {
	case u.ctxCh <- ctx:
	default:
	}
}

func (u *updater[N]) setLast(n N) {
	u.lastGoodValue.Store(&n)
}

// lastGood returns false if no value has ever resolved successfully.
func (u *updater[N]) lastGood() (N, bool) {
	if v := u.lastGoodValue.Load(); v != nil {
		return *v, true
	}
	var zero N
	return zero, false
}

// updateLoop updates the limit either by subscribing via subFn or polling if subFn is not set. It also processes
// contexts.CRE updates. Stopped by Close.
// opt: reap after period of non-use
func (u *updater[N]) updateLoop(ctx context.Context) {
	defer close(u.done)
	ctx, cancel := u.stopCh.Ctx(context.WithoutCancel(ctx))
	defer cancel()

	var updates <-chan settings.Update[N]
	var cancelSub func()
	var c <-chan time.Time
	if u.subFn != nil {
		updates, cancelSub = u.subFn(ctx)
		defer func() { cancelSub() }() // extra func wrapper is required to ensure we get the final cancelSub value
		// opt: poll now to initialize
	} else {
		t := services.TickerConfig{}.NewTicker(pollPeriod)
		defer t.Stop()
		c = t.C
	}
	for {
		select {
		case <-ctx.Done():
			return

		case <-c:
			limit, err := u.getLimitFn(ctx)
			if err != nil {
				if last, ok := u.lastGood(); ok {
					u.lggr.Errorw("Failed to get limit. Using last known value", "value", last, "err", err)
					limit = last
				} else {
					u.lggr.Errorw("Failed to get limit. Using compiled default", "default", limit, "err", err)
				}
			} else {
				u.setLast(limit)
			}
			u.recordLimit(ctx, limit)
			if u.onLimitUpdate != nil {
				u.onLimitUpdate(ctx)
			}

		case update := <-updates:
			limit := update.Value
			if update.Err != nil {
				if last, ok := u.lastGood(); ok {
					u.lggr.Errorw("Failed to update limit. Using last known value", "value", last, "err", update.Err)
					limit = last
				} else {
					u.lggr.Errorw("Failed to update limit. Using compiled default", "default", update.Value, "err", update.Err)
				}
			} else {
				u.setLast(limit)
			}
			u.recordLimit(ctx, limit)
			if u.onLimitUpdate != nil {
				u.onLimitUpdate(ctx)
			}

		case newCtx := <-u.ctxCh:
			cancel()
			ctx, cancel = u.stopCh.Ctx(newCtx) //nolint
			if u.subFn != nil {
				cancelSub()
				updates, cancelSub = u.subFn(ctx)
			}
			// opt: update now
		}
	}
}
