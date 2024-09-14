#!/bin/sh

# Find the root of the Go module
MODULE_ROOT=$(go list -m -f "{{.Dir}}" github.com/smartcontractkit/chainlink-common)

# Install the Go version if necessary
go install golang.org/dl/go1.22.7@latest
go1.22.7 download

# Use the module root for all paths to ensure consistency
GOOS=wasip1 GOARCH=wasm \
go1.22.7 install -trimpath -ldflags="-w -s -buildid= -tags test" \
    github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd

# Move the output binary to the desired location, using the module root for consistent paths
mv "$(go env GOPATH)/bin/wasip1_wasm/cmd" "$MODULE_ROOT/pkg/workflows/wasm/host/test/cmd/testmodule.wasm"
