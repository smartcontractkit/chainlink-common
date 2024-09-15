#!/bin/sh

# Install the Go version if necessary
go install golang.org/dl/go1.22.7@latest
go1.22.7 download

# Dependencies must be generated before building the wasm, or the build won't be consistent
# Even though few or none of these files are used by the wasm build, the values are considered for file sums
GOOS=wasip1 GOARCH=wasm go list -f '{{ join .Deps "\n" }}' github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd | grep '^github.com/smartcontractkit/chainlink-common' | xargs -I % go1.22.7 generate %

GOOS=wasip1 GOARCH=wasm \
go1.22.7 build -trimpath -ldflags="-X 'main.buildTime=00000000' -X 'main.version=1.0.0' -w -s -buildid=" -tags "wasip1" -gcflags=all= -mod=readonly -a -p=1\
    -o testmodule.wasm github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/test/cmd