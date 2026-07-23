package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// resolveCacheDir resolves the cache directory location.
// Priority:
// 1) cliCacheDir parameter
// 2) CL_LOOPINSTALL_CACHE_DIR environment variable
// 3) User cache directory ($HOME/Library/Caches on macOS, $XDG_CACHE_HOME on Linux) or temp directory fallback
func resolveCacheDir(cliCacheDir string) (string, error) {
	if cliCacheDir != "" {
		return cliCacheDir, nil
	}
	if envCache := os.Getenv("CL_LOOPINSTALL_CACHE_DIR"); envCache != "" {
		return envCache, nil
	}
	userCache, err := os.UserCacheDir()
	if err != nil || userCache == "" {
		return filepath.Join(os.TempDir(), "loopinstall-cache"), nil
	}
	return filepath.Join(userCache, "loopinstall-cache"), nil
}

// computeCacheKey calculates a deterministic cache key for a plugin.
func computeCacheKey(pluginType string, plugin PluginDef, defaults DefaultsConfig, goos, goarch string) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s|%s|%s|%s|%s|%s|%s|%s|%s|%v|%v",
		pluginType, plugin.ModuleURI, plugin.GitRef, plugin.InstallPath, plugin.Flags, defaults.GoFlags, plugin.BinaryURL, goos, goarch, plugin.EnvVars, defaults.EnvVars)
	base := filepath.Base(filepath.Clean(plugin.ModuleURI))
	return fmt.Sprintf("%s-%x", base, h.Sum(nil)[:12])
}

