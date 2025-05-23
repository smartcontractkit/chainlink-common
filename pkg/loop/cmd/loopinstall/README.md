# loopinstall

A tool for installing Chainlink LOOP plugins from YAML configuration files.

## Installation

To install from source:

```bash
go install github.com/smartcontractkit/chainlink-common/pkg/loop/cmd/loopinstall@latest
```

Or clone and install locally:

```bash
git clone https://github.com/smartcontractkit/chainlink-common.git
cd chainlink-common
go install ./pkg/loop/cmd/loopinstall
```

## Usage

```bash
# Run `loopinstall --help` to see the options.
loopinstall [options] <plugin-config-file> [<plugin-config-file>...]
```

## Configuration

_(See `plugins.example.yaml` for a complete example configuration.)_

Example configuration file structure:

```yaml
defaults:
  # Go build flags.
  # The `-s` flag is added to strip debug information from the binary to reduce
  # the binary size for releases.
  # See: `go tool link -help`
  goflags: "-ldflags='-s'" # Default Go build flags

plugins:
  cosmos:
    - name: "default"
      moduleURI: "github.com/smartcontractkit/chainlink-cosmos"
      gitRef: "f740e9ae54e79762991bdaf8ad6b50363261c056"
      installPath: "./pkg/cosmos/cmd/chainlink-cosmos"
      libs:
        - "/go/pkg/mod/github.com/!cosm!wasm/wasmvm@v*/internal/api/libwasmvm.*.so"
```

The `installPath` is relative to the root of the downloaded module but it can also be an absolute path like: `github.com/smartcontractkit/chainlink-cosmos/pkg/cosmos/cmd/chainlink-cosmos`. The absolute path will be stripped to the relative path (it is supported for backwards compatibility). The path is cleaned using Go's `filepath.Clean` function before use. This process normalizes the path by:

- Resolving `.` (current directory) and `..` (parent directory) components (e.g., `path/to/../plugin` becomes `path/plugin`).
- Removing redundant slashes (e.g., `path//plugin` becomes `path/plugin`).

The `libs` field is an array of strings representing directory paths, which can include glob patterns for library files that need to be included with the plugin. Docker build will use these paths to copy the libraries into the final container image.

## Private Repository Access

To install plugins from private repositories:

1. Set your GitHub token. Use the `gh` cli and [gh auth setup-git](https://cli.github.com/manual/gh_auth_setup-git).

2. **Optional for CI/CD only**: Configure Git to use HTTPS with token authentication:

   ```bash
   git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
   ```

3. Configure GOPRIVATE via environment variable:
   ```bash
   export GOPRIVATE=github.com/myorg/*,github.com/another-org/*
   ```

The `GOPRIVATE` environment variable supports glob patterns (e.g., `github.com/myorg/*`) and tells Go to bypass the public module proxy for these repositories.
