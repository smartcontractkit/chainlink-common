# Tiny code generator that turns your domain model into:

 - .proto service + messages
 - single Go file (rpc.go) with:
 - typed gRPC Client
 - typed gRPC Server (thin shim over your implementation)
 - pb ↔ domain converters for every user message
 - oneof (interface) adapters
 - safe handling of bytes, repeated fields, and fixed-size arrays
 - an optional rpc_test.go with a single bufconn server and subtests (happy-path roundtrip)
Usage example:
Suppose your interface is in pkg/path and its called MyInterface:
1. Generate proto files + go wrappers
//go:generate bash -c "set -euo pipefail; mkdir -p ./gen/pb ./gen/wrap && go run ./genwiring --pkg pkg/path --interface MyInterface --config config.yaml --service MyService --proto-pkg loop.test --proto-go-package my/proto/package --proto-out ./path/to/my.proto --go-out ./path/to/my/go/wrappers --go-pkg my/go/pkg"
2. Use protoc to generate go proto types
//go:generate bash -c "set -euo pipefail; protoc -I ./gen/pb --go_out=paths=source_relative:./gen/pb --go-grpc_out=paths=source_relative:./gen/pb ./gen/pb/service.proto"
The output will be:
 - A strongly typed Client that does domain <-> pb conversions rpc.go, rpc_test.go 
 - A server wrapper that converts pb and calls your impl.
A full usage example with config in pkg/loop/internal/generator/testdata
