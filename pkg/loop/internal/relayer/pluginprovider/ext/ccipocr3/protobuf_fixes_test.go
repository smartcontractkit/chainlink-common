package ccipocr3

import (
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
