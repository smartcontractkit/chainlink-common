package datafeeds

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr3types "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

var (
	// Test chain selectors
	ethSepoliaChainSelector = chainSelector(16015286601757825753) // Ethereum Sepolia testnet
	bnbTestnetChainSelector = chainSelector(13264668187771770619) // Binance Smart Chain testnet
	solDevnetChainSelector  = chainSelector(16423721717087811551) // Solana devnet
)

func TestSecureMintAggregator_Aggregate(t *testing.T) {
	lggr := logger.Test(t)

	type tcase struct {
		name                 string
		chainSelector        string
		dataID               string
		solAccounts          [][32]byte
		previousOutcome      *types.AggregationOutcome
		observations         map[ocrcommon.OracleID][]values.Value
		f                    int
		expectedShouldReport bool
		expectError          bool
		errorContains        string
		shouldReportAssertFn func(t *testing.T, tc tcase, topLevelMap map[string]any)
	}
	acc1 := [32]byte{4, 5, 6}
	acc2 := [32]byte{3, 2, 1}

	ethReportAssertFn := func(t *testing.T, tc tcase, topLevelMap map[string]any) {
		// Check that we have the expected reports
		reportsList, ok := topLevelMap[TopLevelListOutputFieldName].([]any)
		require.True(t, ok)
		assert.Len(t, reportsList, 1)

		// Check the first (and only) report
		report, ok := reportsList[0].(map[string]any)
		assert.True(t, ok)

		// Verify dataID
		dataIDBytes, ok := report[DataIDOutputFieldName].([]byte)
		assert.True(t, ok, "expected dataID to be []byte but got %T", report[DataIDOutputFieldName])
		assert.Len(t, dataIDBytes, 16)
		assert.Equal(t, tc.dataID, "0x"+hex.EncodeToString(dataIDBytes))

		// Verify other fields exist
		answer, ok := report[AnswerOutputFieldName].(*big.Int)
		assert.True(t, ok)
		assert.NotNil(t, answer)

		timestamp := report[TimestampOutputFieldName].(int64)
		assert.Equal(t, int64(1000), timestamp)
	}

	solReportAssertFn := func(t *testing.T, tc tcase, topLevelMap map[string]any) {
		// Check that we have the expected reports
		reportsList, ok := topLevelMap[TopLevelPayloadListFieldName].([]any)
		assert.True(t, ok)
		assert.Len(t, reportsList, 1)

		// Check that we have expected account hash
		accHash, ok := topLevelMap[TopLevelAccountCtxHashFieldName].([]byte)
		require.True(t, ok, "expected account hash to be []byte but got %T", topLevelMap[TopLevelAccountCtxHashFieldName])
		require.Len(t, accHash, 32)
		expHash := sha256.Sum256(append(acc1[:], acc2[:]...))
		assert.Equal(t, expHash, ([32]byte)(accHash))

		// Check the first (and only) report
		report, ok := reportsList[0].(map[string]any)
		assert.True(t, ok)
		// Verify dataID
		dataIDBytes, ok := report[SolDataIDOutputFieldName].([]byte)
		assert.True(t, ok, "expected dataID to be []byte but got %T", report[DataIDOutputFieldName])
		assert.Len(t, dataIDBytes, 16)
		assert.Equal(t, tc.dataID, "0x"+hex.EncodeToString(dataIDBytes))

		// Verify other fields exist
		answer, ok := report[SolAnswerOutputFieldName].(*big.Int)
		assert.True(t, ok)
		assert.NotNil(t, answer)

		timestamp := report[SolTimestampOutputFieldName].(int64)
		assert.Equal(t, int64(1000), timestamp)
	}

	tests := []tcase{
		{
			name:          "successful eth report extraction",
			chainSelector: "16015286601757825753",
			dataID:        "0x01c508f42b0201320000000000000000",
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: ethSepoliaChainSelector,
					seqNr:         10,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        10,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
				{
					chainSelector: bnbTestnetChainSelector,
					seqNr:         11,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 2, 31: 3},
						SeqNr:        11,
						Block:        1100,
						Mintable:     big.NewInt(200),
					},
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectError:          false,
			shouldReportAssertFn: ethReportAssertFn,
		},
		{
			name:          "no matching chain selector found",
			chainSelector: "16015286601757825753",
			dataID:        "0x01c508f42b0201320000000000000000",
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: bnbTestnetChainSelector,
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
			shouldReportAssertFn: ethReportAssertFn,
		},
		{
			name:          "no observations",
			chainSelector: "16015286601757825753",
			dataID:        "0x01c508f42b0201320000000000000000",
			observations:  map[ocrcommon.OracleID][]values.Value{},
			f:             1,
			expectError:   true,
			errorContains: "no observations",
		},
		{
			name:          "successful sol report extraction",
			chainSelector: "16423721717087811551", // solana devnet
			dataID:        "0x01c508f42b0201320000000000000000",
			solAccounts:   [][32]byte{acc1, acc2},
			observations: createSecureMintObservations(t, []ocrTriggerEventData{
				{
					chainSelector: solDevnetChainSelector,
					seqNr:         10,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 1, 31: 2},
						SeqNr:        10,
						Block:        1000,
						Mintable:     big.NewInt(99),
					},
				},
				{
					chainSelector: bnbTestnetChainSelector,
					seqNr:         11,
					report: &secureMintReport{
						ConfigDigest: ocr2types.ConfigDigest{0: 2, 31: 3},
						SeqNr:        11,
						Block:        1100,
						Mintable:     big.NewInt(200),
					},
				},
			}),
			f:                    1,
			expectedShouldReport: true,
			expectError:          false,
			shouldReportAssertFn: solReportAssertFn,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create aggregator
			rawCfg := map[string]any{
				"targetChainSelector": tc.chainSelector,
				"dataID":              tc.dataID,
			}
			if len(tc.solAccounts) > 0 {
				accountMetaSlice := make(solana.AccountMetaSlice, len(tc.solAccounts))
				for i, acc := range tc.solAccounts {
					accountMetaSlice[i] = &solana.AccountMeta{PublicKey: acc}
				}

				rawCfg["solana"] = map[string]any{
					"remaining_accounts": accountMetaSlice,
				}
			}

			configMap, err := values.WrapMap(rawCfg)
			require.NoError(t, err)
			aggregator, err := NewSecureMintAggregator(*configMap)
			require.NoError(t, err)

			// Run aggregation
			outcome, err := aggregator.Aggregate(lggr, tc.previousOutcome, tc.observations, tc.f)

			// Check error expectations
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedShouldReport, outcome.ShouldReport)

			if outcome.ShouldReport {
				// Verify the output structure matches the feeds aggregator format
				val, err := values.FromMapValueProto(outcome.EncodableOutcome)
				require.NoError(t, err)

				topLevelMap, err := val.Unwrap()
				require.NoError(t, err)
				mm, ok := topLevelMap.(map[string]any)
				require.True(t, ok)

				tc.shouldReportAssertFn(t, tc, mm)
			}
		})
	}
}

