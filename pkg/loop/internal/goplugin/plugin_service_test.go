package goplugin_test

import (
	"context"
	"fmt"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	looptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

func TestPluginService_Reset_RecreatesLogicalService(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	stopCh := make(chan struct{})

	var newServiceCalls atomic.Int32
	var serviceCloseCalls atomic.Int32

	newService := func(context.Context, any) (*testService, services.HealthReporter, error) {
		id := int(newServiceCalls.Add(1))
		svc := &testService{
			name:       fmt.Sprintf("service-%d", id),
			closeCalls: &serviceCloseCalls,
		}
		return svc, svc, nil
	}

	var service goplugin.PluginService[*loop.GRPCPluginRelayer, *testService]
	service.Init(
		loop.PluginRelayerName,
		&loop.GRPCPluginRelayer{BrokerConfig: loop.BrokerConfig{Logger: lggr, StopCh: stopCh}},
		newService,
		lggr,
		func() *exec.Cmd { return newHelperProcessCommand(loop.PluginRelayerName) },
		stopCh,
	)
	hook := service.XXXTestHook()

	require.NoError(t, service.Start(context.Background()))
	t.Cleanup(func() {
		require.NoError(t, service.Close())
	})

	require.Eventually(t, func() bool {
		return service.Ready() == nil && newServiceCalls.Load() == 1
	}, 10*time.Second, 100*time.Millisecond)

	hook.Reset()

	// After a plugin restart, PluginService is expected to build a fresh logical service and
	// close the superseded one.
	require.Eventually(t, func() bool {
		return newServiceCalls.Load() == 2
	}, 3*goplugin.KeepAliveTickDuration, 100*time.Millisecond)
	require.Eventually(t, func() bool {
		return serviceCloseCalls.Load() == 1
	}, 3*goplugin.KeepAliveTickDuration, 100*time.Millisecond)
}

type testService struct {
	name       string
	closeCalls *atomic.Int32
}

func (s *testService) Start(context.Context) error { return nil }

func (s *testService) Close() error {
	s.closeCalls.Add(1)
	return nil
}

func (s *testService) Ready() error { return nil }

func (s *testService) Name() string { return s.name }

func (s *testService) HealthReport() map[string]error {
	return map[string]error{s.name: nil}
}

func newHelperProcessCommand(command string) *exec.Cmd {
	return looptest.HelperProcessCommand{
		CommandLocation: "../test/cmd/main.go",
		Command:         command,
	}.New()
}
