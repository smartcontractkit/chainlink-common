#!/bin/sh
go install golang.org/dl/go1.22.7@latest
go1.22.7 download
# GO111MODULE must be the same to keep the build reproducible
# We shouldn't need CGO, so disabling it assures we don't need have consistent GOGCCFLAGS
GOOS=wasip1 GOARCH=wasm GO111MODULE='' CGO_ENABLED=0 go1.22.7 env
GOOS=wasip1 GOARCH=wasm GO111MODULE='' CGO_ENABLED=0 go1.22.7 build -o ./test/cmd/testmodule.wasm ./test/cmd/main.go