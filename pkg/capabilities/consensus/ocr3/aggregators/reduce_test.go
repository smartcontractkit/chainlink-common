package aggregators_test

import (
	"maps"
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
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
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
				name: "aggregate on uint64 median",
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
						"BenchmarkPrice": uint64(100),
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
							"Price":     uint64(100),
						},
					},
				},
				expectedState: map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": int64(12341414929),
					"Price":     uint64(100),
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
			{
				name: "handle nils gracefully",
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
						"BenchmarkPrice": uint64(100),
						"Timestamp":      12341414929,
					})
					require.NoError(t, err)
					mockValueWithNil, err := values.WrapMap(map[string]any{
						"FeedID":         idABytes[:],
						"BenchmarkPrice": uint64(100),
						"Timestamp":      12341414929,
					})
					mockValueWithNil.Underlying["BenchmarkPrice"] = nil // simulate failed wraping of uint64
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue}, 3: {mockValue}, 4: {mockValueWithNil}}
				},
				shouldReport: true,
				expectedOutcome: map[string]any{
					"Reports": []any{
						map[string]any{
							"FeedID":    idABytes[:],
							"Timestamp": int64(12341414929),
							"Price":     uint64(100),
						},
					},
				},
				expectedState: map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": int64(12341414929),
					"Price":     uint64(100),
				},
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
			errString           string
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
				errString: "consensus failed: insufficient observations, received 0 but need at least 3 (2f+1, f=1). Not enough DON nodes responded in time",
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
				errString: "initializeCurrentState Unmarshal error:", // the proto package discourages full string error comparisons
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
				errString: "consensus failed: insufficient observations for field \"Price\", received 2 but need at least 3 (2f+1, f=1). Not enough DON nodes provided data for this field",
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
				errString: "unable to reduce on method median, err: unable to convert type bool to decimal",
			},
			{
				name:            "reduce error mode with mode quorum of: ocr",
				previousOutcome: nil,
				fields: []aggregators.AggregationField{
					{
						Method:     "mode",
						ModeQuorum: "ocr",
						OutputKey:  "Price",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(1)
					require.NoError(t, err)
					mockValue2, err := values.Wrap(2)
					require.NoError(t, err)
					mockValue3, err := values.Wrap(3)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue2}, 3: {mockValue3}}
				},
				errString: "unable to reduce on method mode, err: consensus failed: mode quorum not reached, 1 nodes agreed but need at least 2 (f+1, f=1). DON nodes disagree too much on the value",
			},
			{
				name:            "reduce error mode with mode quorum of: all",
				previousOutcome: nil,
				fields: []aggregators.AggregationField{
					{
						Method:     "mode",
						ModeQuorum: "all",
						OutputKey:  "Price",
					},
				},
				extraConfig: map[string]any{},
				observationsFactory: func() map[commontypes.OracleID][]values.Value {
					mockValue, err := values.Wrap(1)
					require.NoError(t, err)
					mockValue2, err := values.Wrap(2)
					require.NoError(t, err)
					return map[commontypes.OracleID][]values.Value{1: {mockValue}, 2: {mockValue2}, 3: {mockValue2}}
				},
				errString: "unable to reduce on method mode, err: consensus failed: mode quorum not reached, 2 nodes agreed but need at least 3 (2f+1, f=1). DON nodes disagree too much on the value",
			},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				config := getConfigReduceAggregator(t, tt.fields, tt.extraConfig)
				agg, err := aggregators.NewReduceAggregator(*config)
				require.NoError(t, err)

				_, err = agg.Aggregate(logger.Nop(), tt.previousOutcome, tt.observationsFactory(), 1)
				require.ErrorContains(t, err, tt.errString)
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
								ModeQuorum:      "ocr",
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
			{
				name: "invalid mode quorum",
				configFactory: func() *values.Map {
					vMap, err := values.NewMap(map[string]any{
						"fields": []aggregators.AggregationField{
							{
								InputKey:   "Price",
								Method:     "mode",
								ModeQuorum: "invalid",
								OutputKey:  "Price",
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

func TestAggregateShouldReport(t *testing.T) {
	extraConfig := map[string]any{
		"reportFormat": "array",
	}

	cases := []struct {
		name                    string
		fields                  []aggregators.AggregationField
		mockValueFirstRound     *values.Map
		shouldReportFirstRound  bool
		stateFirstRound         map[string]any
		mockValueSecondRound    *values.Map
		shouldReportSecondRound bool
		stateSecondRound        map[string]any
		mockValueThirdRound     *values.Map
		shouldReportThirdRound  bool
		stateThirdRound         map[string]any
	}{
		{
			name: "OK-report_only_when_deviation_exceeded",
			fields: []aggregators.AggregationField{
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound:        map[string]any{"Time": decimal.NewFromInt(10)},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"Timestamp": decimal.NewFromInt(30),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			// the delta between 10 and 30 is 20, which is less than the deviation of 30, so the state should remain the same
			stateSecondRound: map[string]any(map[string]any{"Time": decimal.NewFromInt(10)}),

			mockValueThirdRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"Timestamp": decimal.NewFromInt(45),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportThirdRound: true,
			// the delta between 10 and 45 is 35, which is more than the deviation of 30, thats why the state is updated
			stateThirdRound: map[string]any{"Time": decimal.NewFromInt(45)},
		},
		{
			name: "NOK-do_not_report_if_deviation_type_any_byte_field_does_not_change",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any(map[string]any{
				"FeedID": idABytes[:],
				"Time":   decimal.NewFromInt(10),
			}),

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": idABytes[:],
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "NOK-do_not_report_if_deviation_type_any_bool_field_does_not_change",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "BoolField",
					OutputKey:     "BoolField",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"BoolField": true,
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"BoolField": true,
				"Time":      decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"BoolField": true,
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"BoolField": true,
				"Time":      decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_byte_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    idABytes[:],
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": idABytes[:],
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    idBBytes[:],
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": idBBytes[:],
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_bool_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "BoolField",
					OutputKey:     "BoolField",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"BoolField": true,
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"BoolField": true,
				"Time":      decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"BoolField": false,
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"BoolField": false,
				"Time":      decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_string_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    "A",
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": "A",
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    "B",
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": "B",
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "NOK-do_not_report_if_deviation_type_any_string_field_does_not_change",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    "A",
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": "A",
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    "A",
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": "A",
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_map_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    map[string]any{"A": "A"},
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": map[string]any{"A": "A"},
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    map[string]any{"A": "B"},
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": map[string]any{"A": "B"},
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "NOK-do_not_report_if_deviation_type_any_map_field_does_not_change",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    map[string]any{"A": "A"},
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": map[string]any{"A": "A"},
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    map[string]any{"A": "A"},
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": map[string]any{"A": "A"},
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_slice_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    []any{"A"},
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": []any{"A"},
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    []any{"B"},
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": []any{"B"},
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "NOK-do_not_report_if_deviation_type_any_slice_field_does_not_change",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    []any{"A"},
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": []any{"A"},
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    []any{"A"},
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": []any{"A"},
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_numeric_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    int64(1),
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": int64(1),
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    int64(2),
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: true,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": int64(2),
				"Time":   decimal.NewFromInt(10),
			}),
		},
		{
			name: "OK-report_if_deviation_type_any_numeric_field_is_changed",
			fields: []aggregators.AggregationField{
				{
					InputKey:      "FeedID",
					OutputKey:     "FeedID",
					Method:        "mode",
					ModeQuorum:    "any",
					DeviationType: "any",
				},
				{
					InputKey:        "Timestamp",
					OutputKey:       "Time",
					Method:          "median",
					DeviationString: "30",
					DeviationType:   "absolute",
				},
			},
			mockValueFirstRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    int64(1),
					"Timestamp": decimal.NewFromInt(10),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportFirstRound: true,
			stateFirstRound: map[string]any{
				"FeedID": int64(1),
				"Time":   decimal.NewFromInt(10),
			},

			mockValueSecondRound: func() *values.Map {
				mockValue, err := values.WrapMap(map[string]any{
					"FeedID":    int64(1),
					"Timestamp": decimal.NewFromInt(20),
				})
				require.NoError(t, err)
				return mockValue
			}(),
			shouldReportSecondRound: false,
			stateSecondRound: map[string]any(map[string]any{
				"FeedID": int64(1),
				"Time":   decimal.NewFromInt(10),
			}),
		},
	}

	for _, tc := range cases {
		config := getConfigReduceAggregator(t, tc.fields, extraConfig)
		agg, err := aggregators.NewReduceAggregator(*config)
		require.NoError(t, err)

		pb := &pb.Map{}

		// 1st round
		firstOutcome, err := agg.Aggregate(logger.Nop(), nil, map[commontypes.OracleID][]values.Value{1: {tc.mockValueFirstRound}, 2: {tc.mockValueFirstRound}, 3: {tc.mockValueFirstRound}}, 1)
		require.NoError(t, err)
		require.Equal(t, tc.shouldReportFirstRound, firstOutcome.ShouldReport)

		// validate metadata
		proto.Unmarshal(firstOutcome.Metadata, pb)
		vmap, err := values.FromMapValueProto(pb)
		require.NoError(t, err)
		state, err := vmap.Unwrap()
		require.NoError(t, err)
		require.Equal(t, map[string]any(tc.stateFirstRound), state)

		// 2nd round
		secondOutcome, err := agg.Aggregate(logger.Nop(), firstOutcome, map[commontypes.OracleID][]values.Value{1: {tc.mockValueSecondRound}, 2: {tc.mockValueSecondRound}, 3: {tc.mockValueSecondRound}}, 1)
		require.NoError(t, err)
		require.Equal(t, tc.shouldReportSecondRound, secondOutcome.ShouldReport)

		// validate metadata
		proto.Unmarshal(secondOutcome.Metadata, pb)
		vmap, err = values.FromMapValueProto(pb)
		require.NoError(t, err)
		state, err = vmap.Unwrap()
		require.NoError(t, err)
		require.Equal(t, tc.stateSecondRound, state)

		// skip if there is no third round
		if tc.mockValueThirdRound == nil {
			continue
		}

		// 3rd round
		thirdOutcome, err := agg.Aggregate(logger.Nop(), secondOutcome, map[commontypes.OracleID][]values.Value{1: {tc.mockValueThirdRound}, 2: {tc.mockValueThirdRound}, 3: {tc.mockValueThirdRound}}, 1)
		require.NoError(t, err)
		require.Equal(t, true, thirdOutcome.ShouldReport)

		// validate metadata
		proto.Unmarshal(thirdOutcome.Metadata, pb)
		vmap, err = values.FromMapValueProto(pb)
		require.NoError(t, err)
		state, err = vmap.Unwrap()
		require.NoError(t, err)
		require.Equal(t, tc.stateThirdRound, state)
	}
}

func getConfigReduceAggregator(t *testing.T, fields []aggregators.AggregationField, override map[string]any) *values.Map {
	unwrappedConfig := map[string]any{
		"fields":          fields,
		"outputFieldName": "Reports",
		"reportFormat":    "array",
	}
	maps.Copy(unwrappedConfig, override)
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