// getAuthToken checks GIT_AUTH_TOKEN first, falling back to GITHUB_TOKEN.
func getAuthToken() string {
	if token := os.Getenv("GIT_AUTH_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

// githubReleaseAsset represents GitHub release JSON structure.
type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubReleaseResponse struct {
	Assets []githubReleaseAsset `json:"assets"`
}

// tryDownloadGitHubRelease attempts to download a pre-compiled binary from GitHub Releases.
func tryDownloadGitHubRelease(plugin PluginDef, binaryName, goos, goarch, githubAPIBaseURL string) ([]byte, error) {
	if githubAPIBaseURL == "" {
		githubAPIBaseURL = "https://api.github.com"
	}
	token := getAuthToken()

	// Extract owner and repo from moduleURI (e.g. github.com/owner/repo)
	parts := strings.Split(strings.TrimPrefix(plugin.ModuleURI, "github.com/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("not a github repository: %s", plugin.ModuleURI)
	}
	owner, repo := parts[0], parts[1]

	ref := plugin.GitRef
	if ref == "" {
		return nil, fmt.Errorf("empty gitRef, release download skipped")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", githubAPIBaseURL, owner, repo, ref)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release lookup HTTP %d", resp.StatusCode)
	}

	var release githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}

	// Candidates for asset naming
	expectedNames := map[string]bool{
		fmt.Sprintf("%s_%s_%s", binaryName, goos, goarch): true,
		fmt.Sprintf("%s-%s-%s", binaryName, goos, goarch): true,
		binaryName: true,
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if expectedNames[asset.Name] {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no matching release asset found for binary %s (%s/%s)", binaryName, goos, goarch)
	}

	// Download asset
	dlReq, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	if token != "" && !strings.Contains(downloadURL, "githubusercontent.com") && !strings.Contains(downloadURL, "amazonaws.com") {
		dlReq.Header.Set("Authorization", "Bearer "+token)
	}

	dlResp, err := client.Do(dlReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dlResp.Body.Close() }()

	if dlResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("asset download HTTP %d", dlResp.StatusCode)
	}

	return io.ReadAll(dlResp.Body)
}

// copyFile helper copies content from src file to dst file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}

// downloadAndInstallPluginWithCache handles binary resolution across 3 tiers:
// Tier 1: Local Disk Cache
// Tier 2: GitHub Release Asset / BinaryURL
// Tier 3: Source compilation fallback
func downloadAndInstallPluginWithCache(pluginType string, pluginIdx int, plugin PluginDef, defaults DefaultsConfig, cacheDir string, githubAPIBaseURL string) error {
	pluginKey := fmt.Sprintf("%s[%d]", pluginType, pluginIdx)

	if !isPluginEnabled(plugin) {
		log.Printf("%s - skipping disabled plugin", pluginKey)
		return nil
	}
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("%s - plugin input validation failed: %w", pluginKey, err)
	}

	resolvedCacheDir, err := resolveCacheDir(cacheDir)
	if err != nil {
		return fmt.Errorf("%s - failed to resolve cache dir: %w", pluginKey, err)
	}

	goos := os.Getenv("GOOS")
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := os.Getenv("GOARCH")
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	cacheKey := computeCacheKey(pluginType, plugin, defaults, goos, goarch)
	cachedBinaryPath := filepath.Join(resolvedCacheDir, cacheKey)

	outputDir := os.Getenv("GOBIN")
	if outputDir == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}
		outputDir = filepath.Join(gopath, "bin")
	}

	// Compute binary name
	isLocal := filepath.IsAbs(plugin.ModuleURI) || strings.HasPrefix(plugin.ModuleURI, "."+string(filepath.Separator))
	installArg := determineInstallArg(plugin.InstallPath, plugin.ModuleURI, isLocal)
	binaryName := filepath.Base(installArg)
	if binaryName == "." {
		binaryName = filepath.Base(filepath.Clean(plugin.ModuleURI))
	}
	outputPath := filepath.Join(outputDir, binaryName)

	// Tier 1: Check Local Disk Cache Hit
	if _, err := os.Stat(cachedBinaryPath); err == nil {
		log.Printf("%s - local disk cache hit (%s)", pluginKey, cacheKey)
		if err := copyFile(cachedBinaryPath, outputPath); err != nil {
			return fmt.Errorf("%s - failed to copy cached binary to target: %w", pluginKey, err)
		}
		return nil
	}

	// Tier 2: Try explicit binary URL override, then GitHub Release Asset fallback
	if plugin.BinaryURL != "" {
		log.Printf("%s - cache miss, attempting binaryURL download (%s)", pluginKey, plugin.BinaryURL)

		client := &http.Client{Timeout: 15 * time.Second}
		req, err := http.NewRequest("GET", plugin.BinaryURL, nil)
		if err == nil {
			if token := getAuthToken(); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			resp, err := client.Do(req)
			if err == nil {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode == http.StatusOK {
					assetData, err := io.ReadAll(resp.Body)
					if err == nil {
						log.Printf("%s - binaryURL downloaded successfully", pluginKey)
						if err := os.MkdirAll(resolvedCacheDir, 0755); err != nil {
							return fmt.Errorf("%s - failed to create cache dir: %w", pluginKey, err)
						}
						if err := os.WriteFile(cachedBinaryPath, assetData, 0755); err != nil {
							return fmt.Errorf("%s - failed to write binary to cache: %w", pluginKey, err)
						}
						if err := copyFile(cachedBinaryPath, outputPath); err != nil {
							return fmt.Errorf("%s - failed to copy binary to output: %w", pluginKey, err)
						}
						return nil
					}
				}
			}
		}
	}

	if strings.HasPrefix(plugin.ModuleURI, "github.com/") && plugin.GitRef != "" {
		log.Printf("%s - cache miss, attempting release asset download for %s@%s", pluginKey, plugin.ModuleURI, plugin.GitRef)
		assetData, relErr := tryDownloadGitHubRelease(plugin, binaryName, goos, goarch, githubAPIBaseURL)
		if relErr == nil {
			log.Printf("%s - release asset downloaded successfully", pluginKey)
			if err := os.MkdirAll(resolvedCacheDir, 0755); err != nil {
				return fmt.Errorf("%s - failed to create cache dir: %w", pluginKey, err)
			}
			if err := os.WriteFile(cachedBinaryPath, assetData, 0755); err != nil {
				return fmt.Errorf("%s - failed to write binary to cache: %w", pluginKey, err)
			}
			if err := copyFile(cachedBinaryPath, outputPath); err != nil {
				return fmt.Errorf("%s - failed to copy binary to output: %w", pluginKey, err)
			}
			return nil
		}
		log.Printf("%s - release asset download not available (%v), falling back to source build", pluginKey, relErr)
	}

	// Tier 3: Source Build Fallback
	if err := downloadAndInstallPlugin(pluginType, pluginIdx, plugin, defaults); err != nil {
		return err
	}

	// Write compiled output to cache dir for future runs
	if _, err := os.Stat(outputPath); err == nil {
		if err := os.MkdirAll(resolvedCacheDir, 0755); err == nil {
			if err := copyFile(outputPath, cachedBinaryPath); err != nil {
				log.Printf("%s - warning: failed to write compiled binary to cache: %v", pluginKey, err)
			}
		}
	}

	return nil
}
