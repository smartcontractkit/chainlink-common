package ccipocr3

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
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
						DestExecData:      []byte("dest-exec-data"),
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
			// With nil BigInt values, protobuf fields should be nil
			if tc.message.FeeTokenAmount.Int == nil {
				assert.Nil(t, pbMessage.FeeTokenAmount)
			} else {
				assert.NotNil(t, pbMessage.FeeTokenAmount)
			}
			if tc.message.FeeValueJuels.Int == nil {
				assert.Nil(t, pbMessage.FeeValueJuels)
			} else {
				assert.NotNil(t, pbMessage.FeeValueJuels)
			}

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

// TestTokenPriceMapConversion tests the new TokenPriceMap conversion functions
func TestTokenPriceMapConversion(t *testing.T) {
	testCases := []struct {
		name     string
		priceMap ccipocr3.TokenPriceMap
	}{
		{
			name: "TokenPriceMap with multiple tokens",
			priceMap: ccipocr3.TokenPriceMap{
				ccipocr3.UnknownEncodedAddress("token1"): ccipocr3.NewBigInt(big.NewInt(1000000)),
				ccipocr3.UnknownEncodedAddress("token2"): ccipocr3.NewBigInt(big.NewInt(2500000)),
				ccipocr3.UnknownEncodedAddress("token3"): ccipocr3.NewBigInt(big.NewInt(500000)),
			},
		},
		{
			name:     "Empty TokenPriceMap",
			priceMap: ccipocr3.TokenPriceMap{},
		},
		{
			name: "TokenPriceMap with zero prices",
			priceMap: ccipocr3.TokenPriceMap{
				ccipocr3.UnknownEncodedAddress("zero-token"): ccipocr3.NewBigInt(big.NewInt(0)),
			},
		},
		{
			name: "TokenPriceMap with large prices",
			priceMap: ccipocr3.TokenPriceMap{
				ccipocr3.UnknownEncodedAddress("large-token"): func() ccipocr3.BigInt {
					val, _ := new(big.Int).SetString("999999999999999999999999999999", 10)
					return ccipocr3.NewBigInt(val)
				}(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbMap := tokenPriceMapToPb(tc.priceMap)

			if tc.priceMap == nil {
				assert.Nil(t, pbMap)
				return
			}

			require.NotNil(t, pbMap)
			assert.Equal(t, len(tc.priceMap), len(pbMap))

			// Verify protobuf values
			for token, price := range tc.priceMap {
				pbPrice, exists := pbMap[string(token)]
				require.True(t, exists, "token %s should exist in protobuf map", string(token))
				require.NotNil(t, pbPrice)
				assert.Equal(t, price.Int.Bytes(), pbPrice.Value)
			}

			// Convert Protobuf -> Go (round-trip)
			convertedMap := pbToTokenPriceMap(pbMap)

			// Verify round-trip conversion
			assert.Equal(t, len(tc.priceMap), len(convertedMap))
			for token, originalPrice := range tc.priceMap {
				convertedPrice, exists := convertedMap[token]
				require.True(t, exists, "token %s should exist in converted map", string(token))
				assert.Equal(t, originalPrice.Int.String(), convertedPrice.Int.String(), "price should survive round-trip for token %s", string(token))
			}
		})
	}
}

func TestTokenPriceMapNilHandling(t *testing.T) {
	t.Run("nil TokenPriceMap to protobuf", func(t *testing.T) {
		pbMap := tokenPriceMapToPb(nil)
		assert.Nil(t, pbMap)
	})

	t.Run("nil protobuf map to TokenPriceMap", func(t *testing.T) {
		priceMap := pbToTokenPriceMap(nil)
		assert.Nil(t, priceMap)
	})

	t.Run("empty protobuf map to TokenPriceMap", func(t *testing.T) {
		emptyPbMap := make(map[string]*ccipocr3pb.BigInt)
		priceMap := pbToTokenPriceMap(emptyPbMap)
		require.NotNil(t, priceMap)
		assert.Equal(t, 0, len(priceMap))
	})
}

// TestMessageTokenIDMapConversion tests the MessageTokenID map conversion functions
func TestMessageTokenIDMapConversion(t *testing.T) {
	testCases := []struct {
		name     string
		tokenMap map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount
	}{
		{
			name: "MessageTokenID map with multiple tokens",
			tokenMap: map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{
				ccipocr3.NewMessageTokenID(1, 0): {
					SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-1"),
					DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-1"),
					ExtraData:         ccipocr3.Bytes("extra-data-1"),
					Amount:            ccipocr3.NewBigInt(big.NewInt(1000)),
					DestExecData:      ccipocr3.Bytes("dest-exec-data-1"),
				},
				ccipocr3.NewMessageTokenID(2, 1): {
					SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-2"),
					DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-2"),
					ExtraData:         ccipocr3.Bytes("extra-data-2"),
					Amount:            ccipocr3.NewBigInt(big.NewInt(2000)),
					DestExecData:      ccipocr3.Bytes("dest-exec-data-2"),
				},
				ccipocr3.NewMessageTokenID(10, 5): {
					SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-10"),
					DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-10"),
					ExtraData:         ccipocr3.Bytes(""),
					Amount:            ccipocr3.NewBigInt(big.NewInt(0)),
					DestExecData:      ccipocr3.Bytes(""),
				},
			},
		},
		{
			name:     "Empty MessageTokenID map",
			tokenMap: map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{},
		},
		{
			name: "Single MessageTokenID with large values",
			tokenMap: map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{
				ccipocr3.NewMessageTokenID(999999, 255): {
					SourcePoolAddress: ccipocr3.UnknownAddress("very-long-source-pool-address-with-many-characters"),
					DestTokenAddress:  ccipocr3.UnknownAddress("very-long-dest-token-address-with-many-characters"),
					ExtraData:         ccipocr3.Bytes("very long extra data with many characters that tests the handling of large data"),
					Amount: func() ccipocr3.BigInt {
						val, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
						return ccipocr3.NewBigInt(val)
					}(),
					DestExecData: ccipocr3.Bytes("very long dest exec data with many characters that tests the handling of large execution data asdlfk(&HEDHSKJ#OIUOIJDL)(@#UE)(#U(R&FH(E&HF0x"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbMap := messageTokenIDMapToPb(tc.tokenMap)

			if tc.tokenMap == nil {
				assert.Nil(t, pbMap)
				return
			}

			require.NotNil(t, pbMap)
			assert.Equal(t, len(tc.tokenMap), len(pbMap))

			// Verify protobuf values
			for tokenID, amount := range tc.tokenMap {
				tokenIDStr := tokenID.String()
				pbAmount, exists := pbMap[tokenIDStr]
				require.True(t, exists, "tokenID %s should exist in protobuf map", tokenIDStr)
				require.NotNil(t, pbAmount)
				assert.Equal(t, []byte(amount.SourcePoolAddress), pbAmount.SourcePoolAddress)
				assert.Equal(t, []byte(amount.DestTokenAddress), pbAmount.DestTokenAddress)
				assert.Equal(t, []byte(amount.ExtraData), pbAmount.ExtraData)
				assert.Equal(t, amount.Amount.Int.Bytes(), pbAmount.Amount.Value)
				assert.Equal(t, []byte(amount.DestExecData), pbAmount.DestExecData)
			}

			// Convert Protobuf -> Go (round-trip)
			convertedMap, err := pbToMessageTokenIDMap(pbMap)
			require.NoError(t, err)

			// Verify round-trip conversion
			assert.Equal(t, len(tc.tokenMap), len(convertedMap))
			for tokenID, originalAmount := range tc.tokenMap {
				convertedAmount, exists := convertedMap[tokenID]
				require.True(t, exists, "tokenID %s should exist in converted map", tokenID.String())
				assert.Equal(t, []byte(originalAmount.SourcePoolAddress), []byte(convertedAmount.SourcePoolAddress))
				assert.Equal(t, []byte(originalAmount.DestTokenAddress), []byte(convertedAmount.DestTokenAddress))
				assert.Equal(t, []byte(originalAmount.ExtraData), []byte(convertedAmount.ExtraData))
				assert.Equal(t, originalAmount.Amount.Int.String(), convertedAmount.Amount.Int.String())
				assert.Equal(t, []byte(originalAmount.DestExecData), []byte(convertedAmount.DestExecData))
			}
		})
	}
}

func TestMessageTokenIDMapErrorHandling(t *testing.T) {
	t.Run("invalid tokenID string should return error", func(t *testing.T) {
		pbMap := map[string]*ccipocr3pb.RampTokenAmount{
			"invalid-format": {
				SourcePoolAddress: []byte("test"),
				DestTokenAddress:  []byte("test"),
				ExtraData:         []byte("test"),
				Amount:            &ccipocr3pb.BigInt{Value: []byte{0x01}},
			},
		}

		_, err := pbToMessageTokenIDMap(pbMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse MessageTokenID")
	})

	t.Run("nil protobuf map should not error", func(t *testing.T) {
		result, err := pbToMessageTokenIDMap(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil Go map should return nil", func(t *testing.T) {
		result := messageTokenIDMapToPb(nil)
		assert.Nil(t, result)
	})
}

// TestMessagesByTokenIDConversion tests the messages by token ID conversion functions
func TestMessagesByTokenIDConversion(t *testing.T) {
	testCases := []struct {
		name     string
		messages map[ccipocr3.MessageTokenID]ccipocr3.Bytes
	}{
		{
			name: "Messages with multiple token IDs",
			messages: map[ccipocr3.MessageTokenID]ccipocr3.Bytes{
				ccipocr3.NewMessageTokenID(1, 0): ccipocr3.Bytes("usdc-message-data-1"),
				ccipocr3.NewMessageTokenID(2, 1): ccipocr3.Bytes("usdc-message-data-2"),
				ccipocr3.NewMessageTokenID(5, 3): ccipocr3.Bytes("usdc-message-data-5"),
			},
		},
		{
			name:     "Empty messages map",
			messages: map[ccipocr3.MessageTokenID]ccipocr3.Bytes{},
		},
		{
			name: "Messages with binary data",
			messages: map[ccipocr3.MessageTokenID]ccipocr3.Bytes{
				ccipocr3.NewMessageTokenID(100, 50): ccipocr3.Bytes([]byte{0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}),
			},
		},
		{
			name: "Messages with empty data",
			messages: map[ccipocr3.MessageTokenID]ccipocr3.Bytes{
				ccipocr3.NewMessageTokenID(0, 0): ccipocr3.Bytes(""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbMap := messagesByTokenIDToPb(tc.messages)

			if tc.messages == nil {
				assert.Nil(t, pbMap)
				return
			}

			require.NotNil(t, pbMap)
			assert.Equal(t, len(tc.messages), len(pbMap))

			// Verify protobuf values
			for tokenID, messageBytes := range tc.messages {
				tokenIDStr := tokenID.String()
				pbBytes, exists := pbMap[tokenIDStr]
				require.True(t, exists, "tokenID %s should exist in protobuf map", tokenIDStr)
				assert.Equal(t, []byte(messageBytes), pbBytes)
			}

			// Convert Protobuf -> Go (round-trip)
			convertedMap, err := pbToMessagesByTokenID(pbMap)
			require.NoError(t, err)

			// Verify round-trip conversion
			assert.Equal(t, len(tc.messages), len(convertedMap))
			for tokenID, originalBytes := range tc.messages {
				convertedBytes, exists := convertedMap[tokenID]
				require.True(t, exists, "tokenID %s should exist in converted map", tokenID.String())
				assert.Equal(t, []byte(originalBytes), []byte(convertedBytes))
			}
		})
	}
}

func TestMessagesByTokenIDErrorHandling(t *testing.T) {
	t.Run("invalid tokenID string should return error", func(t *testing.T) {
		pbMap := map[string][]byte{
			"not-a-valid-format": []byte("test-message"),
		}

		_, err := pbToMessagesByTokenID(pbMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse MessageTokenID")
	})

	t.Run("nil protobuf map should not error", func(t *testing.T) {
		result, err := pbToMessagesByTokenID(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil Go map should return nil", func(t *testing.T) {
		result := messagesByTokenIDToPb(nil)
		assert.Nil(t, result)
	})
}

// TestRampTokenAmountDestExecDataRoundTrip tests that DestExecData is properly preserved in round-trip conversions
func TestRampTokenAmountDestExecDataRoundTrip(t *testing.T) {
	testCases := []struct {
		name        string
		tokenAmount ccipocr3.RampTokenAmount
	}{
		{
			name: "RampTokenAmount with DestExecData",
			tokenAmount: ccipocr3.RampTokenAmount{
				SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-address"),
				DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-address"),
				ExtraData:         ccipocr3.Bytes("extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(12345)),
				DestExecData:      ccipocr3.Bytes("dest-exec-data-content"),
			},
		},
		{
			name: "RampTokenAmount with empty DestExecData",
			tokenAmount: ccipocr3.RampTokenAmount{
				SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-address"),
				DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-address"),
				ExtraData:         ccipocr3.Bytes("extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(54321)),
				DestExecData:      ccipocr3.Bytes(""),
			},
		},
		{
			name: "RampTokenAmount with nil DestExecData",
			tokenAmount: ccipocr3.RampTokenAmount{
				SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-address"),
				DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-address"),
				ExtraData:         ccipocr3.Bytes("extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(98765)),
				DestExecData:      nil,
			},
		},
		{
			name: "RampTokenAmount with large DestExecData",
			tokenAmount: ccipocr3.RampTokenAmount{
				SourcePoolAddress: ccipocr3.UnknownAddress("source-pool-address"),
				DestTokenAddress:  ccipocr3.UnknownAddress("dest-token-address"),
				ExtraData:         ccipocr3.Bytes("extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(11111)),
				DestExecData:      ccipocr3.Bytes("very long dest exec data with many characters that tests the handling of large execution data asdlfk(&HEDHSKJ#OIUOIJDL)(@#UE)(#U(R&FH(E&HF0x"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test via pbToTokenAmounts and messageToPb functions
			message := ccipocr3.Message{
				Header: ccipocr3.RampMessageHeader{
					MessageID:           [32]byte{0x01, 0x02, 0x03, 0x04},
					SourceChainSelector: ccipocr3.ChainSelector(1),
					DestChainSelector:   ccipocr3.ChainSelector(2),
					SequenceNumber:      ccipocr3.SeqNum(1),
					Nonce:               1,
					MsgHash:             [32]byte{0xAA, 0xBB, 0xCC, 0xDD},
					OnRamp:              []byte("onramp"),
					TxHash:              "0x123",
				},
				Sender:         []byte("sender"),
				Data:           []byte("data"),
				Receiver:       []byte("receiver"),
				ExtraArgs:      []byte("extra-args"),
				FeeToken:       []byte("fee-token"),
				FeeTokenAmount: ccipocr3.NewBigInt(big.NewInt(1000)),
				FeeValueJuels:  ccipocr3.NewBigInt(big.NewInt(2000)),
				TokenAmounts:   []ccipocr3.RampTokenAmount{tc.tokenAmount},
			}

			// Convert to protobuf
			pbMessage := messageToPb(message)
			require.NotNil(t, pbMessage)
			require.Len(t, pbMessage.TokenAmounts, 1)

			// Verify DestExecData is preserved in protobuf
			pbTokenAmount := pbMessage.TokenAmounts[0]
			assert.Equal(t, []byte(tc.tokenAmount.DestExecData), pbTokenAmount.DestExecData)

			// Convert back to Go struct
			convertedMessage := pbToMessage(pbMessage)
			require.Len(t, convertedMessage.TokenAmounts, 1)

			// Verify DestExecData is preserved in round-trip
			convertedTokenAmount := convertedMessage.TokenAmounts[0]
			assert.Equal(t, []byte(tc.tokenAmount.DestExecData), []byte(convertedTokenAmount.DestExecData))

			// Test via messageTokenIDMapToPb and pbToMessageTokenIDMap functions
			tokenMap := map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{
				ccipocr3.NewMessageTokenID(1, 0): tc.tokenAmount,
			}

			// Convert to protobuf map
			pbMap := messageTokenIDMapToPb(tokenMap)
			require.NotNil(t, pbMap)
			require.Len(t, pbMap, 1)

			// Verify DestExecData is preserved in protobuf map
			pbAmount := pbMap["1_0"]
			require.NotNil(t, pbAmount)
			assert.Equal(t, []byte(tc.tokenAmount.DestExecData), pbAmount.DestExecData)

			// Convert back to Go map
			convertedMap, err := pbToMessageTokenIDMap(pbMap)
			require.NoError(t, err)
			require.Len(t, convertedMap, 1)

			// Verify DestExecData is preserved in round-trip map
			convertedAmount := convertedMap[ccipocr3.NewMessageTokenID(1, 0)]
			assert.Equal(t, []byte(tc.tokenAmount.DestExecData), []byte(convertedAmount.DestExecData))
		})
	}
}

// TestTokenUpdatesUnixConversion tests the TimestampedUnixBig token updates conversion functions
func TestTokenUpdatesUnixConversion(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		updates map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig
	}{
		{
			name: "Unix token updates with multiple tokens",
			updates: map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig{
				ccipocr3.UnknownEncodedAddress("token1"): {
					Timestamp: uint32(testTime.Unix()),
					Value:     big.NewInt(1500000),
				},
				ccipocr3.UnknownEncodedAddress("token2"): {
					Timestamp: uint32(testTime.Add(time.Hour).Unix()),
					Value:     big.NewInt(2500000),
				},
				ccipocr3.UnknownEncodedAddress("token3"): {
					Timestamp: uint32(testTime.Add(2 * time.Hour).Unix()),
					Value:     big.NewInt(750000),
				},
			},
		},
		{
			name:    "Empty unix token updates",
			updates: map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig{},
		},
		{
			name: "Unix token update with zero value",
			updates: map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig{
				ccipocr3.UnknownEncodedAddress("zero-token"): {
					Timestamp: uint32(testTime.Unix()),
					Value:     big.NewInt(0),
				},
			},
		},
		{
			name: "Unix token update with large value",
			updates: map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig{
				ccipocr3.UnknownEncodedAddress("large-token"): {
					Timestamp: uint32(testTime.Unix()),
					Value: func() *big.Int {
						val, _ := new(big.Int).SetString("999999999999999999999999999999", 10)
						return val
					}(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert Go -> Protobuf
			pbMap := tokenUpdatesUnixToPb(tc.updates)

			if tc.updates == nil {
				assert.Nil(t, pbMap)
				return
			}

			require.NotNil(t, pbMap)
			assert.Equal(t, len(tc.updates), len(pbMap))

			// Verify protobuf values
			for token, update := range tc.updates {
				pbUpdate, exists := pbMap[string(token)]
				require.True(t, exists, "token %s should exist in protobuf map", string(token))
				require.NotNil(t, pbUpdate)
				require.NotNil(t, pbUpdate.Value)

				assert.Equal(t, update.Timestamp, pbUpdate.Timestamp)
				assert.Equal(t, update.Value.Bytes(), pbUpdate.Value.Value)
			}

			// Convert Protobuf -> Go (round-trip)
			convertedMap := pbToTokenUpdatesUnix(pbMap)

			// Verify round-trip conversion
			assert.Equal(t, len(tc.updates), len(convertedMap))
			for token, originalUpdate := range tc.updates {
				convertedUpdate, exists := convertedMap[token]
				require.True(t, exists, "token %s should exist in converted map", string(token))

				assert.Equal(t, originalUpdate.Timestamp, convertedUpdate.Timestamp)
				assert.Equal(t, originalUpdate.Value.String(), convertedUpdate.Value.String())
			}
		})
	}
}

func TestTokenUpdatesUnixNilHandling(t *testing.T) {
	t.Run("nil unix token updates to protobuf", func(t *testing.T) {
		pbMap := tokenUpdatesUnixToPb(nil)
		assert.Nil(t, pbMap)
	})

	t.Run("nil protobuf map to unix token updates", func(t *testing.T) {
		updates := pbToTokenUpdatesUnix(nil)
		assert.Nil(t, updates)
	})

	t.Run("empty protobuf map to unix token updates", func(t *testing.T) {
		emptyPbMap := make(map[string]*ccipocr3pb.TimestampedUnixBig)
		updates := pbToTokenUpdatesUnix(emptyPbMap)
		require.NotNil(t, updates)
		assert.Equal(t, 0, len(updates))
	})

	t.Run("protobuf map with nil value", func(t *testing.T) {
		pbMap := map[string]*ccipocr3pb.TimestampedUnixBig{
			"test-token": {
				Timestamp: 1705320000, // Some valid unix timestamp
				Value:     nil,
			},
		}

		// Should not panic and handle gracefully
		updates := pbToTokenUpdatesUnix(pbMap)
		require.NotNil(t, updates)
		require.Contains(t, updates, ccipocr3.UnknownEncodedAddress("test-token"))

		// When value is nil, should default to zero big.Int
		update := updates[("test-token")]
		assert.Equal(t, uint32(1705320000), update.Timestamp)
		assert.Equal(t, big.NewInt(0), update.Value)
	})
}

// TestPbBigIntToInt tests the pbBigIntToInt function with various inputs
func TestPbBigIntToInt(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ccipocr3pb.BigInt
		expected *big.Int
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty value bytes", // empty bytes are treated as zero
			input:    &ccipocr3pb.BigInt{Value: []byte{}},
			expected: big.NewInt(0),
		},
		{
			name:     "nil value bytes",
			input:    &ccipocr3pb.BigInt{Value: nil},
			expected: nil,
		},
		{
			name:     "zero value",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(0).Bytes()},
			expected: big.NewInt(0),
		},
		{
			name:     "positive small integer",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(42).Bytes()},
			expected: big.NewInt(42),
		},
		{
			name:     "positive large integer",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(1234567890).Bytes()},
			expected: big.NewInt(1234567890),
		},
		{
			name: "very large positive integer",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				val := new(big.Int)
				val.SetString("999999999999999999999999999999", 10)
				return val.Bytes()
			}()},
			expected: func() *big.Int {
				val := new(big.Int)
				val.SetString("999999999999999999999999999999", 10)
				return val
			}(),
		},
		{
			name: "maximum uint64 value",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				val := new(big.Int)
				val.SetUint64(^uint64(0)) // max uint64
				return val.Bytes()
			}()},
			expected: func() *big.Int {
				val := new(big.Int)
				val.SetUint64(^uint64(0))
				return val
			}(),
		},
		{
			name: "256-bit integer (32 bytes)",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				// Create a 256-bit integer (all bits set)
				bytes := make([]byte, 32)
				for i := range bytes {
					bytes[i] = 0xFF
				}
				return bytes
			}()},
			expected: func() *big.Int {
				bytes := make([]byte, 32)
				for i := range bytes {
					bytes[i] = 0xFF
				}
				return new(big.Int).SetBytes(bytes)
			}(),
		},
		{
			name:     "single byte value",
			input:    &ccipocr3pb.BigInt{Value: []byte{0xFF}},
			expected: big.NewInt(255),
		},
		{
			name:     "two byte value",
			input:    &ccipocr3pb.BigInt{Value: []byte{0x01, 0x00}},
			expected: big.NewInt(256),
		},
		{
			name:     "leading zero bytes (should be handled correctly)",
			input:    &ccipocr3pb.BigInt{Value: []byte{0x00, 0x00, 0x01, 0x00}},
			expected: big.NewInt(256),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pbBigIntToInt(tc.input)

			// Handle nil comparisons specially
			if tc.expected == nil {
				assert.Nil(t, result, "result should be nil when expected is nil")
			} else {
				assert.NotNil(t, result, "result should not be nil when expected is not nil")
				assert.Equal(t, tc.expected.String(), result.String(), "converted big.Int should match expected value")
				assert.Equal(t, tc.expected.Cmp(result), 0, "big.Int comparison should be equal")
			}
		})
	}
}

// TestPbToBigInt tests the pbToBigInt function with various inputs
func TestPbToBigInt(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ccipocr3pb.BigInt
		expected ccipocr3.BigInt
	}{
		{
			name:     "nil input should preserve nil",
			input:    nil,
			expected: ccipocr3.BigInt{Int: nil},
		},
		{
			name:     "empty value bytes should return zero BigInt",
			input:    &ccipocr3pb.BigInt{Value: []byte{}},
			expected: ccipocr3.BigInt{Int: big.NewInt(0)},
		},
		{
			name:     "nil value bytes should preserve nil",
			input:    &ccipocr3pb.BigInt{Value: nil},
			expected: ccipocr3.BigInt{Int: nil},
		},
		{
			name:     "zero value",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(0).Bytes()},
			expected: ccipocr3.NewBigInt(big.NewInt(0)),
		},
		{
			name:     "positive small integer",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(123).Bytes()},
			expected: ccipocr3.NewBigInt(big.NewInt(123)),
		},
		{
			name:     "positive large integer",
			input:    &ccipocr3pb.BigInt{Value: big.NewInt(9876543210).Bytes()},
			expected: ccipocr3.NewBigInt(big.NewInt(9876543210)),
		},
		{
			name: "very large positive integer",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				val := new(big.Int)
				val.SetString("123456789012345678901234567890", 10)
				return val.Bytes()
			}()},
			expected: func() ccipocr3.BigInt {
				val := new(big.Int)
				val.SetString("123456789012345678901234567890", 10)
				return ccipocr3.NewBigInt(val)
			}(),
		},
		{
			name: "maximum uint64 value",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				val := new(big.Int)
				val.SetUint64(^uint64(0)) // max uint64
				return val.Bytes()
			}()},
			expected: func() ccipocr3.BigInt {
				val := new(big.Int)
				val.SetUint64(^uint64(0))
				return ccipocr3.NewBigInt(val)
			}(),
		},
		{
			name: "256-bit integer (32 bytes)",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				// Create a 256-bit integer
				bytes := make([]byte, 32)
				for i := range bytes {
					bytes[i] = 0xAA // alternating bit pattern
				}
				return bytes
			}()},
			expected: func() ccipocr3.BigInt {
				bytes := make([]byte, 32)
				for i := range bytes {
					bytes[i] = 0xAA
				}
				return ccipocr3.NewBigInt(new(big.Int).SetBytes(bytes))
			}(),
		},
		{
			name:     "single byte maximum value",
			input:    &ccipocr3pb.BigInt{Value: []byte{0xFF}},
			expected: ccipocr3.NewBigInt(big.NewInt(255)),
		},
		{
			name:     "two byte value",
			input:    &ccipocr3pb.BigInt{Value: []byte{0xFF, 0xFF}},
			expected: ccipocr3.NewBigInt(big.NewInt(65535)),
		},
		{
			name: "ethereum wei amount (18 decimals)",
			input: &ccipocr3pb.BigInt{Value: func() []byte {
				// 1 ETH in wei = 10^18
				val := new(big.Int)
				val.SetString("1000000000000000000", 10)
				return val.Bytes()
			}()},
			expected: func() ccipocr3.BigInt {
				val := new(big.Int)
				val.SetString("1000000000000000000", 10)
				return ccipocr3.NewBigInt(val)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pbToBigInt(tc.input)

			// Handle nil comparisons specially
			if tc.expected.Int == nil {
				assert.Nil(t, result.Int, "result.Int should be nil when expected is nil")
			} else {
				assert.NotNil(t, result.Int, "result.Int should not be nil when expected is not nil")
				assert.Equal(t, tc.expected.Int.String(), result.Int.String(), "converted BigInt should match expected value")
				assert.Equal(t, tc.expected.Int.Cmp(result.Int), 0, "BigInt comparison should be equal")
			}
		})
	}
}

