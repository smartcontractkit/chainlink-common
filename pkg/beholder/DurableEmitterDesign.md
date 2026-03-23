# Durable Event Buffer for ChIP

## Problem Statement

Today there is no persistence in the ChIP pipeline. The `ChipIngressEmitter` calls `chipingress.Client.Publish()` synchronously over gRPC, and the `batch.Client` uses an in-memory channel buffer of 200 messages. If the node crashes, Chip is unreachable, or the buffer fills up, events — including billing records — are silently dropped.

**Drop points in the current architecture:**

```
DualSourceEmitter.Emit()
  ├── OTLP (sync) — errors returned to caller
  └── ChipIngressEmitter (async goroutine)
        │
        └── chipingress.Client.Publish()  ← fire-and-forget gRPC
              │
              ├── If Chip is down → error logged, event LOST
              ├── If node crashes mid-flight → event LOST
              └── batch.Client buffer full → "message buffer is full", event LOST
```

- `batch.Client` drops messages when its 200-message channel is full.
- `DualSourceEmitter` only logs chip-ingress failures — errors are swallowed.
- No event survives a node restart. Nothing is persisted to disk.

**Impact:** Billing records are silently dropped, leading to inconsistent revenue reconciliation. Any customer-facing observability that flows through ChIP is unreliable.

## Requirements

### Functional
- Events must not be lost on node restarts.
- Events must be delivered within a reasonable period of time (seconds under normal operation).
- System must support eventually-consistent billing.
- Node databases must not bloat unboundedly.

### Non-Functional
- Scale to 1k+ TPS per node.
- 4 nines of availability (99.99%).
- `Emit()` must not block workflow execution.

## Architecture

### High-Level Flow

```
Workflow Engine / Billing / Lifecycle Events
      │
      ▼
 DualSourceEmitter.Emit()
      │
      ├── OTLP MessageEmitter (sync, unchanged)
      │
      └── DurableEmitter.Emit()
            │
            ├─ 1. ExtractSourceAndType + build CloudEventPb
            ├─ 2. proto.Marshal → bytes
            ├─ 3. store.Insert(payload)  ← DURABLE GUARANTEE
            ├─ 4. return nil (caller unblocked)
            │
            └─ 5. goroutine: client.Publish(eventPb)
                   ├── Success → store.Delete(id)
                   └── Failure → no-op (retransmit loop handles it)

 ┌─────────────────────────────────────┐
 │  Background Retransmit Loop          │  every 5s (configurable)
 │                                      │
 │  store.ListPending(olderThan 10s)    │
 │  → client.PublishBatch(events)       │
 │  → store.Delete(ids) on success      │
 └─────────────────────────────────────┘

 ┌─────────────────────────────────────┐
 │  Background Expiry Loop              │  every 1min (configurable)
 │                                      │
 │  store.DeleteExpired(ttl=24h)        │
 │  → GC events that could never be     │
 │    delivered (bounds table growth)    │
 └─────────────────────────────────────┘
```

### Key Guarantee

`Emit()` returns `nil` once the Postgres INSERT succeeds. Even if the node crashes immediately after, the event survives in Postgres and will be retransmitted on restart. The gRPC publish is fully asynchronous — `Emit()` latency is dominated by one DB insert (~1ms at typical payloads).

### Design Decision: Standard `chipingress.Client`, Not `batch.Client`

Per Hagen's guidance, we use the standard `chipingress.Client` directly (supports both `Publish` and `PublishBatch`) since we are implementing our own queuing with persistence-backed guarantees. The `batch.Client`'s in-memory buffer is redundant when we have Postgres as the durable queue.

### Service Principal & ACK Guarantees

CRE nodes authenticate to ChIP using the node's **CSA Key** as the `servicePrincipal`. This is NOT the `oti-telemetry-shared` principal which uses a fire-and-forget publish path. With the CSA Key, the gateway waits for Kafka ACKs before returning a gRPC response — a successful response means the event was durably accepted.

## Components

### DurableEventStore Interface (`chainlink-common`)

```go
type DurableEvent struct {
    ID        int64
    Payload   []byte // serialized CloudEventPb proto
    CreatedAt time.Time
}

type DurableEventStore interface {
    Insert(ctx context.Context, payload []byte) (int64, error)
    Delete(ctx context.Context, id int64) error
    ListPending(ctx context.Context, createdBefore time.Time, limit int) ([]DurableEvent, error)
    DeleteExpired(ctx context.Context, ttl time.Duration) (int64, error)
}
```

