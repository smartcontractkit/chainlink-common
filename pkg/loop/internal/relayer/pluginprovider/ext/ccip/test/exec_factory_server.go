package test

import (
	"context"
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	reportingplugintest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func ExecFactoryServer(lggr logger.Logger) execFactoryServer {
	return newExecFactoryServer(lggr, execFactoryServerConfig{
		provider: ExecutionProvider(lggr),
	})
}

type execFactoryServerConfig struct {
	provider ExecProviderEvaluator
}

var _ types.CCIPExecutionFactoryGenerator = execFactoryServer{}

type execFactoryServer struct {
	services.Service
	execFactoryServerConfig
	factory types.ReportingPluginFactory
}

func newExecFactoryServer(lggr logger.Logger, cfg execFactoryServerConfig) execFactoryServer {
	lggr = logger.Named(lggr, "execFactoryServer")
	return execFactoryServer{
		Service:                 test.NewStaticService(lggr),
		execFactoryServerConfig: cfg,
		factory:                 reportingplugintest.Factory(lggr),
	}
}

// NewExecutionFactory implements types.CCIPExecFactoryGenerator.
func (e execFactoryServer) NewExecutionFactory(ctx context.Context, srcProvider types.CCIPExecProvider, dstProvider types.CCIPExecProvider, srcChainID int64, dstChainID int64, sourceTokenAddress string) (types.ReportingPluginFactory, error) {
	err := e.provider.Evaluate(ctx, srcProvider)
	if err != nil {
		return nil, err
	}

	err2 := e.provider.Evaluate(ctx, dstProvider)
	if err2 != nil {
		return nil, err2
	}
	return e.factory, nil
}

func RunExecutionLOOP(t *testing.T, p types.CCIPExecutionFactoryGenerator) {
	lggr := logger.Test(t)
	ExecutionLOOPTester{SrcProvider: ExecutionProvider(lggr), DstProvider: ExecutionProvider(lggr), SrcChainID: 0, DstChainID: 0}.Run(t, p)
}

type ExecutionLOOPTester struct {
	SrcProvider        types.CCIPExecProvider
	DstProvider        types.CCIPExecProvider
	SrcChainID         int64
	DstChainID         int64
	SourceTokenAddress string
}

func (e ExecutionLOOPTester) Run(t *testing.T, p types.CCIPExecutionFactoryGenerator) {
	t.Run("ExecutionLOOP", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewExecutionFactory(ctx, e.SrcProvider, e.DstProvider, e.SrcChainID, e.DstChainID, e.SourceTokenAddress)
		require.NoError(t, err)

		runExecReportingPluginFactory(t, factory)
	})
}

func runExecReportingPluginFactory(t *testing.T, factory types.ReportingPluginFactory) {
	// TODO BCF-3068 de-dupe this with the same function in median/test/median.go
	rpi := libocr.ReportingPluginInfo{
		Name:          "test",
		UniqueReports: true,
		Limits: libocr.ReportingPluginLimits{
			MaxQueryLength:       42,
			MaxObservationLength: 13,
			MaxReportLength:      17,
		},
	}

	t.Run("ReportingPluginFactory", func(t *testing.T) {
		// we expect the static implementation to be used under the covers
		// we can't compare the types directly because the returned reporting plugin may be a grpc client
		// that wraps the static implementation
		var expectedReportingPlugin = reportingplugintest.ReportingPlugin

		rp, gotRPI, err := factory.NewReportingPlugin(tests.Context(t), reportingplugintest.StaticFactoryConfig.ReportingPluginConfig)
		require.NoError(t, err)
		assert.Equal(t, rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })

		t.Run("ReportingPlugin", func(t *testing.T) {
			ctx := tests.Context(t)

			expectedReportingPlugin.AssertEqual(ctx, t, rp)
		})
	})
}
