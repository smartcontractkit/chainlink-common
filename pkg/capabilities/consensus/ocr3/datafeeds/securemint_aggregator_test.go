package datafeeds

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr3types "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	// Test chain selectors
	ethChainSelector = chainSelector(1)
	bnbChainSelector = chainSelector(56)
)

func TestSecureMintAggregator_Aggregate(t *testing.T) {
	lggr := logger.Test(t)

	tests := []struct {
		name                  string
		config                SecureMintAggregatorConfig
		previousOutcome       *types.AggregationOutcome
		observations          map[ocrcommon.OracleID][]values.Value
		f                     int
		expectedShouldReport  bool
		expectedChainSelector chainSelector
		expectError           bool
		errorContains         string
	}{
		{
			name: "successful eth report extraction",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: ethChainSelector,
					seqNr:         10,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        10,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
				{
					chainSelector: bnbChainSelector,
					seqNr:         11,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 2, 31: 3},
						SeqNr:        11,
						Block:        1100,
						Mintable:     big.NewInt(200),
					},
				},
			}),
			f:                     1,
			expectedShouldReport:  true,
			expectedChainSelector: ethChainSelector,
			expectError:           false,
		},
		{
			name: "no matching chain selector found",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: bnbChainSelector,
					seqNr:         10,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        10,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
			}),
			f:                    1,
			expectError:          false,
			expectedShouldReport: false,
		},
		{
			name: "sequence number too low",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			previousOutcome: &types.AggregationOutcome{
				LastSeenAt: 10, // Previous sequence number
			},
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: ethChainSelector,
					seqNr:         9, // Lower than previous
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        9,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
			}),
			f:             1,
			expectError:   true,
			errorContains: "sequence number too low",
		},
		{
			name: "no observations",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			observations:  map[ocrcommon.OracleID][]values.Value{},
			f:             1,
			expectError:   true,
			errorContains: "no observations",
		},
		{
			name: "sequence number equal to previous (should be ignored)",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			previousOutcome: &types.AggregationOutcome{
				LastSeenAt: 10, // Previous sequence number
			},
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: ethChainSelector,
					seqNr:         10, // Equal to previous
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        10,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
			}),
			f:             1,
			expectError:   true,
			errorContains: "sequence number too low",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create config map
			cfgMap, err := tc.config.ToMap()
			require.NoErrorf(t, err, "Failed to convert config %+v to values.Map", tc.config)

			// Create aggregator
			aggregator, err := NewSecureMintAggregator(*cfgMap)
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
				// Verify the output structure matches the feeds aggregator format
				val, err := values.FromMapValueProto(outcome.EncodableOutcome)
				require.NoError(t, err)

				topLevelMap, err := val.Unwrap()
				require.NoError(t, err)
				mm, ok := topLevelMap.(map[string]any)
				require.True(t, ok)

				// Check that we have the expected reports
				reportsList, ok := mm[TopLevelListOutputFieldName].([]any)
				require.True(t, ok)
				require.Len(t, reportsList, 1)

				// Check the first (and only) report
				report, ok := reportsList[0].(map[string]any)
				require.True(t, ok)

				// Verify feed ID (should be the chain selector as bytes)
				feedIDBytes, ok := report[FeedIDOutputFieldName].([]byte)
				require.True(t, ok)
				expectedChainSelectorBytes := big.NewInt(int64(tc.expectedChainSelector)).Bytes()
				require.Equal(t, expectedChainSelectorBytes, feedIDBytes)

				// Verify other fields exist
				_, ok = report[RawReportOutputFieldName].([]byte)
				require.True(t, ok)

				_, ok = report[PriceOutputFieldName].([]byte)
				require.True(t, ok)

				_, ok = report[TimestampOutputFieldName].(int64)
				require.True(t, ok)

				_, ok = report[RemappedIDOutputFieldName].([]byte)
				require.True(t, ok)
			}
		})
	}
}

func TestSecureMintAggregatorConfig_RoundTrip(t *testing.T) {
	testCases := []struct {
		name   string
		config SecureMintAggregatorConfig
	}{
		{
			name: "default eth config",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
		},
		{
			name: "custom target chain selector",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: bnbChainSelector,
			},
		},
		{
			name: "large chain selector",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: 999999,
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
			roundTrippedConfig, err := NewSecureMintConfig(*configMap)
			require.NoError(t, err, "NewSecureMintConfig should not error")

			// Step 3: Compare original and round-tripped configs
			assert.Equal(t, tc.config.TargetChainSelector, roundTrippedConfig.TargetChainSelector,
				"TargetChainSelector should match")
		})
	}
}

func TestSecureMintAggregatorConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      SecureMintAggregatorConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: ethChainSelector,
			},
			expectError: false,
		},
		{
			name: "zero target chain selector",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: 0,
			},
			expectError: true,
			errorMsg:    "targetChainSelector is required",
		},
		{
			name: "negative chain selector",
			config: SecureMintAggregatorConfig{
				TargetChainSelector: -1,
			},
			expectError: true,
			errorMsg:    "targetChainSelector is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configMap, err := tc.config.ToMap()
			require.NoError(t, err)

			_, err = NewSecureMintAggregator(*configMap)
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

// Helper types and functions

type ocrTriggerEventData struct {
	chainSelector chainSelector
	seqNr         uint64
	report        *secureMintReport
}

func createSecureMintObservations(t *testing.T, events []ocrTriggerEventData) map[ocrcommon.OracleID][]values.Value {
	observations := make(map[ocrcommon.OracleID][]values.Value)

	// Create three observations with identical data to ensure f+1 consensus
	for i := ocrcommon.OracleID(1); i <= 3; i++ {
		// For each oracle, create observations for all events
		var oracleObservations []values.Value
		for _, event := range events {
			// Create the ReportWithInfo
			ocr3Report := &ocr3types.ReportWithInfo[chainSelector]{
				Report: createReportBytes(t, event.report),
				Info:   event.chainSelector,
			}

			// Marshal the ReportWithInfo
			jsonReport, err := json.Marshal(ocr3Report)
			require.NoError(t, err)

			// Create the OCRTriggerEvent
			triggerEvent := &capabilities.OCRTriggerEvent{
				ConfigDigest: event.report.ConfigDigest[:],
				SeqNr:        event.seqNr,
				Report:       jsonReport,
				Sigs: []capabilities.OCRAttributedOnchainSignature{
					{
						Signature: []byte("signature1"),
						Signer:    1,
					},
					{
						Signature: []byte("signature2"),
						Signer:    2,
					},
				},
			}

			// Wrap in values.Value
			val, err := values.Wrap(triggerEvent)
			require.NoError(t, err)

			oracleObservations = append(oracleObservations, val)
		}

		observations[i] = oracleObservations
	}

	return observations
}

func createReportBytes(t *testing.T, report *secureMintReport) []byte {
	reportBytes, err := json.Marshal(report)
	require.NoError(t, err)
	return reportBytes
}
