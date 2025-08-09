package ccipocr3

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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

// TestMapConversionUsingValuesPackage tests the conversion between Go maps and protobuf Maps
func TestMapConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "simple values",
			input: map[string]any{
				"string":  "hello",
				"int64":   int64(42),
				"float64": float64(3.14),
				"bool":    true,
				"bytes":   []byte{0x01, 0x02, 0x03},
			},
			expected: map[string]any{
				"string":  "hello",
				"int64":   int64(42),
				"float64": float64(3.14),
				"bool":    true,
				"bytes":   []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name: "big.Int values",
			input: map[string]any{
				"bigint_zero":     big.NewInt(0),
				"bigint_positive": big.NewInt(123456789),
				"bigint_negative": big.NewInt(-987654321),
			},
			expected: map[string]any{
				"bigint_zero":     big.NewInt(0),
				"bigint_positive": big.NewInt(123456789),
				"bigint_negative": big.NewInt(-987654321),
			},
		},
		{
			name: "nested maps and lists",
			input: map[string]any{
				"nested_map": map[string]any{
					"inner_key": "inner_value",
					"inner_num": int64(100),
				},
				"list": []any{
					"item1",
					int64(200),
					big.NewInt(300),
				},
			},
			expected: map[string]any{
				"nested_map": map[string]any{
					"inner_key": "inner_value",
					"inner_num": int64(100),
				},
				"list": []any{
					"item1",
					int64(200),
					big.NewInt(300),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert Go map -> protobuf Map
			pbMap, err := goMapToPbMap(tt.input)
			require.NoError(t, err, "conversion to protobuf Map should not fail")

			// Convert protobuf Map -> Go map
			result, err := pbMapToGoMap(pbMap)
			require.NoError(t, err, "conversion from protobuf Map should not fail")

			// Compare the result
			for key, expectedVal := range tt.expected {
				actualVal, exists := result[key]
				require.True(t, exists, "key %s should exist in result", key)

				if bigInt, ok := expectedVal.(*big.Int); ok {
					resultBigInt, ok := actualVal.(*big.Int)
					require.True(t, ok, "result[%s] should be *big.Int", key)
					assert.Equal(t, bigInt.String(), resultBigInt.String(), "big.Int values should be equal for key %s", key)
				} else {
					assert.Equal(t, expectedVal, actualVal, "values should be equal for key %s", key)
				}
			}
		})
	}
}

func TestMapConversionNilHandling(t *testing.T) {
	t.Run("nil Go map to protobuf", func(t *testing.T) {
		pbMap, err := goMapToPbMap(nil)
		require.NoError(t, err)
		assert.Nil(t, pbMap)
	})

	t.Run("nil protobuf map to Go", func(t *testing.T) {
		goMap, err := pbMapToGoMap(nil)
		require.NoError(t, err)
		assert.Nil(t, goMap)
	})

	t.Run("empty protobuf map to Go", func(t *testing.T) {
		emptyPbMap := &pb.Map{Fields: map[string]*pb.Value{}}
		goMap, err := pbMapToGoMap(emptyPbMap)
		require.NoError(t, err)
		assert.NotNil(t, goMap)
		assert.Equal(t, map[string]any{}, goMap)
	})
}

