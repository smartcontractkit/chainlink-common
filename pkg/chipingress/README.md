# ChipIngress

ChipIngress is a gRPC client library for interacting with the ChipIngress service. It provides functionality for publishing CloudEvents to the ChipIngress service and supports various authentication methods.

## Features

- **CloudEvent Publishing**: Publish single events or batches of events
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
