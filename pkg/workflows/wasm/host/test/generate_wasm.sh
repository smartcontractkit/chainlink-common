#!/bin/sh

# Install the Go version if necessary
go install golang.org/dl/go1.22.7@latest
go1.22.7 download


# Note that these build flags should make the build reproducible,
# running
# go:generate .
# from the host directory and
# go generate ./pkg/workflows/wasm/host/
# produce the same binary but running
# go generate ./...
# from the root directory will produce a different binary
GOOS=wasip1 GOARCH=wasm \
go1.22.7 build -trimpath -ldflags="-X 'main.buildTime=00000000' -X 'main.version=1.0.0' -w -s -buildid=" -tags "wasip1" -gcflags=all= -mod=readonly -a -p=1\
    -o testmodule.wasm github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd
