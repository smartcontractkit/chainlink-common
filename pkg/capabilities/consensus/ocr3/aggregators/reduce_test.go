package aggregators_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

var (
	feedIDA  = datastreams.FeedID("0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292")
	idABytes = feedIDA.Bytes()
	feedIDB  = datastreams.FeedID("0x0003c317fec7fad514c67aacc6366bf2f007ce37100e3cddcacd0ccaa1f3746d")
	idBBytes = feedIDB.Bytes()
	now      = time.Now()
)

func TestReduceAggregator_Aggregate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cases := []struct {
			name                string
			fields              []aggregators.AggregationField
			extraConfig         map[string]any
			observationsFactory func() map[commontypes.OracleID][]values.Value
			shouldReport        bool
			expectedState       any
			expectedOutcome     map[string]any
			previousOutcome     func(t *testing.T) *types.AggregationOutcome
		}{
			{
				name: "aggregate on int64 median",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "FeedID",
						OutputKey: "FeedID",
						Method:    "mode",
					},
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
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"FeedID":         idABytes[:],
						"BenchmarkPrice": int64(100),
						"Timestamp":      12341414929,
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"FeedID":    idABytes[:],
							"Timestamp": int64(12341414929),
							"Price":     int64(100),
						},
					},
				},
				expectedState: map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": int64(12341414929),
					"Price":     int64(100),
				},
			},
			{
				name: "aggregate on decimal median",
				fields: []aggregators.AggregationField{
					{
						InputKey:        "BenchmarkPrice",
						OutputKey:       "Price",
						Method:          "median",
						DeviationString: "10",
						DeviationType:   "percent",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": decimal.NewFromInt(32),
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": decimal.NewFromInt(32),
						},
					},
				},
				expectedState: map[string]any{
					"Price": decimal.NewFromInt(32),
				},
			},
			{
				name: "aggregate on float64 median",
				fields: []aggregators.AggregationField{
					{
						InputKey:        "BenchmarkPrice",
						OutputKey:       "Price",
						Method:          "median",
						DeviationString: "10",
						DeviationType:   "percent",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": float64(1.2),
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": float64(1.2),
						},
					},
				},
				expectedState: map[string]any{
					"Price": float64(1.2),
				},
			},
			{
				name: "aggregate on time median",
				fields: []aggregators.AggregationField{
					{
						InputKey:        "BenchmarkPrice",
						OutputKey:       "Price",
						Method:          "median",
						DeviationString: "10",
						DeviationType:   "percent",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": now,
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": time.Time(now).UTC(),
						},
					},
				},
				expectedState: map[string]any{
					"Price": now.UTC(),
				},
			},
			{
				name: "aggregate on big int median",
				fields: []aggregators.AggregationField{
					{
						InputKey:        "BenchmarkPrice",
						OutputKey:       "Price",
						Method:          "median",
						DeviationString: "10",
						DeviationType:   "percent",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": big.NewInt(100),
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": big.NewInt(100),
						},
					},
				},
				expectedState: map[string]any{
					"Price": big.NewInt(100),
				},
			},
			{
				name: "aggregate with previous outcome",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "FeedID",
						OutputKey: "FeedID",
						Method:    "mode",
					},
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
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"FeedID":         idABytes[:],
						"BenchmarkPrice": int64(100),
						"Timestamp":      12341414929,
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"FeedID":    idABytes[:],
							"Timestamp": int64(12341414929),
							"Price":     int64(100),
						},
					},
				},
				expectedState: map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": int64(12341414929),
					"Price":     int64(100),
				},
				previousOutcome: func(t *testing.T) *types.AggregationOutcome {
					m, err := values.NewMap(map[string]any{})
					require.NoError(t, err)
					pm := values.Proto(m)
					bm, err := proto.Marshal(pm)
					require.NoError(t, err)
					return &types.AggregationOutcome{Metadata: bm}
				},
			},
			{
				name: "aggregate on bytes mode",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "FeedID",
						OutputKey: "FeedID",
						Method:    "mode",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue1, err := values.WrapMap(map[string]any{
						"FeedID": idABytes[:],
					})
					require.NoError(t, err)
					mockValue2, err := values.WrapMap(map[string]any{
						"FeedID": idBBytes[:],
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue1}, 2: {mockValue1}, 3: {mockValue2}, 4: {mockValue1}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"FeedID": idABytes[:],
						},
					},
				},
				expectedState: map[string]any{
					"FeedID": idABytes[:],
				},
			},
			{
				name: "aggregate on string mode",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "BenchmarkPrice",
						OutputKey: "Price",
						Method:    "mode",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue1, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": "1",
					})
					require.NoError(t, err)
					mockValue2, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": "2",
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue1}, 2: {mockValue1}, 3: {mockValue2}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": "1",
						},
					},
				},
				expectedState: map[string]any{
					"Price": "1",
				},
			},
			{
				name: "aggregate on bool mode",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "BenchmarkPrice",
						OutputKey: "Price",
						Method:    "mode",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue1, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": true,
					})
					require.NoError(t, err)
					mockValue2, err := values.WrapMap(map[string]any{
						"BenchmarkPrice": false,
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue1}, 2: {mockValue1}, 3: {mockValue2}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": true,
						},
					},
				},
				expectedState: map[string]any{
					"Price": true,
				},
			},
			{
				name: "aggregate on non-indexable type",
				fields: []aggregators.AggregationField{
					{
						// Omitting "InputKey"
						OutputKey: "Price",
						Method:    "median",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(1)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": int64(1),
						},
					},
				},
				expectedState: map[string]any{"Price": int64(1)},
			},
			{
				name: "aggregate on list type",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "1",
						OutputKey: "Price",
						Method:    "median",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.NewList([]any{"1", "2", "3"})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"Price": "2",
						},
					},
				},
				expectedState: map[string]any{
					"Price": "2",
				},
			},
			{
				name: "submap",
				fields: []aggregators.AggregationField{
					{
						InputKey:  "FeedID",
						OutputKey: "FeedID",
						Method:    "mode",
					},
					{
						InputKey:        "BenchmarkPrice",
						OutputKey:       "Price",
						Method:          "median",
						DeviationString: "10",
						DeviationType:   "percent",
						SubMapField:     true,
					},
					{
						InputKey:        "Timestamp",
						OutputKey:       "Timestamp",
						Method:          "median",
						DeviationString: "100",
						DeviationType:   "absolute",
					},
				},
				extraConfig: map[string]any{
					"SubMapKey": "Report",
				},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{
						"FeedID":         idABytes[:],
						"BenchmarkPrice": int64(100),
						"Timestamp":      12341414929,
					})
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"FeedID":    idABytes[:],
							"Timestamp": int64(12341414929),
							"Report": map[string]any{
								"Price": int64(100),
							},
						},
					},
				},
				expectedState: map[string]any{
					"FeedID":    idABytes[:],
					"Price":     int64(100),
					"Timestamp": int64(12341414929),
				},
			},
			{
				name: "report format value",
				fields: []aggregators.AggregationField{
					{
						OutputKey: "Price",
						Method:    "median",
					},
				},
				extraConfig: map[string]any{
					"reportFormat": "value",
				},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(1)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": int64(1),
				},
				expectedState: map[string]any{"Price": int64(1)},
			},
			{
				name: "report format array",
				fields: []aggregators.AggregationField{
					{
						OutputKey: "Price",
						Method:    "median",
					},
				},
				extraConfig: map[string]any{
					"reportFormat": "array",
				},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(1)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{map[string]any{"Price": int64(1)}},
				},
				expectedState: map[string]any{"Price": int64(1)},
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				config := getConfigReduceAggregator(t, tt.fields, tt.extraConfig)
				agg, err := aggregators.NewReduceAggregator(*config)
				require.NoError(t, err)

				pb := &pb.Map{}

				var po *types.AggregationOutcome
				if tt.previousOutcome != nil {
					po = tt.previousOutcome(t)
				}

				outcome, err := agg.Aggregate(logger.Nop(), po, tt.observationsFactory(), 1)
				require.NoError(t, err)
				require.Equal(t, tt.shouldReport, outcome.ShouldReport)

				// validate metadata
				err = proto.Unmarshal(outcome.Metadata, pb)
				require.NoError(t, err)
				vmap, err := values.FromMapValueProto(pb)
				require.NoError(t, err)
				state, err := vmap.Unwrap()
				require.NoError(t, err)
				require.Equal(t, tt.expectedState, state)

				// validate encodable outcome
				val, err := values.FromMapValueProto(outcome.EncodableOutcome)
				require.NoError(t, err)
				topLevelMap, err := val.Unwrap()
				require.NoError(t, err)
				mm, ok := topLevelMap.(map[string]any)
				require.True(t, ok)

				require.NoError(t, err)

				require.Equal(t, tt.expectedOutcome, mm)
			})
		}
	})

	t.Run("error path", func(t *testing.T) {
		cases := []struct {
			name                string
			previousOutcome     *types.AggregationOutcome
			fields              []aggregators.AggregationField
			extraConfig         map[string]any
			observationsFactory func() map[commontypes.OracleID][]values.Value
		}{
			{
				name:            "not enough observations",
				previousOutcome: nil,
				fields: []aggregators.AggregationField{
					{
						Method:    "median",
						OutputKey: "Price",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					return map[commontypes.OracleID][]values.Value{}
				},
			},
			{
				name: "invalid previous outcome not pb",
				previousOutcome: &types.AggregationOutcome{
					Metadata: []byte{1, 2, 3},
				},
				fields: []aggregators.AggregationField{
					{
						Method:    "median",
						OutputKey: "Price",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(int64(100))
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
			},
			{
				name:            "not enough extracted values",
				previousOutcome: nil,
				fields: []aggregators.AggregationField{
					{
						InputKey:  "Price",
						OutputKey: "Price",
						Method:    "median",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.WrapMap(map[string]any{"Price": int64(100)})
					require.NoError(t, err)
					mockValueEmpty := values.EmptyMap()
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValueEmpty}}
				},
			},
			{
				name:            "reduce error median",
				previousOutcome: nil,
				fields: []aggregators.AggregationField{
					{
						Method:    "median",
						OutputKey: "Price",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(true)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}}
				},
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				config := getConfigReduceAggregator(t, tt.fields, tt.extraConfig)
				agg, err := aggregators.NewReduceAggregator(*config)
				require.NoError(t, err)

				_, err = agg.Aggregate(logger.Nop(), tt.previousOutcome, tt.observationsFactory(), 1)
				require.Error(t, err)
			})
		}
	})
}

func TestInputChanges(t *testing.T) {
	fields := []aggregators.AggregationField{
		{
			InputKey:  "FeedID",
			OutputKey: "FeedID",
			Method:    "mode",
		},
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
	}
	config := getConfigReduceAggregator(t, fields, map[string]any{})
	agg, err := aggregators.NewReduceAggregator(*config)
	require.NoError(t, err)

	// First Round
	mockValue1, err := values.WrapMap(map[string]any{
		"FeedID":         idABytes[:],
		"BenchmarkPrice": int64(100),
		"Timestamp":      12341414929,
	})
	require.NoError(t, err)
	pb := &pb.Map{}
	outcome, err := agg.Aggregate(logger.Nop(), nil, map[commontypes.OracleID][]values.Value{1: {mockValue1}, 2: {mockValue1}, 3: {mockValue1}}, 1)
	require.NoError(t, err)
	shouldReport := true
	require.Equal(t, shouldReport, outcome.ShouldReport)

	// validate metadata
	proto.Unmarshal(outcome.Metadata, pb)
	vmap, err := values.FromMapValueProto(pb)
	require.NoError(t, err)
	state, err := vmap.Unwrap()
	require.NoError(t, err)
	expectedState1 := map[string]any{
		"FeedID":    idABytes[:],
		"Price":     int64(100),
		"Timestamp": int64(12341414929),
	}
	require.Equal(t, expectedState1, state)

	// validate encodable outcome
	val, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)
	topLevelMap, err := val.Unwrap()
	require.NoError(t, err)
	mm, ok := topLevelMap.(map[string]any)
	require.True(t, ok)

	require.NoError(t, err)
	expectedOutcome1 := map[string]any{
		"Reports": []any{
			map[string]any{
				"FeedID":    idABytes[:],
				"Timestamp": int64(12341414929),
				"Price":     int64(100),
			},
		},
	}
	require.Equal(t, expectedOutcome1, mm)

	// Second Round
	mockValue2, err := values.WrapMap(map[string]any{
		"FeedID":         true,
		"Timestamp":      int64(12341414929),
		"BenchmarkPrice": int64(100),
	})
	require.NoError(t, err)
	outcome, err = agg.Aggregate(logger.Nop(), nil, map[commontypes.OracleID][]values.Value{1: {mockValue2}, 2: {mockValue2}, 3: {mockValue2}}, 1)
	require.NoError(t, err)
	require.Equal(t, shouldReport, outcome.ShouldReport)

	// validate metadata
	proto.Unmarshal(outcome.Metadata, pb)
	vmap, err = values.FromMapValueProto(pb)
	require.NoError(t, err)
	state, err = vmap.Unwrap()
	require.NoError(t, err)
	expectedState2 := map[string]any{
		"FeedID":    true,
		"Price":     int64(100),
		"Timestamp": int64(12341414929),
	}
	require.Equal(t, expectedState2, state)

	// validate encodable outcome
	val, err = values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)
	topLevelMap, err = val.Unwrap()
	require.NoError(t, err)
	mm, ok = topLevelMap.(map[string]any)
	require.True(t, ok)

	require.NoError(t, err)
	expectedOutcome2 := map[string]any{
		"Reports": []any{
			map[string]any{
				"FeedID":    true,
				"Timestamp": int64(12341414929),
				"Price":     int64(100),
			},
		},
	}

	require.Equal(t, expectedOutcome2, mm)

}

