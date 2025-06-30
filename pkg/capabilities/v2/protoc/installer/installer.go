package installer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// installProtocGenToDir installs the pkg plugin to .tools. It'll download it from the same commit as sameVersion.
func installProtocGenToDir(pkgName, sameVersion string) error {
	fmt.Printf("Finding version to use for %s\n.", pkgName)
	pluginVersion, err := getVersion(sameVersion, ".")
	if err != nil {
		return err
	}

	pluginDir, err := downloadPlugin(pkgName, pluginVersion)
	if err != nil {
		return err
	}

	if err = verifyValuesVersionsMatch(pluginDir); err != nil {
		return err
	}

	return buildPlugin(pluginDir)
}

func verifyValuesVersionsMatch(pluginDir string) error {
	fmt.Printf("Verifying values versions match")
	pluginValues, err := getVersion("github.com/smartcontractkit/chainlink-common/pkg/values", pluginDir)
	if err != nil {
		return err
	}
	myValues, err := getVersion("github.com/smartcontractkit/chainlink-common/pkg/values", ".")
	if err != nil {
		return err
	}

	if pluginValues != myValues {
		return fmt.Errorf("version mismatch: plugin uses %s, but current project uses %s", pluginValues, myValues)
	}

	return nil
}

func getVersion(of, dir string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Version}}", of)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get version of %s in directory %s: %w\nOutput: %s", of, dir, err, out)
	}

	return strings.TrimSpace(string(out)), nil
}

func downloadPlugin(pkgName, version string) (string, error) {
	fmt.Printf("Downloading plugin version %s\n", version)
	cmd := exec.Command("go", "mod", "download", "-json", fmt.Sprintf("%s@%s", pkgName, version))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to download module: %w\nOutput: %s", err, out)
	}

	var mod struct{ Dir string }
	if err = json.Unmarshal(out, &mod); err != nil {
		return "", fmt.Errorf("failed to parse go mod download output: %w", err)
	}

	return mod.Dir, nil
}

func buildPlugin(pluginDir string) error {
	toolsDir, err := filepath.Abs(".tools")
	if err != nil {
		return fmt.Errorf("failed to get absolute path for .tools: %w", err)
	}

	fmt.Println("Building plugin")
	cmd := exec.Command("go", "build", "-o", toolsDir, ".")
	cmd.Dir = pluginDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build plugin: %w\nOutput: %s", err, out)
	}

	return nil
}
