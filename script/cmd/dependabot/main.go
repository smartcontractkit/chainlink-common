package main

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func main() {
	var exitCode int
	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)

	type bump struct {
		path, packageName string
	}
	bumps := map[bump]*semver.Version{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			log.Fatalf("Invalid line: must have 3 columns but found %d: %s", len(parts), line)
		}
		nv, err := semver.NewVersion("v" + parts[2])
		if err != nil {
			log.Fatalf("Invalid version: %s: %v", parts[2], err)
		}

		k := bump{parts[0], parts[1]}
		if v, ok := bumps[k]; ok {
			if v.GreaterThanEqual(nv) {
				continue
			}
		}
		bumps[k] = nv
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Fatalln("reading standard input:", err)
	}

	keys := slices.SortedFunc(maps.Keys(bumps), func(a bump, b bump) int {
		return cmp.Or(cmp.Compare(a.path, b.path), cmp.Compare(a.packageName, b.packageName))
	})
	for _, k := range keys {
		v := bumps[k]

		var buf bytes.Buffer
		listCmd := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Version}}", k.packageName)
		listCmd.Dir = k.path
		listCmd.Stdout = &buf
		if err := listCmd.Run(); err != nil {
			slog.Error("Failed to run go list", "path", k.path, "package", k.packageName, "err", err)
			exitCode = 1
			continue
		}
		curVersion, err := semver.NewVersion(strings.TrimSpace(buf.String()))
		if err != nil {
			slog.Error("Failed to parse current version", "path", k.path, "package", k.packageName, "version", buf.String(), "err", err)
			exitCode = 1
			continue
		}
		if curVersion.GreaterThanEqual(v) {
			slog.Info("Skipping", "path", k.path, "package", k.packageName, "currentVersion", curVersion, "newVersion", v)
			continue
		}

		slog.Info("Upgrading", "path", k.path, "package", k.packageName, "currentVersion", curVersion, "newVersion", v)
		cmd := exec.CommandContext(ctx, "go", "get", fmt.Sprintf("%s@v%s", k.packageName, v))
		cmd.Dir = k.path
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			slog.Error("Failed to upgrade", "path", k.path, "package", k.packageName, "version", v, "err", err)
			exitCode = 1
		}
	}
	if exitCode != 0 {
		slog.Error("Completed with errors")
	}
	os.Exit(exitCode)
}
