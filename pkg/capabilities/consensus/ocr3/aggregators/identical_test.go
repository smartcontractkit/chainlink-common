package aggregators_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

func TestDataFeedsAggregator_Aggregate(t *testing.T) {
	config := getConfigIdenticalAggregator(t, nil)
	agg, err := aggregators.NewIdenticalAggregator(*config)
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a")},
		1: {values.NewString("a")},
		2: {values.NewString("a")},
		3: {values.NewString("a")},
	}
	outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)
	require.Equal(t, "", outcome.EncoderName)
	require.Nil(t, outcome.EncoderConfig)

	m, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	require.Len(t, m.Underlying, 1)
	require.Equal(t, m.Underlying["0"], values.NewString("a"))
}

func TestDataFeedsAggregator_Aggregate_OverrideWithKeys(t *testing.T) {
	config := getConfigIdenticalAggregator(t, []string{"outcome"})
	agg, err := aggregators.NewIdenticalAggregator(*config)
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a")},
		1: {values.NewString("a")},
		2: {values.NewString("a")},
		3: {values.NewString("a")},
	}
	outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)
	require.Equal(t, "", outcome.EncoderName)
	require.Nil(t, outcome.EncoderConfig)

	m, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	require.Len(t, m.Underlying, 1)
	require.Equal(t, m.Underlying["outcome"], values.NewString("a"))
}

func TestDataFeedsAggregator_Aggregate_NoConsensus(t *testing.T) {
	config := getConfigIdenticalAggregator(t, []string{"outcome"})
	agg, err := aggregators.NewIdenticalAggregator(*config)
	require.NoError(t, err)

	encoderStr := "evm"
	encoderName := values.NewString(encoderStr)
	encoderCfg, err := values.NewMap(map[string]any{"foo": "bar"})
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a"), encoderName, encoderCfg},
		1: {values.NewString("b"), encoderName, encoderCfg},
		2: {values.NewString("b"), encoderName, encoderCfg},
		3: {values.NewString("a"), encoderName, encoderCfg},
	}
	outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
	require.Nil(t, outcome)
	require.ErrorContains(t, err, "consensus failed: cannot reach agreement on observation at index 0")
}

func getConfigIdenticalAggregator(t *testing.T, overrideKeys []string) *values.Map {
	unwrappedConfig := map[string]any{
		"expectedObservationsLen": len(overrideKeys),
		"keyOverrides":            overrideKeys,
	}

	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
