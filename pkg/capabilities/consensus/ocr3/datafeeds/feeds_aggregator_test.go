package datafeeds_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	feedIDA            = mercury.FeedID("0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292")
	deviationA         = decimal.NewFromFloat(0.1)
	heartbeatA         = 60
	mercuryFullReportA = []byte("report")
)

func TestDataFeedsAggregator_Aggregate_TwoRounds(t *testing.T) {
	config := getConfig(t, feedIDA.String(), deviationA, heartbeatA)
	codec := mocks.NewMercuryCodec(t)
	agg, err := datafeeds.NewDataFeedsAggregator(*config, codec, logger.Nop())
	require.NoError(t, err)

	// first round, empty previous Outcome, empty observations
	outcome, err := agg.Aggregate(nil, map[commontypes.OracleID][]values.Value{})
	require.NoError(t, err)
	require.False(t, outcome.ShouldReport)

	// validate metadata
	newState := &datafeeds.DataFeedsOutcomeMetadata{}
	err = proto.Unmarshal(outcome.Metadata, newState)
	require.NoError(t, err)
	require.Equal(t, 1, len(newState.FeedInfo))
	_, ok := newState.FeedInfo[feedIDA.String()]
	require.True(t, ok)
	require.Equal(t, 0.0, newState.FeedInfo[feedIDA.String()].Price)

	// second round, non-empty previous Outcome, one observation
	latestMercuryReports := mercury.ReportSet{
		Reports: map[mercury.FeedID]mercury.Report{
			feedIDA: {
				Info: mercury.ReportInfo{
					Timestamp: 1,
					Price:     1.0,
				},
				FullReport: mercuryFullReportA,
			},
		},
	}
	codec.On("Unwrap", mock.Anything).Return(latestMercuryReports, nil)
	outcome, err = agg.Aggregate(outcome, map[commontypes.OracleID][]values.Value{1: {nil}})
	require.NoError(t, err)
	require.True(t, outcome.ShouldReport)

	// validate metadata
	err = proto.Unmarshal(outcome.Metadata, newState)
	require.NoError(t, err)
	require.Equal(t, 1, len(newState.FeedInfo))
	_, ok = newState.FeedInfo[feedIDA.String()]
	require.True(t, ok)
	require.Equal(t, 1.0, newState.FeedInfo[feedIDA.String()].Price)

	// validate encodable outcome
	val := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)
	topLevelMap, err := val.Unwrap()
	require.NoError(t, err)
	mm, ok := topLevelMap.(map[string]any)
	require.True(t, ok)
	mercuryReports := mm[datafeeds.OutputFieldName]
	reportList, ok := mercuryReports.([]any)
	require.True(t, ok)
	require.Equal(t, 1, len(reportList))
	reportBytes, ok := reportList[0].([]byte)
	require.True(t, ok)
	require.Equal(t, string(mercuryFullReportA), string(reportBytes))
}

func TestDataFeedsAggregator_ParseConfig(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		config := getConfig(t, feedIDA.String(), deviationA, heartbeatA)
		parsedConfig, err := datafeeds.ParseConfig(*config)
		require.NoError(t, err)
		require.Equal(t, deviationA, parsedConfig.Feeds[feedIDA].Deviation)
		require.Equal(t, heartbeatA, parsedConfig.Feeds[feedIDA].Heartbeat)
	})

	t.Run("invalid ID", func(t *testing.T) {
		config := getConfig(t, "bad_id", deviationA, heartbeatA)
		parsedConfig, err := datafeeds.ParseConfig(*config)
		require.NoError(t, err)
		require.Equal(t, deviationA, parsedConfig.Feeds[feedIDA].Deviation)
		require.Equal(t, heartbeatA, parsedConfig.Feeds[feedIDA].Heartbeat)
	})
}

func getConfig(t *testing.T, feedID string, deviation decimal.Decimal, heartbeat int) *values.Map {
	unwrappedConfig := map[string]any{
		feedIDA.String(): map[string]any{
			"Deviation": deviation,
			"Heartbeat": heartbeat,
		},
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
