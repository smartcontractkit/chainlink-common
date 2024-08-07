package loop_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/go-plugin"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	cciptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip/test"
	reportingplugintest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestCommitService(t *testing.T) {
	t.Parallel()

	commit := loop.NewCommitService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return NewHelperProcessCommand(loop.CCIPCommitLOOPName, false, 0)
	}, cciptest.CommitProvider)

	t.Run("service not nil", func(t *testing.T) {
		require.NotPanics(t, func() { commit.Name() })
	})

	hook := commit.PluginService.XXXTestHook()
	servicetest.Run(t, commit)

	t.Run("control", func(t *testing.T) {
		reportingplugintest.RunFactory(t, commit)
	})

	t.Run("Kill", func(t *testing.T) {
		hook.Kill()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		reportingplugintest.RunFactory(t, commit)
	})

	t.Run("Reset", func(t *testing.T) {
		hook.Reset()

		// wait for relaunch
		time.Sleep(2 * goplugin.KeepAliveTickDuration)

		reportingplugintest.RunFactory(t, commit)
	})
}

func TestCommitService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	commit := loop.NewCommitService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		h := HelperProcessCommand{
			Command: loop.CCIPCommitLOOPName,
			Limit:   int(limit.Add(1)),
		}
		return h.New()
	}, cciptest.CommitProvider)
	servicetest.Run(t, commit)

	reportingplugintest.RunFactory(t, commit)
}

func TestCommitLOOP(t *testing.T) {
	// launch the commit loop via the main program
	t.Parallel()
	stopCh := newStopCh(t)
	commit := loop.CommitLoop{BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}
	cc := commit.ClientConfig()
	cc.Cmd = NewHelperProcessCommand(loop.CCIPCommitLOOPName, false, 0)
	c := plugin.NewClient(cc)
	// make sure to kill the commit loop
	t.Cleanup(c.Kill)
	client, err := c.Client()
	require.NoError(t, err)
	defer client.Close()
	require.NoError(t, client.Ping())
	// get a concrete instance of the commit loop
	instance, err := client.Dispense(loop.CCIPCommitLOOPName)
	remoteCommitFactory := instance.(types.CCIPCommitFactoryGenerator)
	require.NoError(t, err)

	cciptest.RunCommitLOOP(t, remoteCommitFactory)

	t.Run("proxy: commit loop <--> relayer loop", func(t *testing.T) {
		// launch the relayer as external process via the main program
		pr := newPluginRelayerCommit(t, false, stopCh)
		remoteProvider, err := newCommitProvider(t, pr)
		require.Error(t, err, "expected error")
		assert.Contains(t, err.Error(), "BCF-3061")
		if err == nil {
			// test to run when BCF-3061 is fixed
			cciptest.CommitLOOPTester{CCIPCommitProvider: remoteProvider}.Run(t, remoteCommitFactory)
		}
	})
}

func newCommitProvider(t *testing.T, pr loop.PluginRelayer) (types.CCIPCommitProvider, error) {
	ctx := tests.Context(t)
	r, err := pr.NewRelayer(ctx, test.ConfigTOML, keystoretest.Keystore, nil)
	require.NoError(t, err)
	servicetest.Run(t, r)

	// TODO: fix BCF-3061. we expect an error here until then.
	p, err := r.NewPluginProvider(ctx, cciptest.CommitRelayArgs, cciptest.CommitPluginArgs)
	if err != nil {
		return nil, err
	}
	// TODO: this shouldn't run until BCF-3061 is fixed
	require.NoError(t, err)
	commitProvider, ok := p.(types.CCIPCommitProvider)
	require.True(t, ok, "got %T", p)
	servicetest.Run(t, commitProvider)
	return commitProvider, nil
}
