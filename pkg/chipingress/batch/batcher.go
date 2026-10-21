package batch

import (
	"context"
	"time"
)

// batchWithInterval reads from a channel and calls processFn with batches based on size or time interval.
// When the context is cancelled, it flushes the current batch and every item already queued on input,
// in batches of at most batchSize, then returns.
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
			timer.Stop()
			drain(input, batch, batchSize, processFn)
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

// drain processes batch plus everything already buffered on input, without waiting for new items.
// The batcher is input's only reader, so a non-zero len guarantees the receive won't block.
func drain[T any](input <-chan T, batch []T, batchSize int, processFn func([]T)) {
	for len(input) > 0 {
		batch = append(batch, <-input)
		if len(batch) >= batchSize {
			processFn(batch)
			batch = nil
		}
	}
	if len(batch) > 0 {
		processFn(batch)
	}
}
