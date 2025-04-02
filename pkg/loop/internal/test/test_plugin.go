package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type HelperProcessCommand struct {
	Limit           int
	CommandLocation string
	Command         string
	StaticChecks    bool
	ThrowErrorType  int
}

func (h HelperProcessCommand) New() *exec.Cmd {
	cmdArgs := []string{
		"go", "run", h.CommandLocation, fmt.Sprintf("-cmd=%s", h.Command),
	}

	if h.Limit != 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-limit=%d", h.Limit))
	}

	if h.StaticChecks {
		cmdArgs = append(cmdArgs, "-static-checks")
	}

	if h.ThrowErrorType != 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-throw-error-type=%d", h.ThrowErrorType))
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...) // #nosec
	cmd.Env = os.Environ()

	return cmd
}

func PluginTest[TB testing.TB, I any](tb TB, name string, p plugin.Plugin, testFn func(TB, I)) {
	ctx, cancel := context.WithCancel(tests.Context(tb))
	defer cancel()

	ch := make(chan *plugin.ReattachConfig, 1)
	closeCh := make(chan struct{})
	go plugin.Serve(&plugin.ServeConfig{
		Test: &plugin.ServeTestConfig{
			Context:          ctx,
			ReattachConfigCh: ch,
			CloseCh:          closeCh,
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Plugins:    map[string]plugin.Plugin{name: p},
	})

	// We should get a config
	var config *plugin.ReattachConfig
	select {
	case config = <-ch:
	case <-time.After(5 * time.Second):
		tb.Fatal("should've received reattach")
	}
	require.NotNil(tb, config)

	c := plugin.NewClient(&plugin.ClientConfig{
		Reattach:    config,
		Plugins:     map[string]plugin.Plugin{name: p},
		SkipHostEnv: true,
	})
	tb.Cleanup(c.Kill)
	clientProtocol, err := c.Client()
	require.NoError(tb, err)
	defer clientProtocol.Close()
	i, err := clientProtocol.Dispense(name)
	require.NoError(tb, err)

	testFn(tb, i.(I))

	// stop plugin
	cancel()
	select {
	case <-closeCh:
	case <-time.After(5 * time.Second):
		tb.Fatal("should've stopped")
	}
	require.Error(tb, clientProtocol.Ping())
}
