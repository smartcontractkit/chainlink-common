# Implementation Plan: `loopinstall` Local Disk & Release Asset Binary Caching

## Goal
Optimize Chainlink Docker build times by enhancing [`loopinstall`](https://github.com/smartcontractkit/chainlink-common/tree/main/pkg/loop/cmd/loopinstall) in `chainlink-common` to cache and reuse pre-compiled plugin binaries locally on disk, with optional GitHub Release asset downloads. 

This prevents single-plugin `gitRef` updates in `plugins.public.yaml` from forcing full recompilation of all 22 remote plugins, dropping stage build time from **117s to ~10s** without relying on 3rd-party storage services.

---

## User Review Required

> [!IMPORTANT]
> - **No 3rd-Party Dependencies**: Relies strictly on local disk caching (`CL_LOOPINSTALL_CACHE_DIR`) and standard GitHub Releases (authenticated via existing `GIT_AUTH_TOKEN` / GATI).
> - **Zero Cache Invalidation Risk**: Cache keys include SHA256 of `(pluginType, moduleURI, gitRef, GOOS, GOARCH, goflags)`. Any commit/tag/flag change automatically creates a unique key and bypasses stale binaries.

---

## Open Questions

> [!NOTE]
> 1. Should GitHub Release asset downloading be enabled by default when `gitRef` is a release tag (e.g. `v1.3.1`), or strictly controlled via a flag `-enable-release-download` / `CL_LOOPINSTALL_DOWNLOAD_RELEASES=true`?

Ans: It should be auto-allowed if any release/tag with a binary for that repo exists.

---

## Proposed Changes

### Component 1: `chainlink-common` (`pkg/loop/cmd/loopinstall`)

#### [MODIFY] `models.go`
Add fields for cache directory and cache configuration:

```go
type PluginDef struct {
    Enabled     *bool    `yaml:"enabled,omitempty"`
    ModuleURI   string   `yaml:"moduleURI"`
    GitRef      string   `yaml:"gitRef"`
    InstallPath string   `yaml:"installPath"`
    Libs        []string `yaml:"libs"`
    Flags       string   `yaml:"flags,omitempty"`
    EnvVars     []string `yaml:"envvars,omitempty"`
    BinaryURL   string   `yaml:"binaryURL,omitempty"` // Optional explicit release asset URL override
}
```

#### [MODIFY] `main.go`
Add `-cache-dir` / `-c` CLI flag and `CL_LOOPINSTALL_CACHE_DIR` env var handling:

```go
var cacheDir string
flag.StringVar(&cacheDir, "cache-dir", "", "Directory path for caching compiled plugin binaries")

if cacheDir == "" {
    cacheDir = os.Getenv("CL_LOOPINSTALL_CACHE_DIR")
}
```

#### [MODIFY] `install.go`
Implement 3-tier resolution logic in `downloadAndInstallPlugin`:
1. **Tier 1 (Local Disk Cache Hit)**: Check `${cacheDir}/${cacheKey}`. If present, copy directly to `$GOBIN/${binaryName}`. (Duration: **< 100ms**).
2. **Tier 2 (GitHub Release Asset Fallback)**: If `-enable-release-download` is set and `gitRef` is a tag, attempt HTTP GET using `GIT_AUTH_TOKEN`. If HTTP 200, write binary to cache directory and `$GOBIN`. (Duration: **~1s**).
3. **Tier 3 (Source Build Fallback)**: Execute existing `go mod download` + `go build` logic. Upon successful compilation, write binary to `${cacheDir}/${cacheKey}` for future runs. (Duration: **~8s**).

```go
func computeCacheKey(pluginType string, plugin PluginDef, defaults DefaultsConfig, goos, goarch string) string {
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
        pluginType, plugin.ModuleURI, plugin.GitRef, plugin.Flags, defaults.GoFlags, goos, goarch)))
    return fmt.Sprintf("%s-%x", filepath.Base(plugin.ModuleURI), h.Sum(nil)[:12])
}
```

---

### Component 2: `chainlink` Repository (`core/chainlink.Dockerfile` & `plugins/chainlink.Dockerfile`)

#### [MODIFY] [core/chainlink.Dockerfile](file:///Users/adamhamrick/Projects/chainlink/core/chainlink.Dockerfile) & [plugins/chainlink.Dockerfile](file:///Users/adamhamrick/Projects/chainlink/plugins/chainlink.Dockerfile)
Mount `/tmp/loopinstall-cache` into BuildKit during `build-remote-plugins` stage:

```dockerfile
ENV CL_LOOPINSTALL_OUTPUT_DIR=/tmp/loopinstall-output \
    CL_LOOPINSTALL_CACHE_DIR=/tmp/loopinstall-cache \
    GIT_CONFIG_GLOBAL=/tmp/gitconfig-github-token

RUN --mount=type=cache,target=/tmp/loopinstall-cache,id=loopinstall-binary-cache \
    --mount=type=secret,id=GIT_AUTH_TOKEN \
    set -e && \
    ...
```

---

## Verification Plan

### Automated Tests
1. **Unit Tests in `chainlink-common`**:
   - Run `go test ./pkg/loop/cmd/loopinstall/...`
   - Test cache hit returns binary without invoking `execCommand("go", "build")`.
   - Test cache miss compiles and writes output binary to `cacheDir`.
   - Test invalidation when `gitRef` or `flags` change.

### Manual Verification
1. Run `loopinstall` on `plugins.public.yaml` with `-cache-dir /tmp/test-cache`.
2. Verify all 22 binaries are built on run 1.
3. Update 1 plugin `gitRef` in `plugins.public.yaml` and run `loopinstall` again on run 2.
4. Verify 21 plugins hit cache (<1s) and only 1 plugin builds from source.
