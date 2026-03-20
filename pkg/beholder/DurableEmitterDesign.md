# Durable Event Buffer for ChIP

## Problem Statement

Today there is no persistence in the ChIP pipeline. The `ChipIngressEmitter` calls `chipingress.Client.Publish()` synchronously over gRPC, and the `batch.Client` uses an in-memory channel buffer of 200 messages. If the node crashes, Chip is unreachable, or the buffer fills up, events (including billing records) are silently dropped.

**Drop points:**
- `batch.Client` returns `"message buffer is full"` and the event is lost.
- `ChipIngressEmitter` propagates the error up, but `DualSourceEmitter` only logs chip-ingress failures.
- Any in-flight events are lost on node crash — nothing is persisted to disk.

## Functional Requirements

- Events must not be lost on node restarts.
- Events must be delivered within a reasonable period of time (seconds, not minutes).
- The system must support eventually-consistent billing.
- Node databases must not bloat unboundedly.

## Non-Functional Requirements

- Scale to 1k+ TPS.
- 4 nines of availability (99.99%).

## Architecture Overview

```
Workflow Engine
      │
      ▼
 DurableEmitter.Emit()
      │
      ├─ 1. Serialize event → proto bytes
      ├─ 2. INSERT into Postgres (durable guarantee)
      ├─ 3. Return nil (caller unblocked)
      │
      └─ 4. Async goroutine: chipingress.Client.Publish()
             │
             ├─ Success → DELETE from Postgres
             └─ Failure → no-op (retransmit loop will handle)

 ┌────────────────────────────┐
 │  Background Retransmit Loop │  (runs every RetransmitInterval)
 │                              │
 │  ListPending(olderThan)      │
 │  → PublishBatch()            │
 │  → DELETE on success         │
 └────────────────────────────┘

 ┌────────────────────────────┐
 │  Background Expiry Loop     │  (runs every ExpiryInterval)
 │                              │
 │  DeleteExpired(ttl)          │
 │  → GC old events            │
 └────────────────────────────┘
```

## Key Design Decision: Use Standard `chipingress.Client`, Not `batch.Client`

Per Hagen's guidance, we use the standard `chipingress.Client` directly (which supports both `Publish` and `PublishBatch`) since we are implementing our own queuing mechanism with persistence-backed guarantees. The `batch.Client`'s in-memory buffer would be redundant.

## Components

### DurableEventStore (interface — `chainlink-common`)

```go
type DurableEventStore interface {
    Insert(ctx context.Context, payload []byte) (int64, error)
    Delete(ctx context.Context, id int64) error
    ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error)
    DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
}
```

- `Insert` persists a serialized `CloudEventPb` and returns an auto-generated ID.
- `Delete` removes a single event after confirmed delivery.
- `ListPending` returns events older than a cutoff (gives the immediate-publish path time to succeed before retransmit picks it up).
- `DeleteExpired` garbage-collects events older than the TTL to bound table growth.

Two implementations:
- **`MemDurableEventStore`** — in-memory, for unit tests (lives in `chainlink-common`).
- **`PgDurableEventStore`** — Postgres-backed ORM (lives in `chainlink`).

### DurableEmitter (`chainlink-common`)

Implements `beholder.Emitter`:
```go
type Emitter interface {
    Emit(ctx context.Context, body []byte, attrKVs ...any) error
    io.Closer
}
```

**`Emit()` flow (synchronous path):**
1. `ExtractSourceAndType(attrKVs...)` — validate and extract source/type.
2. `chipingress.NewEvent(...)` + `EventToProto(...)` — build the CloudEvent proto.
3. `proto.Marshal(eventPb)` — serialize to bytes.
4. `store.Insert(ctx, payload)` — persist. **Emit returns nil here.**

**Async delivery path (goroutine):**
5. `client.Publish(ctx, eventPb)` — attempt immediate delivery.
6. On success: `store.Delete(id)`.
7. On failure: log at debug level; retransmit loop handles it.

**Key guarantee:** `Emit()` returns `nil` once the DB insert succeeds. Even if the node crashes immediately after, the event survives in Postgres and will be retransmitted on restart.

`Emit()` is non-blocking for workflow execution because the expensive operation (gRPC publish) happens asynchronously. The synchronous path is: attribute extraction → proto serialization → one DB insert. At 1k TPS, Postgres handles this trivially.

