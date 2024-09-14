#!/bin/sh

# Install the Go version if necessary
go install golang.org/dl/go1.22.7@latest
go1.22.7 download

GOOS=wasip1 GOARCH=wasm go env

GOOS=wasip1 GOARCH=wasm \
go1.22.7 build -trimpath -ldflags="-w -s -buildid=" -tags "wasip1" \
    -o testmodule.wasm github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd
