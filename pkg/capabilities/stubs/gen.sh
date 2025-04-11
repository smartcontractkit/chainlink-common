#!/bin/bash

(cd ../protoc && go build -o protoc-gen-cre .)

# Go two levels deep
for dir in */ ; do
  if [ -d "$dir" ]; then
    for subdir in "$dir"*/ ; do
      if [ -d "$subdir" ]; then
        echo "Generating in $subdir"
        (cd "$subdir" && go generate .)
      fi
    done
  fi
done
