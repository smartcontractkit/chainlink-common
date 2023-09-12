package reporting_plugins_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/reporting_plugins"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils/tests"
)

func newStopCh(t *testing.T) <-chan struct{} {
	stopCh := make(chan struct{})
	if d, ok := t.Deadline(); ok {
		time.AfterFunc(time.Until(d), func() { close(stopCh) })
	}
	return stopCh
}

func PluginGenericTest(t *testing.T, p types.ReportingPluginClient) {
	t.Run("PluginServer", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewReportingPluginFactory(ctx, types.ReportingPluginServiceConfig{}, test.MockConn{}, &test.StaticErrorLog{})
		require.NoError(t, err)

		test.TestReportingPluginFactory(t, factory)
	})
}

func TestGRPCService_MedianProvider(t *testing.T) {
	t.Parallel()

	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		PluginGenericMedian,
		&reporting_plugins.GRPCService[types.MedianProvider]{
			PluginServer: test.StaticReportingPluginWithMedianProvider{},
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}

func TestGRPCService_PluginProvider(t *testing.T) {
	t.Parallel()

	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		reporting_plugins.PluginServiceName,
		&reporting_plugins.GRPCService[types.PluginProvider]{
			PluginServer: test.StaticReportingPluginWithPluginProvider{},
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}
