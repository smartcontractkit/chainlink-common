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

func CommitFactoryServer(lggr logger.Logger) commitFactoryServer {
	return newCommitFactoryServer(lggr, commitFactoryServerConfig{
		provider: CommitProvider(lggr),
	})
}

type commitFactoryServerConfig struct {
	provider CommitProviderEvaluator
}

var _ types.CCIPCommitFactoryGenerator = commitFactoryServer{}

type commitFactoryServer struct {
	services.Service
	commitFactoryServerConfig
	factory types.ReportingPluginFactory
}

func newCommitFactoryServer(lggr logger.Logger, cfg commitFactoryServerConfig) commitFactoryServer {
	lggr = logger.Named(lggr, "commitFactoryServer")
	return commitFactoryServer{
		Service:                   test.NewStaticService(lggr),
		commitFactoryServerConfig: cfg,
		factory:                   reportingplugintest.Factory(lggr),
	}
}

// NewCommitFactory implements types.CCIPCommitFactoryGenerator.
func (e commitFactoryServer) NewCommitFactory(ctx context.Context, provider types.CCIPCommitProvider) (types.ReportingPluginFactory, error) {
	err := e.provider.Evaluate(ctx, provider)
	if err != nil {
		return nil, err
	}
	return e.factory, nil
}

func RunCommitLOOP(t *testing.T, p types.CCIPCommitFactoryGenerator) {
	CommitLOOPTester{CommitProvider(logger.Test(t))}.Run(t, p)
}

type CommitLOOPTester struct {
	types.CCIPCommitProvider
}

func (e CommitLOOPTester) Run(t *testing.T, p types.CCIPCommitFactoryGenerator) {
	t.Run("CommitLOOP", func(t *testing.T) {
		ctx := tests.Context(t)
		factory, err := p.NewCommitFactory(ctx, e.CCIPCommitProvider)
		require.NoError(t, err)

		runCommitReportingPluginFactory(t, factory)
	})
}

func runCommitReportingPluginFactory(t *testing.T, factory types.ReportingPluginFactory) {
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
		ctx := tests.Context(t)
		// we expect the static implementation to be used under the covers
		// we can't compare the types directly because the returned reporting plugin may be a grpc client
		// that wraps the static implementation
		var expectedReportingPlugin = reportingplugintest.ReportingPlugin

		rp, gotRPI, err := factory.NewReportingPlugin(ctx, reportingplugintest.StaticFactoryConfig.ReportingPluginConfig)
		require.NoError(t, err)
		assert.Equal(t, rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })

		t.Run("ReportingPlugin", func(t *testing.T) {
			ctx := tests.Context(t)

			expectedReportingPlugin.AssertEqual(ctx, t, rp)
		})
	})
}
