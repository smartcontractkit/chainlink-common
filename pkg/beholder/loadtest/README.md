# Beholder node-side load test (INFOPLAT-19737)

An in-process, repeatable load harness for the node-side path
`Client.ForDomain(d).EmitEvent` -> per-domain batch emitter -> `chipingress/batch` client ->
gRPC `PublishBatchNoAck` -> fake chip-ingress. It answers: at what rate, RTT and outage length does
the node drop telemetry, how much memory does it hold, and what does a restart lose. Scenarios follow
sections 2, 3 and 5 of the INFOPLAT-19737 load-test plan.

Every file has `//go:build loadtest`, so normal builds, `go test ./...` and CI ignore the package.

## Run

All commands from the repo root. Add `-v` to see the per-scenario log lines.

```sh
# whole suite, production batching defaults (~2 minutes)
go test -tags loadtest -run TestScenarios -count=1 -v ./pkg/beholder/loadtest/

# one or a few scenarios (regex), results to JSON + markdown
LOADTEST_SCENARIO='^(steady_peak|headroom_)' LOADTEST_OUT=/tmp/lt/results.json \
  go test -tags loadtest -run TestScenarios -count=1 -v ./pkg/beholder/loadtest/

# enforce the plan section 5 pass bar
LOADTEST_STRICT=1 go test -tags loadtest -run TestScenarios -count=1 ./pkg/beholder/loadtest/

# longer soak at 1.2x
LOADTEST_DURATION=60s LOADTEST_RATE_SCALE=1.2 LOADTEST_SCENARIO='^steady_peak$' \
  go test -tags loadtest -run TestScenarios -count=1 -v ./pkg/beholder/loadtest/

# data race check on a short scenario
LOADTEST_SCENARIO='^outage_1s$' go test -race -tags loadtest -run TestScenarios -count=1 ./pkg/beholder/loadtest/
```

The test binds `127.0.0.1:0`; in a sandbox that blocks local port binding it fails with
`operation not permitted` and must be run outside the sandbox.

## Testing local changes to `pkg/chipingress/batch`

`pkg/chipingress` is a separate Go module pinned in the root `go.mod`, so by default the harness
compiles the module-cache copy of `chipingress/batch`, not the worktree's. `TestScenarios` logs the
resolved directory first (`chipingress/batch under test: ...`); if it is under `~/go/pkg/mod`, edits
in the worktree are not being tested. To test them, use a throwaway workspace, never committed:

```sh
cd <repo root>
go work init . ./pkg/chipingress          # writes go.work; delete it afterwards (or gitignore locally)
GOWORK=$PWD/go.work go test -tags loadtest -run TestScenarios -count=1 -v ./pkg/beholder/loadtest/
rm go.work go.work.sum
```

Alternatively `GOWORK=/tmp/lt.work` created with `go work init` pointing at both directories, or a
temporary `replace` in `go.mod` that you revert. Confirm the log line shows the worktree path.

Unset `GOFLAGS` (or at least drop `-mod=mod`) when using a workspace: in workspace mode `go test`
fails with "-mod may only be set to readonly or vendor when in workspace mode".

## Environment variables

| Variable | Meaning |
|---|---|
| `LOADTEST_SCENARIO` | Regex on scenario names; default all. |
| `LOADTEST_DURATION` | Go duration replacing each scenario's load phase (e.g. `30s`). Outage windows are absolute (2s..3s, 2s..7s), so keep it longer than they end. |
| `LOADTEST_RATE_SCALE` | Multiplies every domain's rate (default 1). |
| `LOADTEST_OUT` | Path of a JSON results file; a markdown table is written beside it (same name, `.md`). |
| `LOADTEST_EXTENDED` | `1` also runs the extended scenarios (they are also run when `LOADTEST_SCENARIO` matches them). |
| `LOADTEST_STRICT` | `1` fails the test on pass-bar violations (below). Default: report only. |
| `LOADTEST_MEM_LIMIT_MIB` | Optional RSS-proxy limit for `memory_all_domains` (enforced under `LOADTEST_STRICT=1`); no limit is agreed yet, so unset by default. |

