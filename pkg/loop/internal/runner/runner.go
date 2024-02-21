package runner

import (
	"context"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/runner"
)

type logMessage struct {
	Level     string    `json:"@level"`
	Timestamp time.Time `json:"@timestamp"`
	Logger    string    `json:"logger"`
	Caller    string    `json:"caller"`
	Message   string    `json:"@message"`
}

var _ runner.Runner = (*Runner)(nil)

type Runner struct {
	logger  hclog.Logger
	cmd     *exec.Cmd
	cmdPath string
	cmdArgs []string
	cmdEnv  []string

	stdout io.ReadCloser
	stderr io.ReadCloser

	pid int
}

var _ io.ReadCloser = (*stderrWrapper)(nil)

type stderrWrapper struct {
	r    io.ReadCloser
	lggr hclog.Logger
}

func NewIOWrapper(l hclog.Logger, r io.ReadCloser) *stderrWrapper {
	return &stderrWrapper{
		r:    r,
		lggr: l,
	}
}
func (i *stderrWrapper) logIfPanic(m []byte) {
	//Normal logs are pushed in here as well, to distinguish them from panics we try to unmarshall them as json
	l := &logMessage{}
	err := json.Unmarshal(m[:len(m)-1], l) //Remove trailing newline
	if err != nil {
		i.lggr.Error(string(m))
	}
}

func (i *stderrWrapper) Read(p []byte) (n int, err error) {
	n, err = i.r.Read(p)
	if err != nil {
		return 0, nil
	}
	i.logIfPanic(p[:n])
	return
}

func (i *stderrWrapper) Close() error {
	i.r.Close()
	return nil
}

func NewRunnerFunc(path string, args, env []string) func(l hclog.Logger, cmd *exec.Cmd, tmpDir string) (runner.Runner, error) {
	cr := &Runner{
		cmdPath: path,
		cmdArgs: args,
		cmdEnv:  env,
	}
	return cr.NewCmdRunner
}

func (c *Runner) NewCmdRunner(logger hclog.Logger, cmd *exec.Cmd, tmpDir string) (runner.Runner, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	c.logger = logger
	c.cmd = cmd
	c.cmd.Path = c.cmdPath
	c.cmd.Args = c.cmdArgs
	c.cmd.Env = append(c.cmd.Env, c.cmdEnv...)
	c.stdout = stdout
	c.stderr = NewIOWrapper(logger, stderr)

	return c, nil
}

func (c *Runner) Start(ctx context.Context) error {
	c.logger.Debug("starting plugin", "path", c.cmd.Path, "args", c.cmd.Args)
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	c.pid = c.cmd.Process.Pid
	c.logger.Debug("plugin started", "path", c.cmdPath, "pid", c.pid)
	return nil
}

func (c *Runner) Diagnose(ctx context.Context) string {
	return fmt.Sprintf(`This usually means   the plugin was not compiled for this architecture,   the plugin is missing dynamic-link libraries necessary to run,   the plugin is not executable by this process due to file permissions, or   the plugin failed to negotiate the initial go-plugin protocol handshake %s`, additionalNotesAboutCommand(c.cmd.Path))
}

func (c *Runner) Stdout() io.ReadCloser {
	io.MultiReader()
	return c.stdout
}

func (c *Runner) Stderr() io.ReadCloser {
	return c.stderr
}

func (c *Runner) Name() string {
	return c.cmdPath
}

func (c *Runner) Wait(ctx context.Context) error {
	return c.cmd.Wait()
}

func (c *Runner) Kill(ctx context.Context) error {
	if c.cmd.Process != nil {
		err := c.cmd.Process.Kill()
		// Swallow ErrProcessDone, we support calling Kill multiple times.
		if !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return nil
	}

	return nil
}

func (c *Runner) ID() string {
	return fmt.Sprintf("%d", c.pid)
}

func (c *Runner) PluginToHost(pluginNet, pluginAddr string) (hostNet string, hostAddr string, err error) {
	return pluginNet, pluginAddr, nil
}

func (c *Runner) HostToPlugin(hostNet, hostAddr string) (pluginNet string, pluginAddr string, err error) {
	return hostNet, hostAddr, nil
}

var peTypes = map[uint16]string{
	0x14c:  "386",
	0x1c0:  "arm",
	0x6264: "loong64",
	0x8664: "amd64",
	0xaa64: "arm64",
}

func additionalNotesAboutCommand(path string) string {
	notes := ""
	stat, err := os.Stat(path)
	if err != nil {
		return notes
	}

	notes += "\nAdditional notes about plugin:\n"
	notes += fmt.Sprintf("  Path: %s\n", path)
	notes += fmt.Sprintf("  Mode: %s\n", stat.Mode())
	statT, ok := stat.Sys().(*syscall.Stat_t)
	if ok {
		currentUsername := "?"
		if u, err := user.LookupId(strconv.FormatUint(uint64(os.Getuid()), 10)); err == nil {
			currentUsername = u.Username
		}
		currentGroup := "?"
		if g, err := user.LookupGroupId(strconv.FormatUint(uint64(os.Getgid()), 10)); err == nil {
			currentGroup = g.Name
		}
		username := "?"
		if u, err := user.LookupId(strconv.FormatUint(uint64(statT.Uid), 10)); err == nil {
			username = u.Username
		}
		group := "?"
		if g, err := user.LookupGroupId(strconv.FormatUint(uint64(statT.Gid), 10)); err == nil {
			group = g.Name
		}
		notes += fmt.Sprintf("  Owner: %d [%s] (current: %d [%s])\n", statT.Uid, username, os.Getuid(), currentUsername)
		notes += fmt.Sprintf("  Group: %d [%s] (current: %d [%s])\n", statT.Gid, group, os.Getgid(), currentGroup)
	}

	if elfFile, err := elf.Open(path); err == nil {
		defer elfFile.Close()
		notes += fmt.Sprintf("  ELF architecture: %s (current architecture: %s)\n", elfFile.Machine, runtime.GOARCH)
	} else if machoFile, err := macho.Open(path); err == nil {
		defer machoFile.Close()
		notes += fmt.Sprintf("  MachO architecture: %s (current architecture: %s)\n", machoFile.Cpu, runtime.GOARCH)
	} else if peFile, err := pe.Open(path); err == nil {
		defer peFile.Close()
		machine, ok := peTypes[peFile.Machine]
		if !ok {
			machine = "unknown"
		}
		notes += fmt.Sprintf("  PE architecture: %s (current architecture: %s)\n", machine, runtime.GOARCH)
	}
	return notes
}