func TestMedianAggregator_ParseConfig(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cases := []struct {
			name          string
			inputFactory  func() map[string]any
			outputFactory func() aggregators.ReduceAggConfig
		}{
			{
				name: "no inputkey",
				inputFactory: func() map[string]any {
					return map[string]any{
						"fields": []aggregators.AggregationField{
							{
								Method:    "median",
								OutputKey: "Price",
							},
						},
					}
				},
				outputFactory: func() aggregators.ReduceAggConfig {
					return aggregators.ReduceAggConfig{
						Fields: []aggregators.AggregationField{
							{
								InputKey:        "",
								OutputKey:       "Price",
								Method:          "median",
								DeviationString: "",
								Deviation:       decimal.Decimal{},
								DeviationType:   "none",
							},
						},
						OutputFieldName: "Reports",
						ReportFormat:    "map",
					}
				},
			},
			{
				name: "reportFormat map, aggregation method mode, deviation",
				inputFactory: func() map[string]any {
					return map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedId",
								Method:          "mode",
								DeviationString: "1.1",
								DeviationType:   "absolute",
							},
						},
					}
				},
				outputFactory: func() aggregators.ReduceAggConfig {
					return aggregators.ReduceAggConfig{
						Fields: []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedId",
								Method:          "mode",
								DeviationString: "1.1",
								Deviation:       decimal.NewFromFloat(1.1),
								DeviationType:   "absolute",
							},
						},
						OutputFieldName: "Reports",
						ReportFormat:    "map",
					}
				},
			},
			{
				name: "reportFormat array, aggregation method median, no deviation",
				inputFactory: func() map[string]any {
					return map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedId",
								Method:    "median",
							},
						},
						"outputFieldName": "Reports",
						"reportFormat":    "array",
					}
				},
				outputFactory: func() aggregators.ReduceAggConfig {
					return aggregators.ReduceAggConfig{
						Fields: []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedId",
								Method:          "median",
								DeviationString: "",
								Deviation:       decimal.Decimal{},
								DeviationType:   "none",
							},
						},
						OutputFieldName: "Reports",
						ReportFormat:    "array",
					}
				},
			},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				vMap, err := values.NewMap(tt.inputFactory())
				require.NoError(t, err)
				parsedConfig, err := aggregators.ParseConfigReduceAggregator(*vMap)
				require.NoError(t, err)
				require.Equal(t, tt.outputFactory(), parsedConfig)
			})
		}
	})

	t.Run("unhappy path", func(t *testing.T) {
		cases := []struct {
			name          string
			configFactory func() *values.Map
		}{
			{
				name: "empty",
				configFactory: func() *values.Map {
					return values.EmptyMap()
				},
			},
			{
				name: "invalid report format",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "median",
							},
						},
						"reportFormat": "invalid",
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with no method",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with empty method",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with invalid method",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "invalid",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with deviation string but no deviation type",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedID",
								Method:          "median",
								DeviationString: "1",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with deviation string but empty deviation type",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedID",
								Method:          "median",
								DeviationString: "1",
								DeviationType:   "",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with invalid deviation type",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedID",
								Method:          "median",
								DeviationString: "1",
								DeviationType:   "invalid",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with deviation type but no deviation string",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:      "FeedID",
								OutputKey:     "FeedID",
								Method:        "median",
								DeviationType: "absolute",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with deviation type but empty deviation string",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedID",
								Method:          "median",
								DeviationType:   "absolute",
								DeviationString: "",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with invalid deviation string",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:        "FeedID",
								OutputKey:       "FeedID",
								Method:          "median",
								DeviationType:   "absolute",
								DeviationString: "1-1",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "field with sub report, but no sub report key",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:    "FeedID",
								OutputKey:   "FeedID",
								Method:      "median",
								SubMapField: true,
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "sub report key, but no sub report fields",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"subMapKey": "Report",
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "median",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "clashing output keys",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "median",
							},
							{
								InputKey:  "FeedID",
								OutputKey: "FeedID",
								Method:    "median",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "map/array type, no output key",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey: "FeedID",
								Method:   "median",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
			{
				name: "report type value with multiple fields",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"reportFormat": "value",
						"fields": []aggregators.AggregationField{
							{
								InputKey:  "FeedID",
								Method:    "median",
								OutputKey: "FeedID",
							},
							{
								InputKey:  "Price",
								Method:    "median",
								OutputKey: "Price",
							},
						},
					})
					require.NoError(t, err)
					return vMap
				},
			},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := aggregators.ParseConfigReduceAggregator(*tt.configFactory())
				require.Error(t, err)
			},
			)
		}
	})
}

func getConfigReduceAggregator(t *testing.T, fields []aggregators.AggregationField, override map[string]any) *values.Map {
	unwrappedConfig := map[string]any{
		"fields":          fields,
		"outputFieldName": "Reports",
		"reportFormat":    "array",
	}
	for key, val := range override {
		unwrappedConfig[key] = val
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
