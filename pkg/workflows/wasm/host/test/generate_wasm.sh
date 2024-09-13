#!/bin/sh
go install golang.org/dl/go1.22.7@latest
go1.22.7 download
GOOS=wasip1 GOARCH=wasm go1.22.7 build -trimpath -ldflags="-w -s" -tags "" -o ./test/cmd/testmodule.wasm ./test/cmd/main.go