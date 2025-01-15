package ocr3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	errorlogtest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog/test"
	keyvaluestoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keyvalue/test"
	pipelinetest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/pipeline/test"
	relayersettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/relayerset/test"
	ocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3/test"
	telemetrytest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/telemetry/test"
	nettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
	reportingplugintest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/test"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func newStopCh(t *testing.T) <-chan struct{} {
	stopCh := make(chan struct{})
	if d, ok := t.Deadline(); ok {
		time.AfterFunc(time.Until(d), func() { close(stopCh) })
	}
	return stopCh
}

func PluginGenericTest(t *testing.T, p core.OCR3ReportingPluginClient) {
	t.Run("PluginServer", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewReportingPluginFactory(ctx,
			core.ReportingPluginServiceConfig{},
			nettest.MockConn{},
			pipelinetest.PipelineRunner,
			telemetrytest.Telemetry,
			errorlogtest.ErrorLog,
			core.CapabilitiesRegistry(nil),
			keyvaluestoretest.KeyValueStore{},
			relayersettest.RelayerSet{})
		require.NoError(t, err)

		ocr3test.OCR3ReportingPluginFactory(t, factory)
	})
	t.Run("ValidationService", func(t *testing.T) {
		ctx := tests.Context(t)
		validationService, err := p.NewValidationService(ctx)
		require.NoError(t, err)

		reportingplugintest.RunValidation(t, validationService)
	})
}

func TestGRPCService_MedianProvider(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		ocr3test.OCR3ReportingPluginWithMedianProviderName,
		&GRPCService[types.MedianProvider]{
			PluginServer: ocr3test.MedianServer(lggr),
			BrokerConfig: loop.BrokerConfig{
				Logger: lggr,
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}

func TestGRPCService_PluginProvider(t *testing.T) {
	t.Parallel()

	lggr := logger.Test(t)
	stopCh := newStopCh(t)
	test.PluginTest(
		t,
		PluginServiceName,
		&GRPCService[types.PluginProvider]{
			PluginServer: ocr3test.AgnosticPluginServer(lggr),
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh,
			},
		},
		PluginGenericTest,
	)
}
