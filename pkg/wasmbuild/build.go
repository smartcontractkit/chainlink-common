package wasmbuild

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
)

const (
	defaultCacheDirName = ".wasm-cache"
	defaultGOOS         = "wasip1"
	defaultGOARCH       = "wasm"
)

var defaultBuildFlags = []string{"-trimpath", "-buildvcs=false"}

// Config controls a single WASM build invocation.
type Config struct {
	// PkgDir is the directory containing the Go package to compile to WASM.
	// May be absolute or relative to the current working directory.
	PkgDir string

	// RepoRoot is the git repository root, used for fingerprinting source
	// paths, locating .wasm-cache/, and filtering transitive deps to in-repo
	// files. If empty, auto-discovered via `git rev-parse --show-toplevel`.
	RepoRoot string

	// Compress controls whether the returned bytes are brotli-compressed.
	// The on-disk cache always stores compressed bytes; when Compress is false
	// the bytes are decompressed before returning.
	Compress bool

	// BuildFlags passed to `go build`. Defaults to [-trimpath -buildvcs=false].
	BuildFlags []string

	// GOOS for cross-compilation. Defaults to "wasip1".
	GOOS string

	// GOARCH for cross-compilation. Defaults to "wasm".
	GOARCH string
}

// Build compiles the Go package in cfg.PkgDir to a WASM binary and returns it.
// The binary is content-addressed cached in .wasm-cache/ under the repo root.
// Concurrent calls for the same package are serialized; different packages
// build in parallel.
func Build(ctx context.Context, cfg Config) ([]byte, error) {
	if cfg.GOOS == "" {
		cfg.GOOS = defaultGOOS
	}
	if cfg.GOARCH == "" {
		cfg.GOARCH = defaultGOARCH
	}
	if cfg.BuildFlags == nil {
		cfg.BuildFlags = defaultBuildFlags
	}

	repoRoot := cfg.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = discoverRepoRoot()
		if err != nil {
			return nil, fmt.Errorf("discover repo root: %w", err)
		}
	}

	absPkgDir, err := filepath.Abs(cfg.PkgDir)
	if err != nil {
		return nil, fmt.Errorf("resolve package dir: %w", err)
	}
	if absPkgDir, err = filepath.EvalSymlinks(absPkgDir); err != nil {
		return nil, fmt.Errorf("resolve package dir symlinks: %w", err)
	}

	cacheKey := fmt.Sprintf("%s:%t", absPkgDir, cfg.Compress)
	if cached, ok := loadBinaryCache(cacheKey); ok {
		return cached, nil
	}

	mu := buildLock(absPkgDir)
	mu.Lock()
	defer mu.Unlock()

	if cached, ok := loadBinaryCache(cacheKey); ok {
		return cached, nil
	}

	binary, err := getOrBuildBinary(ctx, absPkgDir, repoRoot, cfg)
	if err != nil {
		return nil, err
	}

	storeBinaryCache(cacheKey, binary)
	return binary, nil
}

var (
	binaryCache   = make(map[string][]byte)
	binaryCacheMu sync.RWMutex

	fingerprintCache   = make(map[string]string)
	fingerprintCacheMu sync.RWMutex

	buildLocksMu sync.Mutex
	buildLocks   = make(map[string]*sync.Mutex)
)

func loadBinaryCache(cacheKey string) ([]byte, bool) {
	binaryCacheMu.RLock()
	defer binaryCacheMu.RUnlock()
	cached, ok := binaryCache[cacheKey]
	if !ok {
		return nil, false
	}
	res := make([]byte, len(cached))
	copy(res, cached)
	return res, true
}

func storeBinaryCache(cacheKey string, binary []byte) {
	cp := make([]byte, len(binary))
	copy(cp, binary)
	binaryCacheMu.Lock()
	binaryCache[cacheKey] = cp
	binaryCacheMu.Unlock()
}

func buildLock(key string) *sync.Mutex {
	buildLocksMu.Lock()
	defer buildLocksMu.Unlock()
	mu, ok := buildLocks[key]
	if !ok {
		mu = &sync.Mutex{}
		buildLocks[key] = mu
	}
	return mu
}

func clearFingerprintCache() {
	fingerprintCacheMu.Lock()
	fingerprintCache = make(map[string]string)
	fingerprintCacheMu.Unlock()
}

