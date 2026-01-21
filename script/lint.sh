#!/usr/bin/env bash

# run_linter.sh
# This script runs golangci-lint either locally or in a Docker container based on version matching.

# Parameters
GOLANGCI_LINT_VERSION="$1"
COMMON_OPTS="$2"
DIRECTORY="$3"
EXTRA_OPTS="$4"
OUTPUT_FILE="$DIRECTORY/$(date +%Y-%m-%d_%H:%M:%S).txt"

# Require GOLANGCI_LINT_VERSION as first argument
if [[ -z "${GOLANGCI_LINT_VERSION:-}" ]]; then
    echo "Error: GOLANGCI_LINT_VERSION env var is required." >&2
    echo "Usage: $0 <golangci-lint-version> <common-opts> <directory> [extra-opts]" >&2
    exit 1
fi


# Prepare the lint directory
mkdir -p "$DIRECTORY"

# Check if user provided a local config path via env var
if [[ -n "${GOLANGCI_LINT_CONFIG:-}" ]]; then
    CONFIG_FILE="$GOLANGCI_LINT_CONFIG"
    echo "Using user-provided config: $CONFIG_FILE"
else
    # NOTE: Keep this version in sync with the action tag in /.github/workflows/golangci_lint.yml
    ACTION_CI_LINT_GO_GIT_TAG="${CI_LINT_GO_VERSION:-ci-lint-go/v4}"
    # Download remote golangci-lint config to gitignored directory
    REMOTE_CONFIG_URL="https://raw.githubusercontent.com/smartcontractkit/.github/refs/tags/${ACTION_CI_LINT_GO_GIT_TAG}/actions/ci-lint-go/files/golangci-default.yml"
    CONFIG_FILE="$DIRECTORY/golangci.remote.yml"
    echo "Downloading remote config from: $REMOTE_CONFIG_URL"
    curl -sfL "$REMOTE_CONFIG_URL" -o "$CONFIG_FILE"
fi

# Always use the selected config
COMMON_OPTS="$COMMON_OPTS --config $CONFIG_FILE"

DOCKER_CMD="docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v$GOLANGCI_LINT_VERSION golangci-lint run $COMMON_OPTS $EXTRA_OPTS"

if command -v golangci-lint >/dev/null 2>&1; then
    LOCAL_VERSION=$(golangci-lint version 2>&1 | grep -oE "version [0-9.]+" |  sed "s|version ||")

    if [ "$LOCAL_VERSION" = "$GOLANGCI_LINT_VERSION" ]; then
        echo "Local golangci-lint version ($LOCAL_VERSION) matches desired version ($GOLANGCI_LINT_VERSION). Using local version."
        # shellcheck disable=SC2086
        golangci-lint run $COMMON_OPTS $EXTRA_OPTS | tee "$OUTPUT_FILE"

    else
        echo "Local golangci-lint version ($LOCAL_VERSION) does not match desired version ($GOLANGCI_LINT_VERSION). Using Docker version."
        $DOCKER_CMD | tee "$OUTPUT_FILE"
    fi

else
    echo "Local golangci-lint not found. Using Docker version."
    $DOCKER_CMD | tee "$OUTPUT_FILE"
fi

echo "Linting complete. Results saved to $OUTPUT_FILE"
