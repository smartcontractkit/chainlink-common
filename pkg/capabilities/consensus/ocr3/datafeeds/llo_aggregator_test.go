package datafeeds_test

import (
	"encoding/hex"
	"fmt"
	"sort"
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

	testTime := time.Unix(3164233, 0)

	tests := []struct {
		name              string
		streamIDs         []uint32
		events            map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent
		f                 int
		expectedTimestamp time.Time
		expectedPrices    map[uint32]decimal.Decimal
		expectError       bool
	}{
		{
			name:      "successful price consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				3: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
			},
			f:                 1,
			expectedTimestamp: testTime,
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
				1: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(101.5), // Different value
					2: decimal.NewFromFloat(201.5), // Different value
				}),
				3: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(102.5), // Different value
					2: decimal.NewFromFloat(202.5), // Different value
				}),
			},
			f:                 1,
			expectedTimestamp: testTime,
			expectedPrices:    map[uint32]decimal.Decimal{}, // No consensus
			expectError:       false,
		},
		{
			name:      "mixed consensus",
			streamIDs: []uint32{1, 2},
			events: map[ocrcommon.OracleID]*datastreams.LLOStreamsTriggerEvent{
				1: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(201.5), // Different value
				}),
				3: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(202.5), // Different value
				}),
			},
			f:                 1,
			expectedTimestamp: testTime,
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
				1: createLLOEvent(t, testTime, map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				2: createLLOEvent(t, testTime.Add(time.Second), map[uint32]decimal.Decimal{
					1: decimal.NewFromFloat(100.5),
					2: decimal.NewFromFloat(200.5),
				}),
				3: createLLOEvent(t, testTime.Add(2*time.Second), map[uint32]decimal.Decimal{
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

			require.NoError(t, err)
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

	testStartTime := time.Now()
	remappedHex1 := "0x680084f7347baFfb5C323c2982dfC90e04F9F918"

	remapped1, err := hex.DecodeString(remappedHex1[2:])
	require.NoError(t, err)
	remappedHex2 := "0x00001237347baFfb5C323c1112dfC90e0789FFFF"
	remapped2, err := hex.DecodeString(remappedHex2[2:])
	require.NoError(t, err)
	remappedHex3 := "0xaaaa59b7347baFfb5C323c1112dfC90e0789FEDC"

	tests := []struct {
		name                 string
		config               datafeeds.LLOAggregatorConfig
		previousOutcome      *types.AggregationOutcome
		observations         map[ocrcommon.OracleID][]values.Value
		f                    int
		expectedShouldReport bool
		expectedStreamIDs    []uint32
		expectError          bool
		wantUpdates          []*datafeeds.EVMEncodableStreamUpdate
	}{

		{
			name: "update due to no previous outcome",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": datafeeds.FeedConfig{
						Deviation:     decimal.NewFromFloat(0.01).String(), // 1%
						Heartbeat:     3600,                                // 1 hour
						RemappedIDHex: remappedHex1,
					},
				},
			},
			observations: createObservations(t, testStartTime, map[uint32]decimal.Decimal{ //nolint: gosec // G115
				1: decimal.NewFromFloat(102.123), // 2% change, exceeds 1% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1},
			wantUpdates: []*datafeeds.EVMEncodableStreamUpdate{
				{
					StreamID:   1,
					Price:      datafeeds.DecimalToBigInt(decimal.NewFromFloat(102.123)),
					Timestamp:  uint32(testStartTime.Unix()), //nolint: gosec // G115
					RemappedID: remapped1,
				},
			},

			expectError: false,
		},

		{
			name: "update due to deviation",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						Deviation:     decimal.NewFromFloat(0.01).String(), // 1%
						Heartbeat:     3600,                                // 1 hour
						RemappedIDHex: remappedHex1,
					},
				},
			},
			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: testStartTime.Add(-10 * time.Minute).UnixNano(),
				},
			}),
			observations: createObservations(t, testStartTime, map[uint32]decimal.Decimal{ //nolint: gosec // G115
				1: decimal.NewFromFloat(102.00000000001), // 2% change, exceeds 1% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1},
			wantUpdates: []*datafeeds.EVMEncodableStreamUpdate{
				{
					StreamID:   1,
					Price:      datafeeds.DecimalToBigInt(decimal.NewFromFloat(102.00000000001)),
					Timestamp:  uint32(testStartTime.Unix()), //nolint: gosec // G115
					RemappedID: remapped1,
				},
			},

			expectError: false,
		},

		{
			name: "update due to heartbeat",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     300,                                // 5 min
						RemappedIDHex: remappedHex1,
					},
				},
			},
			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: testStartTime.Add(-6 * time.Minute).UnixNano(), // Over heartbeat
				},
			}),
			observations: createObservations(t, testStartTime, map[uint32]decimal.Decimal{ //nolint: gosec // G115
				1: decimal.NewFromFloat(101), // 1% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1},
			wantUpdates: []*datafeeds.EVMEncodableStreamUpdate{
				{
					StreamID:   1,
					Price:      datafeeds.DecimalToBigInt(decimal.NewFromFloat(101)),
					Timestamp:  uint32(testStartTime.Unix()), //nolint: gosec // G115
					RemappedID: remapped1,
				},
			},
			expectError: false,
		},

		{
			name: "no update needed",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     3600,                               // 1 hour
						RemappedIDHex: remappedHex1,
					},
				},
			},
			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromInt(100),
					timestamp: time.Now().Add(-30 * time.Minute).UnixNano(), // Under heartbeat
				},
			}),
			observations: createObservations(t, time.Now(), map[uint32]decimal.Decimal{ //nolint: gosec // G115
				1: decimal.NewFromInt(105), // 5% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: false, // No update needed
			expectedStreamIDs:    []uint32{},
			expectError:          false,
		},

		{
			name: "partial staleness optimization",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     3600,                               // 1 hour
						RemappedIDHex: remappedHex1,
					},
					"2": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     300,                                // 5 min
						RemappedIDHex: remappedHex2,
					},
					"3": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     300,                                // 5 min
						RemappedIDHex: remappedHex3,
					},
				},
				AllowedPartialStaleness: "0.2", // 20% allowed partial staleness
			},
			previousOutcome: createPreviousOutcome(t, map[uint32]struct {
				price     decimal.Decimal
				timestamp int64
			}{
				1: {
					price:     decimal.NewFromFloat(100),
					timestamp: testStartTime.Add(-50 * time.Minute).UnixNano(), // 83% of heartbeat, within 20% staleness
				},
				2: {
					price:     decimal.NewFromFloat(200),
					timestamp: testStartTime.Add(-6 * time.Minute).UnixNano(), // Over heartbeat
				},
				3: {
					price:     decimal.NewFromFloat(200),
					timestamp: testStartTime.Add(-1 * time.Minute).UnixNano(), // Under heartbeat, outside optimization
				},
			}),
			observations: createObservations(t, testStartTime, map[uint32]decimal.Decimal{ //nolint: gosec // G115
				1: decimal.NewFromFloat(105), // 5% change, under 10% threshold
				2: decimal.NewFromFloat(202), // 1% change, under 10% threshold
				3: decimal.NewFromFloat(205), // 2.5% change, under 10% threshold
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedStreamIDs:    []uint32{1, 2}, // Both update due to optimization
			wantUpdates: []*datafeeds.EVMEncodableStreamUpdate{
				{
					StreamID:   1,
					Price:      datafeeds.DecimalToBigInt(decimal.NewFromFloat(105)), //big.NewInt(105),
					Timestamp:  uint32(testStartTime.Unix()),                         //nolint: gosec // G115
					RemappedID: remapped1,
				},
				{
					StreamID:   2,
					Price:      datafeeds.DecimalToBigInt(decimal.NewFromFloat(202)), //big.NewInt(202),
					Timestamp:  uint32(testStartTime.Unix()),                         //nolint: gosec // G115
					RemappedID: remapped2,
				},
			},

			expectError: false,
		},

		{
			name: "empty observations",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						Deviation:     decimal.NewFromFloat(0.1).String(), // 10%
						Heartbeat:     3600,                               // 1 hour
						RemappedIDHex: remappedHex1,
					},
				},
			},
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
			cfgMap, err := tc.config.ToMap()
			require.NoError(t, err, "Failed to convert config to values.Map")
			aggregator, err := datafeeds.NewLLOAggregator(*cfgMap)
			require.NoError(t, err)

			outcome, err := aggregator.Aggregate(lggr, tc.previousOutcome, tc.observations, tc.f)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedShouldReport, outcome.ShouldReport)

			if outcome.ShouldReport {
				// Verify that the correct streams were updated
				reportedStreams, reports := extractUpdatedStreamIDs(t, outcome)
				assert.ElementsMatch(t, tc.expectedStreamIDs, reportedStreams)
				assert.Len(t, reports, len(tc.expectedStreamIDs))
				sort.Slice(reports, func(i, j int) bool {
					return reports[i].StreamID < reports[j].StreamID
				})
				sort.Slice(tc.wantUpdates, func(i, j int) bool {
					return tc.wantUpdates[i].StreamID < tc.wantUpdates[j].StreamID
				})
				for i, report := range reports {
					assert.Equal(t, tc.wantUpdates[i].StreamID, report.StreamID)
					assert.Equal(t, tc.wantUpdates[i].Price, report.Price)
					assert.Equal(t, tc.wantUpdates[i].Timestamp, report.Timestamp)
				}
			}
		})
	}
}

