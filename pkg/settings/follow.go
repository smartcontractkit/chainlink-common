package settings

import (
	"context"
	"time"

	"github.com/jpillora/backoff"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// Follow stores every Snapshot b broadcasts in dst until ctx is done, resubscribing with backoff whenever a
// subscription fails or its stream ends.
func Follow(ctx context.Context, lggr logger.Logger, b Broadcaster, dst *Atomic) {
	bo := backoff.Backoff{Min: 500 * time.Millisecond, Max: time.Minute, Factor: 2, Jitter: true}
	follow(ctx, lggr, b, dst, &bo)
}

func follow(ctx context.Context, lggr logger.Logger, b Broadcaster, dst *Atomic, bo *backoff.Backoff) {
	for {
		ch, err := b.Subscribe(ctx)
		if err != nil {
			lggr.Errorw("Failed to subscribe to settings updates", "err", err)
		} else {
			lggr.Info("Subscribed to settings updates")
			// Reset only once a stream has delivered something: a stream that ends straight away must still
			// back off, or a failing host is resubscribed to in a tight loop.
			if apply(ctx, lggr, ch, dst) {
				bo.Reset()
			}
			if local, ok := b.(LocalBroadcaster); ok {
				local.Unsubscribe(ch)
			}
		}
		if !timeutil.Sleep(ctx.Done(), bo.Duration()) {
			return
		}
	}
}

// apply stores each update from ch in dst until ch closes or ctx is done, and reports whether any arrived.
func apply(ctx context.Context, lggr logger.Logger, ch <-chan Snapshot, dst *Atomic) (received bool) {
	for {
		select {
		case <-ctx.Done():
			return received
		case update, ok := <-ch:
			if !ok {
				return received
			}
			received = true
			if err := dst.Store(update); err != nil {
				lggr.Errorw("Failed to update settings", "hash", update.Hash, "err", err)
				continue
			}
			lggr.Infow("Updated settings", "hash", update.Hash)
		}
	}
}
