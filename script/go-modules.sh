#!/usr/bin/env bash
# Go modules in this repository, relative to its root, the root module as ".".
#
# A module under a directory named examples is an example of its nearest enclosing module that is
# not one. It models a downstream consumer, local replace directives and all, so it is built,
# tidied and generated as part of that module rather than checked on its own.
#
#   go-modules.sh                  every module that is not an example
#   go-modules.sh examples MODULE  MODULE's examples
#   go-modules.sh parent DIR       the module DIR belongs to, an example belonging to its parent
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

all() {
  git ls-files --cached --others --exclude-standard '*go.mod' | xargs -n1 dirname | sort -u
}

is_example() {
  [[ "$1" =~ (^|/)examples(/|$) ]]
}

parent() {
  local dir=$1
  while [[ "$dir" != . ]] && { is_example "$dir" || [[ ! -f "$dir/go.mod" ]]; }; do
    dir=$(dirname "$dir")
  done
  echo "$dir"
}

case "${1:-}" in
  "")
    all | while read -r module; do
      if ! is_example "$module"; then echo "$module"; fi
    done
    ;;
  examples)
    all | while read -r module; do
      if is_example "$module" && [[ "$(parent "$module")" == "$2" ]]; then echo "$module"; fi
    done
    ;;
  parent)
    parent "$2"
    ;;
  *)
    echo "usage: $0 [examples MODULE | parent DIR]" >&2
    exit 2
    ;;
esac