// TestPbBigIntRoundTrip tests round-trip conversion between protobuf BigInt and Go big.Int
func TestPbBigIntRoundTrip(t *testing.T) {
	testValues := []*big.Int{
		nil, // Test nil value
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(42),
		big.NewInt(255),
		big.NewInt(256),
		big.NewInt(65535),
		big.NewInt(65536),
		big.NewInt(1234567890),
		func() *big.Int {
			val := new(big.Int)
			val.SetString("999999999999999999999999999999", 10)
			return val
		}(),
		func() *big.Int {
			val := new(big.Int)
			val.SetUint64(^uint64(0)) // max uint64
			return val
		}(),
	}

	for i, originalValue := range testValues {
		t.Run(fmt.Sprintf("round_trip_%d", i), func(t *testing.T) {
			// Go big.Int -> protobuf BigInt -> Go big.Int
			pbBigInt := intToPbBigInt(originalValue)
			convertedValue := pbBigIntToInt(pbBigInt)

			// Handle nil case specially
			if originalValue == nil {
				assert.Nil(t, convertedValue, "nil should round-trip to nil")
			} else {
				assert.NotNil(t, convertedValue, "non-nil should round-trip to non-nil")
				assert.Equal(t, originalValue.String(), convertedValue.String(), "round-trip conversion should preserve value")
				assert.Equal(t, originalValue.Cmp(convertedValue), 0, "round-trip big.Int comparison should be equal")
			}
		})
	}
}

