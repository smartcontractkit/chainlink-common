package ccipocr3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

func TestExecutePluginReportConversions(t *testing.T) {
	t.Run("Empty Report", func(t *testing.T) {
		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		assert.Empty(t, pb.ChainReports)

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Single Chain Report with Empty OffchainTokenData", func(t *testing.T) {
		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages:            []ccipocr3.Message{},
					OffchainTokenData:   [][][]byte{}, // Empty 3D array
					Proofs:              nil,
					ProofFlagBits:       ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		assert.Equal(t, uint64(12345), chainReport.SourceChainSelector)
		assert.Empty(t, chainReport.Messages)
		assert.Empty(t, chainReport.OffchainTokenData)
		assert.Empty(t, chainReport.Proofs)

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Single Message with Single Token", func(t *testing.T) {
		tokenData := []byte("token-data-1")

		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages: []ccipocr3.Message{
						createTestMessage("msg1"),
					},
					OffchainTokenData: [][][]byte{
						{tokenData}, // Message 1 has 1 token
					},
					Proofs:        []ccipocr3.Bytes32{createTestBytes32("proof1")},
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		assert.Equal(t, uint64(12345), chainReport.SourceChainSelector)
		require.Len(t, chainReport.Messages, 1)
		require.Len(t, chainReport.OffchainTokenData, 1)

		// Verify the token data structure
		messageTokenData := chainReport.OffchainTokenData[0]
		require.Len(t, messageTokenData.TokenData, 1)
		assert.Equal(t, tokenData, messageTokenData.TokenData[0])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Single Message with Multiple Tokens", func(t *testing.T) {
		token1 := []byte("token-data-1")
		token2 := []byte("token-data-2")
		token3 := []byte("token-data-3")

		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages: []ccipocr3.Message{
						createTestMessage("msg1"),
					},
					OffchainTokenData: [][][]byte{
						{token1, token2, token3}, // Message 1 has 3 tokens
					},
					Proofs:        []ccipocr3.Bytes32{createTestBytes32("proof1")},
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		require.Len(t, chainReport.OffchainTokenData, 1)

		// Verify the token data structure
		messageTokenData := chainReport.OffchainTokenData[0]
		require.Len(t, messageTokenData.TokenData, 3)
		assert.Equal(t, token1, messageTokenData.TokenData[0])
		assert.Equal(t, token2, messageTokenData.TokenData[1])
		assert.Equal(t, token3, messageTokenData.TokenData[2])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Multiple Messages with Varying Token Counts", func(t *testing.T) {
		// Message 1: 2 tokens
		msg1Token1 := []byte("msg1-token1")
		msg1Token2 := []byte("msg1-token2")

		// Message 2: 0 tokens (empty)

		// Message 3: 1 token
		msg3Token1 := []byte("msg3-token1")

		// Message 4: 3 tokens
		msg4Token1 := []byte("msg4-token1")
		msg4Token2 := []byte("msg4-token2")
		msg4Token3 := []byte("msg4-token3")

		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages: []ccipocr3.Message{
						createTestMessage("msg1"),
						createTestMessage("msg2"),
						createTestMessage("msg3"),
						createTestMessage("msg4"),
					},
					OffchainTokenData: [][][]byte{
						{msg1Token1, msg1Token2},             // Message 1: 2 tokens
						{},                                   // Message 2: 0 tokens
						{msg3Token1},                         // Message 3: 1 token
						{msg4Token1, msg4Token2, msg4Token3}, // Message 4: 3 tokens
					},
					Proofs: []ccipocr3.Bytes32{
						createTestBytes32("proof1"),
						createTestBytes32("proof2"),
					},
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		require.Len(t, chainReport.Messages, 4)
		require.Len(t, chainReport.OffchainTokenData, 4)

		// Verify Message 1 tokens
		msg1Data := chainReport.OffchainTokenData[0]
		require.Len(t, msg1Data.TokenData, 2)
		assert.Equal(t, msg1Token1, msg1Data.TokenData[0])
		assert.Equal(t, msg1Token2, msg1Data.TokenData[1])

		// Verify Message 2 tokens (empty)
		msg2Data := chainReport.OffchainTokenData[1]
		assert.Empty(t, msg2Data.TokenData)

		// Verify Message 3 tokens
		msg3Data := chainReport.OffchainTokenData[2]
		require.Len(t, msg3Data.TokenData, 1)
		assert.Equal(t, msg3Token1, msg3Data.TokenData[0])

		// Verify Message 4 tokens
		msg4Data := chainReport.OffchainTokenData[3]
		require.Len(t, msg4Data.TokenData, 3)
		assert.Equal(t, msg4Token1, msg4Data.TokenData[0])
		assert.Equal(t, msg4Token2, msg4Data.TokenData[1])
		assert.Equal(t, msg4Token3, msg4Data.TokenData[2])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Multiple Chain Reports", func(t *testing.T) {
		// Chain 1 data
		chain1Token1 := []byte("chain1-msg1-token1")
		chain1Token2 := []byte("chain1-msg2-token1")

		// Chain 2 data
		chain2Token1 := []byte("chain2-msg1-token1")
		chain2Token2 := []byte("chain2-msg1-token2")

		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(111),
					Messages: []ccipocr3.Message{
						createTestMessage("chain1-msg1"),
						createTestMessage("chain1-msg2"),
					},
					OffchainTokenData: [][][]byte{
						{chain1Token1}, // Chain 1, Message 1: 1 token
						{chain1Token2}, // Chain 1, Message 2: 1 token
					},
					Proofs:        []ccipocr3.Bytes32{createTestBytes32("chain1-proof1")},
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
				{
					SourceChainSelector: ccipocr3.ChainSelector(222),
					Messages: []ccipocr3.Message{
						createTestMessage("chain2-msg1"),
					},
					OffchainTokenData: [][][]byte{
						{chain2Token1, chain2Token2}, // Chain 2, Message 1: 2 tokens
					},
					Proofs:        []ccipocr3.Bytes32{createTestBytes32("chain2-proof1")},
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 2)

		// Verify Chain 1
		chain1Report := pb.ChainReports[0]
		assert.Equal(t, uint64(111), chain1Report.SourceChainSelector)
		require.Len(t, chain1Report.OffchainTokenData, 2)

		assert.Len(t, chain1Report.OffchainTokenData[0].TokenData, 1)
		assert.Equal(t, chain1Token1, chain1Report.OffchainTokenData[0].TokenData[0])

		assert.Len(t, chain1Report.OffchainTokenData[1].TokenData, 1)
		assert.Equal(t, chain1Token2, chain1Report.OffchainTokenData[1].TokenData[0])

		// Verify Chain 2
		chain2Report := pb.ChainReports[1]
		assert.Equal(t, uint64(222), chain2Report.SourceChainSelector)
		require.Len(t, chain2Report.OffchainTokenData, 1)

		assert.Len(t, chain2Report.OffchainTokenData[0].TokenData, 2)
		assert.Equal(t, chain2Token1, chain2Report.OffchainTokenData[0].TokenData[0])
		assert.Equal(t, chain2Token2, chain2Report.OffchainTokenData[0].TokenData[1])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Edge Case: Large Token Data", func(t *testing.T) {
		// Create large token data to test memory handling
		largeToken := make([]byte, 10000)
		for i := range largeToken {
			largeToken[i] = byte(i % 256)
		}

		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages: []ccipocr3.Message{
						createTestMessage("msg1"),
					},
					OffchainTokenData: [][][]byte{
						{largeToken},
					},
					Proofs:        nil,
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		require.Len(t, chainReport.OffchainTokenData, 1)
		require.Len(t, chainReport.OffchainTokenData[0].TokenData, 1)

		// Verify large token data integrity
		assert.Equal(t, largeToken, chainReport.OffchainTokenData[0].TokenData[0])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})

	t.Run("Edge Case: Nil Token Data Elements", func(t *testing.T) {
		original := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
				{
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					Messages: []ccipocr3.Message{
						createTestMessage("msg1"),
					},
					OffchainTokenData: [][][]byte{
						{nil, []byte("token2"), nil}, // Mix of nil and non-nil
					},
					Proofs:        nil,
					ProofFlagBits: ccipocr3.BigInt{Int: nil},
				},
			},
		}

		// Test Go -> Protobuf
		pb := executePluginReportToPb(original)
		require.NotNil(t, pb)
		require.Len(t, pb.ChainReports, 1)

		chainReport := pb.ChainReports[0]
		require.Len(t, chainReport.OffchainTokenData, 1)
		require.Len(t, chainReport.OffchainTokenData[0].TokenData, 3)

		// Verify nil handling
		assert.Nil(t, chainReport.OffchainTokenData[0].TokenData[0])
		assert.Equal(t, []byte("token2"), chainReport.OffchainTokenData[0].TokenData[1])
		assert.Nil(t, chainReport.OffchainTokenData[0].TokenData[2])

		// Test round-trip
		roundTrip := pbToExecutePluginReport(pb)
		assert.Equal(t, original, roundTrip)
	})
}

