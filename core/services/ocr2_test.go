package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median/evmreportcodec"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// can be used to generate test payloads
func TestOCR2_NumericalMedian(t *testing.T) {
	t.Parallel()

	// create plugin factory
	factory := median.NumericalMedianFactory{
		ContractTransmitter:   &test.MockMedianContract{},
		DataSource:            &test.MockDataSource{},
		JuelsPerEthDataSource: &test.MockDataSource{},
		Logger:                test.MockOCRLogger,
		ReportCodec:           evmreportcodec.ReportCodec{},
	}

	// get onchainConfig
	onchainConfig, err := test.MockOnchainConfig()
	require.NoError(t, err)

	// create new plugin
	plugin, _, err := factory.NewReportingPlugin(types.ReportingPluginConfig{
		OnchainConfig: onchainConfig,
	})
	require.NoError(t, err)

	// create timestamp
	timestamp := types.ReportTimestamp{
		ConfigDigest: test.MockDigest,
		Epoch:        test.MockEpoch,
		Round:        test.MockRound,
	}

	// fetch observation from mock
	observation, err := plugin.Observation(context.TODO(), timestamp, []byte{})
	require.NoError(t, err)

	// copy observation across X nodes
	observations := []types.AttributedObservation{}
	for i := 0; i < 4; i++ {
		observations = append(observations, types.AttributedObservation{observation, commontypes.OracleID(i)})
	}

	// generate report from observations
	ok, report, err := plugin.Report(context.TODO(), timestamp, []byte{}, observations)
	require.NoError(t, err)
	assert.True(t, ok)

	fmt.Println("Report payload", report)
}
