#!/bin/bash

for dir in */ ; do
  if [ -d "$dir" ]; then
    echo "Building in $dir"
    (cd "$dir/cmd" && GOOS=wasip1 GOARCH=wasm go build -o testmodule.wasm .)
  fi
done

