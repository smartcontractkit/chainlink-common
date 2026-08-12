package wasmbuild

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompressBinary(t *testing.T) {
	t.Parallel()

	want := []byte("wasm-bytes")
	var b bytes.Buffer
	bwr := brotli.NewWriter(&b)
	_, err := bwr.Write(want)
	require.NoError(t, err)
	require.NoError(t, bwr.Close())

	got, err := decompressBinary(b.Bytes())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCacheFilePath(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	path, err := cacheFilePath(repoRoot, "core/capabilities/compute/test/simple/cmd", "abc123")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(repoRoot, ".wasm-cache", "core_capabilities_compute_test_simple_cmd-abc123.wasm.br"), path)
}

func TestPkgSlugRejectsTraversal(t *testing.T) {
	t.Parallel()

	_, err := pkgSlug("../escape")
	require.Error(t, err)
}

func TestReadCacheFile(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	want := []byte("compressed-bytes")
	cachePath := filepath.Join(cacheDir, "fixture.wasm.br")
	require.NoError(t, os.WriteFile(cachePath, want, 0o600))

	got, err := readCacheFile(cachePath)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestBuildLockPerKey(t *testing.T) {
	t.Parallel()

	lockA := buildLock("pkgA")
	assert.Same(t, lockA, buildLock("pkgA"), "same key must return the same lock")
	assert.NotSame(t, lockA, buildLock("pkgB"), "different keys must return different locks")
}

func TestBuildLockSerializesSameKey(t *testing.T) {
	t.Parallel()

	const goroutines = 25
	var active, maxActive int32
	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			mu := buildLock("serialize-test")
			mu.Lock()
			defer mu.Unlock()

			cur := atomic.AddInt32(&active, 1)
			for {
				prev := atomic.LoadInt32(&maxActive)
				if cur <= prev || atomic.CompareAndSwapInt32(&maxActive, prev, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&active, -1)
		})
	}
	wg.Wait()

	assert.Equal(t, int32(1), maxActive, "same-key builds must not run concurrently")
}

func TestFindModuleRoot(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "sub", "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/test\n"), 0o600))

	got, err := findModuleRoot(pkgDir)
	require.NoError(t, err)
	assert.Equal(t, repoRoot, got)
}

func TestFindModuleRootNested(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	nested := filepath.Join(repoRoot, "a", "b", "c")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(filepath.Join(repoRoot, "a"), "go.mod"), []byte("module example.com/test\n"), 0o600))

	got, err := findModuleRoot(nested)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(repoRoot, "a"), got)
}

func TestComputeBuildFingerprintStable(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/test\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\n"), 0o600))

	cfg := Config{GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	ctx := t.Context()

	first, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)
	second, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestComputeBuildFingerprintInvalidatesOnSourceEdit(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/test\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o600))

	goFile := filepath.Join(pkgDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))

	cfg := Config{GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	ctx := t.Context()

	base, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	clearCaches()

	require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nvar X = 1\n"), 0o600))
	afterEdit, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)
	assert.NotEqual(t, base, afterEdit, "go source edit must bust the fingerprint")
}

