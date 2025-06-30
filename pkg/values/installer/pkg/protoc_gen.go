package pkg

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type ProtocGen struct {
	packageNames map[string]string
	sources      []string
	init         bool
	Plugins      []Plugin
}

func (p *ProtocGen) LinkPackage(pkgs Packages) {
	if p.packageNames == nil {
		p.packageNames = make(map[string]string)
	}
	p.packageNames[pkgs.Proto] = pkgs.Go
}

func (p *ProtocGen) AddSourceDirectories(sources ...string) {
	p.sources = append(p.sources, sources...)
}

func (p *ProtocGen) Generate(file, from string) error {
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

func (p *ProtocGen) GenerateMany(fileToFrom map[string]string) map[string]error {
	errors := map[string]error{}
	for file, from := range fileToFrom {
		if err := p.Generate(file, from); err != nil {
			errors[file] = err
		}
	}

	return errors
}

func (p *ProtocGen) doInit() error {
	if p.init {
		return nil
	}

	root, err := run("git", ".", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("failed to get root directory: %v", err)
	}

	protoDir := path.Join(root, "proto_vendor")
	if err = os.MkdirAll(protoDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create proto vendor directory: %v", err)
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

	// TODO check why this didn't fetch...
	if _, err := run("git", repoPath, "rev-parse", "--verify", "--quiet", chainlinkProtosVersion); err != nil {
		if out, err := run("git", repoPath, "fetch", "origin"); err != nil {
			return fmt.Errorf("failed to fetch: %v\n%s", err, out)
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
