# ChipIngress

ChipIngress is a gRPC client library for interacting with the ChipIngress service. It provides functionality for publishing CloudEvents to the ChipIngress service and supports various authentication methods.

## Features

- **CloudEvent Publishing**: Publish single events or batches of events
- **Authentication Support**: Multiple authentication methods including basic auth and token-based auth
- **Secure Communication**: Support for both TLS and insecure connections
- **Event Management**: Utilities for creating, converting, and managing CloudEvents

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

## Dependencies

- `github.com/cloudevents/sdk-go/v2` - CloudEvents SDK
- `github.com/google/uuid` - UUID generation
- `google.golang.org/grpc` - gRPC communication
- `google.golang.org/protobuf` - Protocol buffer support

## License

This project is licensed under the MIT License.
