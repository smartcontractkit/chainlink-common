package reportingplugin_test

import (
	"testing"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var FactoryImpl = StaticFactory{
	StaticFactoryConfig: StaticFactoryConfig{
		config: libocr.ReportingPluginConfig{},
		rpi: libocr.ReportingPluginInfo{
			Name: "test-impl",
		},
		reportingPlugin: ReportingPluginImpl,
	},
}

type StaticFactoryConfig struct {
	config          libocr.ReportingPluginConfig
	rpi             libocr.ReportingPluginInfo
	reportingPlugin ReportingPluginTester
}

type StaticFactory struct {
	StaticFactoryConfig
}

func Factory(t *testing.T, factory libocr.ReportingPluginFactory) {
	expectedFactory := FactoryImpl
	ctx := tests.Context(t)
	t.Run("ReportingPluginFactory", func(t *testing.T) {
		rp, gotRPI, err := factory.NewReportingPlugin(expectedFactory.config)
		require.NoError(t, err)
		assert.Equal(t, expectedFactory.rpi, gotRPI)
		t.Cleanup(func() { assert.NoError(t, rp.Close()) })
		t.Run("ReportingPlugin", func(t *testing.T) {
			expectedFactory.reportingPlugin.AssertEqual(t, ctx, rp)
		})
	})
}