func TestMessageOffchainTokenDataStructure(t *testing.T) {
	t.Run("Direct MessageOffchainTokenData Creation", func(t *testing.T) {
		tokenData := [][]byte{
			[]byte("token1"),
			[]byte("token2"),
			nil,
			[]byte("token4"),
		}

		msgData := &ccipocr3pb.MessageOffchainTokenData{
			TokenData: tokenData,
		}

		assert.Equal(t, tokenData, msgData.TokenData)
		assert.Len(t, msgData.TokenData, 4)
		assert.Equal(t, []byte("token1"), msgData.TokenData[0])
		assert.Equal(t, []byte("token2"), msgData.TokenData[1])
		assert.Nil(t, msgData.TokenData[2])
		assert.Equal(t, []byte("token4"), msgData.TokenData[3])
	})

	t.Run("Empty MessageOffchainTokenData", func(t *testing.T) {
		msgData := &ccipocr3pb.MessageOffchainTokenData{
			TokenData: [][]byte{},
		}

		assert.Empty(t, msgData.TokenData)
	})

	t.Run("Nil MessageOffchainTokenData", func(t *testing.T) {
		msgData := &ccipocr3pb.MessageOffchainTokenData{
			TokenData: nil,
		}

		assert.Nil(t, msgData.TokenData)
	})
}

