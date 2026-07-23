package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeCacheKey(t *testing.T) {
	defaults := DefaultsConfig{GoFlags: "-tags=foo"}
	plugin1 := PluginDef{
		ModuleURI: "github.com/smartcontractkit/chainlink-cosmos",
		GitRef:    "v1.0.0",
		Flags:     "-s -w",
	}
	plugin2 := PluginDef{
		ModuleURI: "github.com/smartcontractkit/chainlink-cosmos",
		GitRef:    "v1.0.1", // different gitRef
		Flags:     "-s -w",
	}

	key1 := computeCacheKey("cosmos", plugin1, defaults, "linux", "amd64")
	key2 := computeCacheKey("cosmos", plugin2, defaults, "linux", "amd64")

	assert.NotEqual(t, key1, key2, "expected different cache keys for different gitRef")

	// Verify key format: baseName-hash
	expectedPrefix := "chainlink-cosmos-"
	assert.True(t, strings.HasPrefix(key1, expectedPrefix), "expected key %q to start with prefix %q", key1, expectedPrefix)

	// Hash component check
	h := sha256.New()
	h.Write([]byte("cosmos|github.com/smartcontractkit/chainlink-cosmos|v1.0.0|-s -w|-tags=foo|linux|amd64"))
	expectedKey := fmt.Sprintf("chainlink-cosmos-%x", h.Sum(nil)[:12])
	assert.Equal(t, expectedKey, key1)
}

func TestResolveCacheDir(t *testing.T) {
	tests := []struct {
		name       string
		cliFlag    string
		envVar     string
		wantPrefix string
	}{
		{
			name:       "cli_flag_takes_precedence",
			cliFlag:    "/tmp/cli-cache",
			envVar:     "/tmp/env-cache",
			wantPrefix: "/tmp/cli-cache",
		},
		{
			name:       "env_var_fallback",
			cliFlag:    "",
			envVar:     "/tmp/env-cache",
			wantPrefix: "/tmp/env-cache",
		},
		{
			name:       "default_system_cache",
			cliFlag:    "",
			envVar:     "",
			wantPrefix: "loopinstall-cache",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CL_LOOPINSTALL_CACHE_DIR", tc.envVar)
			got, err := resolveCacheDir(tc.cliFlag)
			require.NoError(t, err)
			require.NotEmpty(t, got)

			if tc.cliFlag != "" {
				assert.Equal(t, tc.cliFlag, got)
			} else if tc.envVar != "" {
				assert.Equal(t, tc.envVar, got)
			} else {
				assert.Contains(t, got, tc.wantPrefix)
			}
		})
	}
}

func TestDownloadAndInstallPlugin_Tier1_LocalCacheHit(t *testing.T) {
	tempCacheDir := t.TempDir()
	tempBinDir := t.TempDir()
	t.Setenv("GOBIN", tempBinDir)

	plugin := PluginDef{
		ModuleURI:   "github.com/smartcontractkit/chainlink-solana",
		GitRef:      "v1.0.0",
		InstallPath: ".",
	}
	defaults := DefaultsConfig{}

	cacheKey := computeCacheKey("solana", plugin, defaults, runtime.GOOS, runtime.GOARCH)
	cachedBinaryPath := filepath.Join(tempCacheDir, cacheKey)
	expectedContent := []byte("cached-binary-content")
	err := os.WriteFile(cachedBinaryPath, expectedContent, 0755)
	require.NoError(t, err)

	// Mock execCommand to fail if go build is called (should NOT be called on Tier 1 hit)
	withMockExec(t, func(cmd *exec.Cmd) error {
		require.Failf(t, "execCommand invoked", "execCommand should not be called on cache hit, command was: %v", cmd.Args)
		return fmt.Errorf("should not be called")
	}, func() {
		err := downloadAndInstallPluginWithCache("solana", 0, plugin, defaults, tempCacheDir, "")
		require.NoError(t, err)

		installedPath := filepath.Join(tempBinDir, "chainlink-solana")
		gotContent, err := os.ReadFile(installedPath)
		require.NoError(t, err)
		assert.Equal(t, string(expectedContent), string(gotContent))
	})
}