## Scenarios

Durations are the load phase; the client is then closed (drained) and late callbacks settle.
"RTT" is the fake server's response delay. Defaults are the production batching defaults
(buffer 10000, batch 1000, interval 500ms, send timeout 10s, drain timeout 30s, 10 concurrent sends, 10 MiB).
Per-domain steady-state ceiling from the plan: `MaxConcurrentSends x MaxBatchSize / RTT`.

| Name | Load | Fake chip-ingress | Duration | Strict pass bar |
|---|---|---|---|---|
| `steady_peak` | data-streams 16k/s, data-feeds 40/s | RTT 50-100ms (50ms + 0-50ms jitter) | 8s | 0 dropped, 0 pending |
| `headroom_1_5x` | 1.5x peak (24k/s, feeds 60/s) | RTT 100ms | 6s | 0 dropped |
| `headroom_1_5x_rtt250` | 1.5x peak (24k/s, feeds 60/s) | RTT 250ms: the plan section 5 pass-bar condition | 6s | 0 dropped |
| `headroom_2x` | 2x peak (32k/s) | RTT 100ms | 6s | none (drop rate recorded) |
| `slow_rtt_250ms` | peak | RTT 250ms | 6s | 0 dropped (plan: zero drops at RTT <= 250ms) |
| `slow_rtt_500ms`, `slow_rtt_1s`, `slow_rtt_5s` | peak | RTT 500ms / 1s / 5s | 6s | none (drop onset recorded) |
| `outage_1s`, `outage_5s` | peak | `Unavailable` for [2s,3s) / [2s,7s), then normal | 8s / 12s | none (drops and recovery recorded) |
| `outage_30s` (extended) | peak | `Unavailable` for [2s,32s) | 40s | none |
| `slow_rtt_9s` (extended) | peak | RTT 9s (SendTimeout is 10s) | 6s + long drain | none |
| `peak_split_domains` (extended) | 16k/s split evenly over data-streams, data-feeds, ccip | RTT 50-100ms | 8s | 0 dropped |
| `restart_full_buffer` | peak, buffer full | RTT 1s; Close with DrainTimeout 5s (shortened from 30s) | 6s | none |
| `restart_full_buffer_short_drain` | peak, buffer full | RTT 2s; Close with DrainTimeout 1s | 6s | none |
| `memory_all_domains` | all 3 domains at 16k/s, payloads 0.5-32 KiB (mean ~2.7 KiB) | RTT 1s so every buffer stays full | 8s | RSS proxy <= `LOADTEST_MEM_LIMIT_MIB` if set |
| `oversized_messages` | data-streams 2k/s, every 500th event 11 MiB; data-feeds 40/s | RTT 50ms | 5s | always enforced (see below) |

Extended scenarios stay out of the default run to keep it near 2 minutes:
`LOADTEST_EXTENDED=1` or `LOADTEST_SCENARIO='^outage_30s$'`.

Plan section 5 bars. `headroom_1_5x_rtt250` is the plan's actual condition (zero drops at 1.5x with
RTT <= 250ms) and is in the strict set; on the reference laptop it currently FAILS (about 3% buffer-full
drops at 24k/s although the formula allows 40k/s: see per-batch client CPU below). The drop count
varies between runs (about 8k to 15k of ~144k), so treat it as approximate; add a tolerance before
using it as anything other than a zero-drop gate. The bar "no drops in a
5 s outage" is known NOT MET and is not enforced under strict: this path has no retry, so every event
emitted during an outage is dropped regardless of buffer size. `outage_5s` always prints a
`plan bar ... NOT MET: N dropped` note.

`oversized_messages` is a correctness check, so it is enforced without `LOADTEST_STRICT`: oversized
events must all fail locally with the max-request-size error, none may reach the server, and
every drop must be an oversized event.

## Invariants (always asserted)

Per domain, from independent counters:

