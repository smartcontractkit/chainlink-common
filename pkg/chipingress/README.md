# ChipIngress

ChipIngress is a gRPC client library for interacting with the ChipIngress service. It provides functionality for publishing CloudEvents to the ChipIngress service and supports various authentication methods.

## Features

- **CloudEvent Publishing**: Publish single events or batches of events
- **Batch Client with Partial Delivery**: Accumulate and flush events in batches, with per-event delivery results or all-or-nothing semantics
- **Authentication Support**: Multiple authentication methods including basic auth and token-based auth
- **Secure Communication**: Support for both TLS and insecure connections
- **Event Management**: Utilities for creating, converting, and managing CloudEvents
- **OpenTelemetry Integration**: Built-in OpenTelemetry support for distributed tracing and metrics

## Installation

```bash
go get github.com/smartcontractkit/chainlink-common/pkg/chipingress
```

## Usage

### Basic Client Creation

```go
import "github.com/smartcontractkit/chainlink-common/pkg/chipingress"

// Create client with default settings (insecure connection)
client, err := chipingress.NewClient("localhost:9090")
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Client with TLS

```go
client, err := chipingress.NewClient("example.com:9090", chipingress.WithTLS())
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Client with Authentication

```go
// Basic Auth
client, err := chipingress.NewClient("example.com:9090", 
    chipingress.WithTLS(),
    chipingress.WithBasicAuth("username", "password"))

// Token Auth
tokenProvider := &myTokenProvider{}
client, err := chipingress.NewClient("example.com:9090", 
    chipingress.WithTLS(),
    chipingress.WithTokenAuth(tokenProvider))
```

### Creating and Publishing Events

```go
// Create a new event
event, err := chipingress.NewEvent(
    "my-domain", 
    "my-entity", 
    []byte("event payload"),
    map[string]any{
        "subject": "test-subject",
        "time": time.Now(),
    })
if err != nil {
    log.Fatal(err)
}

// Convert to protobuf
eventPb, err := chipingress.EventToProto(event)
if err != nil {
    log.Fatal(err)
}

// Publish the event
response, err := client.Publish(context.Background(), eventPb)
if err != nil {
    log.Fatal(err)
}
```

### Batch Publishing & Partial Delivery

For higher throughput, the `batch` subpackage provides a client that accumulates
events and flushes them by batch size, byte budget, or time interval.

```go
import (
    "github.com/smartcontractkit/chainlink-common/pkg/chipingress"
    "github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
)

client, err := chipingress.NewClient("example.com:9090", chipingress.WithTLS())
if err != nil {
    log.Fatal(err)
}

bc, err := batch.NewBatchClient(client,
    batch.WithBatchSize(100),
    batch.WithBatchInterval(100*time.Millisecond),
    batch.WithMaxGRPCRequestSize(10*1024*1024),
)
if err != nil {
    log.Fatal(err)
}
bc.Start(context.Background())
defer bc.Stop() // flushes the pending batch and waits for callbacks

eventPb, _ := chipingress.EventToProto(event)
err = bc.QueueMessage(eventPb, func(err error) {
    if err != nil {
        // per-event delivery failure (see PublishError below)
    }
})
```

`QueueMessage` returns immediately and drops the message (returning an error) if
the internal buffer is full. The optional callback is invoked once the batch
containing the event has been sent.

#### Delivery mode

`PublishBatch` supports two modes, selected by `transaction_enabled`:

- **Partial delivery (default)** — valid events are produced and per-event errors
  are returned for the invalid ones, rather than failing the whole batch. The
  callback for a failed event receives a `*batch.PublishError`.
- **All-or-nothing** — enable with `batch.WithTransactionEnabled(true)`. Any
  per-event failure fails the entire batch and every callback receives the error.

The single-shot client exposes the same option via
`chipingress.EventsToBatchWithOpts(events, chipingress.WithTransactionEnabled(true))`.

#### Per-event error codes

A `*batch.PublishError` carries a `Code` and `Reason`:

| Code | Meaning |
| --- | --- |
| `ErrCodeValidationFailed` | CloudEvent structure is invalid (missing required fields, bad attribute). |
| `ErrCodeSchemaMissing` | No schema registered for the event's subject. Register it first. |
| `ErrCodeEncodeError` | Payload could not be encoded against its registered schema. |
| `ErrCodeDomainMisconfiguration` | Event source does not map to a known domain. |
| `ErrCodeResultsMismatch` | Client-side: server returned fewer results than events sent. |

Results are positional — `results[i]` corresponds to the i-th queued event. A
gRPC-level error (e.g. connection failure) fails every callback in the batch
regardless of delivery mode.

#### Sizing

`WithMaxGRPCRequestSize` bounds the serialized request size; oversized batches are
split into multiple requests, and a single event larger than the limit fails its
callback rather than being sent. Each queued event is also stamped with a
monotonic `seqnum` extension per `(source, type)` pair so downstream consumers can
detect gaps.

### OpenTelemetry Integration

The client automatically instruments gRPC calls with OpenTelemetry for distributed tracing and metrics. 

#### Using Global Providers

By default, the client uses the global OpenTelemetry providers:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Set up your OpenTelemetry tracer provider
tp := trace.NewTracerProvider(...)
otel.SetTracerProvider(tp)

// Create client - it will automatically use the global tracer
client, err := chipingress.NewClient("example.com:9090", chipingress.WithTLS())
```

#### Using Custom Providers

You can also pass custom MeterProvider and TracerProvider instances to the client:

```go
import (
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/sdk/metric"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
)

// Create custom providers
meterProvider := metric.NewMeterProvider(...)
tracerProvider := trace.NewTracerProvider(...)

// Create client with custom providers
client, err := chipingress.NewClient("example.com:9090",
    chipingress.WithTLS(),
    chipingress.WithMeterProvider(meterProvider),
    chipingress.WithTracerProvider(tracerProvider))
```

The client uses `otelgrpc.NewClientHandler()` to automatically create spans for all gRPC calls, including metrics for request duration, message sizes, and error rates.

## Dependencies

- `github.com/cloudevents/sdk-go/v2` - CloudEvents SDK
- `github.com/google/uuid` - UUID generation
- `google.golang.org/grpc` - gRPC communication
- `google.golang.org/protobuf` - Protocol buffer support
- `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` - OpenTelemetry gRPC instrumentation

## License

This project is licensed under the MIT License.
