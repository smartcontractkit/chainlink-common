package aggregators_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func TestDataFeedsAggregator_Aggregate(t *testing.T) {
	config := getConfig(t, false, nil)
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
	require.Equal(t, "", outcome.EncoderName)
	require.Nil(t, outcome.EncoderConfig)

	m, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	require.Len(t, m.Underlying, 1)
	require.Equal(t, m.Underlying["0"], values.NewString("a"))
}

func TestDataFeedsAggregator_Aggregate_OverrideEncoder(t *testing.T) {
	config := getConfig(t, true, nil)
	agg, err := aggregators.NewIdenticalAggregator(*config, logger.Nop())
	require.NoError(t, err)

	encoderStr := "evm"
	encoderName := values.NewString(encoderStr)
	encoderCfg, err := values.NewMap(map[string]any{"foo": "bar"})
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a"), encoderName, encoderCfg},
		1: {values.NewString("a"), encoderName, encoderCfg},
		2: {values.NewString("a"), encoderName, encoderCfg},
		3: {values.NewString("a"), encoderName, encoderCfg},
	}
	outcome, err := agg.Aggregate(nil, observations, 1)
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)
	require.Equal(t, outcome.EncoderName, encoderStr)
	require.Equal(t, outcome.EncoderConfig, values.ProtoMap(encoderCfg))

	m, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	require.Len(t, m.Underlying, 3)
	require.Equal(t, m.Underlying["0"], values.NewString("a"))
	require.Equal(t, m.Underlying["1"], encoderName)
	require.Equal(t, m.Underlying["2"], encoderCfg)
}

func TestDataFeedsAggregator_Aggregate_OverrideWithKeys(t *testing.T) {
	config := getConfig(t, true, []string{"outcome", "encoderName", "encoderConfig"})
	agg, err := aggregators.NewIdenticalAggregator(*config, logger.Nop())
	require.NoError(t, err)

	encoderStr := "evm"
	encoderName := values.NewString(encoderStr)
	encoderCfg, err := values.NewMap(map[string]any{"foo": "bar"})
	require.NoError(t, err)

	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("a"), encoderName, encoderCfg},
		1: {values.NewString("a"), encoderName, encoderCfg},
		2: {values.NewString("a"), encoderName, encoderCfg},
		3: {values.NewString("a"), encoderName, encoderCfg},
	}
	outcome, err := agg.Aggregate(nil, observations, 1)
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)
	require.Equal(t, outcome.EncoderName, encoderStr)
	require.Equal(t, outcome.EncoderConfig, values.ProtoMap(encoderCfg))

	m, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	require.Len(t, m.Underlying, 3)
	require.Equal(t, m.Underlying["outcome"], values.NewString("a"))
	require.Equal(t, m.Underlying["encoderName"], encoderName)
	require.Equal(t, m.Underlying["encoderConfig"], encoderCfg)
}

func TestDataFeedsAggregator_Aggregate_NoConsensus(t *testing.T) {
	config := getConfig(t, true, []string{"outcome", "encoderName", "encoderConfig"})
	agg, err := aggregators.NewIdenticalAggregator(*config, logger.Nop())
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
	outcome, err := agg.Aggregate(nil, observations, 1)
	require.Nil(t, outcome)
	require.ErrorContains(t, err, "can't reach consensus on observations with index 0")
}

func getConfig(t *testing.T, override bool, overrideKeys []string) *values.Map {
	obsLen := 1

	unwrappedConfig := map[string]any{
		"expectedObservationsLen": obsLen,
	}

	if override {
		unwrappedConfig["expectedObservationsLen"] = 3
		unwrappedConfig["encoderNameObservationIndex"] = 1
		unwrappedConfig["encoderConfigObservationIndex"] = 2
		unwrappedConfig["overrideEncoder"] = true
		unwrappedConfig["keyOverrides"] = overrideKeys
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
