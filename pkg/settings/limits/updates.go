package limits

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

type updater[N any] struct {
	lggr        logger.Logger
	getLimitFn  func(context.Context) (N, error)
	subFn       func(ctx context.Context) (<-chan settings.Update[N], func()) // optional
	recordLimit func(context.Context, N)

	creCh chan struct{}
	cre   atomic.Value

	stopOnce  sync.Once
	stopCh    services.StopChan
	done      chan struct{}
	cancelSub func() // optional
}

func newUpdater[N any](lggr logger.Logger, getLimitFn func(context.Context) (N, error), subFn func(ctx context.Context) (<-chan settings.Update[N], func())) *updater[N] {
	if lggr == nil {
		lggr = logger.Nop()
	}
	return &updater[N]{
		lggr:        lggr,
		getLimitFn:  getLimitFn,
		subFn:       subFn,
		recordLimit: func(ctx context.Context, n N) {}, // no-op
		creCh:       make(chan struct{}, 1),
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

func (u *updater[N]) updateCRE(cre contexts.CRE) {
	cur := u.cre.Load().(contexts.CRE)
	if cur == cre {
		return
	}
	u.cre.Store(cre)
	select {
	case u.creCh <- struct{}{}:
	default:
	}
}

// opt: reap after period of non-use
func (u *updater[N]) updateLoop(cre contexts.CRE) {
	defer close(u.done)
	ctx, cancel := u.stopCh.NewCtx()
	defer cancel()

	var updates <-chan settings.Update[N]
	var cancelSub func()
	var c <-chan time.Time
	if u.subFn != nil {
		updates, cancelSub = u.subFn(contexts.WithCRE(ctx, cre))
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
			limit, err := u.getLimitFn(contexts.WithCRE(ctx, cre))
			if err != nil {
				u.lggr.Errorw("Failed to get limit. Using default value", "default", limit, "err", err)
			}
			u.recordLimit(ctx, limit)

		case update := <-updates:
			if update.Err != nil {
				u.lggr.Errorw("Failed to update limit. Using default value", "default", update.Value, "err", update.Err)
			}
			u.recordLimit(ctx, update.Value)

		case <-u.creCh:
			cre = u.cre.Load().(contexts.CRE)
			if u.subFn != nil {
				cancelSub()
				updates, cancelSub = u.subFn(contexts.WithCRE(ctx, cre))
			}
		}
	}
}