* `emitted == sync errors + accepted + callback errors + pending`. Sync errors are `EmitEvent`
  returning an error (callback not invoked); accepted is `callback(nil)`; callback errors are keyed by
  `batch.ErrorCodeFor` (a buffer-full drop is a callback error, `message buffer is full`); pending is
  queued with no callback yet after Close and a 10s settle. Pending > 0 at the end is a failure
  (a callback was lost).
* Every callback fires exactly once, and never after `EmitEvent` returned an error.
* Server accepted >= client accepted, and server accepted <= events queued. No-ack means a nil
  callback is a success response, which the fake only sends after counting the batch, so the server
  has seen at least every accepted event. It can have accepted more: a response lost to a client send
  timeout or a force-close cancel is a client error for events the server did accept. It never
  implies persistence: real chip-ingress answers after handing events to its Kafka producer with acks=0.
* No event reaches the server on another domain's route, and no batch uses `PublishBatch`
  (domain emitters must use `PublishBatchNoAck`).

## Reading the results

`-v` prints per scenario: emitted, accepted, dropped (by code), pending, server accepted, achieved rate,
p50/p99 of emit -> `callback(nil)` latency, peak heap, goroutines, CPU cores; then Close and
per-domain lines. JSON (`LOADTEST_OUT`) also has the 250ms timeline, server batch-size histogram and config.

* `dropped` is sync + callback errors. `message buffer is full` is a local drop (buffer overflow);
  `Unavailable`/`Internal` are server errors; `DeadlineExceeded` is a send timeout; `Canceled` is a
  drain cut off by DrainTimeout; `client_error` is a local failure such as an oversized request.
* `achieved` vs `target` shows whether the load generator kept up; a note flags a shortfall below 95%
  (then the scenario is not valid for a pass bar).
* `predicted max/s` is the plan formula for the server RTT. Compare with `accepted/duration`; the first
  error time marks drop onset. The formula ignores the client's own per-batch CPU time (see limits).
* `first error s` / `last error s` are seconds from load start. For outages, `recovery: first accept
  after outage` is the time from the end of the outage to the first accepted event.
* `pending at stop`, `delivered during Close`, `failed during Close`: what was queued or in flight when
  the load stopped and what happened to it. In the restart scenarios this is the DrainTimeout result.
* `peak pending events` (queued plus in flight, sampled every 250ms) times mean payload is a direct
  measure of buffered payload bytes. `peak heap` and `rss~` (MemStats `Sys - HeapReleased`) are
  process-wide.

## Known limits

* The chainlink node metrics `telemetry_messages_dropped_total`, `telemetry_messages_send_errors_total`
  and the batch client send latency are not read. The harness counts callback outcomes by
  `batch.ErrorCodeFor`, which is what feeds them (buffer-full and send errors by `error_code`).
* Latency percentiles come from a log-scale histogram and are bucket upper edges, up to ~19% high.
  Memory is a `runtime.MemStats` proxy for the whole process, including the fake server and the
  generator; it is not the node's RSS.

* In-process, loopback network: no real TLS, auth, kernel or LB behaviour; RTT is an injected delay,
  not network latency. The fake server answers instantly apart from that delay and never applies
  backpressure.
* Generator, beholder client and fake server share one process and one machine, so CPU and heap
  include the server's request decoding and the generator's event building (an upper bound for the
  client alone). Results depend on the machine; compare runs on the same one. Rates are driven by a
  1ms pacing loop and are reported as achieved.
* Fake server: no schema validation, no Kafka. Outages return `Unavailable` immediately (fast
  failure), not a hang; a hang would run into SendTimeout instead.
* Payload sizes are synthetic until real per-domain distributions are collected. Each event gets a
  freshly allocated payload so buffered memory is not shared.
* Not covered: chainlink's `domain_manager` (it is not in this repo) and how it wraps or retries the
  emitters; the default (non-domain) emitter; the OTLP metric/trace exporters (they point at the fake
  and are swallowed, the log exporter is a noop).
* Wall-clock timing, not `testing/synctest`: scenarios take real seconds and timings vary a little
  between runs.
* One run per scenario, short durations: good for drop onset and gross memory, not for tail latency
  or leaks.