func TestCommitPluginReportConversions(t *testing.T) {
	testCases := []struct {
		name   string
		report ccipocr3.CommitPluginReport
	}{
		{
			name:   "Empty Report",
			report: ccipocr3.CommitPluginReport{},
		},
		{
			name: "Report with OnRampAddress and RMN Signatures",
			report: ccipocr3.CommitPluginReport{
				PriceUpdates: ccipocr3.PriceUpdates{
					TokenPriceUpdates: []ccipocr3.TokenPrice{
						{
							TokenID: ccipocr3.UnknownEncodedAddress("0x1234"),
							Price:   ccipocr3.NewBigInt(big.NewInt(1000)),
						},
					},
					GasPriceUpdates: []ccipocr3.GasPriceChain{
						{
							ChainSel: ccipocr3.ChainSelector(1),
							GasPrice: ccipocr3.NewBigInt(big.NewInt(2000)),
						},
					},
				},
				BlessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(1),
						OnRampAddress: []byte("onramp123"), // Test OnRampAddress field
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(1), ccipocr3.SeqNum(10)),
						MerkleRoot:    [32]byte{1, 2, 3, 4, 5},
					},
				},
				UnblessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(2),
						OnRampAddress: []byte("onramp456"), // Test OnRampAddress field
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(11), ccipocr3.SeqNum(20)),
						MerkleRoot:    [32]byte{6, 7, 8, 9, 10},
					},
				},
				RMNSignatures: []ccipocr3.RMNECDSASignature{ // Test RMN signatures
					{
						R: [32]byte{0xaa, 0xbb, 0xcc, 0xdd},
						S: [32]byte{0xee, 0xff, 0x00, 0x11},
					},
					{
						R: [32]byte{0x22, 0x33, 0x44, 0x55},
						S: [32]byte{0x66, 0x77, 0x88, 0x99},
					},
				},
			},
		},
		{
			name: "Blessed vs Unblessed Separation Test",
			report: ccipocr3.CommitPluginReport{
				BlessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(100),
						OnRampAddress: []byte("blessed-onramp"),
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(1), ccipocr3.SeqNum(5)),
						MerkleRoot:    [32]byte{0xB1, 0xE5, 0x5E, 0xD}, // BLESSED in hex
					},
					{
						ChainSel:      ccipocr3.ChainSelector(200),
						OnRampAddress: []byte("blessed-onramp-2"),
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(6), ccipocr3.SeqNum(10)),
						MerkleRoot:    [32]byte{0xB1, 0xE5, 0x5E, 0xD2}, // BLESSED 2 in hex
					},
				},
				UnblessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(300),
						OnRampAddress: []byte("unblessed-onramp"),
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(11), ccipocr3.SeqNum(15)),
						MerkleRoot:    [32]byte{0xBA, 0xD}, // BAD in hex
					},
				},
				RMNSignatures: []ccipocr3.RMNECDSASignature{
					{
						R: [32]byte{0xDE, 0xAD, 0xBE, 0xEF},
						S: [32]byte{0xCA, 0xFE, 0xBA, 0xBE},
					},
				},
			},
		},
		{
			name: "Only Blessed Merkle Roots",
			report: ccipocr3.CommitPluginReport{
				BlessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(500),
						OnRampAddress: []byte("only-blessed"),
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(100), ccipocr3.SeqNum(200)),
						MerkleRoot:    [32]byte{0x0B, 0x1E, 0x55},
					},
				},
				// Note: No unblessed merkle roots
			},
		},
		{
			name: "Only Unblessed Merkle Roots",
			report: ccipocr3.CommitPluginReport{
				// Note: No blessed merkle roots
				UnblessedMerkleRoots: []ccipocr3.MerkleRootChain{
					{
						ChainSel:      ccipocr3.ChainSelector(600),
						OnRampAddress: []byte("only-unblessed"),
						SeqNumsRange:  ccipocr3.NewSeqNumRange(ccipocr3.SeqNum(300), ccipocr3.SeqNum(400)),
						MerkleRoot:    [32]byte{0xFA, 0xDE, 0xD},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to protobuf
			pbReport := commitPluginReportToPb(tc.report)
			require.NotNil(t, pbReport)

			// Convert back to Go struct
			convertedReport := pbToCommitPluginReportDetailed(pbReport)

			// Verify round-trip conversion preserves all data
			assert.Equal(t, len(tc.report.PriceUpdates.TokenPriceUpdates), len(convertedReport.PriceUpdates.TokenPriceUpdates))
			assert.Equal(t, len(tc.report.PriceUpdates.GasPriceUpdates), len(convertedReport.PriceUpdates.GasPriceUpdates))

			// Verify blessed and unblessed merkle roots are preserved separately
			assert.Equal(t, len(tc.report.BlessedMerkleRoots), len(convertedReport.BlessedMerkleRoots))
			assert.Equal(t, len(tc.report.UnblessedMerkleRoots), len(convertedReport.UnblessedMerkleRoots))

			assert.Equal(t, len(tc.report.RMNSignatures), len(convertedReport.RMNSignatures))

			// Verify specific field values (OnRampAddress and RMN signatures)
			if len(tc.report.BlessedMerkleRoots) > 0 {
				original := tc.report.BlessedMerkleRoots[0]
				converted := convertedReport.BlessedMerkleRoots[0]
				assert.Equal(t, original.ChainSel, converted.ChainSel)
				assert.Equal(t, original.OnRampAddress, converted.OnRampAddress) // Key test for OnRampAddress
				assert.Equal(t, original.SeqNumsRange, converted.SeqNumsRange)
				assert.Equal(t, original.MerkleRoot, converted.MerkleRoot)
			}

			// Verify unblessed merkle roots are preserved correctly
			if len(tc.report.UnblessedMerkleRoots) > 0 {
				original := tc.report.UnblessedMerkleRoots[0]
				converted := convertedReport.UnblessedMerkleRoots[0]
				assert.Equal(t, original.ChainSel, converted.ChainSel)
				assert.Equal(t, original.OnRampAddress, converted.OnRampAddress) // Key test for OnRampAddress
				assert.Equal(t, original.SeqNumsRange, converted.SeqNumsRange)
				assert.Equal(t, original.MerkleRoot, converted.MerkleRoot)
			}

			if len(tc.report.RMNSignatures) > 0 {
				original := tc.report.RMNSignatures[0]
				converted := convertedReport.RMNSignatures[0]
				assert.Equal(t, original.R, converted.R) // Key test for RMN signature R
				assert.Equal(t, original.S, converted.S) // Key test for RMN signature S
			}
		})
	}
}

