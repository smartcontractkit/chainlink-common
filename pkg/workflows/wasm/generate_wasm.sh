#!/bin/sh
GOOS=wasip1 GOARCH=wasm go build -o ./test/cmd/testmodule.wasm ./test/cmd/main.go