func discoverRepoRoot() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output() // #nosec
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func getOrBuildBinary(ctx context.Context, absPkgDir, repoRoot string, cfg Config) ([]byte, error) {
	repoRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repo root symlinks: %w", err)
	}

	pkgRel, err := filepath.Rel(repoRoot, absPkgDir)
	if err != nil {
		return nil, fmt.Errorf("compute package relative path: %w", err)
	}
	if pkgRel == "." || strings.HasPrefix(pkgRel, "..") {
		return nil, fmt.Errorf("package dir %s is not under repo root %s", absPkgDir, repoRoot)
	}

	fingerprint, err := buildFingerprint(ctx, absPkgDir, repoRoot, cfg)
	if err != nil {
		return nil, fmt.Errorf("compute build fingerprint for %s: %w", pkgRel, err)
	}

	cachePath, err := cacheFilePath(repoRoot, pkgRel, fingerprint)
	if err != nil {
		return nil, err
	}

	compressed, err := readCacheFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read WASM cache %s: %w", cachePath, err)
		}

		compressed, err = buildAndCacheBinary(ctx, absPkgDir, repoRoot, pkgRel, cachePath, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Compress {
		res := make([]byte, len(compressed))
		copy(res, compressed)
		return res, nil
	}

	return decompressBinary(compressed)
}

func cacheFilePath(repoRoot, pkgRelPath, fingerprint string) (string, error) {
	slug, err := pkgSlug(pkgRelPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, defaultCacheDirName, fmt.Sprintf("%s-%s.wasm.br", slug, fingerprint)), nil
}

func pkgSlug(pkgRelPath string) (string, error) {
	clean := filepath.Clean(pkgRelPath)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("invalid package path: %s", pkgRelPath)
	}
	return strings.ReplaceAll(clean, string(filepath.Separator), "_"), nil
}

func readCacheFile(cachePath string) ([]byte, error) {
	return os.ReadFile(cachePath)
}

func decompressBinary(compressed []byte) ([]byte, error) {
	var b bytes.Buffer
	bwr := brotli.NewReader(bytes.NewReader(compressed))
	if _, err := io.Copy(&b, bwr); err != nil {
		return nil, fmt.Errorf("decompress WASM cache failed: %w", err)
	}
	return b.Bytes(), nil
}

func buildAndCacheBinary(ctx context.Context, absPkgDir, repoRoot, pkgRel, cachePath string, cfg Config) ([]byte, error) {
	compressed, err := buildBinary(ctx, absPkgDir, cfg)
	if err != nil {
		return nil, fmt.Errorf("build WASM for %s: %w", pkgRel, err)
	}

	cacheDir := filepath.Join(repoRoot, defaultCacheDirName)
	if err = os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("create WASM cache dir: %w", err)
	}

	tmp, err := os.CreateTemp(cacheDir, filepath.Base(cachePath)+".tmp.*")
	if err != nil {
		return nil, fmt.Errorf("create WASM cache temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(compressed); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("write WASM cache temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return nil, fmt.Errorf("close WASM cache temp file: %w", err)
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		_ = os.Remove(tmpPath)
		if existing, readErr := os.ReadFile(cachePath); readErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("write WASM cache file: %w", err)
	}

	return compressed, nil
}

func buildFingerprint(ctx context.Context, absPkgDir, repoRoot string, cfg Config) (string, error) {
	cacheKey := absPkgDir + cfg.GOOS + cfg.GOARCH
	fingerprintCacheMu.RLock()
	if fp, ok := fingerprintCache[cacheKey]; ok {
		fingerprintCacheMu.RUnlock()
		return fp, nil
	}
	fingerprintCacheMu.RUnlock()

	fp, err := computeBuildFingerprint(ctx, absPkgDir, repoRoot, cfg)
	if err != nil {
		return "", err
	}

	fingerprintCacheMu.Lock()
	fingerprintCache[cacheKey] = fp
	fingerprintCacheMu.Unlock()

	return fp, nil
}

type listPackage struct {
	Dir        string   `json:"Dir"`
	GoFiles    []string `json:"GoFiles"`
	EmbedFiles []string `json:"EmbedFiles"`
}

type fileDigest struct {
	path string
	hash string
}

func computeBuildFingerprint(ctx context.Context, absPkgDir, repoRoot string, cfg Config) (string, error) {
	repoRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repo root symlinks: %w", err)
	}

	goVersion, err := goEnv(ctx, "GOVERSION")
	if err != nil {
		return "", err
	}

	moduleRoot, err := findModuleRoot(absPkgDir)
	if err != nil {
		return "", fmt.Errorf("find module root for %s: %w", absPkgDir, err)
	}

	goSumPath := filepath.Join(moduleRoot, "go.sum")
	goSum, err := os.ReadFile(goSumPath)
	if err != nil {
		return "", fmt.Errorf("read go.sum at %s: %w", goSumPath, err)
	}
	goSumDigest := sha256.Sum256(goSum)

	pkgs, err := listDeps(ctx, absPkgDir, cfg)
	if err != nil {
		return "", err
	}

	var digests []fileDigest

	for _, pkg := range pkgs {
		if pkg.Dir == "" {
			continue
		}
		if pkg.Dir != repoRoot && !strings.HasPrefix(pkg.Dir, repoRoot+string(filepath.Separator)) {
			continue
		}

		if err := hashPackageFiles(repoRoot, pkg, &digests); err != nil {
			return "", fmt.Errorf("hash sources in %s: %w", pkg.Dir, err)
		}
	}

	sort.Slice(digests, func(i, j int) bool {
		return digests[i].path < digests[j].path
	})

	h := sha256.New()
	if err := writeFingerprint(h, goVersion, cfg.GOOS, cfg.GOARCH, cfg.BuildFlags, goSumDigest, digests); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)[:16]), nil
}