func TestArrayAlignmentInvariant(t *testing.T) {
	t.Run("Messages and OffchainTokenData Must Have Same Length", func(t *testing.T) {
		testCases := []struct {
			name         string
			messageCount int
			tokenArrays  [][][]byte
		}{
			{
				name:         "Zero messages, zero token arrays",
				messageCount: 0,
				tokenArrays:  [][][]byte{},
			},
			{
				name:         "One message, one token array",
				messageCount: 1,
				tokenArrays:  [][][]byte{{}},
			},
			{
				name:         "Three messages, three token arrays",
				messageCount: 3,
				tokenArrays: [][][]byte{
					{[]byte("msg1-token1")},
					{},
					{[]byte("msg3-token1"), []byte("msg3-token2")},
				},
			},
			{
				name:         "Five messages with varying token counts",
				messageCount: 5,
				tokenArrays: [][][]byte{
					{[]byte("msg1-token1"), []byte("msg1-token2")},
					{},
					{[]byte("msg3-token1")},
					{},
					{[]byte("msg5-token1"), []byte("msg5-token2"), []byte("msg5-token3")},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create messages
				messages := make([]ccipocr3.Message, tc.messageCount)
				for i := 0; i < tc.messageCount; i++ {
					messages[i] = createTestMessage("msg" + string(rune('1'+i)))
				}

				original := ccipocr3.ExecutePluginReport{
					ChainReports: []ccipocr3.ExecutePluginReportSingleChain{
						{
							SourceChainSelector: ccipocr3.ChainSelector(12345),
							Messages:            messages,
							OffchainTokenData:   tc.tokenArrays,
							Proofs:              nil,
							ProofFlagBits:       ccipocr3.BigInt{Int: nil},
						},
					},
				}

				// Verify alignment invariant before conversion
				require.Equal(t, len(messages), len(tc.tokenArrays),
					"Messages and OffchainTokenData arrays must have the same length")

				// Test conversion
				pb := executePluginReportToPb(original)
				require.NotNil(t, pb)
				require.Len(t, pb.ChainReports, 1)

				chainReport := pb.ChainReports[0]

				// Verify alignment invariant after conversion
				assert.Equal(t, len(chainReport.Messages), len(chainReport.OffchainTokenData),
					"Messages and OffchainTokenData arrays must maintain alignment after conversion")

				// Verify each token array matches expected structure
				for i, expectedTokenArray := range tc.tokenArrays {
					actualTokenData := chainReport.OffchainTokenData[i].TokenData
					assert.Equal(t, expectedTokenArray, actualTokenData,
						"Token data for message %d should match", i)
				}

				// Test round-trip
				roundTrip := pbToExecutePluginReport(pb)
				assert.Equal(t, original, roundTrip, "Round-trip conversion should preserve data")
			})
		}
	})
}

