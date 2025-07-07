package datafeeds

import (
	"encoding/binary"
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

				// Verify dataID
				dataIDBytes, ok := report[FeedIDOutputFieldName].([]byte)
				require.True(t, ok)
				// Should be 0x04 + chain selector as bytes + right padded with 0s
				var expectedChainSelectorBytes [32]byte
				expectedChainSelectorBytes[0] = 0x04
				binary.BigEndian.PutUint64(expectedChainSelectorBytes[1:], uint64(tc.expectedChainSelector))
				for i := 9; i < 32; i++ {
					expectedChainSelectorBytes[i] = 0x00
				}
				require.Equal(t, expectedChainSelectorBytes[:], dataIDBytes)

				// Verify other fields exist
				price, ok := report[PriceOutputFieldName].(*big.Int)
				require.True(t, ok)
				require.NotNil(t, price)

				timestamp := report[TimestampOutputFieldName].(int64)
				require.Equal(t, int64(1000), timestamp)
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

func TestPackSecureMintReportForIntoUint224(t *testing.T) {
	tests := []struct {
		name        string
		mintable    *big.Int
		blockNumber uint64
		expected    *big.Int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "zero values",
			mintable:    big.NewInt(0),
			blockNumber: 0,
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "small positive values",
			mintable:    big.NewInt(100),
			blockNumber: 12345,
			expected:    new(big.Int).Add(big.NewInt(100), new(big.Int).Lsh(big.NewInt(12345), 128)),
			expectError: false,
		},
		{
			name:        "maximum mintable value (2^128 - 1)",
			mintable:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)),
			blockNumber: 999999,
			expected: new(big.Int).Add(
				new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)),
				new(big.Int).Lsh(big.NewInt(999999), 128),
			),
			expectError: false,
		},
		{
			name:        "large block number",
			mintable:    big.NewInt(500),
			blockNumber: 18446744073709551615, // max uint64
			expected:    new(big.Int).Add(big.NewInt(500), new(big.Int).Lsh(new(big.Int).SetUint64(18446744073709551615), 128)),
			expectError: false,
		},
		{
			name:        "mintable exceeds 128 bits",
			mintable:    new(big.Int).Lsh(big.NewInt(1), 128), // 2^128
			blockNumber: 1000,
			expectError: true,
			errorMsg:    "mintable amount",
		},
		{
			name:        "very large mintable that exceeds 128 bits",
			mintable:    new(big.Int).Lsh(big.NewInt(1), 256), // 2^256
			blockNumber: 1000,
			expectError: true,
			errorMsg:    "mintable amount",
		},
		{
			name:        "nil mintable",
			mintable:    nil,
			blockNumber: 1000,
			expectError: true,
			errorMsg:    "mintable cannot be nil",
		},
		{
			name:        "bit pattern verification - mintable 1, block 1",
			mintable:    big.NewInt(1),
			blockNumber: 1,
			expected:    new(big.Int).Add(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 128)),
			expectError: false,
		},
		{
			name:        "bit pattern verification - mintable 0xFFFFFFFF, block 0xFFFFFFFF",
			mintable:    big.NewInt(0xFFFFFFFF),
			blockNumber: 0xFFFFFFFF,
			expected:    new(big.Int).Add(big.NewInt(0xFFFFFFFF), new(big.Int).Lsh(big.NewInt(0xFFFFFFFF), 128)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := packSecureMintReportIntoUint224ForEVM(tt.mintable, tt.blockNumber)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			if tt.expected != nil {
				assert.Equal(t, tt.expected, result)
			}

			// Additional validation: ensure the result fits in 224 bits
			maxUint224 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 224), big.NewInt(1))
			assert.True(t, result.Cmp(maxUint224) <= 0, "result should fit in 224 bits")

			// Verify bit layout if we have expected values and not a large block number
			if tt.expected != nil {
				verifyBitLayout(t, result, tt.mintable, tt.blockNumber)
			}
		})
	}
}

func TestPackSecureMintReportForIntoUint224_EdgeCases(t *testing.T) {
	// Test edge cases and boundary conditions
	tests := []struct {
		name        string
		mintable    *big.Int
		blockNumber uint64
		expectError bool
	}{
		{
			name:        "mintable exactly at 128-bit boundary",
			mintable:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)), // 2^128 - 1
			blockNumber: 1000,
			expectError: false,
		},
		{
			name:        "mintable one over 128-bit boundary",
			mintable:    new(big.Int).Lsh(big.NewInt(1), 128), // 2^128
			blockNumber: 1000,
			expectError: true,
		},
		{
			name:        "block number at max uint64",
			mintable:    big.NewInt(100),
			blockNumber: 0xFFFFFFFFFFFFFFFF,
			expectError: false,
		},
		{
			name:        "both values at maximum",
			mintable:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)),
			blockNumber: 0xFFFFFFFFFFFFFFFF,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := packSecureMintReportIntoUint224ForEVM(tt.mintable, tt.blockNumber)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify the result is within uint224 bounds
			maxUint224 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 224), big.NewInt(1))
			assert.True(t, result.Cmp(maxUint224) <= 0, "result should fit in 224 bits")
		})
	}
}

// verifyBitLayout verifies that the packed result has the correct bit layout
// mintable should be in bits 0-127, block number in bits 128-191
func verifyBitLayout(t *testing.T, packed *big.Int, mintable *big.Int, blockNumber uint64) {
	// Extract mintable from lower 128 bits
	mintableMask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	extractedMintable := new(big.Int).And(packed, mintableMask)

	// Extract block number from bits 128-191
	blockNumberMask := new(big.Int).Lsh(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)), 128)
	extractedBlockNumber := new(big.Int).And(packed, blockNumberMask)
	extractedBlockNumber = new(big.Int).Rsh(extractedBlockNumber, 128)

	// Always use big.NewInt(0) for zero-value mintable
	expectedMintable := mintable
	if mintable == nil || (mintable != nil && mintable.Sign() == 0) {
		expectedMintable = big.NewInt(0)
	}

	assert.Equal(t, expectedMintable, extractedMintable, "mintable bits should match")
	assert.Equal(t, new(big.Int).SetUint64(blockNumber), extractedBlockNumber, "block number bits should match")
}

func TestMaxMintableConstant(t *testing.T) {
	// Verify the maxMintable constant is correctly defined
	expectedMax := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	assert.Equal(t, expectedMax, maxMintable, "maxMintable should be 2^128 - 1")

	// Verify it's exactly 128 bits
	bitLen := maxMintable.BitLen()
	assert.Equal(t, 128, bitLen, "maxMintable should be exactly 128 bits")
}