### Configuration

```go
type DurableEmitterConfig struct {
    RetransmitInterval  time.Duration // default 5s
    RetransmitAfter     time.Duration // default 10s  (min age before retry)
    RetransmitBatchSize int           // default 100
    ExpiryInterval      time.Duration // default 1min
    EventTTL            time.Duration // default 24h
    PublishTimeout      time.Duration // default 5s
}
```

## Postgres Table

```sql
CREATE TABLE IF NOT EXISTS cre.chip_durable_events (
    id         BIGSERIAL   PRIMARY KEY,
    payload    BYTEA       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chip_durable_events_created_at
    ON cre.chip_durable_events (created_at ASC);
```

The table lives in the existing `cre` schema. The `created_at` index supports efficient `ListPending` and `DeleteExpired` queries.

## Service Principal & ACK Guarantees

CRE nodes authenticate to ChIP using the node's CSA Key as the `servicePrincipal`. This is **not** the `oti-telemetry-shared` principal, which uses a fire-and-forget publish path. With the CSA Key principal, the gateway waits for Kafka ACKs before returning a response, so a successful gRPC response (`200`) means the event was durably accepted by the gateway.

## Files

| Repo | File | Purpose |
|------|------|---------|
| chainlink-common | `pkg/beholder/durable_event_store.go` | `DurableEventStore` interface + `MemDurableEventStore` |
| chainlink-common | `pkg/beholder/durable_emitter.go` | `DurableEmitter` struct, `Emit`, retransmit loop, expiry loop |
| chainlink-common | `pkg/beholder/durable_emitter_test.go` | Unit tests with in-memory store |
| chainlink | `core/services/beholder/durable_event_store_orm.go` | `PgDurableEventStore` implementing `DurableEventStore` |
| chainlink | `core/store/migrate/migrations/0294_chip_durable_events.sql` | Postgres migration |

## Wiring (TODO)

In `core/cmd/shell.go`, where the beholder client is created, replace or wrap the `ChipIngressEmitter` with `DurableEmitter`:

```go
// After creating chipIngressClient...
pgStore := beholder.NewPgDurableEventStore(ds)
durableEmitter, err := beholder.NewDurableEmitter(pgStore, chipIngressClient, cfg, log)
durableEmitter.Start(ctx)
// Use durableEmitter as the chip-ingress emitter in DualSourceEmitter
```

## Metrics to Instrument (Future)

| Metric | Description |
|--------|-------------|
| `durable_emitter.queue_depth` | Number of events currently in the store |
| `durable_emitter.insert_rate` | Events persisted per second |
| `durable_emitter.publish_rate` | Events successfully delivered per second |
| `durable_emitter.retransmit_rate` | Events retransmitted per second |
| `durable_emitter.publish_latency` | Time from insert to confirmed delivery |
| `durable_emitter.oldest_pending` | Age of the longest-waiting event |
| `durable_emitter.expired_count` | Events expired (dropped after TTL) |
| `durable_emitter.error_rate` | Failed publish attempts per second |

## Open Questions

1. **Chip gateway idempotency** — Does the gateway deduplicate re-sent events? If the retransmit loop re-sends an event that the immediate path already delivered (race window), will the gateway de-dup or create a duplicate billing record? The CloudEvent `id` field (UUID) could serve as a dedup key.

2. **DB load at scale** — At 1k TPS: ~1k inserts/sec + ~1k deletes/sec. The delete-heavy workload will produce dead tuples requiring autovacuum tuning. Potential optimizations:
   - Batch deletes (delete by ID list instead of per-row).
   - Two-table approach (queued + recently-sent) to reduce churn.
   - CDC streaming as an alternative to the insert/delete pattern entirely.

3. **Exponential backoff** — Current PoC uses a fixed retransmit interval. Production should implement per-event exponential backoff using `attempts` and `last_sent_at` columns (schema extension).

4. **rmq / Redis alternative** — Patrick raised using [rmq](https://github.com/wellle/rmq) backed by our own DB instead of re-implementing a queue. Worth evaluating if the Postgres-backed approach has scaling issues.

5. **CDC streaming** — Could stream WAL changes directly rather than polling the table, avoiding the insert/delete churn entirely. Matthew Gardener and Clement can advise on CDC implementation within the existing data analytics pipeline.