// TestPbToBigIntRoundTrip tests round-trip conversion between protobuf BigInt and ccipocr3.BigInt
func TestPbToBigIntRoundTrip(t *testing.T) {
	testValues := []ccipocr3.BigInt{
		ccipocr3.NewBigInt(big.NewInt(0)),
		ccipocr3.NewBigInt(big.NewInt(1)),
		ccipocr3.NewBigInt(big.NewInt(42)),
		ccipocr3.NewBigInt(big.NewInt(255)),
		ccipocr3.NewBigInt(big.NewInt(256)),
		ccipocr3.NewBigInt(big.NewInt(65535)),
		ccipocr3.NewBigInt(big.NewInt(65536)),
		ccipocr3.NewBigInt(big.NewInt(1234567890)),
		func() ccipocr3.BigInt {
			val := new(big.Int)
			val.SetString("999999999999999999999999999999", 10)
			return ccipocr3.NewBigInt(val)
		}(),
		func() ccipocr3.BigInt {
			val := new(big.Int)
			val.SetUint64(^uint64(0)) // max uint64
			return ccipocr3.NewBigInt(val)
		}(),
		// Test nil value specially
		ccipocr3.BigInt{Int: nil},
	}

	for i, originalValue := range testValues {
		t.Run(fmt.Sprintf("round_trip_%d", i), func(t *testing.T) {
			// ccipocr3.BigInt -> protobuf BigInt -> ccipocr3.BigInt
			pbBigInt := intToPbBigInt(originalValue.Int)
			convertedValue := pbToBigInt(pbBigInt)

			// Handle nil case specially
			if originalValue.Int == nil {
				assert.Nil(t, convertedValue.Int, "nil should round-trip to nil")
			} else {
				assert.NotNil(t, convertedValue.Int, "converted value should not be nil for non-nil input")
				assert.Equal(t, originalValue.Int.String(), convertedValue.Int.String(), "round-trip conversion should preserve value")
				assert.Equal(t, originalValue.Int.Cmp(convertedValue.Int), 0, "round-trip BigInt comparison should be equal")
			}
		})
	}
}

