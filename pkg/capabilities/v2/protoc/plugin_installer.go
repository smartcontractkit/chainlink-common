package protoc

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallProtocGenToDir installs the pkg plugin to .tools. It'll download it from the same commit as sameVersion.
func InstallProtocGenToDir(pkgName, sameVersion string) error {
	fmt.Printf("Finding version to use for %s\n.", pkgName)
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Version}}", sameVersion)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get module version: %w\nOutput: %s", err, out)
	}
	version := strings.TrimSpace(string(out))

	fmt.Printf("Downloading protoc-gen-cre version %s\n", version)
	cmd = exec.Command("go", "mod", "download", "-json", fmt.Sprintf("%s@%s\n", pkgName, version))
	out, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to download module: %w\nOutput: %s", err, out)
	}

	var mod struct{ Dir string }
	if err = json.Unmarshal(out, &mod); err != nil {
		return fmt.Errorf("failed to parse go mod download output: %w", err)
	}

	absDir, err := filepath.Abs(".tools")
	if err != nil {
		return fmt.Errorf("failed to get absolute path for .tools directory: %w", err)
	}

	fmt.Println("Building plugin")
	cmd = exec.Command("go", "build", "-o", absDir, ".")
	cmd.Dir = mod.Dir
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build plugin: %w\nOutput: %s", err, out)
	}

	return nil
}
