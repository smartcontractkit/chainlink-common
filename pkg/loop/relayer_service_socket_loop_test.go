package loop_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
)

func TestRelayerService_RepeatedRecoveryKeepsCurrentPluginSocketsBounded(t *testing.T) {
	t.Parallel()

	// Use a short path under /tmp so HashiCorp go-plugin can bind Unix sockets without
	// tripping OS path-length limits.
	tmpDir, err := os.MkdirTemp("/tmp", "cl-plugin-")
	require.NoError(t, err)

	var launchCount atomic.Int32
	relayer := loop.NewRelayerService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		launchCount.Add(1)
		cmd := HelperProcessCommand{
			Command: loop.PluginRelayerName,
			// The helper plugin is configured to exit after a single gRPC call.
			// The polling loop below repeatedly calls Ready(), which consumes that one
			// allowed request and forces the next keepalive tick to relaunch the plugin.
			Limit: 1,
		}.New()
		cmd.Env = append(cmd.Env, "TMPDIR="+tmpDir)
		return cmd
	}, test.ConfigTOML, keystoretest.Keystore, keystoretest.Keystore, nil)

	stopPolling := make(chan struct{})
	pollingDone := make(chan struct{})
	var stopOnce sync.Once
	stopPollingLoop := func() {
		stopOnce.Do(func() {
			close(stopPolling)
			<-pollingDone
		})
	}

	require.NoError(t, relayer.Start(context.Background()))
	t.Cleanup(func() {
		stopPollingLoop()
		err := relayer.Close()
		if err != nil {
			require.Truef(t, isExpectedRelayerCloseError(err), "unexpected close error: %v", err)
		}
		require.NoError(t, os.RemoveAll(tmpDir))
	})

	go func() {
		defer close(pollingDone)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopPolling:
				return
			case <-ticker.C:
				// This is the failure driver for the test. Each successful call helps push the
				// helper process over its one-request limit, after which PluginService should
				// notice the dead plugin on keepalive and start a fresh process with a fresh
				// plugin socket path.
				_ = relayer.Ready()
			}
		}
	}()
	const maxCurrentPluginSockets = 2

	// Expected behavior:
	// 1. Crashes should trigger fresh plugin launches.
	// 2. Once the crash loop stops, old sockets should be cleaned up so TMPDIR settles back down to
	//    a small fixed set for the current plugin and its live broker wiring, rather than growing
	//    with every crash.
	require.Eventually(t, func() bool {
		return launchCount.Load() >= 2
	}, 5*goplugin.KeepAliveTickDuration, 100*time.Millisecond)
	stopPollingLoop()

	deadline := time.Now().Add(5 * time.Second)
	lastSockets := pluginSocketNames(t, tmpDir)
	for len(lastSockets) > maxCurrentPluginSockets && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		lastSockets = pluginSocketNames(t, tmpDir)
	}
	require.LessOrEqualf(t, len(lastSockets), maxCurrentPluginSockets, "leftover plugin sockets: %v", lastSockets)
}

func isExpectedRelayerCloseError(err error) bool {
	return errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "broker closed")
}

func pluginSocketNames(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var socketNames []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "plugin") {
			continue
		}
		socketNames = append(socketNames, filepath.Join(dir, name))
	}
	return socketNames
}
