package reduce_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/reduce"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

var (
	feedIDA     = datastreams.FeedID("0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292")
	remappedIDA = "0x680084f7347baFfb5C323c2982dfC90e04F9F918"
	deviationA  = decimal.NewFromFloat(0.1)
	heartbeatA  = 60
)

var ExpectedReportFields = []string{"FeedId", "RawReport", "Timestamp", "Price"}

func TestReduceAggregator_Aggregate_TwoRounds(t *testing.T) {
	cases := []struct {
		name                string
		observationsFactory func() map[commontypes.OracleID][]values.Value
		priceOutput1        any
		priceOutput2        any
	}{
		// {
		// 	name: "success - big int",
		// 	observationsFactory: func() map[commontypes.OracleID][]values.Value {
		// 		value, err := values.WrapMap(map[string]any{
		// 			"BenchmarkPrice": big.NewInt(100).Bytes(),
		// 		})
		// 		require.NoError(t, err)
		// 		mockEvent, err := values.Wrap(capabilities.CapabilityResponse{
		// 			Value: value,
		// 		})
		// 		require.NoError(t, err)
		// 		return map[commontypes.OracleID][]values.Value{1: {mockEvent}, 2: {mockEvent}}
		// 	},
		// 	priceOutput1: big.NewInt(100).Bytes(),
		// 	priceOutput2: big.NewInt(100),
		// },
		// {
		// 	name: "success - buffer",
		// 	observationsFactory: func() map[commontypes.OracleID][]values.Value {
		// 		buff := new(bytes.Buffer)
		// 		bigOrLittleEndian := binary.BigEndian
		// 		err := binary.Write(buff, bigOrLittleEndian, int64(100))
		// 		if err != nil {
		// 			log.Panic(err)
		// 		}
		// 		value, err := values.WrapMap(map[string]any{
		// 			"BenchmarkPrice": buff.Bytes(),
		// 		})
		// 		require.NoError(t, err)
		// 		mockEvent, err := values.Wrap(capabilities.CapabilityResponse{
		// 			Value: value,
		// 		})
		// 		require.NoError(t, err)
		// 		return map[commontypes.OracleID][]values.Value{1: {mockEvent}, 2: {mockEvent}}
		// 	},
		// 	priceOutput1: big.NewInt(100).Bytes(),
		// 	priceOutput2: big.NewInt(100),
		// },
		{
			name: "success - int64",
			observationsFactory: func() map[commontypes.OracleID][]values.Value {
				mockValue, err := values.WrapMap(map[string]any{
					"BenchmarkPrice": int64(100),
					"Timestamp":      12341414929,
				})
				require.NoError(t, err)
				return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
			},
			priceOutput1: int64(100),
			priceOutput2: int64(100),
		},
		// {
		// 	name: "success - string",
		// 	observationsFactory: func() map[commontypes.OracleID][]values.Value {
		// 		value, err := values.WrapMap(map[string]any{
		// 			"BenchmarkPrice": "d",
		// 		})
		// 		require.NoError(t, err)
		// 		mockEvent, err := values.Wrap(capabilities.CapabilityResponse{
		// 			Value: value,
		// 		})
		// 		require.NoError(t, err)
		// 		return map[commontypes.OracleID][]values.Value{1: {mockEvent}, 2: {mockEvent}}
		// 	},
		// 	priceOutput1: []byte("d"),
		// 	priceOutput2: big.NewInt(100),
		// },
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			config := getConfig(t, feedIDA, []reduce.AggregationField{
				{
					InputKey:        "BenchmarkPrice",
					OutputKey:       "Price",
					Method:          "median",
					DeviationString: "10",
					DeviationType:   "percent",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Timestamp",
					Method:          "median",
					DeviationString: "100",
					DeviationType:   "absolute",
				},
			})
			agg, err := reduce.NewReduceAggregator(*config, logger.Nop())
			require.NoError(t, err)

			// first round, empty previous Outcome, empty observations, expect should not report
			outcome, err := agg.Aggregate(nil, map[commontypes.OracleID][]values.Value{}, 1)
			require.NoError(t, err)
			require.False(t, outcome.ShouldReport)

			// validate metadata
			pb := &pb.Map{}
			proto.Unmarshal(outcome.Metadata, pb)
			vmap, err := values.FromMapValueProto(pb)
			require.NoError(t, err)
			newState, err := vmap.Unwrap()
			require.NoError(t, err)
			for _, val := range newState.(map[string]any) {
				require.Equal(t, val, nil)
			}

			// second round, expect should report
			outcome, err = agg.Aggregate(outcome, tt.observationsFactory(), 1)
			require.NoError(t, err)
			require.True(t, outcome.ShouldReport)

			// validate metadata
			proto.Unmarshal(outcome.Metadata, pb)
			vmap, err = values.FromMapValueProto(pb)
			require.NoError(t, err)
			newState, err = vmap.Unwrap()
			require.NoError(t, err)
			require.Equal(t, tt.priceOutput1, newState.(map[string]any)["BenchmarkPrice"])

			// validate encodable outcome
			val, err := values.FromMapValueProto(outcome.EncodableOutcome)
			require.NoError(t, err)
			topLevelMap, err := val.Unwrap()
			require.NoError(t, err)
			mm, ok := topLevelMap.(map[string]any)
			require.True(t, ok)

			idBytes := feedIDA.Bytes()
			require.NoError(t, err)
			expected := map[string]any{
				"Reports": []any{
					map[string]any{
						"FeedID":    idBytes[:],
						"RawReport": nil,
						"Timestamp": int64(12341414929),
						"Price":     tt.priceOutput2,
					},
				},
			}
			require.Equal(t, expected, mm)
		})
	}
}

// func TestMedianAggregator_ParseConfig(t *testing.T) {
// 	t.Run("happy path", func(t *testing.T) {
// 		config := getConfig(t, feedIDA.String(), "0.1", heartbeatA, "BenchmarkPrice")
// 		parsedConfig, err := reduce.ParseConfig(*config)
// 		require.NoError(t, err)
// 		require.Equal(t, feedIDA, parsedConfig.FeedID)
// 		require.Equal(t, deviationA, parsedConfig.Deviation)
// 		require.Equal(t, heartbeatA, parsedConfig.Heartbeat)
// 	})

// 	t.Run("invalid ID", func(t *testing.T) {
// 		config := getConfig(t, "bad_id", "0.1", heartbeatA, "BenchmarkPrice")
// 		_, err := reduce.ParseConfig(*config)
// 		require.Error(t, err)
// 	})

// 	t.Run("invalid deviation string", func(t *testing.T) {
// 		config := getConfig(t, feedIDA.String(), "bad_number", heartbeatA, "BenchmarkPrice")
// 		_, err := reduce.ParseConfig(*config)
// 		require.Error(t, err)
// 	})

// 	t.Run("no value key", func(t *testing.T) {
// 		config := getConfig(t, feedIDA.String(), "0.1", heartbeatA, "")
// 		_, err := reduce.ParseConfig(*config)
// 		require.Error(t, err)
// 	})
// }

func getConfig(t *testing.T, feedID datastreams.FeedID, fields []reduce.AggregationField) *values.Map {
	unwrappedConfig := map[string]any{
		"additionalData": map[string]any{
			"FeedID":    feedID.Bytes(),
			"RawReport": nil,
		},
		"fields":          fields,
		"outputFieldName": "Reports",
		"reportFormat":    "array",
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
