package datafeeds_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestGetLatestPrices(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)

	tests := []struct {
		name              string
		streamIDs         []uint32
		events            map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent
		f                 int
		expectedTimestamp uint64
		expectedPrices    map[uint32]decimal.Decimal
		expectError       bool
	}{
		{
			name:      "successful price consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				3: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
			},
			f:                 1,
			expectedTimestamp: 1000,
			expectedPrices: map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(100.5),
				2: decimal.NewFromFloat(200.5),
			},
			expectError: false,
		},
		{
			name:      "insufficient price consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(101.5), // Different value
					2: decimal.NewFromFloat(201.5), // Different value
				}),
				3: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(102.5), // Different value
					2: decimal.NewFromFloat(202.5), // Different value
				}),
			},
			f:                 1,
			expectedTimestamp: 1000,
			expectedPrices:    map[uint32]decimal.Decimal{}, // No consensus
			expectError:       false,
		},
		{
			name:      "mixed consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(201.5), // Different value
				}),
				3: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(202.5), // Different value
				}),
			},
			f:                 1,
			expectedTimestamp: 1000,
			expectedPrices: map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(100.5), // Consensus for stream 1
				// No consensus for stream 2
			},
			expectError: false,
		},
		{
			name:      "no timestamp consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, 1000, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, 1001, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				3: createLLOEvent(t, 1002, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
			},
			f:           1,
			expectError: true, // No timestamp consensus
		},
		{
			name:        "empty event list",
			streamIDs:   []uint32{1, 2},
			events:      map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{},
			f:           1,
			expectError: true, // No events to check
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts, prices, err := datafeeds.LLOStreamPrices(lggr, tc.streamIDs, tc.events, tc.f)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTimestamp, ts)

			// Check all expected prices
			for streamID, expectedPrice := range tc.expectedPrices {
				actualPrice, exists := prices[streamID]
				assert.True(t, exists, "Expected price for stream %d not found", streamID)
				assert.True(t, expectedPrice.Equal(actualPrice),
					"Expected price %s for stream %d, got %s",
					expectedPrice.String(), streamID, actualPrice.String())
			}

			// Ensure no extra prices
			assert.Equal(t, len(tc.expectedPrices), len(prices), "Unexpected number of prices")
		})
	}
}

func TestLLOAggregator_Aggregate(t *testing.T) {
	lggr := logger.Test(t)

	tests := []struct {
		name                 string
		config               values.Map //datafeeds.LLOAggregatorConfig
		previousOutcome      *types.AggregationOutcome
		observations         map[ocrcommon.OracleID][]values.Value
		f                    int
		expectedShouldReport bool
		expectedStreamIDs    []uint32
		expectError          bool
	}{
		{
			name: "update due to deviation",
			config: zz(t, map[uint32]datafeeds.FeedConfig{
				1: {
					Deviation: decimal.NewFromFloat(0.01), // 1%
					Heartbeat: 3600,                       // 1 hour
				},
			}),

			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: time.Now().Add(-10 * time.Minute).UnixNano(),
				},
			}),
			observations: createObservations(t, 1000, map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(102), // 2% change, exceeds 1% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1},
			expectError:          false,
		},
		{
			name: "update due to heartbeat",
			config: zz(t, map[uint32]datafeeds.FeedConfig{
				1: {
					Deviation: decimal.NewFromFloat(0.1), // 10%
					Heartbeat: 300,                       // 5 minutes
				},
			}),

			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: time.Now().Add(-6 * time.Minute).UnixNano(), // Over heartbeat
				},
			}),
			observations: createObservations(t, 1000, map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(101), // 1% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1},
			expectError:          false,
		},
		{
			name: "no update needed",
			config: zz(t, map[uint32]datafeeds.FeedConfig{

				1: {
					Deviation: decimal.NewFromFloat(0.1), // 10%
					Heartbeat: 3600,                      // 1 hour
				},
			}),
			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: time.Now().Add(-30 * time.Minute).UnixNano(), // Under heartbeat
				},
			}),
			observations: createObservations(t, 1000, map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(105), // 5% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: false, // No update needed
			expectedStreamIDs:    []uint32{},
			expectError:          false,
		},
		{
			name: "partial staleness optimization",
			config: zz(t, map[uint32]datafeeds.FeedConfig{
				1: {
					Deviation: decimal.NewFromFloat(0.1), // 10%
					Heartbeat: 3600,                      // 1 hour
				},
				2: {
					Deviation: decimal.NewFromFloat(0.1), // 10%
					Heartbeat: 300,                       // 5 minutes
				},
			}, lloConfigAllowStaleness(0.2)), // 20% allowed partial staleness

			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: time.Now().Add(-50 * time.Minute).UnixNano(), // 83% of heartbeat
				},
				2: {
					price:     decimal.NewFromFloat(200),
					timestamp: time.Now().Add(-6 * time.Minute).UnixNano(), // Over heartbeat
				},
			}),
			observations: createObservations(t, 1000, map[uint32]decimal.Decimal{
				1: decimal.NewFromFloat(105), // 5% change, under 10% threshold
				2: decimal.NewFromFloat(202), // 1% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1, 2}, // Both update due to optimization
			expectError:          false,
		},
		{
			name: "empty observations",
			config: zz(t, map[uint32]datafeeds.FeedConfig{
				1: {
					Deviation: decimal.NewFromFloat(0.1),
					Heartbeat: 3600,
				},
			}),

			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{}),
			observations:         map[ocrcommon.OracleID][]values.Value{},
			f:                    1,
			expectedShouldReport: false,
			expectError:          true, // Should error with empty observations
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aggregator, err := datafeeds.NewLLOAggregator(tc.config)

			outcome, err := aggregator.Aggregate(lggr, tc.previousOutcome, tc.observations, tc.f)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedShouldReport, outcome.ShouldReport)

			if outcome.ShouldReport {
				// Verify that the correct streams were updated
				reportedStreams := extractUpdatedStreamIDs(t, outcome)
				assert.ElementsMatch(t, tc.expectedStreamIDs, reportedStreams)
			}
		})
	}
}