func TestDownloadAndInstallPlugin_Tier2_GitHubReleaseAsset(t *testing.T) {
	tempCacheDir := t.TempDir()
	tempBinDir := t.TempDir()
	t.Setenv("GOBIN", tempBinDir)
	t.Setenv("GITHUB_TOKEN", "mock-token")

	// Set up mock HTTP server for GitHub release lookup & asset download
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		if r.URL.Path == "/repos/smartcontractkit/chainlink-starknet/releases/tags/v1.2.0" {
			assetName := fmt.Sprintf("chainlink-starknet_%s_%s", runtime.GOOS, runtime.GOARCH)
			jsonResp := fmt.Sprintf(`{
				"assets": [
					{
						"name": "%s",
						"browser_download_url": "http://%s/download/%s"
					}
				]
			}`, assetName, r.Host, assetName)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(jsonResp))
			return
		}
		if filepath.Base(r.URL.Path) == fmt.Sprintf("chainlink-starknet_%s_%s", runtime.GOOS, runtime.GOARCH) {
			_, _ = w.Write([]byte("release-asset-binary"))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	plugin := PluginDef{
		ModuleURI:   "github.com/smartcontractkit/chainlink-starknet",
		GitRef:      "v1.2.0",
		InstallPath: ".",
	}
	defaults := DefaultsConfig{}

	withMockExec(t, func(cmd *exec.Cmd) error {
		require.Failf(t, "execCommand invoked", "execCommand should not be called on GitHub Release hit, command was: %v", cmd.Args)
		return fmt.Errorf("should not be called")
	}, func() {
		err := downloadAndInstallPluginWithCache("starknet", 0, plugin, defaults, tempCacheDir, "http://"+mockServer.Listener.Addr().String())
		require.NoError(t, err)

		installedPath := filepath.Join(tempBinDir, "chainlink-starknet")
		gotContent, err := os.ReadFile(installedPath)
		require.NoError(t, err)
		assert.Equal(t, "release-asset-binary", string(gotContent))

		// Also check binary was saved to cache
		cacheKey := computeCacheKey("starknet", plugin, defaults, runtime.GOOS, runtime.GOARCH)
		cachedPath := filepath.Join(tempCacheDir, cacheKey)
		cachedContent, err := os.ReadFile(cachedPath)
		require.NoError(t, err)
		assert.Equal(t, "release-asset-binary", string(cachedContent))
	})
}

func TestDownloadAndInstallPlugin_Tier3_SourceBuildAndCacheWrite(t *testing.T) {
	tempCacheDir := t.TempDir()
	tempBinDir := t.TempDir()
	t.Setenv("GOBIN", tempBinDir)

	plugin := PluginDef{
		ModuleURI:   "github.com/smartcontractkit/chainlink-common",
		GitRef:      "v2.0.0",
		InstallPath: ".",
	}
	defaults := DefaultsConfig{}

	cacheKey := computeCacheKey("common", plugin, defaults, runtime.GOOS, runtime.GOARCH)
	targetBinPath := filepath.Join(tempBinDir, "chainlink-common")

	withMockExec(t, func(cmd *exec.Cmd) error {
		// Mock go mod download and go build
		if len(cmd.Args) >= 3 && cmd.Args[1] == "mod" && cmd.Args[2] == "download" {
			type dl struct{ Dir string }
			_ = json.NewEncoder(cmd.Stdout).Encode(dl{Dir: t.TempDir()})
			return nil
		}
		if len(cmd.Args) >= 2 && cmd.Args[0] == "go" && cmd.Args[1] == "build" {
			return os.WriteFile(targetBinPath, []byte("compiled-source-binary"), 0755)
		}
		return fmt.Errorf("unexpected command: %v", cmd.Args)
	}, func() {
		err := downloadAndInstallPluginWithCache("common", 0, plugin, defaults, tempCacheDir, "http://invalid-host")
		require.NoError(t, err)

		// Verify binary copied to cache dir after build
		cachedPath := filepath.Join(tempCacheDir, cacheKey)
		cachedContent, err := os.ReadFile(cachedPath)
		require.NoError(t, err)
		assert.Equal(t, "compiled-source-binary", string(cachedContent))
	})
}