func TestLLOAggregatorConfig_RoundTrip(t *testing.T) {
	testCases := []struct {
		name   string
		config datafeeds.LLOAggregatorConfig
	}{
		{
			name: "typical config with multiple streams",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						RemappedIDHex: "0x680084f7347baFfb5C323c2982dfC90e04F9F918",
						Deviation:     "0.01",
						Heartbeat:     3600,
					},
					"2": {
						RemappedIDHex: "0x00001237347baFfb5C323c1112dfC90e0789FFFF",
						Deviation:     "0.02",
						Heartbeat:     1800,
					},
					"42": {
						RemappedIDHex: "0xF8D170535513B67Ce18aF7A45E9a1F1A93c0F9ac",
						Deviation:     "0.005",
						Heartbeat:     7200,
					},
				},
				AllowedPartialStaleness: "0.2",
			},
		},
		{
			name: "config with single stream",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						RemappedIDHex: "0x680084f7347baFfb5C323c2982dfC90e04F9F918",
						Deviation:     "0.01",
						Heartbeat:     3600,
					},
				},
				AllowedPartialStaleness: "0.1",
			},
		},
		{
			name: "config with empty staleness",
			config: datafeeds.LLOAggregatorConfig{
				Streams: map[string]datafeeds.FeedConfig{
					"1": {
						RemappedIDHex: "0x680084f7347baFfb5C323c2982dfC90e04F9F918",
						Deviation:     "0.01",
						Heartbeat:     3600,
					},
				},
				AllowedPartialStaleness: "",
			},
		},
		{
			name: "config with empty streams",
			config: datafeeds.LLOAggregatorConfig{
				Streams:                 map[string]datafeeds.FeedConfig{},
				AllowedPartialStaleness: "0.2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Convert original config to values.Map
			configMap, err := tc.config.ToMap()
			require.NoError(t, err, "ToMap should not error")
			require.NotNil(t, configMap, "ToMap should return non-nil map")

			// Step 2: Convert values.Map back to config
			roundTrippedConfig, err := datafeeds.NewLLOConfig(*configMap)
			require.NoError(t, err, "NewLLOConfig should not error")

			// Step 3: Compare original and round-tripped configs
			// We need to compare each field specifically since the derived fields won't be compared correctly
			assert.Equal(t, len(tc.config.Streams), len(roundTrippedConfig.Streams),
				"Number of streams should match")

			// Compare streams
			for streamID, origFeed := range tc.config.Streams {
				roundTrippedFeed, exists := roundTrippedConfig.Streams[streamID]
				assert.True(t, exists, "Stream %s should exist in round-tripped config", streamID)

				if exists {
					assert.Equal(t, origFeed.RemappedIDHex, roundTrippedFeed.RemappedIDHex,
						"RemappedIDHex should match for stream %s", streamID)
					assert.Equal(t, origFeed.Deviation, roundTrippedFeed.Deviation,
						"DeviationString should match for stream %s", streamID)
					assert.Equal(t, origFeed.Heartbeat, roundTrippedFeed.Heartbeat,
						"Heartbeat should match for stream %s", streamID)
				}
			}

			// Compare staleness
			assert.Equal(t, tc.config.AllowedPartialStaleness, roundTrippedConfig.AllowedPartialStaleness,
				"AllowedPartialStalenessStr should match")
		})
	}
}