func TestComputeBuildFingerprintInvalidatesOnGoSumEdit(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/test\n"), 0o600))
	goSumPath := filepath.Join(repoRoot, "go.sum")
	require.NoError(t, os.WriteFile(goSumPath, []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\n"), 0o600))

	cfg := Config{GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	ctx := t.Context()

	base, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	clearCaches()

	require.NoError(t, os.WriteFile(goSumPath, []byte("example.com/dep v1.0.0 h1:def=\n"), 0o600))
	afterEdit, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)
	assert.NotEqual(t, base, afterEdit, "go.sum edit must bust the fingerprint")
}

func TestWriteFingerprintIncludesPlatform(t *testing.T) {
	t.Parallel()

	digests := []fileDigest{{path: "main.go", hash: "abc"}}
	goSumDigest := sha256.Sum256([]byte("gosum"))

	h1 := sha256.New()
	require.NoError(t, writeFingerprint(h1, "go1.26.5", "wasip1", "wasm", defaultBuildFlags, sha256.Sum256([]byte("gomod")), goSumDigest, digests))
	fp1 := hex.EncodeToString(h1.Sum(nil)[:16])

	h2 := sha256.New()
	require.NoError(t, writeFingerprint(h2, "go1.26.5", "linux", "amd64", defaultBuildFlags, sha256.Sum256([]byte("gomod")), goSumDigest, digests))
	fp2 := hex.EncodeToString(h2.Sum(nil)[:16])

	assert.NotEqual(t, fp1, fp2, "different GOOS/GOARCH must produce different fingerprints")
}

func fingerprintFiles(t *testing.T, repoRoot string, pkg listPackage) string {
	t.Helper()
	var digests []fileDigest
	require.NoError(t, hashPackageFiles(repoRoot, pkg, &digests))
	sort.Slice(digests, func(i, j int) bool { return digests[i].path < digests[j].path })
	h := sha256.New()
	require.NoError(t, writeFingerprint(h, "go1.26.5", "wasip1", "wasm", defaultBuildFlags, sha256.Sum256([]byte("gomod")), sha256.Sum256([]byte("gosum")), digests))
	return hex.EncodeToString(h.Sum(nil)[:16])
}

func TestFingerprintInvalidationSequence(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	assetsDir := filepath.Join(pkgDir, "assets")
	require.NoError(t, os.MkdirAll(assetsDir, 0o755))

	goFile := filepath.Join(pkgDir, "main.go")
	embedFile := filepath.Join(assetsDir, "data.txt")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))
	require.NoError(t, os.WriteFile(embedFile, []byte("v1"), 0o600))

	pkg := listPackage{
		Dir:        pkgDir,
		GoFiles:    []string{"main.go"},
		EmbedFiles: []string{"assets/data.txt"},
	}

	base := fingerprintFiles(t, repoRoot, pkg)

	clearCaches()

	require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nvar X = 1\n"), 0o600))
	afterGoEdit := fingerprintFiles(t, repoRoot, pkg)
	assert.NotEqual(t, base, afterGoEdit, "go source edit must bust the fingerprint")

	clearCaches()

	require.NoError(t, os.WriteFile(embedFile, []byte("v2-changed"), 0o600))
	afterEmbedEdit := fingerprintFiles(t, repoRoot, pkg)
	assert.NotEqual(t, afterGoEdit, afterEmbedEdit, "embedded asset edit must bust the fingerprint")
}

func TestFingerprintStableWhenUnchanged(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\n"), 0o600))

	pkg := listPackage{Dir: pkgDir, GoFiles: []string{"main.go"}}

	first := fingerprintFiles(t, repoRoot, pkg)
	second := fingerprintFiles(t, repoRoot, pkg)
	assert.Equal(t, first, second)
}