Two implementations:
- **`MemDurableEventStore`** — in-memory map, for unit/integration tests. Lives in `chainlink-common`.
- **`PgDurableEventStore`** — Postgres-backed ORM using `sqlutil.DataSource`. Lives in `chainlink`.

### DurableEmitter (`chainlink-common`)

Implements `beholder.Emitter` (`Emit` + `Close`). Core logic:

```go
func (d *DurableEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
    // 1. Validate and extract source/type from attributes
    sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)

    // 2. Build CloudEvent and serialize to proto bytes
    event, _ := chipingress.NewEvent(sourceDomain, entityType, body, newAttributes(attrKVs...))
    eventPb, _ := chipingress.EventToProto(event)
    payload, _ := proto.Marshal(eventPb)

    // 3. Persist — this is the durable guarantee
    id, err := d.store.Insert(ctx, payload)
    if err != nil {
        return fmt.Errorf("failed to persist event: %w", err)
    }

    // 4. Async delivery attempt
    go d.publishAndDelete(id, eventPb)
    return nil
}
```

### Configuration

```go
type DurableEmitterConfig struct {
    RetransmitInterval  time.Duration // default 5s — retransmit loop tick rate
    RetransmitAfter     time.Duration // default 10s — min age before retry
    RetransmitBatchSize int           // default 100 — max events per batch
    ExpiryInterval      time.Duration // default 1min — expiry loop tick rate
    EventTTL            time.Duration // default 24h — max event age
    PublishTimeout      time.Duration // default 5s — per-RPC deadline
}
```

## Postgres Schema

