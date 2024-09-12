#!/bin/sh
go install golang.org/dl/go1.22.7@latest
go1.22.7 download

# leave the -ffile-prefix-map= field for $GOGCCFLAGS so that symbol tables are changed to /tmp/go-build by default
# otherwise the table will vary based on file location.
# Disable all other flags to make a consistent build.
FFILE_PREFIX_MAP=$(echo $GOGCCFLAGS | grep -oP '-ffile-prefix-map=[^ ]+')
GOOS=wasip1 GOARCH=wasm GO111MODULE='' CGO_ENABLED=0 GONOPROXY='' GONOSUMDB='' GOGCCFLAGS="$FFILE_PREFIX_MAP" go1.22.7 env
GOOS=wasip1 GOARCH=wasm GO111MODULE='' CGO_ENABLED=0 GONOPROXY='' GONOSUMDB='' GOGCCFLAGS="$FFILE_PREFIX_MAP" go1.22.7 build -o ./test/cmd/testmodule.wasm ./test/cmd/main.go