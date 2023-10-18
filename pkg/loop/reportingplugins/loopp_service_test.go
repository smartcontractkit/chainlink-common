package reportingplugins_test

import (
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	utilstests "github.com/smartcontractkit/chainlink-relay/pkg/utils/tests"
)

func HelperProcess(command string, opts ...func(o *test.HelperProcessCommand)) *exec.Cmd {
	return test.NewHelperProcess("../internal/test/cmd/main.go", command, opts...)
}

func TestLOOPPService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Plugin string
	}{
		// A generic plugin with a median provider
		{Plugin: test.ReportingPluginWithMedianProviderName},
		// A generic plugin with a plugin provider
		{Plugin: reportingplugins.PluginServiceName},
	}
	for _, ts := range tests {
		looppSvc := reportingplugins.NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
			return HelperProcess(ts.Plugin)
		}, types.ReportingPluginServiceConfig{}, test.MockConn{}, &test.StaticErrorLog{})
		hook := looppSvc.XXXTestHook()
		require.NoError(t, looppSvc.Start(utilstests.Context(t)))
		t.Cleanup(func() { assert.NoError(t, looppSvc.Close()) })

		t.Run("control", func(t *testing.T) {
			test.TestReportingPluginFactory(t, looppSvc)
		})

		t.Run("Kill", func(t *testing.T) {
			hook.Kill()

			// wait for relaunch
			time.Sleep(2 * internal.KeepAliveTickDuration)

			test.TestReportingPluginFactory(t, looppSvc)
		})

		t.Run("Reset", func(t *testing.T) {
			hook.Reset()

			// wait for relaunch
			time.Sleep(2 * internal.KeepAliveTickDuration)

			test.TestReportingPluginFactory(t, looppSvc)
		})
	}
}

func TestLOOPPService_recovery(t *testing.T) {
	t.Parallel()
	var limit atomic.Int32
	looppSvc := reportingplugins.NewLOOPPService(logger.Test(t), loop.GRPCOpts{}, func() *exec.Cmd {
		return HelperProcess(test.ReportingPluginWithMedianProviderName, test.WithLimit(int(limit.Add(1))))
	}, types.ReportingPluginServiceConfig{}, test.MockConn{}, &test.StaticErrorLog{})
	require.NoError(t, looppSvc.Start(utilstests.Context(t)))
	t.Cleanup(func() { assert.NoError(t, looppSvc.Close()) })

	test.TestReportingPluginFactory(t, looppSvc)
}
