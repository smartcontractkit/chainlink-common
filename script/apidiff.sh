#!/usr/bin/env bash
set -euo pipefail

# script/apidiff.sh
# Compare API of only modified Go packages between PR HEAD and default branch using apidiff, formatted for GitHub Actions.
# Usage: ./script/apidiff.sh

# Ensure apidiff is installed
if ! command -v apidiff &> /dev/null; then
  echo "::warning::apidiff not found, installing..."
  go install golang.org/x/exp/apidiff@latest
fi

# Locate repository root
repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

echo "Repository root: $repo_root"

# Determine default branch
default_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||' || echo "main")
echo "Default branch: $default_branch"

# Fetch latest origin
git fetch origin "$default_branch"

# Determine modified Go files relative to base
raw_changes=$(git diff --name-only "origin/$default_branch...HEAD" -- '*.go')
if [[ -z "$raw_changes" ]]; then
  echo "No modified Go files detected relative to origin/$default_branch. Exiting."
  exit 0
fi

# Derive unique directories
readarray -t changed_dirs < <(
  printf '%s
' "$raw_changes" |
  xargs -n1 dirname |
  sort -u
)

echo "::group::Modified directories"
echo "${changed_dirs[@]}"
echo "::endgroup::"

# Create temporary workspace
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

# Add worktrees for base and head
git worktree add "$tmp_dir/base" "origin/$default_branch"
git worktree add "$tmp_dir/head" HEAD

# Resolve import paths for changed directories
declare -a pkgs
for dir in "${changed_dirs[@]}"; do
  pkg_path=$(cd "$tmp_dir/head" && go list "./$dir")
  pkgs+=("$pkg_path")
done
# Deduplicate
readarray -t pkgs < <(printf '%s
' "${pkgs[@]}" | sort -u)

echo "::group::Packages to compare"
echo "${pkgs[@]}"
echo "::endgroup::"

# Prepare export dirs
exports_dir="$tmp_dir/exports"
base_exports="$exports_dir/base"
head_exports="$exports_dir/head"
mkdir -p "$base_exports" "$head_exports"

# Generate exports
generate_pkg_exports() {
  local tree=$1 dest=$2
  pushd "$tree" > /dev/null
  for pkg in "${pkgs[@]}"; do
    local file=${pkg//\//_}.export
    apidiff -w "$dest/$file" "$pkg"
  done
  popd > /dev/null
}

generate_pkg_exports "$tmp_dir/base" "$base_exports"
generate_pkg_exports "$tmp_dir/head" "$head_exports"

# Compare exports for breaking changes
echo "::group::API Comparison"
broken=false
declare -a broken_pkgs
for pkg in "${pkgs[@]}"; do
  file=${pkg//\//_}.export
  echo "::group::Checking $pkg"
  # show full apidiff output
  output=$(apidiff "$base_exports/$file" "$head_exports/$file" 2>&1)
  echo "${output}"
  if grep -q "Incompatible changes" <<< "$output"; then
    broken=true
    broken_pkgs+=("$pkg")
    echo "::error title=API break detected::$pkg has breaking changes"
  else
    echo "::notice::$pkg: no incompatible changes"
  fi
  echo "::endgroup::"
done

# Summary of broken packages
if [[ "$broken" == true ]]; then
  echo "::group::Breaking API Summary"
  for pkg in "${broken_pkgs[@]}"; do
    echo "::error::$pkg has breaking changes"
  done
  echo "::endgroup::"
fi
# end API Comparison
echo "::endgroup::"

# Clean up worktrees
git worktree remove "$tmp_dir/base" --force
git worktree remove "$tmp_dir/head" --force

# Final status
if [[ "$broken" == true ]]; then
  exit 1
else
  echo "::notice::No breaking API changes detected."
fi