func createTestMessage(id string) ccipocr3.Message {
	messageID := ccipocr3.Bytes32{}
	copy(messageID[:], id)

	msgHash := ccipocr3.Bytes32{}
	copy(msgHash[:], id+"-hash")

	return ccipocr3.Message{
		Header: ccipocr3.RampMessageHeader{
			MessageID:           messageID,
			SourceChainSelector: ccipocr3.ChainSelector(1),
			DestChainSelector:   ccipocr3.ChainSelector(2),
			SequenceNumber:      ccipocr3.SeqNum(1),
			Nonce:               1,
			MsgHash:             msgHash,
			OnRamp:              []byte("onramp"),
		},
		Sender:         []byte("sender"),
		Data:           []byte("data"),
		Receiver:       []byte("receiver"),
		ExtraArgs:      []byte("extraargs"),
		FeeToken:       []byte("feetoken"),
		FeeTokenAmount: ccipocr3.NewBigInt(big.NewInt(1000)),
		TokenAmounts: []ccipocr3.RampTokenAmount{
			{
				SourcePoolAddress: []byte("source-pool-address"),
				DestTokenAddress:  []byte("dest-token-address"),
				ExtraData:         []byte("token-extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(500)),
			},
		},
	}
}

func createTestBytes32(data string) ccipocr3.Bytes32 {
	var b32 ccipocr3.Bytes32
	copy(b32[:], data)
	return b32
}