// TestPbBigIntEdgeCases tests edge cases and error conditions
func TestPbBigIntEdgeCases(t *testing.T) {
	t.Run("empty bytes should not panic", func(t *testing.T) {
		input := &ccipocr3pb.BigInt{Value: []byte{}}

		// Should not panic
		result1 := pbBigIntToInt(input)
		result2 := pbToBigInt(input)

		assert.Equal(t, big.NewInt(0).String(), result1.String())
		assert.Equal(t, big.NewInt(0).String(), result2.Int.String())
	})

	t.Run("single zero byte should equal zero", func(t *testing.T) {
		input := &ccipocr3pb.BigInt{Value: []byte{0x00}}

		result1 := pbBigIntToInt(input)
		result2 := pbToBigInt(input)

		assert.Equal(t, big.NewInt(0).String(), result1.String())
		assert.Equal(t, big.NewInt(0).String(), result2.Int.String())
	})

	t.Run("multiple zero bytes should equal zero", func(t *testing.T) {
		input := &ccipocr3pb.BigInt{Value: []byte{0x00, 0x00, 0x00, 0x00}}

		result1 := pbBigIntToInt(input)
		result2 := pbToBigInt(input)

		assert.Equal(t, big.NewInt(0).String(), result1.String())
		assert.Equal(t, big.NewInt(0).String(), result2.Int.String())
	})

	t.Run("large byte array should work correctly", func(t *testing.T) {
		// Create a 64-byte array (512 bits)
		bytes := make([]byte, 64)
		bytes[0] = 0x01 // Set the most significant bit to 1

		input := &ccipocr3pb.BigInt{Value: bytes}

		result1 := pbBigIntToInt(input)
		result2 := pbToBigInt(input)

		expected := new(big.Int).SetBytes(bytes)

		assert.Equal(t, expected.String(), result1.String())
		assert.Equal(t, expected.String(), result2.Int.String())
	})
}

