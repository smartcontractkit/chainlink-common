package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- helpers ---

// withMockExec temporarily swaps execCommand and restores it after the test.
func withMockExec(t *testing.T, f func(cmd *exec.Cmd) error, body func()) {
	t.Helper()
	orig := execCommand
	execCommand = f
	defer func() { execCommand = orig }()
	body()
}

// normalize slashes for stable asserts across platforms.
func toSlash(p string) string { return filepath.ToSlash(p) }

// --- determineModuleDirectoryLocal ---

func TestDetermineModuleDirectoryLocal(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string // returns the path to test
		wantErrSub string                    // substring expected in error, or empty for success
	}{
		{
			name: "success_directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErrSub: "",
		},
		{
			name: "not_a_directory",
			setup: func(t *testing.T) string {
				td := t.TempDir()
				fp := filepath.Join(td, "file.txt")
				if err := os.WriteFile(fp, []byte("hi"), 0o600); err != nil {
					t.Fatal(err)
				}
				return fp
			},
			wantErrSub: "is not a directory",
		},
		{
			name: "not_accessible",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist-foo-bar-baz")
			},
			wantErrSub: "not accessible",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup(t)
			got, err := determineModuleDirectoryLocal("plugin[0]", path)

			// error cases
			if err != nil {
				if tc.wantErrSub == "" {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tc.wantErrSub) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrSub, err)
				}
				return
			}

			// success case
			want, _ := filepath.Abs(path)
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

// --- determineModuleDirectoryRemote ---

func TestDetermineModuleDirectoryRemote_Success(t *testing.T) {
	wantDir := filepath.Join(t.TempDir(), "gomodcache", "module")
	mod := "github.com/acme/thing"
	ref := "v1.2.3"

	withMockExec(t, func(cmd *exec.Cmd) error {
		// Basic command shape
		if len(cmd.Args) < 5 ||
			cmd.Args[0] != "go" ||
			cmd.Args[1] != "mod" ||
			cmd.Args[2] != "download" ||
			cmd.Args[3] != "-json" ||
			cmd.Args[4] != fmt.Sprintf("%s@%s", mod, ref) {
			t.Fatalf("unexpected args: %v", cmd.Args)
		}
		// GOPRIVATE propagates only if provided
		found := false
		for _, e := range cmd.Env {
			if strings.HasPrefix(e, "GOPRIVATE=") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected GOPRIVATE in env")
		}

		// Simulate go's JSON
		type dl struct{ Dir string }
		enc := json.NewEncoder(cmd.Stdout)
		_ = enc.Encode(dl{Dir: wantDir})
		return nil
	}, func() {
		got, err := determineModuleDirectoryRemote("plugin[0]", mod, ref, "github.com/private/*")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != wantDir {
			t.Fatalf("got %q, want %q", got, wantDir)
		}
	})
}

func TestDetermineModuleDirectoryRemote_EmptyDir(t *testing.T) {
	withMockExec(t, func(cmd *exec.Cmd) error {
		// write empty object (no Dir)
		_, _ = fmt.Fprint(cmd.Stdout, `{}`)
		return nil
	}, func() {
		_, err := determineModuleDirectoryRemote("plugin[0]", "github.com/acme/thing", "main", "")
		if err == nil || !strings.Contains(err.Error(), "empty module directory") {
			t.Fatalf("expected empty module directory error, got %v", err)
		}
	})
}

func TestDetermineModuleDirectoryRemote_CommandError(t *testing.T) {
	withMockExec(t, func(cmd *exec.Cmd) error {
		return errors.New("boom")
	}, func() {
		_, err := determineModuleDirectoryRemote("plugin[0]", "github.com/acme/thing", "main", "")
		if err == nil || !strings.Contains(err.Error(), "failed to download module") {
			t.Fatalf("expected download failure, got %v", err)
		}
	})
}

func TestDetermineModuleDirectoryRemote_InvalidJSON(t *testing.T) {
	withMockExec(t, func(cmd *exec.Cmd) error {
		_, _ = fmt.Fprint(cmd.Stdout, `not-json`)
		return nil
	}, func() {
		_, err := determineModuleDirectoryRemote("plugin[0]", "github.com/acme/thing", "main", "")
		if err == nil || !strings.Contains(err.Error(), "failed to parse") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

// --- determineInstallArg ---

func TestDetermineInstallArg_LocalModule(t *testing.T) {
	// Use an absolute module root to exercise local path handling
	modRoot := filepath.Clean(t.TempDir())

	tests := []struct {
		name        string
		installPath string
		want        string
	}{
		{"root equals module", modRoot, "."},
		{"dot", ".", "."},
		{"subdir in module", filepath.Join(modRoot, "cmd", "tool"), "./cmd/tool"},
		{"relative subdir", "cmd/tool", "./cmd/tool"},
		{
			"absolute outside module becomes relative-ish",
			func() string {
				// a path that is surely outside modRoot
				base := string(os.PathSeparator) + "opt" + string(os.PathSeparator) + "other"
				return filepath.Clean(base)
			}(),
			func() string {
				abs := func() string {
					base := string(os.PathSeparator) + "opt" + string(os.PathSeparator) + "other"
					return filepath.Clean(base)
				}()
				return "./" + toSlash(strings.TrimLeft(abs, string(os.PathSeparator)))
			}(),
		},
		{"already ./ prefixed", "./cmd/tool", "./cmd/tool"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := determineInstallArg(tc.installPath, modRoot, true)
			if toSlash(got) != toSlash(tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetermineInstallArg_RemoteModule(t *testing.T) {
	module := "github.com/acme/repo"

	tests := []struct {
		name        string
		installPath string
		want        string
	}{
		{"root equals module", module, "."},
		{"subpackage absolute import", module + "/sub/pkg", "./sub/pkg"},
		{"dot", ".", "."},
		{"plain relative", "sub/pkg", "./sub/pkg"},
		{"already ./ prefixed", "./sub/pkg", "./sub/pkg"},
		{"normalize leading slash", "/sub/pkg", "./sub/pkg"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := determineInstallArg(tc.installPath, module, false)
			if toSlash(got) != toSlash(tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