**Migration `0295_chip_durable_events.sql`** in the existing `cre` schema:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS cre.chip_durable_events (
    id         BIGSERIAL   PRIMARY KEY,
    payload    BYTEA       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chip_durable_events_created_at
    ON cre.chip_durable_events (created_at ASC);

-- +goose Down
DROP INDEX IF EXISTS cre.idx_chip_durable_events_created_at;
DROP TABLE IF EXISTS cre.chip_durable_events;
```

The table lives in each node's existing Postgres database. Under normal operation it is **transient** — events are inserted and deleted within milliseconds. Under Chip outage, events accumulate until delivery resumes.

## Node Wiring

### Config Flag

```toml
[Telemetry]
DurableEmitterEnabled = true
```

Added to `config.Telemetry` interface, `toml.Telemetry` struct, and `telemetryConfig` implementation.

### Integration Point (`application.go`)

Wired in `NewApplication` after the DB is available but before CRE services start:

```go
func setupDurableEmitter(ctx context.Context, ds sqlutil.DataSource, lggr logger.SugaredLogger) error {
    client := beholder.GetClient()
    chipClient := client.Chip

    pgStore := beholdersvc.NewPgDurableEventStore(ds)
    durableEmitter, _ := beholder.NewDurableEmitter(pgStore, chipClient, beholder.DefaultDurableEmitterConfig(), lggr)

    // Preserve OTLP path alongside durable chip delivery
    messageLogger := client.MessageLoggerProvider.Logger("durable-emitter")
    otlpEmitter := beholder.NewMessageEmitter(messageLogger)
    dualEmitter, _ := beholder.NewDualSourceEmitter(durableEmitter, otlpEmitter)

    durableEmitter.Start(ctx)
    client.Emitter = dualEmitter
    return nil
}
```

This replaces the global beholder emitter, covering **all** emission paths:
- `events.emitProtoMessage()` — billing, workflow execution lifecycle
- `custmsg.Labeler.Emit()` — workflow user logs
- `BridgeStatusReporter` — bridge status events
- Any other `beholder.GetEmitter()` caller

### CRE Environment Auto-Enable

`system-tests/lib/cre/don/config/config.go` sets `DurableEmitterEnabled = true` for all Docker-based nodesets, so it activates automatically in local CRE environments.

## File Manifest

| Repo | File | Purpose |
|------|------|---------|
| chainlink-common | `pkg/beholder/durable_event_store.go` | `DurableEventStore` interface + `MemDurableEventStore` |
| chainlink-common | `pkg/beholder/durable_emitter.go` | `DurableEmitter` — Emit, retransmit loop, expiry loop |
| chainlink-common | `pkg/beholder/durable_emitter_test.go` | Unit tests (in-memory store) |
| chainlink-common | `pkg/beholder/durable_emitter_integration_test.go` | Integration tests (mock gRPC server) |
| chainlink | `core/services/beholder/durable_event_store_orm.go` | `PgDurableEventStore` (Postgres ORM) |
| chainlink | `core/services/beholder/durable_event_store_orm_test.go` | ORM tests + Postgres benchmarks + load tests |
| chainlink | `core/services/beholder/durable_emitter_load_test.go` | TPS ramp/sustained/payload tests (Postgres + mock or external Chip via `CHIP_INGRESS_TEST_ADDR`) |
| chainlink | `core/store/migrate/migrations/0295_chip_durable_events.sql` | Postgres migration |
| chainlink | `core/config/telemetry_config.go` | `DurableEmitterEnabled()` on Telemetry interface |
| chainlink | `core/config/toml/types.go` | TOML field + setFrom merge |
| chainlink | `core/services/chainlink/config_telemetry.go` | Config accessor |
| chainlink | `core/services/chainlink/application.go` | `setupDurableEmitter` wiring |
| chainlink | `system-tests/lib/cre/don/config/config.go` | Auto-enable in CRE Docker envs |
| chainlink | `system-tests/tests/smoke/cre/v2_durable_emitter_test.go` | CRE smoke tests + load test |
| chainlink | `system-tests/tests/smoke/cre/cre_suite_test.go` | Test entry points |

## Testing

### Unit Tests (`chainlink-common`, in-memory store)

| Test | What It Proves |
|------|---------------|
| `TestDurableEmitter_EmitPersistsAndPublishes` | Happy path: emit → publish → delete |
| `TestDurableEmitter_EmitReturnSuccessEvenWhenPublishFails` | Emit succeeds on DB insert even when gRPC fails |
| `TestDurableEmitter_RetransmitLoopDeliversFailedEvents` | Background loop retries failed events |
| `TestDurableEmitter_ExpiryLoopDeletesOldEvents` | TTL-based garbage collection |
| `TestDurableEmitter_EmitRejectsInvalidAttributes` | Validation before DB insert |
| `TestDurableEmitter_MultipleEvents` | 50 concurrent events all delivered |

### Integration Tests (`chainlink-common`, mock gRPC server)

Real gRPC server with controllable failure injection:

| Test | What It Proves |
|------|---------------|
| `TestIntegration_HappyPath` | Events delivered over real gRPC + proto round-trip |
| `TestIntegration_ServerUnavailable_RetransmitRecovers` | Server returns UNAVAILABLE → retransmit delivers via PublishBatch |
| `TestIntegration_ServerDown_EventsSurvive` | **Crash recovery**: server stopped → events persist → new emitter (same store) retransmits on "restart" |
| `TestIntegration_HighThroughput` | 500 events delivered concurrently |
| `TestIntegration_EventExpiry` | Undeliverable events expired after TTL |
| `TestIntegration_RetransmitUsesBatch` | Retransmit path uses PublishBatch, not individual Publish |
| `TestIntegration_GRPCConnection` | Source/type arrive correctly on server side |

### Postgres ORM Tests + Benchmarks (`chainlink`, real Postgres)

| Test / Benchmark | What It Measures |
|---|---|
| `TestPgDurableEventStore_*` | ORM correctness (insert, list, delete, expiry) |
| `Benchmark_Insert` | Raw INSERT throughput |
| `Benchmark_InsertDelete` | Insert+delete cycle (happy-path hot loop) |
| `Benchmark_InsertPayloadSizes` | INSERT at 64B, 256B, 1KB, 4KB |
| `Benchmark_ListPending` | Query performance at 100 and 1000 queue depth |
| `TestLoad_SustainedInsertDelete` | 2000 events, 10-way concurrent insert+delete, measures ops/sec |
| `TestLoad_BurstThenDrain` | 1000-event burst, then drain via ListPending+Delete batches |
| `TestLoad_ConcurrentInsertWithListPending` | 3s of concurrent inserts + ListPending (real contention) |

### Full-Stack Load Tests (`chainlink`, Postgres + mock gRPC)

| Test / Benchmark | What It Measures |
|---|---|
| `TestFullStack_SustainedThroughput` | 1000 events, 10 concurrent emitters, end-to-end rate |
| `TestFullStack_ChipOutage` | 3-phase: normal → Chip goes UNAVAILABLE → recovery. Measures accumulation and drain rate |
| `TestFullStack_SlowChip` | 50ms gRPC latency. Proves Emit() stays fast while server is slow |
| `Benchmark_FullStack_EmitThroughput` | Upper bound events/sec through full pipeline |
| `Benchmark_FullStack_EmitPayloadSizes` | Full emit at 64B, 256B, 1KB, 4KB |

### Durable emitter TPS load tests (`chainlink/core/services/beholder/durable_emitter_load_test.go`)

These tests exercise **Postgres + `DurableEmitter` + Chip Ingress** (in-process mock **or** a real gateway). They are heavier than the ORM benchmarks and require a **real Postgres** (not `txdb`).

#### Prerequisites

- **`CL_DATABASE_URL`** — must point at a Postgres instance where migration **`0295_chip_durable_events`** has been applied (`cre.chip_durable_events` exists). Same URL pattern as other chainlink DB tests.
- **Short tests skipped** — if your test runner uses `-short`, these tests are skipped (`SkipShortDB`); run **without** `-short`.

#### Mock Chip vs real Chip Ingress

| Mode | How | Notes |
|------|-----|--------|
| **Mock** (default) | Do **not** set `CHIP_INGRESS_TEST_ADDR` | In-process gRPC server; tests can count **Server recv** events and inject failures (outage, slow Chip). |
| **Real Chip** | Set `CHIP_INGRESS_TEST_ADDR=host:port` | Dials external Chip Ingress. Optional: `CHIP_INGRESS_TEST_TLS`, `CHIP_INGRESS_TEST_BASIC_AUTH_*`, `CHIP_INGRESS_TEST_SKIP_BASIC_AUTH`, `CHIP_INGRESS_TEST_SKIP_SCHEMA_REGISTRATION`. You need Kafka/Redpanda, topic **`chip-demo`**, and schema subject **`chip-demo-pb.DemoClientPayload`** (e.g. Atlas `make create-topic-and-schema` under `atlas/chip-ingress`). |

Tests that **inject** Chip failures or rely on **in-process** receive counts are **skipped** when `CHIP_INGRESS_TEST_ADDR` is set.

#### How to run

From the `chainlink` repo root (examples):

```bash
# All beholder tests including TPS (requires CL_DATABASE_URL)
export CL_DATABASE_URL='postgres://...'
go test -v -count=1 ./core/services/beholder/ -run 'TestTPS_|TestChipIngressExternalPing'

# Ramp-up only (100 → 500 → 1k → 2k TPS levels)
go test -v -count=1 ./core/services/beholder/ -run TestTPS_RampUp

# Sustained 1k TPS for 60s + drain check
go test -v -count=1 ./core/services/beholder/ -run TestTPS_Sustained1k

# Payload size scaling (fixed duration per size)
go test -v -count=1 ./core/services/beholder/ -run TestTPS_PayloadSizeScaling

# External Chip smoke (with addr set)
export CHIP_INGRESS_TEST_ADDR='localhost:50051'
go test -v -count=1 ./core/services/beholder/ -run TestChipIngressExternalPing
```

After a full package run, **`TestMain`** prints a **TPS LOAD TEST SUMMARY** block aggregating result blocks from **`TestTPS_RampUp`**, **`TestTPS_Sustained1k`**, **`TestTPS_1k_WithChipOutage`** (mock only; skipped with external Chip), and **`TestTPS_PayloadSizeScaling`**.

#### Reading the tables (column glossary)

| Column | Meaning |
|--------|---------|
| **Target TPS** | Requested emit rate (token-bucket style scheduling across workers). |
| **Achieved TPS** | `Total emits ÷ window duration` — realized successful `Emit()` throughput. |
| **Total emits** | Count of **`Emit()` calls that returned `nil`** in the measurement window (successful Postgres insert path). Does not count failures. |
| **Emit p50 / p99** | Latency of successful `Emit()` calls (dominated by DB insert). |
| **Failures** | `Emit()` calls that returned an error (e.g. DB failure). |
| **Server recv** | **Mock only:** number of events observed by the in-process gRPC server (`Publish` / `PublishBatch`). |
| **Queue depth** | Rows remaining in `cre.chip_durable_events` after the emit phase (+ short settle), i.e. backlog not yet deleted after successful publish. |

#### Why **Server recv** shows **N/A** with real Chip

The **Server recv** column is implemented by counting events on the **in-process mock** `ChipIngress` server. When you use **`CHIP_INGRESS_TEST_ADDR`**, there is no mock — the client talks to a **real** gateway — so the test **cannot** count server-side receives in-process. Use **Kafka / Chip / gateway metrics** (or consumer verification) to validate end-to-end delivery instead. **Total emits** and **Achieved TPS** still reflect client-side durable insert success; they are not replaced by N/A.

### CRE Smoke Tests (live Docker environment)

Tests connect to the node's Postgres and query `cre.chip_durable_events` directly, using `pg_stat_user_tables` for insert/delete statistics — the same pattern used by the EVM LogTrigger test for `trigger_pending_events`.

| Test | What It Does |
|------|-------------|
| `Test_CRE_V2_DurableEmitter` | Deploys a cron workflow (every 5s), waits for 30+ insert+delete cycles, verifies queue drains to near-empty |
| `Test_CRE_V2_DurableEmitter_Load` | Deploys 5 cron workflows (every 1s each), runs for 3 minutes. Logs insert/delete rates, max queue depth, and prints summary table |

**Running CRE smoke tests:**
```bash
# Basic correctness
go test -v -run Test_CRE_V2_DurableEmitter$ -timeout 10m

# Load test (5 workflows × 1s cron, 3min observation)
go test -v -run Test_CRE_V2_DurableEmitter_Load -timeout 10m
```

**Example load test output:**
```
╔════════════════════════════════════════════════╗
║        DURABLE EMITTER LOAD TEST RESULTS      ║
╠════════════════════════════════════════════════╣
║ Workflows deployed:   5                       ║
║ Observation period:   3m0s                    ║
║ Total inserts:        1842                    ║
║ Total deletes:        1840                    ║
║ Avg insert rate:      10.2  events/sec        ║
║ Avg delete rate:      10.2  events/sec        ║
║ Max queue depth:      12                      ║
║ Final pending:        2                       ║
╚════════════════════════════════════════════════╝
```

## Metrics to Instrument (Future)

| Metric | Description |
|--------|-------------|
| `durable_emitter.queue_depth` | Current row count in `chip_durable_events` |
| `durable_emitter.insert_rate` | Events persisted per second |
| `durable_emitter.publish_rate` | Events successfully delivered per second |
| `durable_emitter.retransmit_rate` | Events retransmitted via background loop |
| `durable_emitter.publish_latency` | Time from insert to confirmed delivery |
| `durable_emitter.oldest_pending` | Age of the longest-waiting event |
| `durable_emitter.expired_count` | Events expired (dropped after TTL) |
| `durable_emitter.error_rate` | Failed publish attempts per second |

## Open Questions & Future Work

### 1. Chip Gateway Idempotency
Does the gateway deduplicate re-sent events? If the retransmit loop re-sends an event that the immediate path already delivered (race window), the gateway should de-dup using the CloudEvent `id` (UUID). Needs server-side confirmation.

### 2. DB Load at Scale
At 1k TPS: ~1k inserts/sec + ~1k deletes/sec = ~2k write ops/sec on the node's Postgres. This produces dead tuples requiring autovacuum tuning. Potential optimizations:
- **Batch deletes** — delete by ID list instead of per-row.
- **Two-table approach** — queued + recently-sent to reduce churn on the hot table.
- **CDC streaming** — stream WAL changes directly, avoiding the insert/delete pattern entirely. Matthew Gardener and Clement can advise on CDC implementation.

### 3. Exponential Backoff
Current PoC uses a fixed retransmit interval. Production should implement per-event exponential backoff using `attempts` and `last_sent_at` columns (schema extension).

### 4. rmq / Redis Alternative
Patrick raised using [rmq](https://github.com/wellle/rmq) backed by our own DB instead of re-implementing a queue. Worth evaluating if the Postgres-backed approach shows scaling issues in load testing.

### 5. CDC Streaming
Could stream WAL changes directly rather than polling the table, avoiding the insert/delete churn entirely. This would also enable real-time analytics on event flow. Requires infrastructure coordination with the data analytics pipeline team.

### 6. DurableEmitter Lifecycle Management
Currently the `DurableEmitter` is started in `application.go` and its background loops are tied to the application context. For production, it should be registered as a proper `services.ServiceCtx` with Start/Close lifecycle management, health checks, and graceful shutdown (flush pending events before stopping).