func findModuleRoot(pkgDir string) (string, error) {
	dir := pkgDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find go.mod walking up from package dir")
		}
		dir = parent
	}
}

func hashPackageFiles(repoRoot string, pkg listPackage, digests *[]fileDigest) error {
	root, err := os.OpenRoot(pkg.Dir)
	if err != nil {
		return err
	}
	defer func() {
		_ = root.Close()
	}()

	names := make([]string, 0, len(pkg.GoFiles)+len(pkg.EmbedFiles))
	names = append(names, pkg.GoFiles...)
	names = append(names, pkg.EmbedFiles...)

	for _, name := range names {
		data, err := root.ReadFile(name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, err := filepath.Rel(repoRoot, filepath.Join(pkg.Dir, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		*digests = append(*digests, fileDigest{
			path: filepath.ToSlash(rel),
			hash: hex.EncodeToString(sum[:]),
		})
	}
	return nil
}

func writeFingerprint(h io.Writer, goVersion, goos, goarch string, buildFlags []string, goSumDigest [sha256.Size]byte, digests []fileDigest) error {
	if _, err := io.WriteString(h, goVersion); err != nil {
		return fmt.Errorf("hash go version: %w", err)
	}
	if _, err := fmt.Fprintf(h, "\n%s\n%s\n", goos, goarch); err != nil {
		return fmt.Errorf("hash platform: %w", err)
	}
	if _, err := io.WriteString(h, strings.Join(buildFlags, " ")); err != nil {
		return fmt.Errorf("hash build flags: %w", err)
	}
	if _, err := io.WriteString(h, "\n"); err != nil {
		return fmt.Errorf("hash separator: %w", err)
	}
	if _, err := h.Write(goSumDigest[:]); err != nil {
		return fmt.Errorf("hash go.sum digest: %w", err)
	}
	if _, err := io.WriteString(h, "\n"); err != nil {
		return fmt.Errorf("hash separator: %w", err)
	}
	for _, d := range digests {
		if _, err := io.WriteString(h, d.path); err != nil {
			return fmt.Errorf("hash source path: %w", err)
		}
		if _, err := io.WriteString(h, "\n"); err != nil {
			return fmt.Errorf("hash separator: %w", err)
		}
		if _, err := io.WriteString(h, d.hash); err != nil {
			return fmt.Errorf("hash source digest: %w", err)
		}
		if _, err := io.WriteString(h, "\n"); err != nil {
			return fmt.Errorf("hash separator: %w", err)
		}
	}
	return nil
}

func listDeps(ctx context.Context, pkgDir string, cfg Config) ([]listPackage, error) {
	listCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	cmd := exec.CommandContext(listCtx, "go", "list", "-deps", "-json", ".") // #nosec
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOOS="+cfg.GOOS, "GOARCH="+cfg.GOARCH)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list -deps in %s failed: %s: %w", pkgDir, strings.TrimSpace(stderr.String()), err)
	}

	var pkgs []listPackage
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var pkg listPackage
		if err := dec.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode go list output: %w", err)
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func goEnv(ctx context.Context, key string) (string, error) {
	envCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(envCtx, "go", "env", key) // #nosec
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env %s: %w", key, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func buildBinary(ctx context.Context, pkgDir string, cfg Config) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "wasmbuild-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	buildCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	buildPath := filepath.Join(tmpDir, "output.wasm")
	args := append([]string{"build"}, cfg.BuildFlags...)
	args = append(args, "-o", buildPath, ".")

	cmd := exec.CommandContext(buildCtx, "go", args...) // #nosec
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOOS="+cfg.GOOS, "GOARCH="+cfg.GOARCH, "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed: %s %w", string(output), err)
	}

	binary, err := os.ReadFile(buildPath)
	if err != nil {
		return nil, fmt.Errorf("read built binary: %w", err)
	}

	var b bytes.Buffer
	bwr := brotli.NewWriter(&b)
	if _, err = bwr.Write(binary); err != nil {
		return nil, err
	}
	if err = bwr.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
