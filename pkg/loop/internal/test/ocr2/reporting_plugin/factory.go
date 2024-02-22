package reportingplugin_test

import (
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StaticFactoryConfig struct {
	Config  libocr.ReportingPluginConfig
	RPI     libocr.ReportingPluginInfo
	Factory libocr.ReportingPluginFactory
}

type StaticFactory struct {
	StaticFactoryConfig
}

func Factory(t *testing.T, s StaticFactory) {
	t.Run("ReportingPluginFactory", func(t *testing.T) {
		rp, gotRPI, err := s.Factory.NewReportingPlugin(s.Config)
		require.NoError(t, err)
		assert.Equal(t, s.RPI, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })
		t.Run("ReportingPlugin", func(t *testing.T) {
			ctx := tests.Context(t)
			gotQuery, err := rp.Query(ctx, reportContext.ReportTimestamp)
			require.NoError(t, err)
			assert.Equal(t, query, []byte(gotQuery))
			gotObs, err := rp.Observation(ctx, reportContext.ReportTimestamp, query)
			require.NoError(t, err)
			assert.Equal(t, observation, gotObs)
			gotOk, gotReport, err := rp.Report(ctx, reportContext.ReportTimestamp, query, obs)
			require.NoError(t, err)
			assert.True(t, gotOk)
			assert.Equal(t, report, gotReport)
			gotShouldAccept, err := rp.ShouldAcceptFinalizedReport(ctx, reportContext.ReportTimestamp, report)
			require.NoError(t, err)
			assert.True(t, gotShouldAccept)
			gotShouldTransmit, err := rp.ShouldTransmitAcceptedReport(ctx, reportContext.ReportTimestamp, report)
			require.NoError(t, err)
			assert.True(t, gotShouldTransmit)
		})
	})
}