func TestFingerprintCacheKeySeparatesByBuildFlags(t *testing.T) {
	t.Parallel()

	pkgDir := "/fake/pkg"
	base := Config{GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	other := base
	other.BuildFlags = []string{"-tags", "foo"}

	key1 := fingerprintCacheKey(pkgDir, base)
	key2 := fingerprintCacheKey(pkgDir, other)

	assert.NotEqual(t, key1, key2, "different BuildFlags must produce different fingerprint cache keys")
}

func TestComputeBuildFingerprintInvalidatesOnGoModEdit(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))

	goModPath := filepath.Join(repoRoot, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module example.com/test\n\ngo 1.23\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\n"), 0o600))

	cfg := Config{GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	ctx := t.Context()

	base, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	clearCaches()

	require.NoError(t, os.WriteFile(goModPath, []byte("module example.com/test/v2\n\ngo 1.23\n"), 0o600))
	afterEdit, err := computeBuildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	assert.NotEqual(t, base, afterEdit, "go.mod edit must bust the fingerprint")
}

func TestFileHashCacheAvoidsRedundantDiskReads(t *testing.T) {
	clearCaches()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\n"), 0o600))

	pkg := listPackage{Dir: pkgDir, GoFiles: []string{"main.go"}}

	var digests1 []fileDigest
	require.NoError(t, hashPackageFiles(repoRoot, pkg, &digests1))

	goFile := filepath.Join(pkgDir, "main.go")
	require.NoError(t, os.Remove(goFile))

	var digests2 []fileDigest
	require.NoError(t, hashPackageFiles(repoRoot, pkg, &digests2), "second call must use file hash cache, not disk")

	assert.Equal(t, digests1, digests2, "cached file hashes must match")
}

func TestFileHashCacheSharedAcrossPackages(t *testing.T) {
	clearCaches()

	repoRoot := t.TempDir()
	sharedDir := filepath.Join(repoRoot, "shared")
	pkgADir := filepath.Join(repoRoot, "pkgA")
	pkgBDir := filepath.Join(repoRoot, "pkgB")
	require.NoError(t, os.MkdirAll(sharedDir, 0o755))
	require.NoError(t, os.MkdirAll(pkgADir, 0o755))
	require.NoError(t, os.MkdirAll(pkgBDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(sharedDir, "lib.go"), []byte("package shared\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgADir, "main.go"), []byte("package main\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgBDir, "main.go"), []byte("package main\n"), 0o600))

	sharedPkg := listPackage{Dir: sharedDir, GoFiles: []string{"lib.go"}}

	var digests1 []fileDigest
	require.NoError(t, hashPackageFiles(repoRoot, sharedPkg, &digests1))

	sharedFile := filepath.Join(sharedDir, "lib.go")
	require.NoError(t, os.Remove(sharedFile))

	var digests2 []fileDigest
	require.NoError(t, hashPackageFiles(repoRoot, sharedPkg, &digests2), "second call must use cache for shared file")

	assert.Equal(t, digests1, digests2, "shared dep file hash must be cached across calls")
}

func TestPruneRemovesOldFiles(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	cacheDir := filepath.Join(repoRoot, defaultCacheDirName)
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	oldFile := filepath.Join(cacheDir, "old-pkg-deadbeef.wasm.br")
	recentFile := filepath.Join(cacheDir, "recent-pkg-cafebabe.wasm.br")
	require.NoError(t, os.WriteFile(oldFile, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(recentFile, []byte("recent"), 0o600))

	oldTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	removed, err := Prune(repoRoot, 24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, removed, "only the old file should be pruned")

	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err), "old file must be removed")

	_, err = os.Stat(recentFile)
	assert.NoError(t, err, "recent file must survive prune")
}

func TestPruneEmptyDirIsNoop(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()

	removed, err := Prune(repoRoot, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 0, removed)
}

func TestEnsureCacheDirIdempotent(t *testing.T) {
	clearCaches()

	repoRoot := t.TempDir()

	dir1, err := ensureCacheDir(repoRoot)
	require.NoError(t, err)

	dir2, err := ensureCacheDir(repoRoot)
	require.NoError(t, err)

	assert.Equal(t, dir1, dir2)

	_, ok := cacheDirsCreated.Load(dir1)
	assert.True(t, ok, "cache dir must be marked as created")
}

func TestWriteFingerprintBuildFlagsNullDelimiter(t *testing.T) {
	t.Parallel()

	goModDigest := sha256.Sum256([]byte("gomod"))
	goSumDigest := sha256.Sum256([]byte("gosum"))
	digests := []fileDigest{{path: "main.go", hash: "abc"}}

	h1 := sha256.New()
	require.NoError(t, writeFingerprint(h1, "go1.26.5", "wasip1", "wasm", []string{"-tags", "foo bar"}, goModDigest, goSumDigest, digests))
	fp1 := hex.EncodeToString(h1.Sum(nil)[:16])

	h2 := sha256.New()
	require.NoError(t, writeFingerprint(h2, "go1.26.5", "wasip1", "wasm", []string{"-tags foo", "bar"}, goModDigest, goSumDigest, digests))
	fp2 := hex.EncodeToString(h2.Sum(nil)[:16])

	assert.NotEqual(t, fp1, fp2, "different flag boundaries with spaces must produce different fingerprints")
}

func TestBuildInvalidatesCacheOnSourceEdit(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	pkgDir := filepath.Join(repoRoot, "pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module example.com/test\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, "go.sum"), []byte("example.com/dep v1.0.0 h1:abc=\n"), 0o600))

	goFile := filepath.Join(pkgDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o600))

	cfg := Config{PkgDir: pkgDir, RepoRoot: repoRoot, GOOS: "wasip1", GOARCH: "wasm", BuildFlags: defaultBuildFlags}
	ctx := t.Context()

	fp1, err := buildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	clearCaches()

	require.NoError(t, os.WriteFile(goFile, []byte("package main\nvar Modified = true\n"), 0o600))

	fp2, err := buildFingerprint(ctx, pkgDir, repoRoot, cfg)
	require.NoError(t, err)

	assert.NotEqual(t, fp1, fp2, "source edit must produce new build fingerprint")
}



