#!/bin/bash

for dir in */ ; do
  if [ -d "$dir" ]; then
    echo "Genereating in $dir"
    (cd "$dir" && go generate .)
  fi
done