// TestPbBigIntConsistency tests that both functions handle the same inputs consistently
func TestPbBigIntConsistency(t *testing.T) {
	testInputs := []*ccipocr3pb.BigInt{
		nil,
		{Value: nil},
		{Value: []byte{}},
		{Value: []byte{0x00}},
		{Value: []byte{0x01}},
		{Value: []byte{0xFF}},
		{Value: []byte{0x01, 0x00}},
		{Value: []byte{0xFF, 0xFF}},
		{Value: big.NewInt(42).Bytes()},
		{Value: big.NewInt(1234567890).Bytes()},
	}

	for i, input := range testInputs {
		t.Run(fmt.Sprintf("consistency_test_%d", i), func(t *testing.T) {
			result1 := pbBigIntToInt(input)
			result2 := pbToBigInt(input)

			// For non-nil inputs, both functions should produce equivalent numeric results
			// (except pbToBigInt may preserve nil differently)
			if input != nil {
				if result2.Int != nil {
					assert.Equal(t, result1.String(), result2.Int.String(),
						"pbBigIntToInt and pbToBigInt should produce equivalent numeric results")
				}
			}
		})
	}
}

// TestTokenInfoConversion tests the TokenInfo conversion functions
func TestTokenInfoConversion(t *testing.T) {
	testCases := []struct {
		name     string
		info     ccipocr3.TokenInfo
		expected ccipocr3.TokenInfo
	}{
		{
			name: "complete TokenInfo",
			info: ccipocr3.TokenInfo{
				AggregatorAddress: ccipocr3.UnknownEncodedAddress("0x1234567890123456789012345678901234567890"),
				DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(1000000000)), // 1%
				Decimals:          18,
			},
			expected: ccipocr3.TokenInfo{
				AggregatorAddress: ccipocr3.UnknownEncodedAddress("0x1234567890123456789012345678901234567890"),
				DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(1000000000)),
				Decimals:          18,
			},
		},
		{
			name: "minimal TokenInfo",
			info: ccipocr3.TokenInfo{
				AggregatorAddress: ccipocr3.UnknownEncodedAddress("0xabc"),
				DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(1)),
				Decimals:          6,
			},
			expected: ccipocr3.TokenInfo{
				AggregatorAddress: ccipocr3.UnknownEncodedAddress("0xabc"),
				DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(1)),
				Decimals:          6,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to protobuf and back
			pbInfo := tokenInfoToPb(tc.info)
			convertedInfo := pbToTokenInfo(pbInfo)

			assert.Equal(t, tc.expected.AggregatorAddress, convertedInfo.AggregatorAddress)
			assert.Equal(t, tc.expected.DeviationPPB.String(), convertedInfo.DeviationPPB.String())
			assert.Equal(t, tc.expected.Decimals, convertedInfo.Decimals)
		})
	}
}

