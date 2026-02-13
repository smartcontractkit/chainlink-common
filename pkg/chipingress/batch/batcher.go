package batch

import (
	"context"
	"time"
)

// batchWithInterval reads from a channel and calls processFn with batches based on size or time interval.
// When the context is cancelled, it automatically flushes any remaining items in the batch.
func batchWithInterval[T any](
	ctx context.Context,
	input <-chan T,
	batchSize int,
	interval time.Duration,
	processFn func([]T),
) {
	var batch []T
	timer := time.NewTimer(interval)
	timer.Stop()

	flush := func() {
		if len(batch) > 0 {
			processFn(batch)
			batch = nil
			timer.Stop()
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case msg, ok := <-input:
			if !ok {
				// Channel closed
				flush()
				return
			}

			// Start timer on first message in batch
			if len(batch) == 0 {
				timer.Reset(interval)
			}

			batch = append(batch, msg)

			// Flush when batch is full
			if len(batch) >= batchSize {
				processFn(batch)
				batch = nil
				timer.Stop()
			}
		case <-timer.C:
			flush()
		}
	}
}
