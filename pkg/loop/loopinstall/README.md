# loopinstall

A tool for installing Chainlink LOOP plugins from YAML configuration files.

## Installation

To install from source:

```bash
go install github.com/smartcontractkit/chainlink-common/pkg/loop/loopinstall@latest
```

Or clone and install locally:

```bash
git clone https://github.com/smartcontractkit/chainlink-common.git
cd chainlink-common
go install ./pkg/loop/loopinstall
```

## Usage

```bash
loopinstall [options] <plugin-config-file> [<plugin-config-file>...]
```

## Configuration

_(See `plugins.example.yaml` for a complete example configuration.)_

Example configuration file structure:

```yaml
defaults:
  goflags: "-ldflags='-s -w'" # Default Go build flags
  goprivate: "github.com/myorg/myrepo,github.com/private-org/*" # Comma-separated list of private repo patterns

plugins:
  cosmos:
    - name: "default"
      enabled: true
      moduleURI: "github.com/smartcontractkit/chainlink-cosmos"
      gitRef: "f740e9ae54e79762991bdaf8ad6b50363261c056"
      installPath: "github.com/smartcontractkit/chainlink-cosmos/pkg/cosmos/cmd/chainlink-cosmos"
      additionalFiles:
        - src: "/go/pkg/mod/github.com/!cosm!wasm/wasmvm@v*/internal/api/libwasmvm.*.so"
          dest: "/usr/lib/"
```

## Private Repository Access

To install plugins from private repositories:

1. Set your GitHub token:

   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

   > **Note**: You can use `GITHUB_TOKEN=$(gh auth token)` if you have the GitHub CLI installed.

2. **Optional for CI/CD only**: Configure Git to use HTTPS with token authentication:

   ```bash
   git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
   ```

3. Configure GOPRIVATE using one of these methods:
   - In your config file under `defaults`:
     ```yaml
     defaults:
       goprivate: "github.com/myorg/*,github.com/private-org/*"
     ```
   - Or via environment variable:
     ```bash
     export GOPRIVATE=github.com/myorg/*,github.com/another-org/*
     ```

The `goprivate` setting supports glob patterns (e.g., `github.com/myorg/*`) and tells Go to bypass the public module proxy for these repositories.
