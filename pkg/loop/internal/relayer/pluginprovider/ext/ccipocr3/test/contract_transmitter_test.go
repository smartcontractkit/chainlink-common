package test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestContractTransmitter(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	// Create a provider to get the contract transmitter
	provider := CCIPProvider(lggr)
	transmitter := provider.ContractTransmitter()

	t.Run("FromAccount", func(t *testing.T) {
		account, err := transmitter.FromAccount(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, account)
		assert.Equal(t, "some-account", string(account))
	})

	t.Run("Transmit", func(t *testing.T) {
		// Test data from the OCR3 contract transmitter test
		configDigest := [32]byte{1: 7, 13: 11, 31: 23}
		seqNr := uint64(3)
		reportWithInfo := ocr3types.ReportWithInfo[[]byte]{
			Report: []byte{41: 131},
			Info:   []byte("some-info"),
		}
		sigs := []libocr.AttributedOnchainSignature{{Signature: []byte{9: 8, 7: 6}, Signer: commontypes.OracleID(54)}}

		err := transmitter.Transmit(ctx, configDigest, seqNr, reportWithInfo, sigs)
		assert.NoError(t, err)
	})

	t.Run("TransmitWithWrongData", func(t *testing.T) {
		// Test with wrong data to verify error handling
		wrongConfigDigest := [32]byte{1: 1, 2: 2, 3: 3}
		seqNr := uint64(999) // Wrong sequence number
		reportWithInfo := ocr3types.ReportWithInfo[[]byte]{
			Report: []byte("wrong-report"),
			Info:   []byte("wrong-info"),
		}
		sigs := []libocr.AttributedOnchainSignature{{Signature: []byte("wrong"), Signer: commontypes.OracleID(99)}}

		err := transmitter.Transmit(ctx, wrongConfigDigest, seqNr, reportWithInfo, sigs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected configDigest")
	})
}

func TestContractTransmitterIntegration(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	// Create two providers and test their transmitters
	provider1 := CCIPProvider(lggr)
	provider2 := CCIPProvider(lggr)

	transmitter1 := provider1.ContractTransmitter()
	transmitter2 := provider2.ContractTransmitter()

	t.Run("SameImplementation", func(t *testing.T) {
		// Both should have the same account
		account1, err := transmitter1.FromAccount(ctx)
		assert.NoError(t, err)

		account2, err := transmitter2.FromAccount(ctx)
		assert.NoError(t, err)

		assert.Equal(t, account1, account2)
	})

	t.Run("CrossEvaluation", func(t *testing.T) {
		// Test that they can evaluate each other (they use the same static implementation)
		// This tests the compatibility
		configDigest := [32]byte{1: 7, 13: 11, 31: 23}
		seqNr := uint64(3)
		reportWithInfo := ocr3types.ReportWithInfo[[]byte]{
			Report: []byte{41: 131},
			Info:   []byte("some-info"),
		}
		sigs := []libocr.AttributedOnchainSignature{{Signature: []byte{9: 8, 7: 6}, Signer: commontypes.OracleID(54)}}

		// Both transmitters should accept the same valid data
		err1 := transmitter1.Transmit(ctx, configDigest, seqNr, reportWithInfo, sigs)
		assert.NoError(t, err1)

		err2 := transmitter2.Transmit(ctx, configDigest, seqNr, reportWithInfo, sigs)
		assert.NoError(t, err2)
	})
}
