#!/usr/bin/env bash
set -euo pipefail

# script/apidiff.sh
# Compare API of only modified Go packages between PR HEAD and default branch using apidiff.
# Usage: ./script/apidiff.sh

# Ensure apidiff is installed
if ! command -v apidiff &> /dev/null; then
  echo "apidiff not found, installing..."
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

echo "Modified directories: ${changed_dirs[*]}"

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

echo "Packages to compare: ${pkgs[*]}"

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
    echo "Exporting $pkg -> $dest/$file"
    apidiff -w "$dest/$file" "$pkg"
  done
  popd > /dev/null
}

generate_pkg_exports "$tmp_dir/base" "$base_exports"
generate_pkg_exports "$tmp_dir/head" "$head_exports"

# Compare exports for breaking changes
echo -e "\nComparing API for breaking changes..."
broken=false
for pkg in "${pkgs[@]}"; do
  file=${pkg//\//_}.export
  echo -e "\nChecking $pkg"
  if ! apidiff "$base_exports/$file" "$head_exports/$file"; then
    broken=true
  fi
done

# Clean up worktrees
git worktree remove "$tmp_dir/base" --force
git worktree remove "$tmp_dir/head" --force

# Final status
if [[ "$broken" == true ]]; then
  echo -e "\nBreaking API changes detected."
  exit 1
else
  echo -e "\nNo breaking API changes detected."
fi
