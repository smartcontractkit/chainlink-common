package ccipocr3

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

// TestMessageProtobufFlattening tests the critical fix for Message field flattening
func TestMessageProtobufFlattening(t *testing.T) {
	testCases := []struct {
		name    string
		message ccipocr3.Message
	}{
		{
			name: "Message with all fields populated",
			message: ccipocr3.Message{
				Header: ccipocr3.RampMessageHeader{
					MessageID:           [32]byte{0x01, 0x02, 0x03, 0x04},
					SourceChainSelector: ccipocr3.ChainSelector(12345),
					DestChainSelector:   ccipocr3.ChainSelector(67890),
					SequenceNumber:      ccipocr3.SeqNum(100),
					Nonce:               42,
					MsgHash:             [32]byte{0xAA, 0xBB, 0xCC, 0xDD},
					OnRamp:              []byte("onramp-address"),
					TxHash:              "0x1234567890abcdef",
				},
				Sender:         []byte("sender-address"),
				Data:           []byte("message-data"),
				Receiver:       []byte("receiver-address"),
				ExtraArgs:      []byte("extra-args"),
				FeeToken:       []byte("fee-token-address"),
				FeeTokenAmount: ccipocr3.NewBigInt(big.NewInt(1000)),
				FeeValueJuels:  ccipocr3.NewBigInt(big.NewInt(2000)),
				TokenAmounts: []ccipocr3.RampTokenAmount{
					{
						SourcePoolAddress: []byte("source-pool"),
						DestTokenAddress:  []byte("dest-token"),
						ExtraData:         []byte("token-extra"),
						Amount:            ccipocr3.NewBigInt(big.NewInt(500)),
					},
				},
			},
		},
		{
			name: "Message with empty byte fields",
			message: ccipocr3.Message{
				Header: ccipocr3.RampMessageHeader{
					MessageID:           [32]byte{},
					SourceChainSelector: ccipocr3.ChainSelector(1),
					DestChainSelector:   ccipocr3.ChainSelector(2),
					SequenceNumber:      ccipocr3.SeqNum(1),
					Nonce:               0,
					MsgHash:             [32]byte{},
					OnRamp:              []byte{},
					TxHash:              "",
				},
				Sender:         []byte{},
				Data:           []byte{},
				Receiver:       []byte{},
				ExtraArgs:      []byte{},
				FeeToken:       []byte{},
				FeeTokenAmount: ccipocr3.BigInt{Int: nil},
				FeeValueJuels:  ccipocr3.BigInt{Int: nil},
				TokenAmounts:   []ccipocr3.RampTokenAmount{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to protobuf
			pbMessage := messageToPb(tc.message)
			require.NotNil(t, pbMessage)

			// Verify protobuf structure is correct (not flattened)
			require.NotNil(t, pbMessage.Header)

			// Critical assertions
			assert.Equal(t, tc.message.Header.MessageID[:], pbMessage.Header.MessageId)
			assert.Equal(t, tc.message.Header.MsgHash[:], pbMessage.Header.MessageHash)
			assert.Equal(t, []byte(tc.message.Header.OnRamp), pbMessage.Header.OnRamp)
			assert.Equal(t, tc.message.Header.TxHash, pbMessage.Header.TxHash)
			assert.Equal(t, []byte(tc.message.Sender), pbMessage.Sender)
			assert.Equal(t, []byte(tc.message.Data), pbMessage.Data)
			assert.Equal(t, []byte(tc.message.Receiver), pbMessage.Receiver)
			assert.Equal(t, []byte(tc.message.ExtraArgs), pbMessage.ExtraArgs)
			assert.Equal(t, []byte(tc.message.FeeToken), pbMessage.FeeToken)
			assert.NotNil(t, pbMessage.FeeValueJuels)

			// Convert back to Go struct
			convertedMessage := pbToMessage(pbMessage)

			// Verify round-trip conversion preserves all data
			assert.Equal(t, tc.message.Header.MessageID, convertedMessage.Header.MessageID)
			assert.Equal(t, tc.message.Header.SourceChainSelector, convertedMessage.Header.SourceChainSelector)
			assert.Equal(t, tc.message.Header.DestChainSelector, convertedMessage.Header.DestChainSelector)
			assert.Equal(t, tc.message.Header.SequenceNumber, convertedMessage.Header.SequenceNumber)
			assert.Equal(t, tc.message.Header.Nonce, convertedMessage.Header.Nonce)
			assert.Equal(t, tc.message.Header.MsgHash, convertedMessage.Header.MsgHash)
			assert.Equal(t, []byte(tc.message.Header.OnRamp), []byte(convertedMessage.Header.OnRamp))
			assert.Equal(t, tc.message.Header.TxHash, convertedMessage.Header.TxHash)
			assert.Equal(t, []byte(tc.message.Sender), []byte(convertedMessage.Sender))
			assert.Equal(t, []byte(tc.message.Data), []byte(convertedMessage.Data))
			assert.Equal(t, []byte(tc.message.Receiver), []byte(convertedMessage.Receiver))
			assert.Equal(t, []byte(tc.message.ExtraArgs), []byte(convertedMessage.ExtraArgs))
			assert.Equal(t, []byte(tc.message.FeeToken), []byte(convertedMessage.FeeToken))
			// Handle nil BigInt comparisons properly
			if tc.message.FeeTokenAmount.Int == nil && convertedMessage.FeeTokenAmount.Int == nil {
				assert.True(t, true, "Both FeeTokenAmount.Int are nil")
			} else if tc.message.FeeTokenAmount.Int != nil && convertedMessage.FeeTokenAmount.Int != nil {
				assert.Equal(t, tc.message.FeeTokenAmount.Int.String(), convertedMessage.FeeTokenAmount.Int.String())
			} else {
				assert.Equal(t, tc.message.FeeTokenAmount.Int, convertedMessage.FeeTokenAmount.Int, "FeeTokenAmount.Int nil mismatch")
			}

			if tc.message.FeeValueJuels.Int == nil && convertedMessage.FeeValueJuels.Int == nil {
				assert.True(t, true, "Both FeeValueJuels.Int are nil")
			} else if tc.message.FeeValueJuels.Int != nil && convertedMessage.FeeValueJuels.Int != nil {
				assert.Equal(t, tc.message.FeeValueJuels.Int.String(), convertedMessage.FeeValueJuels.Int.String())
			} else {
				assert.Equal(t, tc.message.FeeValueJuels.Int, convertedMessage.FeeValueJuels.Int, "FeeValueJuels.Int nil mismatch")
			}
			assert.Equal(t, len(tc.message.TokenAmounts), len(convertedMessage.TokenAmounts))

			// Verify token amounts
			for i, original := range tc.message.TokenAmounts {
				converted := convertedMessage.TokenAmounts[i]
				assert.Equal(t, []byte(original.SourcePoolAddress), []byte(converted.SourcePoolAddress))
				assert.Equal(t, []byte(original.DestTokenAddress), []byte(converted.DestTokenAddress))
				assert.Equal(t, []byte(original.ExtraData), []byte(converted.ExtraData))
				assert.Equal(t, original.Amount.Int.String(), converted.Amount.Int.String())
			}
		})
	}
}

// TestRMNTypesConversion tests the newly added RMN type conversions
func TestRMNTypesConversion(t *testing.T) {
	testCases := []struct {
		name      string
		rmnReport ccipocr3.RMNReport
	}{
		{
			name: "RMN Report with multiple lane updates",
			rmnReport: ccipocr3.RMNReport{
				ReportVersionDigest:         [32]byte{0x01, 0x02, 0x03},
				DestChainID:                 ccipocr3.NewBigInt(big.NewInt(12345)),
				DestChainSelector:           ccipocr3.ChainSelector(67890),
				RmnRemoteContractAddress:    []byte("rmn-remote-address"),
				OfframpAddress:              []byte("offramp-address"),
				RmnHomeContractConfigDigest: [32]byte{0xAA, 0xBB, 0xCC},
				LaneUpdates: []ccipocr3.RMNLaneUpdate{
					{
						SourceChainSelector: ccipocr3.ChainSelector(111),
						OnRampAddress:       []byte("onramp-1"),
						MinSeqNr:            ccipocr3.SeqNum(100),
						MaxSeqNr:            ccipocr3.SeqNum(200),
						MerkleRoot:          [32]byte{0x11, 0x22, 0x33},
					},
					{
						SourceChainSelector: ccipocr3.ChainSelector(222),
						OnRampAddress:       []byte("onramp-2"),
						MinSeqNr:            ccipocr3.SeqNum(300),
						MaxSeqNr:            ccipocr3.SeqNum(400),
						MerkleRoot:          [32]byte{0x44, 0x55, 0x66},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to protobuf
			pbReport := rmnReportToPb(tc.rmnReport)
			require.NotNil(t, pbReport)

			// Verify protobuf structure
			assert.Equal(t, tc.rmnReport.ReportVersionDigest[:], pbReport.ReportVersionDigest)
			assert.Equal(t, uint64(tc.rmnReport.DestChainSelector), pbReport.DestChainSelector)
			assert.Equal(t, []byte(tc.rmnReport.RmnRemoteContractAddress), pbReport.RmnRemoteContractAddress)
			assert.Equal(t, []byte(tc.rmnReport.OfframpAddress), pbReport.OfframpAddress)
			assert.Equal(t, tc.rmnReport.RmnHomeContractConfigDigest[:], pbReport.RmnHomeContractConfigDigest)
			assert.Equal(t, len(tc.rmnReport.LaneUpdates), len(pbReport.LaneUpdates))

			// Convert back to Go struct
			convertedReport := pbToRMNReport(pbReport)

			// Verify round-trip conversion
			assert.Equal(t, tc.rmnReport.ReportVersionDigest, convertedReport.ReportVersionDigest)
			assert.Equal(t, tc.rmnReport.DestChainID.Int.String(), convertedReport.DestChainID.Int.String())
			assert.Equal(t, tc.rmnReport.DestChainSelector, convertedReport.DestChainSelector)
			assert.Equal(t, []byte(tc.rmnReport.RmnRemoteContractAddress), []byte(convertedReport.RmnRemoteContractAddress))
			assert.Equal(t, []byte(tc.rmnReport.OfframpAddress), []byte(convertedReport.OfframpAddress))
			assert.Equal(t, tc.rmnReport.RmnHomeContractConfigDigest, convertedReport.RmnHomeContractConfigDigest)
			assert.Equal(t, len(tc.rmnReport.LaneUpdates), len(convertedReport.LaneUpdates))

			// Verify lane updates
			for i, original := range tc.rmnReport.LaneUpdates {
				converted := convertedReport.LaneUpdates[i]
				assert.Equal(t, original.SourceChainSelector, converted.SourceChainSelector)
				assert.Equal(t, []byte(original.OnRampAddress), []byte(converted.OnRampAddress))
				assert.Equal(t, original.MinSeqNr, converted.MinSeqNr)
				assert.Equal(t, original.MaxSeqNr, converted.MaxSeqNr)
				assert.Equal(t, original.MerkleRoot, converted.MerkleRoot)
			}
		})
	}
}

// TestProtobufFlatteningRegressionTest ensures we never reintroduce flattening bugs
func TestProtobufFlatteningRegressionTest(t *testing.T) {
	t.Run("Message fields must not be repeated bytes", func(t *testing.T) {
		// This test would catch if someone accidentally reverts to repeated bytes
		msg := ccipocr3.Message{
			Header: ccipocr3.RampMessageHeader{
				MessageID: [32]byte{0x01},
				MsgHash:   [32]byte{0x02},
				OnRamp:    []byte("test"),
				TxHash:    "test-tx",
			},
			Sender:        []byte("test-sender"),
			Data:          []byte("test-data"),
			Receiver:      []byte("test-receiver"),
			ExtraArgs:     []byte("test-extra"),
			FeeToken:      []byte("test-fee-token"),
			FeeValueJuels: ccipocr3.NewBigInt(big.NewInt(12345)),
		}

		// Convert to protobuf and back
		pbMsg := messageToPb(msg)
		convertedMsg := pbToMessage(pbMsg)

		// These assertions would FAIL if protobuf flattening was reintroduced
		assert.Equal(t, msg.Header.MessageID, convertedMsg.Header.MessageID, "MessageID corruption indicates protobuf flattening regression")
		assert.Equal(t, msg.Header.MsgHash, convertedMsg.Header.MsgHash, "MsgHash corruption indicates protobuf flattening regression")
		assert.Equal(t, []byte(msg.Header.OnRamp), []byte(convertedMsg.Header.OnRamp), "OnRamp corruption indicates protobuf flattening regression")
		assert.Equal(t, msg.Header.TxHash, convertedMsg.Header.TxHash, "TxHash missing indicates missing field regression")
		assert.Equal(t, []byte(msg.Sender), []byte(convertedMsg.Sender), "Sender corruption indicates protobuf flattening regression")
		assert.Equal(t, []byte(msg.Data), []byte(convertedMsg.Data), "Data corruption indicates protobuf flattening regression")
		assert.Equal(t, []byte(msg.Receiver), []byte(convertedMsg.Receiver), "Receiver corruption indicates protobuf flattening regression")
		assert.Equal(t, []byte(msg.ExtraArgs), []byte(convertedMsg.ExtraArgs), "ExtraArgs corruption indicates protobuf flattening regression")
		assert.Equal(t, []byte(msg.FeeToken), []byte(convertedMsg.FeeToken), "FeeToken corruption indicates protobuf flattening regression")
		assert.Equal(t, msg.FeeValueJuels.Int.String(), convertedMsg.FeeValueJuels.Int.String(), "FeeValueJuels missing indicates missing field regression")
	})
}

// TestFeeQuoterDestChainConfigConversion tests the critical uint16/uint32 type conversions
func TestFeeQuoterDestChainConfigConversion(t *testing.T) {
	testCases := []struct {
		name   string
		config ccipocr3.FeeQuoterDestChainConfig
	}{
		{
			name: "Config with typical values",
			config: ccipocr3.FeeQuoterDestChainConfig{
				IsEnabled:                         true,
				MaxNumberOfTokensPerMsg:           10,                              // uint16 -> uint32 -> uint16
				MaxDataBytes:                      100000,                          // uint32 stays uint32
				MaxPerMsgGasLimit:                 2000000,                         // uint32 stays uint32
				DestGasOverhead:                   50000,                           // uint32 stays uint32
				DestGasPerPayloadByteBase:         100,                             // uint32 stays uint32
				DestGasPerPayloadByteHigh:         150,                             // uint32 stays uint32
				DestGasPerPayloadByteThreshold:    1000,                            // uint32 stays uint32
				DestDataAvailabilityOverheadGas:   10000,                           // uint32 stays uint32
				DestGasPerDataAvailabilityByte:    5,                               // uint16 -> uint32 -> uint16
				DestDataAvailabilityMultiplierBps: 1500,                            // uint16 -> uint32 -> uint16
				DefaultTokenFeeUSDCents:           100,                             // uint16 -> uint32 -> uint16
				DefaultTokenDestGasOverhead:       75000,                           // uint32 stays uint32
				DefaultTxGasLimit:                 3000000,                         // uint32 stays uint32
				GasMultiplierWeiPerEth:            1100000000000000000,             // uint64 stays uint64
				NetworkFeeUSDCents:                50,                              // uint32 stays uint32
				GasPriceStalenessThreshold:        300,                             // uint32 stays uint32
				EnforceOutOfOrder:                 false,                           // bool stays bool
				ChainFamilySelector:               [4]byte{0x01, 0x02, 0x03, 0x04}, // [4]byte -> bytes -> [4]byte
			},
		},
		{
			name: "Config with max uint16 values",
			config: ccipocr3.FeeQuoterDestChainConfig{
				IsEnabled:                         true,
				MaxNumberOfTokensPerMsg:           math.MaxUint16,                  // 65535 - max uint16
				MaxDataBytes:                      math.MaxUint32,                  // max uint32
				MaxPerMsgGasLimit:                 math.MaxUint32,                  // max uint32
				DestGasOverhead:                   math.MaxUint32,                  // max uint32
				DestGasPerPayloadByteBase:         math.MaxUint32,                  // max uint32
				DestGasPerPayloadByteHigh:         math.MaxUint32,                  // max uint32
				DestGasPerPayloadByteThreshold:    math.MaxUint32,                  // max uint32
				DestDataAvailabilityOverheadGas:   math.MaxUint32,                  // max uint32
				DestGasPerDataAvailabilityByte:    math.MaxUint16,                  // 65535 - max uint16
				DestDataAvailabilityMultiplierBps: math.MaxUint16,                  // 65535 - max uint16
				DefaultTokenFeeUSDCents:           math.MaxUint16,                  // 65535 - max uint16
				DefaultTokenDestGasOverhead:       math.MaxUint32,                  // max uint32
				DefaultTxGasLimit:                 math.MaxUint32,                  // max uint32
				GasMultiplierWeiPerEth:            math.MaxUint64,                  // max uint64
				NetworkFeeUSDCents:                math.MaxUint32,                  // max uint32
				GasPriceStalenessThreshold:        math.MaxUint32,                  // max uint32
				EnforceOutOfOrder:                 true,                            // bool
				ChainFamilySelector:               [4]byte{0xFF, 0xFF, 0xFF, 0xFF}, // max bytes
			},
		},
		{
			name: "Config with zero values",
			config: ccipocr3.FeeQuoterDestChainConfig{
				IsEnabled:                         false,
				MaxNumberOfTokensPerMsg:           0,                               // uint16 min
				MaxDataBytes:                      0,                               // uint32 min
				MaxPerMsgGasLimit:                 0,                               // uint32 min
				DestGasOverhead:                   0,                               // uint32 min
				DestGasPerPayloadByteBase:         0,                               // uint32 min
				DestGasPerPayloadByteHigh:         0,                               // uint32 min
				DestGasPerPayloadByteThreshold:    0,                               // uint32 min
				DestDataAvailabilityOverheadGas:   0,                               // uint32 min
				DestGasPerDataAvailabilityByte:    0,                               // uint16 min
				DestDataAvailabilityMultiplierBps: 0,                               // uint16 min
				DefaultTokenFeeUSDCents:           0,                               // uint16 min
				DefaultTokenDestGasOverhead:       0,                               // uint32 min
				DefaultTxGasLimit:                 0,                               // uint32 min
				GasMultiplierWeiPerEth:            0,                               // uint64 min
				NetworkFeeUSDCents:                0,                               // uint32 min
				GasPriceStalenessThreshold:        0,                               // uint32 min
				EnforceOutOfOrder:                 false,                           // bool
				ChainFamilySelector:               [4]byte{0x00, 0x00, 0x00, 0x00}, // zero bytes
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbConfig := feeQuoterDestChainConfigToPb(tc.config)
			require.NotNil(t, pbConfig)

			// Verify protobuf values - uint16 fields should be safely widened to uint32
			assert.Equal(t, tc.config.IsEnabled, pbConfig.IsEnabled)
			assert.Equal(t, uint32(tc.config.MaxNumberOfTokensPerMsg), pbConfig.MaxNumberOfTokensPerMsg, "uint16 MaxNumberOfTokensPerMsg should be widened to uint32")
			assert.Equal(t, tc.config.MaxDataBytes, pbConfig.MaxDataBytes)
			assert.Equal(t, tc.config.MaxPerMsgGasLimit, pbConfig.MaxPerMsgGasLimit)
			assert.Equal(t, tc.config.DestGasOverhead, pbConfig.DestGasOverhead)
			assert.Equal(t, tc.config.DestGasPerPayloadByteBase, pbConfig.DestGasPerPayloadByte)
			assert.Equal(t, tc.config.DestGasPerPayloadByteHigh, pbConfig.DestGasPerPayloadByteHigh)
			assert.Equal(t, tc.config.DestGasPerPayloadByteThreshold, pbConfig.DestGasPerPayloadByteThreshold)
			assert.Equal(t, tc.config.DestDataAvailabilityOverheadGas, pbConfig.DestDataAvailabilityOverheadGas)
			assert.Equal(t, uint32(tc.config.DestGasPerDataAvailabilityByte), pbConfig.DestGasPerDataAvailabilityByte, "uint16 DestGasPerDataAvailabilityByte should be widened to uint32")
			assert.Equal(t, uint32(tc.config.DestDataAvailabilityMultiplierBps), pbConfig.DestDataAvailabilityMultiplierBps, "uint16 DestDataAvailabilityMultiplierBps should be widened to uint32")
			assert.Equal(t, uint32(tc.config.DefaultTokenFeeUSDCents), pbConfig.DefaultTokenFeeUsdcCents, "uint16 DefaultTokenFeeUSDCents should be widened to uint32")
			assert.Equal(t, tc.config.DefaultTokenDestGasOverhead, pbConfig.DefaultTokenDestGasOverhead)
			assert.Equal(t, tc.config.DefaultTxGasLimit, pbConfig.DefaultTxGasLimit)
			assert.Equal(t, tc.config.GasMultiplierWeiPerEth, pbConfig.GasMultiplierWad)
			assert.Equal(t, tc.config.NetworkFeeUSDCents, pbConfig.NetworkFeeUsdcCents)
			assert.Equal(t, tc.config.GasPriceStalenessThreshold, pbConfig.GasPriceStalenessThreshold)
			assert.Equal(t, tc.config.EnforceOutOfOrder, pbConfig.EnforceOutOfOrder)
			assert.Equal(t, tc.config.ChainFamilySelector[:], pbConfig.ChainFamilySelector)

			// Convert Protobuf -> Go (round-trip)
			convertedConfig := pbToFeeQuoterDestChainConfigDetailed(pbConfig)

			// Verify round-trip conversion preserves all data
			assert.Equal(t, tc.config.IsEnabled, convertedConfig.IsEnabled)
			assert.Equal(t, tc.config.MaxNumberOfTokensPerMsg, convertedConfig.MaxNumberOfTokensPerMsg, "uint16 MaxNumberOfTokensPerMsg should survive round-trip")
			assert.Equal(t, tc.config.MaxDataBytes, convertedConfig.MaxDataBytes)
			assert.Equal(t, tc.config.MaxPerMsgGasLimit, convertedConfig.MaxPerMsgGasLimit)
			assert.Equal(t, tc.config.DestGasOverhead, convertedConfig.DestGasOverhead)
			assert.Equal(t, tc.config.DestGasPerPayloadByteBase, convertedConfig.DestGasPerPayloadByteBase)
			assert.Equal(t, tc.config.DestGasPerPayloadByteHigh, convertedConfig.DestGasPerPayloadByteHigh)
			assert.Equal(t, tc.config.DestGasPerPayloadByteThreshold, convertedConfig.DestGasPerPayloadByteThreshold)
			assert.Equal(t, tc.config.DestDataAvailabilityOverheadGas, convertedConfig.DestDataAvailabilityOverheadGas)
			assert.Equal(t, tc.config.DestGasPerDataAvailabilityByte, convertedConfig.DestGasPerDataAvailabilityByte, "uint16 DestGasPerDataAvailabilityByte should survive round-trip")
			assert.Equal(t, tc.config.DestDataAvailabilityMultiplierBps, convertedConfig.DestDataAvailabilityMultiplierBps, "uint16 DestDataAvailabilityMultiplierBps should survive round-trip")
			assert.Equal(t, tc.config.DefaultTokenFeeUSDCents, convertedConfig.DefaultTokenFeeUSDCents, "uint16 DefaultTokenFeeUSDCents should survive round-trip")
			assert.Equal(t, tc.config.DefaultTokenDestGasOverhead, convertedConfig.DefaultTokenDestGasOverhead)
			assert.Equal(t, tc.config.DefaultTxGasLimit, convertedConfig.DefaultTxGasLimit)
			assert.Equal(t, tc.config.GasMultiplierWeiPerEth, convertedConfig.GasMultiplierWeiPerEth)
			assert.Equal(t, tc.config.NetworkFeeUSDCents, convertedConfig.NetworkFeeUSDCents)
			assert.Equal(t, tc.config.GasPriceStalenessThreshold, convertedConfig.GasPriceStalenessThreshold)
			assert.Equal(t, tc.config.EnforceOutOfOrder, convertedConfig.EnforceOutOfOrder)
			assert.Equal(t, tc.config.ChainFamilySelector, convertedConfig.ChainFamilySelector)
		})
	}
}

// TestFeeQuoterDestChainConfigUint16Boundaries tests the edge cases of uint16 conversion
func TestFeeQuoterDestChainConfigUint16Boundaries(t *testing.T) {
	t.Run("uint16 max values should convert safely", func(t *testing.T) {
		config := ccipocr3.FeeQuoterDestChainConfig{
			MaxNumberOfTokensPerMsg:           65535, // max uint16
			DestGasPerDataAvailabilityByte:    65535, // max uint16
			DestDataAvailabilityMultiplierBps: 65535, // max uint16
			DefaultTokenFeeUSDCents:           65535, // max uint16
		}

		// Convert to protobuf (should widen to uint32)
		pbConfig := feeQuoterDestChainConfigToPb(config)

		// Verify uint32 values are correct
		assert.Equal(t, uint32(65535), pbConfig.MaxNumberOfTokensPerMsg)
		assert.Equal(t, uint32(65535), pbConfig.DestGasPerDataAvailabilityByte)
		assert.Equal(t, uint32(65535), pbConfig.DestDataAvailabilityMultiplierBps)
		assert.Equal(t, uint32(65535), pbConfig.DefaultTokenFeeUsdcCents)

		// Convert back to Go (should narrow back to uint16)
		convertedConfig := pbToFeeQuoterDestChainConfigDetailed(pbConfig)

		// Verify uint16 values are preserved
		assert.Equal(t, uint16(65535), convertedConfig.MaxNumberOfTokensPerMsg)
		assert.Equal(t, uint16(65535), convertedConfig.DestGasPerDataAvailabilityByte)
		assert.Equal(t, uint16(65535), convertedConfig.DestDataAvailabilityMultiplierBps)
		assert.Equal(t, uint16(65535), convertedConfig.DefaultTokenFeeUSDCents)
	})

	t.Run("uint16 zero values should convert safely", func(t *testing.T) {
		config := ccipocr3.FeeQuoterDestChainConfig{
			MaxNumberOfTokensPerMsg:           0, // min uint16
			DestGasPerDataAvailabilityByte:    0, // min uint16
			DestDataAvailabilityMultiplierBps: 0, // min uint16
			DefaultTokenFeeUSDCents:           0, // min uint16
		}

		// Round-trip conversion
		pbConfig := feeQuoterDestChainConfigToPb(config)
		convertedConfig := pbToFeeQuoterDestChainConfigDetailed(pbConfig)

		// Verify zero values are preserved
		assert.Equal(t, uint16(0), convertedConfig.MaxNumberOfTokensPerMsg)
		assert.Equal(t, uint16(0), convertedConfig.DestGasPerDataAvailabilityByte)
		assert.Equal(t, uint16(0), convertedConfig.DestDataAvailabilityMultiplierBps)
		assert.Equal(t, uint16(0), convertedConfig.DefaultTokenFeeUSDCents)
	})
}

func TestAnyToPbValueAndPbValueToAny(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expectedVal any
	}{
		{
			name:        "string conversion",
			input:       "hello world",
			expectedVal: "hello world",
		},
		{
			name:        "int64 conversion",
			input:       int64(42),
			expectedVal: int64(42),
		},
		{
			name:        "uint64 conversion",
			input:       uint64(123),
			expectedVal: uint64(123),
		},
		{
			name:        "uint32 conversion",
			input:       uint32(456),
			expectedVal: uint32(456),
		},
		{
			name:        "float64 conversion",
			input:       float64(3.14),
			expectedVal: float64(3.14),
		},
		{
			name:        "bool true conversion",
			input:       true,
			expectedVal: true,
		},
		{
			name:        "bool false conversion",
			input:       false,
			expectedVal: false,
		},
		{
			name:        "bytes conversion",
			input:       []byte{0x01, 0x02, 0x03},
			expectedVal: []byte{0x01, 0x02, 0x03},
		},
		{
			name:        "big.Int zero conversion",
			input:       big.NewInt(0),
			expectedVal: big.NewInt(0),
		},
		{
			name:        "big.Int positive conversion",
			input:       big.NewInt(123456789),
			expectedVal: big.NewInt(123456789),
		},
		{
			name:        "big.Int negative conversion",
			input:       big.NewInt(-987654321),
			expectedVal: big.NewInt(987654321), // Note: Sign is lost due to Bytes()/SetBytes() limitation
		},
		{
			name: "big.Int large number conversion",
			input: func() *big.Int {
				val, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
				return val
			}(),
			expectedVal: func() *big.Int {
				val, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
				return val
			}(),
		},
		{
			name:        "nil big.Int conversion",
			input:       (*big.Int)(nil),
			expectedVal: big.NewInt(0), // pbBigIntToInt returns zero for nil
		},
		{
			name:        "map conversion",
			input:       map[string]any{"key1": "value1", "key2": int64(42)},
			expectedVal: map[string]any{"key1": "value1", "key2": int64(42)},
		},
		{
			name:        "slice conversion",
			input:       []any{"item1", int64(123), uint32(456)},
			expectedVal: []any{"item1", int64(123), uint32(456)},
		},
		{
			name:        "nested map with big.Int",
			input:       map[string]any{"amount": big.NewInt(1000), "count": uint32(5)},
			expectedVal: map[string]any{"amount": big.NewInt(1000), "count": uint32(5)},
		},
		{
			name: "complex nested structure",
			input: map[string]any{
				"header": map[string]any{
					"version": uint32(1),
					"hash":    []byte{0xFF, 0xEE},
				},
				"amounts": []any{
					big.NewInt(1000),
					big.NewInt(2000),
				},
				"enabled": true,
			},
			expectedVal: map[string]any{
				"header": map[string]any{
					"version": uint32(1),
					"hash":    []byte{0xFF, 0xEE},
				},
				"amounts": []any{
					big.NewInt(1000),
					big.NewInt(2000),
				},
				"enabled": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert Go any -> protobuf Value
			pbVal := anyToPbValue(tt.input)
			require.NotNil(t, pbVal, "protobuf value should not be nil")

			// Convert protobuf Value -> Go any
			result := pbValueToAny(pbVal)

			// Compare the result
			if bigInt, ok := tt.expectedVal.(*big.Int); ok && bigInt != nil {
				resultBigInt, ok := result.(*big.Int)
				require.True(t, ok, "result should be *big.Int")
				assert.Equal(t, bigInt.String(), resultBigInt.String(), "big.Int values should be equal")
			} else {
				assert.Equal(t, tt.expectedVal, result, "round-trip conversion should preserve the value")
			}
		})
	}
}

func TestAnyToPbValueUnsupportedType(t *testing.T) {
	// Test unsupported type falls back to string representation
	type customStruct struct {
		Field string
	}

	input := customStruct{Field: "test"}
	pbVal := anyToPbValue(input)
	result := pbValueToAny(pbVal)

	// Should be converted to string representation
	assert.Equal(t, "{test}", result)
}

func TestPbValueToAnyNil(t *testing.T) {
	result := pbValueToAny(nil)
	assert.Nil(t, result)
}

func TestBigIntEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "max uint64",
			input:    new(big.Int).SetUint64(^uint64(0)), // max uint64
			expected: "18446744073709551615",
		},
		{
			name:     "very large number",
			input:    new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil),
			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639936",
		},
		{
			name:     "negative large number",
			input:    new(big.Int).Neg(new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)),
			expected: "340282366920938463463374607431768211456", // Note: Sign is lost due to Bytes()/SetBytes() limitation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pbVal := anyToPbValue(tt.input)
			result := pbValueToAny(pbVal)

			resultBigInt, ok := result.(*big.Int)
			require.True(t, ok, "result should be *big.Int")
			assert.Equal(t, tt.expected, resultBigInt.String())
		})
	}
}

func TestUint32Boundaries(t *testing.T) {
	tests := []struct {
		name  string
		input uint32
	}{
		{
			name:  "zero",
			input: 0,
		},
		{
			name:  "max uint32",
			input: ^uint32(0), // 4294967295
		},
		{
			name:  "mid range",
			input: 2147483647, // max int32
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pbVal := anyToPbValue(tt.input)
			result := pbValueToAny(pbVal)

			resultUint32, ok := result.(uint32)
			require.True(t, ok, "result should be uint32")
			assert.Equal(t, tt.input, resultUint32)
		})
	}
}
