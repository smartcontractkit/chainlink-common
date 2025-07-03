package pkg

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

var values = Packages{
	Go:    "github.com/smartcontractkit/chainlink-common/pkg/values/pb",
	Proto: "values/v1/values.proto",
}

type ProtocGen struct {
	ProtocHelper
	packageNames map[string]string
	sources      []string
	init         bool
	Plugins      []Plugin
}

// LinkPackage directly links a package and does not require ProtocHelper to be set
func (p *ProtocGen) LinkPackage(pkgs Packages) {
	if p.packageNames == nil {
		p.packageNames = make(map[string]string)
	}
	p.packageNames[pkgs.Proto] = pkgs.Go
}

func (p *ProtocGen) LinkCapabilities(config *CapabilityConfig) {
	for _, file := range config.FullProtoFiles() {
		goPkg := p.FullGoPackageName(config)
		p.LinkPackage(Packages{Go: goPkg, Proto: file})
	}
}

func (p *ProtocGen) AddSourceDirectories(sources ...string) {
	p.sources = append(p.sources, sources...)
}

// GenerateFile generates a single file using protoc with the provided plugins and sources.
// Calling this method directly does not require ProtocHelper to be set.
func (p *ProtocGen) GenerateFile(file, from string) error {
	if err := p.doInit(); err != nil {
		return err
	}

	var args []string
	for _, pkg := range p.sources {
		args = append(args, "-I", pkg)
	}

	for _, plugin := range p.Plugins {
		prefix := fmt.Sprintf("%s_", plugin.Name)
		if plugin.Path != "" {
			sep := string(filepath.Separator)
			upDir := ""
			if from != "." {
				upLen := len(strings.Split(from, string([]byte{filepath.Separator})))
				upDir = strings.Repeat(".."+sep, upLen)
			}

			args = append(args, fmt.Sprintf("--plugin=protoc-gen-%s=%s%s%sprotoc-gen-%s", plugin.Name, upDir, plugin.Path, sep, plugin.Name))
		}

		args = append(args, fmt.Sprintf("--%sout=.", prefix))
		args = append(args, fmt.Sprintf("--%sopt=paths=source_relative", prefix))

		for proto, goPkg := range p.packageNames {
			args = append(args, fmt.Sprintf("--%sopt=M%s=%s", prefix, proto, goPkg))
		}
	}

	args = append(args, file)

	if out, err := run("protoc", from, args...); err != nil {
		return fmt.Errorf("failed to run protoc: %v\n%s", err, out)
	}

	return nil
}

func (p *ProtocGen) Generate(config *CapabilityConfig) error {
	return p.GenerateMany(map[string]*CapabilityConfig{".": config})
}

func (p *ProtocGen) GenerateMany(dirToConfig map[string]*CapabilityConfig) error {
	for _, config := range dirToConfig {
		p.LinkCapabilities(config)
	}

	fmt.Println("Generating capabilities")
	errMap := map[string]error{}
	for from, config := range dirToConfig {
		for _, file := range config.FullProtoFiles() {
			if err := p.GenerateFile(file, from); err != nil {
				errMap[file] = err
			}
		}
	}

	if len(errMap) > 0 {
		var errStrings []string
		for file, err := range errMap {
			if err != nil {
				errStrings = append(errStrings, fmt.Sprintf("file %s\n%v\n", file, err))
			}
		}

		return errors.New(strings.Join(errStrings, ""))
	}

	err := p.moveGeneratedFiles(dirToConfig)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProtocGen) moveGeneratedFiles(dirToConfig map[string]*CapabilityConfig) error {
	fmt.Println("Moving generated files to correct locations")
	for from, config := range dirToConfig {
		for i, file := range config.FullProtoFiles() {
			file = strings.Replace(file, ".proto", ".pb.go", 1)
			to := strings.Replace(config.Files[i], ".proto", ".pb.go", 1)
			if err := os.Rename(path.Join(from, file), path.Join(from, to)); err != nil {
				return fmt.Errorf("failed to move generated file %s: %w", file, err)
			}
		}

		if err := os.RemoveAll(path.Join(from, "capabilities")); err != nil {
			return fmt.Errorf("failed to remove capabilities directory %w", err)
		}
	}
	return nil
}

func (p *ProtocGen) doInit() error {
	if p.init {
		return nil
	}

	p.LinkPackage(values)

	if p.ProtocHelper != nil {
		p.LinkPackage(Packages{Go: p.SdkPgk(), Proto: "sdk/v1alpha/sdk.proto"})
		p.LinkPackage(Packages{Go: p.SdkPgk(), Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	}

	root, err := run("git", ".", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("failed to get root directory: %v", err)
	}

	protoDir := path.Join(root, "proto_vendor")
	if err = os.MkdirAll(protoDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create proto vendor directory: %v", err)
	}

	// Ensure that multiple generations running in parallel do not interfere with each other.
	// This can happen if both try to reset hard at once, clone the repo at once, or checkout different versions.
	if err = lockProtos(protoDir); err != nil {
		return fmt.Errorf("failed to lock protos: %v", err)
	}

	clProtos := path.Join(protoDir, "chainlink-protos")
	if err = checkoutClProtosRef(clProtos); err != nil {
		return fmt.Errorf("failed to checkout chainlink-protos: %v", err)
	}

	p.Plugins = append(p.Plugins, Plugin{Name: "go"})
	p.AddSourceDirectories(path.Join(clProtos, "cre"))
	p.init = true
	return nil
}

func run(command string, path string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	outuptBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed running command\n%s\nfrom path: %s\n%v", cmd.String(), path, err)
	}
	return strings.TrimSpace(string(outuptBytes)), nil
}

func checkoutClProtosRef(repoPath string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
		if err = cloneClProtosRepo(repoPath); err != nil {
			return fmt.Errorf("failed to clone chainlink-protos repo: %v", err)
		}
	}

	// Reset to head so that fetch and checkout won't fail
	if _, err := run("git", repoPath, "reset", "--hard", "HEAD"); err != nil {
		return fmt.Errorf("failed to reset chainlink-protos repo: %v", err)
	}

	if _, err := run("git", repoPath, "rev-parse", "--verify", "--quiet", chainlinkProtosVersion); err != nil {
		if out, err := run("git", repoPath, "fetch"); err != nil {
			return fmt.Errorf("failed to fetch: %v\nIf you're not working on the main branch, you may need to track that branch in proto_vendor/chainlink-protos, this tool will not do that for you to avoid accidental non-main commits\n%s", err, out)
		}
	}

	fmt.Println("checking out chainlink-protos version:", chainlinkProtosVersion)
	if out, err := run("git", repoPath, "checkout", chainlinkProtosVersion); err != nil {
		return fmt.Errorf("failed to checkout: %v\n%s", err, out)
	}

	return nil
}

func cloneClProtosRepo(repo string) error {
	repoGit := path.Join(repo, ".git")
	if _, err := os.Stat(repoGit); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check if chainlink-protos repo exists: %v", err)
		}

		if _, err = run("git", path.Dir(repo), "clone", "https://github.com/smartcontractkit/chainlink-protos"); err != nil {
			return fmt.Errorf("failed to clone chainlink-protos repo: %v", err)
		}
	}
	return nil
}

func lockProtos(protoRepo string) error {
	lockPath := path.Join(protoRepo, "protos.lock")

	// Open or create the lock file
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("failed to open protos.lock file: %w", err)
	}

	fmt.Println("Waiting for lock...")
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		panic(fmt.Sprintf("Failed to acquire lock: %v", err))
	}
	fmt.Println("Acquired lock")

	return nil
}
