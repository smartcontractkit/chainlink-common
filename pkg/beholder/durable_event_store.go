package beholder

import (
	"context"
	"time"
)

// Backward-compatibility types for downstream consumers (e.g. chainlink node).
// These were moved to pkg/durableemitter in PR #2081. They are re-declared here
// (not aliases) to avoid an import cycle through test utilities.
//
// TODO(SHARED-2644): Remove this file once chainlink node migrates imports to
// pkg/durableemitter (tracked in chainlink PR #22562). These duplicate declarations
// have no compile-time sync check against the canonical types.

// DurableEvent represents a persisted event awaiting delivery to Chip.
type DurableEvent struct {
	ID        int64
	Payload   []byte
	CreatedAt time.Time
}

// DurableQueueStats is a point-in-time snapshot of the pending queue for metrics.
type DurableQueueStats struct {
	Depth            int64
	PayloadBytes     int64
	OldestPendingAge time.Duration
	NearTTLCount     int64
}

// DurableQueueObserver is optionally implemented by DurableEventStore implementations
// so DurableEmitter can export queue depth and age gauges when metrics are enabled.
type DurableQueueObserver interface {
	ObserveDurableQueue(ctx context.Context, eventTTL, nearExpiryLead time.Duration) (DurableQueueStats, error)
}

// BatchInserter is optionally implemented by DurableEventStore implementations
// to support multi-row inserts for higher throughput.
type BatchInserter interface {
	InsertBatch(ctx context.Context, payloads [][]byte) ([]int64, error)
}

// DurableEventStore abstracts the persistence layer for durable chip events.
type DurableEventStore interface {
	Insert(ctx context.Context, payload []byte) (int64, error)
	Delete(ctx context.Context, id int64) error
	MarkDelivered(ctx context.Context, id int64) error
	MarkDeliveredBatch(ctx context.Context, ids []int64) (int64, error)
	PurgeDelivered(ctx context.Context, batchLimit int) (deleted int64, err error)
	ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error)
	DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
}
