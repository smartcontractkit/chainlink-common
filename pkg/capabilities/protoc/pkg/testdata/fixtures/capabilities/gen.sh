#!/bin/bash

(cd ../../../../ && go build -o protoc-gen-cre .)

for dir in */ ; do
  if [ -d "$dir" ]; then
    echo "Genereating in $dir"
    (cd "$dir" && go generate .)
  fi
done

