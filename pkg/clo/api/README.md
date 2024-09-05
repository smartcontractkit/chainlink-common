# Orchestrator API Client

The `api` package encapsulates the API client implementation for interacting with the Chainlink Orchestrator via its GraphQL API.

## Features
- Standalone library intended to be imported by other packages or applications
- Automated type-safe code generation and validation
- Reference CLI implementation available at `pkg/cli/client`

## Directory Structure
```
api
 ├── client.go
 ├── client_test.go
 ├── internal
 │   ├── auth
 │   │   ├── token.go
 │   │   ├── token_test.go
 │   │   ├── transport.go
 │   │   └── transport_test.go
 │   ├── graph
 │   │   ├── aggregator.graphql
 │   │   ├── app.graphql
 │   │   ├── authentication.graphql
 │   │   ├── build_info.graphql
 │   │   ├── category.graphql
 │   │   ├── ccip.graphql
 │   │   ├── contract.graphql
 │   │   ├── feed.graphql
 │   │   ├── job.graphql
 │   │   ├── network.graphql
 │   │   ├── node.graphql
 │   │   ├── nodeOperator.graphql
 │   │   ├── profile.graphql
 │   │   └── user.graphql
 │   └── tools
 │       ├── generate.go
 │       ├── genqlient.yml
 │       └── gqlgen.yml
 ├── models
 │   ├── models.go
 │   └── models_gen.go
 └── operations
     ├── operations.go
     └── operations_gen.go
```

`client.go` exposes the API client and configuration

`models` provides generated structs and interfaces that reflect the API's GraphQL types

`operations` provides generated type-safe go wrappers over the GraphQL API calls supplied to `graph`

`auth` contains the API session authentication logic

`graph` contains the raw GraphQL queries and mutations that serve as inputs to the code generation pipeline

`tools` contains the code generation script and associated configuration files


## Usage

Import the package(s) into another package or into another application
```go
import(
    "github.com/smartcontractkit/feeds-manager/api"
    "github.com/smartcontractkit/feeds-manager/api/models"
    "github.com/smartcontractkit/feeds-manager/api/operations"
)
```

Instantiate the API client, providing a context and a configuration
```go
client, err := api.NewClient(ctx, config)
```

Authenticate your session
```go
resp, err := client.Login()
```

Execute a query
```go
resp, err := operations.GetFeed(client.Ctx(), client.Gql(), feedId)
```

Execute a mutation
```go
input := models.ProposeJobInput{
	ID: jobId,
}

resp, err := operations.ProposeJob(client.Ctx(), client.Gql(), input)
```

Terminate the session
```go
err := client.logout()
```

*Usage example available in `pkg/cli/client`*

## Development
To add new operations to the API client:
1. Add the desired GraphQL query or mutation to `internal/graph`
2. `go run api/internal/tools/generate.go` (or just `make client` from the repo root)

The client code generation pipeline will parse the GraphQL schemas in `pkg/graph/schemas` and then generate the `models` and `operations`.
