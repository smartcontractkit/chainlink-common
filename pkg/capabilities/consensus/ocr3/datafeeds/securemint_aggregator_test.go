package datafeeds_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	// Feed IDs that actually contain 'eth' and 'btc' as substrings
	ethFeedID = datastreams.FeedID("0x0000000000000000000000000000000000006574680000000000000000000000") // contains 'eth'
	btcFeedID = datastreams.FeedID("0x0000000000000000000000000000000000006274630000000000000000000000") // contains 'btc'
	ethReport = []byte("eth_report")
	btcReport = []byte("btc_report")
)

func TestSecureMintAggregator_Aggregate(t *testing.T) {
	lggr := logger.Test(t)

	tests := []struct {
		name                 string
		config               datafeeds.SecureMintAggregatorConfig
		previousOutcome      *types.AggregationOutcome
		observations         map[ocrcommon.OracleID][]values.Value
		f                    int
		expectedShouldReport bool
		expectedFeedID       string
		expectError          bool
		errorContains        string
	}{
		{
			name: "successful eth report extraction",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
			observations: createSecureMintObservations(t, []datastreams.FeedReport{
				{
					FeedID:               ethFeedID.String(),
					ObservationTimestamp: 1000,
					BenchmarkPrice:       []byte{100},
					FullReport:           ethReport,
				},
				{
					FeedID:               btcFeedID.String(),
					ObservationTimestamp: 1100,
					BenchmarkPrice:       []byte{200},
					FullReport:           btcReport,
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedFeedID:       ethFeedID.String(),
			expectError:          false,
		},
		{
			name: "case insensitive eth search",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "ETH",
			},
			observations: createSecureMintObservations(t, []datastreams.FeedReport{
				{
					FeedID:               "0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292", // contains "eth"
					ObservationTimestamp: 1000,
					BenchmarkPrice:       []byte{100},
					FullReport:           ethReport,
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedFeedID:       "0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292",
			expectError:          false,
		},
		{
			name: "no eth report found",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
			observations: createSecureMintObservations(t, []datastreams.FeedReport{
				{
					FeedID:               btcFeedID.String(),
					ObservationTimestamp: 1100,
					BenchmarkPrice:       []byte{200},
					FullReport:           btcReport,
				},
			}),
			f:             1,
			expectError:   true,
			errorContains: "no eth report found",
		},
		{
			name: "empty observations",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
			observations:  map[ocrcommon.OracleID][]values.Value{},
			f:             1,
			expectError:   true,
			errorContains: "empty observation",
		},
		{
			name: "custom target feed ID",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "btc",
			},
			observations: createSecureMintObservations(t, []datastreams.FeedReport{
				{
					FeedID:               ethFeedID.String(),
					ObservationTimestamp: 1000,
					BenchmarkPrice:       []byte{100},
					FullReport:           ethReport,
				},
				{
					FeedID:               btcFeedID.String(),
					ObservationTimestamp: 1100,
					BenchmarkPrice:       []byte{200},
					FullReport:           btcReport,
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedFeedID:       btcFeedID.String(),
			expectError:          false,
		},
		{
			name: "partial match in feed ID",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
			observations: createSecureMintObservations(t, []datastreams.FeedReport{
				{
					FeedID:               "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // no eth
					ObservationTimestamp: 1000,
					BenchmarkPrice:       []byte{100},
					FullReport:           []byte("other_report"),
				},
				{
					FeedID:               "0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292", // contains eth
					ObservationTimestamp: 1100,
					BenchmarkPrice:       []byte{200},
					FullReport:           ethReport,
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectedFeedID:       "0x0001013ebd4ed3f5889fb5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292",
			expectError:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock report codec
			codec := mocks.NewReportCodec(t)

			// Set up mock expectations
			for _, nodeObservations := range tc.observations {
				if len(nodeObservations) > 0 {
					// Extract the reports that would be returned by the codec
					triggerEvent := &datastreams.StreamsTriggerEvent{}
					err := nodeObservations[0].UnwrapTo(triggerEvent)
					require.NoError(t, err)

					codec.On("Unwrap", nodeObservations[0]).Return(triggerEvent.Payload, nil)
				}
			}

			// Create config map
			cfgMap, err := tc.config.ToMap()
			require.NoError(t, err, "Failed to convert config to values.Map")

			// Create aggregator
			aggregator, err := datafeeds.NewSecureMintAggregator(*cfgMap, codec)
			require.NoError(t, err)

			// Run aggregation
			outcome, err := aggregator.Aggregate(lggr, tc.previousOutcome, tc.observations, tc.f)

			// Check error expectations
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedShouldReport, outcome.ShouldReport)

			if outcome.ShouldReport {
				// Verify the output structure
				val, err := values.FromMapValueProto(outcome.EncodableOutcome)
				require.NoError(t, err)

				topLevelMap, err := val.Unwrap()
				require.NoError(t, err)
				mm, ok := topLevelMap.(map[string]any)
				require.True(t, ok)

				// Check that we have the expected reports
				reportsList, ok := mm[datafeeds.TopLevelListOutputFieldName].([]any)
				require.True(t, ok)
				require.Len(t, reportsList, 1)

				// Check the first (and only) report
				report, ok := reportsList[0].(map[string]any)
				require.True(t, ok)

				// Verify feed ID
				feedIDBytes, ok := report[datafeeds.FeedIDOutputFieldName].([]byte)
				require.True(t, ok)
				require.Equal(t, tc.expectedFeedID, string(feedIDBytes))

				// Verify other fields exist
				_, ok = report[datafeeds.RawReportOutputFieldName].([]byte)
				require.True(t, ok)

				_, ok = report[datafeeds.PriceOutputFieldName].([]byte)
				require.True(t, ok)

				_, ok = report[datafeeds.TimestampOutputFieldName].(int64)
				require.True(t, ok)

				_, ok = report[datafeeds.RemappedIDOutputFieldName].([]byte)
				require.True(t, ok)
			}

			// Verify mock expectations
			codec.AssertExpectations(t)
		})
	}
}

func TestSecureMintAggregatorConfig_RoundTrip(t *testing.T) {
	testCases := []struct {
		name   string
		config datafeeds.SecureMintAggregatorConfig
	}{
		{
			name: "default eth config",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
		},
		{
			name: "custom target feed ID",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "btc",
			},
		},
		{
			name: "uppercase target feed ID",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "ETH",
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
			roundTrippedConfig, err := datafeeds.NewSecureMintConfig(*configMap)
			require.NoError(t, err, "NewSecureMintConfig should not error")

			// Step 3: Compare original and round-tripped configs
			assert.Equal(t, tc.config.TargetFeedID, roundTrippedConfig.TargetFeedID,
				"TargetFeedID should match")
		})
	}
}

func TestSecureMintAggregatorConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      datafeeds.SecureMintAggregatorConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "eth",
			},
			expectError: false,
		},
		{
			name: "empty target feed ID",
			config: datafeeds.SecureMintAggregatorConfig{
				TargetFeedID: "",
			},
			expectError: true,
			errorMsg:    "targetFeedId is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configMap, err := tc.config.ToMap()
			require.NoError(t, err)

			_, err = datafeeds.NewSecureMintAggregator(*configMap, nil)
			if tc.expectError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions

func createSecureMintObservations(t *testing.T, reports []datastreams.FeedReport) map[ocrcommon.OracleID][]values.Value {
	observations := make(map[ocrcommon.OracleID][]values.Value)

	// Create trigger event with the reports
	triggerEvent := &datastreams.StreamsTriggerEvent{
		Payload: reports,
		Metadata: datastreams.Metadata{
			Signers:               [][]byte{newSigner(t), newSigner(t)},
			MinRequiredSignatures: 1,
		},
		Timestamp: 1000,
	}

	// Create three observations with identical data to ensure f+1 consensus
	for i := ocrcommon.OracleID(1); i <= 3; i++ {
		val, err := values.Wrap(triggerEvent)
		require.NoError(t, err)

		observations[i] = []values.Value{val}
	}

	return observations
}
