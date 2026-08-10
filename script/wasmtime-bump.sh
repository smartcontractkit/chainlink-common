#!/bin/bash

set -e

export WASMTIME_VERSION=$(git ls-remote --tags --refs https://github.com/bytecodealliance/wasmtime-go | awk -F/ '{print $3}' | grep -E '^v?[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1);
echo "Latest wasmtime-go version:"$WASMTIME_VERSION;
export WASMTIME_MAJOR_VERSION=$(echo $WASMTIME_VERSION | sed 's/^v//' | cut -d. -f1);
echo "Latest major version:"$WASMTIME_MAJOR_VERSION;
export WASMTIME_MAJOR_VERSION_CURRENT=$(go list -m -json github.com/bytecodealliance/wasmtime-go/... | jq -r .Version | sed 's/^v//' | cut -d. -f1);
echo "Current major version:"$WASMTIME_MAJOR_VERSION_CURRENT;
if [ "$WASMTIME_MAJOR_VERSION" != "$WASMTIME_MAJOR_VERSION_CURRENT" ]; then
  echo "Bumping major version...";
  gofmt -w -r "\"github.com/bytecodealliance/wasmtime-go/v$WASMTIME_MAJOR_VERSION_CURRENT\" -> \"github.com/bytecodealliance/wasmtime-go/v$WASMTIME_MAJOR_VERSION\"" .
fi
echo "Getting latest version...";
go get -v github.com/bytecodealliance/wasmtime-go/v$WASMTIME_MAJOR_VERSION@$WASMTIME_VERSION;
git add **/*.go **/go.*