// Helper functions

func createLLOEvent(t *testing.T, ts uint64, prices map[uint32]decimal.Decimal) *datastreams.LLOStreamsTriggerEvent {
	event := &datastreams.LLOStreamsTriggerEvent{
		ObservationTimestampNanoseconds: ts,
		Payload:                         make([]*datastreams.LLOStreamDecimal, 0, len(prices)),
	}

	for streamID, price := range prices {
		binary, err := price.MarshalBinary()
		require.NoError(t, err)

		event.Payload = append(event.Payload, &datastreams.LLOStreamDecimal{
			StreamID: streamID,
			Decimal:  binary,
		})
	}

	return event
}

func createPreviousOutcome(t *testing.T, streams map[uint32]struct {
	price     decimal.Decimal
	timestamp int64
}) *types.AggregationOutcome {
	state := &datafeeds.LLOOutcomeMetadata{
		StreamInfo: make(map[uint32]*datafeeds.LLOStreamInfo),
	}

	for streamID, info := range streams {
		priceBytes, err := info.price.MarshalBinary()
		require.NoError(t, err)

		state.StreamInfo[streamID] = &datafeeds.LLOStreamInfo{
			Timestamp: info.timestamp,
			Price:     priceBytes,
		}
	}

	marshalledState, err := proto.Marshal(state)
	require.NoError(t, err)

	return &types.AggregationOutcome{
		Metadata: marshalledState,
	}
}

func createObservations(t *testing.T, ts uint64, prices map[uint32]decimal.Decimal) map[ocrcommon.OracleID][]values.Value {
	observations := make(map[ocrcommon.OracleID][]values.Value)

	// Create three observations with identical data to ensure f+1 consensus
	for i := ocrcommon.OracleID(1); i <= 3; i++ {
		event := createLLOEvent(t, ts, prices)

		val, err := values.Wrap(event)
		require.NoError(t, err)

		observations[i] = []values.Value{val}
	}

	return observations
}

func extractUpdatedStreamIDs(t *testing.T, outcome *types.AggregationOutcome) []uint32 {

	//outcome.EncodableOutcome.Fields[datafeeds.TopLevelListOutputFieldName].GetListValue()
	decodedMap, err := values.FromMapValueProto(outcome.EncodableOutcome)
	require.NoError(t, err)

	reportsAny, ok := decodedMap.Underlying[datafeeds.TopLevelListOutputFieldName]
	require.True(t, ok)

	var reportsList []any // each element is a WrappableUpdate
	err = reportsAny.UnwrapTo(reportsList)
	require.NoError(t, err)

	streamIDs := make([]uint32, 0, len(reportsList))
	for _, reportAny := range reportsList {
		report := reportAny.(*datafeeds.WrappableUpdate)

		/*
			streamIDAny, exists :=  reports[datafeeds.StreamIDOutputFieldName] //report.Get(StreamIDOutputFieldName)
			require.True(t, exists)

			streamID, err := streamIDAny.AsUint()
			require.NoError(t, err)
		*/
		streamIDs = append(streamIDs, report.StreamID)
	}

	return streamIDs
}

func lloFeedConfig(t *testing.T, feedID string, deviation string, heartbeat int) *values.Map {
	unwrappedConfig := map[string]any{
		"feeds": map[string]any{
			feedID: map[string]any{
				"deviation":  deviation,
				"heartbeat":  heartbeat,
				"remappedID": remappedIDA,
			},
			feedIDB.String(): map[string]any{
				"deviation": deviationB.String(),
				"heartbeat": heartbeatB,
			},
		},
		"allowedPartialStaleness": "0.2",
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}

func zz(t *testing.T, m map[uint32]datafeeds.FeedConfig, opts ...lloConfigOpt) values.Map {
	unwrappedConfig := map[string]any{
		"feeds": map[string]any{},
	}
	for feedID, cfg := range m {
		unwrappedConfig["feeds"].(map[string]any)[strconv.FormatUint(uint64(feedID), 10)] = map[string]any{
			"deviation": cfg.Deviation.String(),
			"heartbeat": cfg.Heartbeat,
			//			"remappedID": cfg.RemappedID,
		}
	}
	for _, opt := range opts {
		opt(t, unwrappedConfig)
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return *config
}

type lloConfigOpt = func(t *testing.T, m map[string]any)

func lloConfigAllowStaleness(staleness float64) lloConfigOpt {
	return func(t *testing.T, m map[string]any) {
		m["allowedPartialStaleness"] = staleness
	}
}