// TestTokenInfoMapConversion tests the TokenInfo map conversion functions
func TestTokenInfoMapConversion(t *testing.T) {
	testMap := map[ccipocr3.UnknownEncodedAddress]ccipocr3.TokenInfo{
		"token1": {
			AggregatorAddress: ccipocr3.UnknownEncodedAddress("0x1111111111111111111111111111111111111111"),
			DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(2000000000)), // 2%
			Decimals:          18,
		},
		"token2": {
			AggregatorAddress: ccipocr3.UnknownEncodedAddress("0x2222222222222222222222222222222222222222"),
			DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(5000000000)), // 5%
			Decimals:          6,
		},
	}

	// Convert to protobuf and back
	pbMap := tokenInfoMapToPb(testMap)
	convertedMap := pbToTokenInfoMap(pbMap)

	assert.Len(t, convertedMap, 2)

	for token, expectedInfo := range testMap {
		convertedInfo, exists := convertedMap[token]
		assert.True(t, exists, "Token %s should exist in converted map", token)
		assert.Equal(t, expectedInfo.AggregatorAddress, convertedInfo.AggregatorAddress)
		assert.Equal(t, expectedInfo.DeviationPPB.String(), convertedInfo.DeviationPPB.String())
		assert.Equal(t, expectedInfo.Decimals, convertedInfo.Decimals)
	}
}