// TestSourceChainConfigConversion tests the new SourceChainConfig conversion functions
func TestSourceChainConfigConversion(t *testing.T) {
	testCases := []struct {
		name   string
		config ccipocr3.SourceChainConfig
	}{
		{
			name: "SourceChainConfig with all fields populated",
			config: ccipocr3.SourceChainConfig{
				Router:                    []byte("router-address-123"),
				IsEnabled:                 true,
				IsRMNVerificationDisabled: false,
				MinSeqNr:                  uint64(1000),
				OnRamp:                    ccipocr3.UnknownAddress("onramp-address-456"),
			},
		},
		{
			name: "SourceChainConfig with disabled state",
			config: ccipocr3.SourceChainConfig{
				Router:                    []byte("disabled-router"),
				IsEnabled:                 false,
				IsRMNVerificationDisabled: true,
				MinSeqNr:                  uint64(0),
				OnRamp:                    ccipocr3.UnknownAddress("disabled-onramp"),
			},
		},
		{
			name: "SourceChainConfig with empty addresses",
			config: ccipocr3.SourceChainConfig{
				Router:                    []byte{},
				IsEnabled:                 true,
				IsRMNVerificationDisabled: false,
				MinSeqNr:                  uint64(9999999),
				OnRamp:                    ccipocr3.UnknownAddress{},
			},
		},
		{
			name: "SourceChainConfig with max sequence number",
			config: ccipocr3.SourceChainConfig{
				Router:                    []byte("max-seq-router"),
				IsEnabled:                 true,
				IsRMNVerificationDisabled: true,
				MinSeqNr:                  ^uint64(0), // max uint64
				OnRamp:                    ccipocr3.UnknownAddress("max-seq-onramp"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbConfig := sourceChainConfigToPb(tc.config)
			require.NotNil(t, pbConfig, "protobuf config should not be nil")

			// Verify protobuf values
			assert.Equal(t, tc.config.Router, pbConfig.Router, "Router should be preserved")
			assert.Equal(t, tc.config.IsEnabled, pbConfig.IsEnabled, "IsEnabled should be preserved")
			assert.Equal(t, tc.config.IsRMNVerificationDisabled, pbConfig.IsRmnVerificationDisabled, "IsRMNVerificationDisabled should be preserved")
			assert.Equal(t, tc.config.MinSeqNr, pbConfig.MinSeqNr, "MinSeqNr should be preserved")
			assert.Equal(t, []byte(tc.config.OnRamp), pbConfig.OnRamp, "OnRamp should be preserved as bytes")

			// Convert Protobuf -> Go (round-trip)
			convertedConfig := pbToSourceChainConfig(pbConfig)

			// Verify round-trip conversion preserves all data
			assert.Equal(t, tc.config.Router, convertedConfig.Router, "Router should survive round-trip")
			assert.Equal(t, tc.config.IsEnabled, convertedConfig.IsEnabled, "IsEnabled should survive round-trip")
			assert.Equal(t, tc.config.IsRMNVerificationDisabled, convertedConfig.IsRMNVerificationDisabled, "IsRMNVerificationDisabled should survive round-trip")
			assert.Equal(t, tc.config.MinSeqNr, convertedConfig.MinSeqNr, "MinSeqNr should survive round-trip")
			assert.Equal(t, []byte(tc.config.OnRamp), []byte(convertedConfig.OnRamp), "OnRamp should survive round-trip as UnknownAddress")
		})
	}
}

// TestSourceChainConfigNilHandling tests nil handling for SourceChainConfig conversion
func TestSourceChainConfigNilHandling(t *testing.T) {
	t.Run("pbToSourceChainConfig with nil input", func(t *testing.T) {
		result := pbToSourceChainConfig(nil)
		expected := ccipocr3.SourceChainConfig{}
		assert.Equal(t, expected, result, "nil protobuf should convert to zero value SourceChainConfig")
	})

	t.Run("sourceChainConfigToPb with zero value", func(t *testing.T) {
		zeroConfig := ccipocr3.SourceChainConfig{}
		pbConfig := sourceChainConfigToPb(zeroConfig)
		require.NotNil(t, pbConfig, "protobuf should not be nil even for zero value input")

		// Verify zero values are preserved
		assert.Equal(t, []byte(nil), pbConfig.Router)
		assert.Equal(t, false, pbConfig.IsEnabled)
		assert.Equal(t, false, pbConfig.IsRmnVerificationDisabled)
		assert.Equal(t, uint64(0), pbConfig.MinSeqNr)
		assert.Equal(t, []byte(nil), pbConfig.OnRamp)
	})
}
