package reportingplugins_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	resources_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources"
	coreapi_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/core_api"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
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
		factory, err := p.NewReportingPluginFactory(ctx,
			types.ReportingPluginServiceConfig{},
			resources_test.MockConn{},
			resources_test.PipelineRunnerImpl,
			resources_test.TelemetryImpl,
			&resources_test.ErrorLogImpl)
		require.NoError(t, err)

		reportingplugin_test.Factory(t, factory)
	})
}

func TestGRPCService_MedianProvider(t *testing.T) {
	t.Parallel()

	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		coreapi_test.MedianID,
		&reportingplugins.GRPCService[types.MedianProvider]{
			PluginServer: coreapi_test.MedianProviderServerImpl,
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
		reportingplugins.PluginServiceName,
		&reportingplugins.GRPCService[types.PluginProvider]{
			PluginServer: coreapi_test.AgnosticProviderServerImpl,
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}
