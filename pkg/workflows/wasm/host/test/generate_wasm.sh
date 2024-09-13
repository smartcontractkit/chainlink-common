#!/bin/sh
go install golang.org/dl/go1.22.7@latest
go1.22.7 download

GOOS=wasip1 GOARCH=wasm GO111MODULE='' CGO_ENABLED=0 GONOPROXY='' GONOSUMDB='' GOGCCFLAGS="-fPIC -fno-caret-diagnostics -Qunused-arguments -fmessage-length=0 -ffile-prefix-map=/path/to/source=/tmp/build -gno-record-gcc-switches"  go1.22.7 build -trimpath -ldflags=-buildid= -o ./test/cmd/testmodule.wasm ./test/cmd/main.go