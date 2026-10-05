# Retry policy

The ChipIngress client can retry failed unary RPCs at the gRPC level. **Retries are disabled by
default**; opt in per client with `WithRetryPolicy`.

## Enabling

```go
// Default: NewClient installs no retry config.
client, err := chipingress.NewClient(addr)

// Opt in with the recommended policy.
client, err = chipingress.NewClient(addr, chipingress.WithRetryPolicy(chipingress.DefaultRetryPolicy()))

// Or tune it. ParseStatusCodes maps codes.Code.String() names ("Unavailable") from string config.
retryable, err := chipingress.ParseStatusCodes([]string{"Unavailable", "ResourceExhausted"})
policy := chipingress.DefaultRetryPolicy()
policy.RetryableStatusCodes = retryable
client, err = chipingress.NewClient(addr, chipingress.WithRetryPolicy(policy))
```

## Why it is opt-in

The policy applies to the whole `ChipIngress` service, so it also replays `Publish`,
`PublishBatch` and `RegisterSchema`. An `UNAVAILABLE` can be returned after the server has already
done the work, so a retried batch can produce **duplicate events**. Before enabling retries for a
publisher, set the `idempotencykey` CloudEvent extension (`chipingress.IdempotencyKeyAttr`) so
downstream consumers can deduplicate.

## Default policy

`DefaultRetryPolicy()`:

| Field | Value |
|---|---|
| `MaxAttempts` | 3 (1 original attempt + 2 retries) |
| `InitialBackoff` | 100ms |
| `MaxBackoff` | 1s |
| `BackoffMultiplier` | 2 |
| `RetryableStatusCodes` | `Unavailable`, `ResourceExhausted` |

`MaxAttempts` counts the original attempt, so the default makes **at most 2 retries**.

### Timing

The delay before retry *n* is a random value between 0 and `min(InitialBackoff * BackoffMultiplier^(n-1), MaxBackoff)`
(gRPC applies jitter so clients do not retry in lockstep):

| Attempt | Runs | Wait before it |
|---|---|---|
| 1 | immediately | none |
| 2 (retry 1) | after attempt 1 fails | 0-100ms |
| 3 (retry 2) | after attempt 2 fails | 0-200ms |

Worst case adds 300ms of waiting plus the time of each attempt. `MaxBackoff` (1s) is only reached
if `MaxAttempts` is raised.

### What is retried

- Only calls that fail with a status in `RetryableStatusCodes`. Any other status ends the call
  immediately.
- Retries share the caller's context deadline. If the deadline expires during a backoff, the call
  fails with `DEADLINE_EXCEEDED` and no further attempts are made.

## Retry throttling

Whenever a retry policy is set, the client also installs gRPC retry throttling so retries cannot
pile load onto a failing server (see
[gRPC A6](https://github.com/grpc/proposal/blob/master/A6-client-retries.md#integration-with-service-config)):

- Bucket of 10 tokens, refilled by 0.1 per successful RPC.
- Each failed attempt removes 1 token.
- Retries are only attempted while the bucket is above half full (5 tokens). After about 5
  consecutive failures, retries stop until successes refill the bucket.

Throttling is not configurable through `WithRetryPolicy` and grpc-go exposes no metric or hook
for it, so suppressed retries are not directly observable.

## Implementation notes

- The service config is built from typed structs (`buildRetryServiceConfigJSON`), not a hand-written
  JSON literal. A previous hand-written literal was missing the `methodConfig[].retryPolicy`
  nesting, which gRPC's parser silently discards, so the client never retried. Durations must be
  protobuf-JSON seconds (`"0.1s"`), not Go durations (`"100ms"`).
- Status codes are emitted as numbers. `codes.Code.String()` output is not valid parser input;
  use `ParseStatusCodes` for name-based config.
- Tests in `retry_policy_test.go` run the generated config through gRPC's real service-config
  parser rather than asserting "no error", which is how the original bug evaded detection.