// Helper functions

func createLLOEvent(t *testing.T, obs time.Time, prices map[uint32]decimal.Decimal) *datastreams.LLOStreamsTriggerEvent {
	event := &datastreams.LLOStreamsTriggerEvent{
		ObservationTimestampNanoseconds: uint64(obs.UnixNano()), //nolint: gosec // G115
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
	timestamp int64 // UnixNano
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

func createObservations(t *testing.T, ts time.Time, prices map[uint32]decimal.Decimal) map[ocrcommon.OracleID][]values.Value {
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

func extractUpdatedStreamIDs(t *testing.T, outcome *types.AggregationOutcome) ([]uint32, []*datafeeds.EVMEncodableStreamUpdate) {
	streamIDs, reports, err := processOutcome(outcome)
	require.NoError(t, err)

	return streamIDs, reports
}

func processOutcome(outcome *types.AggregationOutcome) ([]uint32, []*datafeeds.EVMEncodableStreamUpdate, error) {
	// TODOD here add the decoder of the slice
	decodedMap, err := values.FromMapValueProto(outcome.EncodableOutcome)
	if err != nil {
		return nil, nil, err
	}

	reportsAny, ok := decodedMap.Underlying[datafeeds.TopLevelListOutputFieldName]
	if !ok {
		return nil, nil, fmt.Errorf("missing field %s", datafeeds.TopLevelListOutputFieldName)
	}

	var reportsList []*datafeeds.EVMEncodableStreamUpdate // each element is a WrappableUpdate
	err = reportsAny.UnwrapTo(&reportsList)
	if err != nil {
		return nil, nil, err
	}

	streamIDs := make([]uint32, 0, len(reportsList))
	for _, reportAny := range reportsList {
		streamIDs = append(streamIDs, reportAny.StreamID)
	}

	return streamIDs, reportsList, nil
}