func TestSecureMintAggregatorConfig_Validation(t *testing.T) {
	acc1 := [32]byte{4, 5, 6}

	tests := []struct {
		name                  string
		chainSelector         string
		dataID                string
		solanaAccounts        solana.AccountMetaSlice
		expectedChainSelector chainSelector
		expectedDataID        [16]byte
		expectError           bool
		errorMsg              string
	}{
		{
			name:                  "valid chain selector, dataID and solana accounts",
			chainSelector:         "1",
			dataID:                "0x01c508f42b0201320000000000000000",
			solanaAccounts:        solana.AccountMetaSlice{&solana.AccountMeta{PublicKey: acc1, IsWritable: true, IsSigner: false}},
			expectedChainSelector: 1,
			expectedDataID:        [16]byte{0x01, 0xc5, 0x08, 0xf4, 0x2b, 0x02, 0x01, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:           false,
		},
		{
			name:                  "large chain selector",
			chainSelector:         "16015286601757825753", // ethereum-testnet-sepolia
			dataID:                "0x01c508f42b0201320000000000000000",
			expectedChainSelector: 16015286601757825753,
			expectedDataID:        [16]byte{0x01, 0xc5, 0x08, 0xf4, 0x2b, 0x02, 0x01, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:           false,
		},
		{
			name:                  "dataID without 0x prefix",
			chainSelector:         "1",
			dataID:                "01c508f42b0201320000000000000000",
			expectedChainSelector: 1,
			expectedDataID:        [16]byte{0x01, 0xc5, 0x08, 0xf4, 0x2b, 0x02, 0x01, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectError:           false},
		{
			name:          "invalid chain selector",
			chainSelector: "invalid",
			expectError:   true,
			errorMsg:      "invalid chain selector",
		},
		{
			name:          "negative chain selector",
			chainSelector: "-1",
			dataID:        "0x01c508f42b0201320000000000000000",
			expectError:   true,
			errorMsg:      "invalid chain selector",
		},
		{
			name:          "invalid dataID",
			chainSelector: "1",
			dataID:        "invalid_data_id",
			expectError:   true,
			errorMsg:      "invalid dataID",
		},
		{
			name:          "dataID too short",
			chainSelector: "1",
			dataID:        "0x0000",
			expectError:   true,
			errorMsg:      "dataID must be 16 bytes",
		},
		{
			name:          "dataID with odd length",
			chainSelector: "1",
			dataID:        "0x0",
			expectError:   true,
			errorMsg:      "odd length hex string",
		},
		{
			name:          "dataID too long",
			chainSelector: "1",
			dataID:        "0x01111111111111111111111111111111111111111111",
			expectError:   true,
			errorMsg:      "dataID must be 16 bytes",
		},
		{
			name:           "solana account context with invalid public key",
			chainSelector:  "1",
			dataID:         "0x01c508f42b0201320000000000000000",
			solanaAccounts: solana.AccountMetaSlice{&solana.AccountMeta{PublicKey: [32]byte{}}},
			expectError:    true,
			errorMsg:       "solana account context public key must not be all zeros",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawCfg := map[string]any{
				"targetChainSelector": tt.chainSelector,
				"dataID":              tt.dataID,
			}
			if len(tt.solanaAccounts) > 0 {
				rawCfg["solana"] = map[string]any{
					"remaining_accounts": tt.solanaAccounts,
				}
			}

			configMap, err := values.WrapMap(rawCfg)
			require.NoError(t, err)

			aggregator, err := NewSecureMintAggregator(*configMap)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedChainSelector, aggregator.(*SecureMintAggregator).config.TargetChainSelector)
			assert.Equal(t, tt.expectedDataID, aggregator.(*SecureMintAggregator).config.DataID)
			assert.Equal(t, tt.solanaAccounts, aggregator.(*SecureMintAggregator).config.Solana.AccountContext)
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
	assert.Equal(t, expectedMax, maxMintableEVM, "maxMintable should be 2^128 - 1")

	// Verify it's exactly 128 bits
	bitLen := maxMintableEVM.BitLen()
	assert.Equal(t, 128, bitLen, "maxMintable should be exactly 128 bits")
}