// TestTokenInfoMapNilHandling tests nil handling for TokenInfo map conversion
func TestTokenInfoMapNilHandling(t *testing.T) {
	t.Run("nil map to protobuf", func(t *testing.T) {
		result := tokenInfoMapToPb(nil)
		assert.Nil(t, result)
	})

	t.Run("nil protobuf map to Go", func(t *testing.T) {
		result := pbToTokenInfoMap(nil)
		assert.Nil(t, result)
	})

	t.Run("empty map round-trip", func(t *testing.T) {
		emptyMap := make(map[ccipocr3.UnknownEncodedAddress]ccipocr3.TokenInfo)
		pbMap := tokenInfoMapToPb(emptyMap)
		convertedMap := pbToTokenInfoMap(pbMap)
		assert.NotNil(t, convertedMap)
		assert.Len(t, convertedMap, 0)
	})
}

// TestExtraDataCodecBundleConversion tests the map conversion logic used by ExtraDataCodecBundle methods
// This test verifies that round-trip conversion preserves structure and semantic meaning
func TestExtraDataCodecBundleConversion(t *testing.T) {
	t.Run("Nil map round-trip", func(t *testing.T) {
		// Test that nil maps remain nil after round-trip
		pbMap, err := goMapToPbMap(nil)
		require.NoError(t, err)

		result, err := pbMapToGoMap(pbMap)
		require.NoError(t, err)
		assert.Nil(t, result, "nil input should result in nil output")
	})

	t.Run("Empty map round-trip", func(t *testing.T) {
		// Test that empty maps remain empty after round-trip
		emptyMap := map[string]any{}
		pbMap, err := goMapToPbMap(emptyMap)
		require.NoError(t, err)

		result, err := pbMapToGoMap(pbMap)
		require.NoError(t, err)
		assert.NotNil(t, result, "empty map should not become nil")
		assert.Equal(t, 0, len(result), "empty map should remain empty")
	})

	t.Run("ExtraArgs-like data structure preservation", func(t *testing.T) {
		// Test typical ExtraArgs structure with basic types
		input := map[string]any{
			"gasLimit": uint64(100000),
			"gasPrice": uint64(20000000000),
			"enabled":  true,
			"data":     []byte{0x01, 0x02, 0x03},
		}

		pbMap, err := goMapToPbMap(input)
		require.NoError(t, err, "conversion to protobuf should succeed")

		result, err := pbMapToGoMap(pbMap)
		require.NoError(t, err, "conversion from protobuf should succeed")

		// Verify structure is preserved
		assert.Equal(t, len(input), len(result), "map size should be preserved")

		// Verify all keys exist
		for key := range input {
			assert.Contains(t, result, key, "key %s should exist after round-trip", key)
		}

		// Verify specific values that should be exactly preserved
		assert.Equal(t, input["enabled"], result["enabled"], "bool values should be preserved")
		assert.Equal(t, input["data"], result["data"], "byte slice values should be preserved")

		// For numeric values, verify semantic equivalence (protobuf may normalize types)
		assert.Equal(t, fmt.Sprintf("%v", input["gasLimit"]), fmt.Sprintf("%v", result["gasLimit"]), "gasLimit should be semantically equal")
		assert.Equal(t, fmt.Sprintf("%v", input["gasPrice"]), fmt.Sprintf("%v", result["gasPrice"]), "gasPrice should be semantically equal")
	})

	t.Run("BigInt preservation", func(t *testing.T) {
		// Test various BigInt values that are commonly used in ExtraArgs/DestExecData
		input := map[string]any{
			"zero":     big.NewInt(0),
			"small":    big.NewInt(42),
			"large":    big.NewInt(1000000000000000000), // 1 ETH in wei
			"negative": big.NewInt(-123456789),
		}

		pbMap, err := goMapToPbMap(input)
		require.NoError(t, err)

		result, err := pbMapToGoMap(pbMap)
		require.NoError(t, err)

		// Verify all BigInt values are preserved exactly
		for key, originalVal := range input {
			resultVal, exists := result[key]
			require.True(t, exists, "key %s should exist", key)

			originalBigInt := originalVal.(*big.Int)
			resultBigInt, ok := resultVal.(*big.Int)
			require.True(t, ok, "result[%s] should be *big.Int", key)
			assert.Equal(t, originalBigInt.String(), resultBigInt.String(), "BigInt value should be preserved for key %s", key)
		}
	})

	t.Run("Nested structure preservation", func(t *testing.T) {
		// Test nested maps and arrays as might appear in DestExecData
		input := map[string]any{
			"version": uint32(1),
			"config": map[string]any{
				"maxGas":     uint64(500000),
				"multiplier": float64(1.5),
			},
			"amounts": []any{
				big.NewInt(1000),
				big.NewInt(2000),
			},
		}

		pbMap, err := goMapToPbMap(input)
		require.NoError(t, err)

		result, err := pbMapToGoMap(pbMap)
		require.NoError(t, err)

		// Verify top-level structure
		assert.Equal(t, len(input), len(result), "top-level map size should be preserved")

		// Verify nested map structure
		configResult, ok := result["config"].(map[string]any)
		require.True(t, ok, "config should remain a map")
		configInput := input["config"].(map[string]any)
		assert.Equal(t, len(configInput), len(configResult), "nested map size should be preserved")

		// Verify array structure
		amountsResult, ok := result["amounts"].([]any)
		require.True(t, ok, "amounts should remain an array")
		amountsInput := input["amounts"].([]any)
		assert.Equal(t, len(amountsInput), len(amountsResult), "array size should be preserved")

		// Verify BigInt values in array are preserved
		for i, originalAmount := range amountsInput {
			originalBigInt := originalAmount.(*big.Int)
			resultBigInt, ok := amountsResult[i].(*big.Int)
			require.True(t, ok, "array element should remain *big.Int")
			assert.Equal(t, originalBigInt.String(), resultBigInt.String(), "BigInt value should be preserved")
		}
	})
}
