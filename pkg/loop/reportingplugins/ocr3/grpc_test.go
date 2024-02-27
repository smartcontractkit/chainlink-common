package ocr3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	ocr3_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr3"
	pipeline_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/pipeline"
	telemetry_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/resources/telemetry"
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

func PluginGenericTest(t *testing.T, p types.OCR3ReportingPluginClient) {
	t.Run("PluginServer", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewReportingPluginFactory(ctx,
			types.ReportingPluginServiceConfig{},
			test.MockConn{},
			pipeline_test.PipelineRunnerImpl,
			telemetry_test.TelemetryImpl,
			&test.StaticErrorLog{},
			types.CapabilitiesRegistry(nil))
		require.NoError(t, err)

		ocr3_test.OCR3ReportingPluginFactory(t, factory)
	})
}

func TestGRPCService_MedianProvider(t *testing.T) {
	t.Parallel()

	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		test.ReportingPluginWithMedianProviderName,
		&GRPCService[types.MedianProvider]{
			PluginServer: ocr3_test.MedianGeneratorImpl,
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
		PluginServiceName,
		&GRPCService[types.PluginProvider]{
			PluginServer: ocr3_test.AgnosticPluginGeneratorImpl,
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}
