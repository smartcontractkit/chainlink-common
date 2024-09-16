package aggregators_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func TestDataFeedsAggregator_Aggregate_TwoRounds(t *testing.T) {
	config := getConfig(t)
	agg, err := aggregators.NewIdenticalAggregator(*config, logger.Nop())
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a")},
		1: {values.NewString("a")},
		2: {values.NewString("a")},
		3: {values.NewString("a")},
	}
	outcome, err := agg.Aggregate(nil, observations, 1)
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)
}

func getConfig(t *testing.T) *values.Map {
	unwrappedConfig := map[string]any{
		"expectedObservationsLen": 1,
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
