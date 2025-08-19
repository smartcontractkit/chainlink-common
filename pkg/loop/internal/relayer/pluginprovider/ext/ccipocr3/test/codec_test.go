package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

func TestCodec(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := Codec(lggr)

	t.Run("ChainSpecificAddressCodec", func(t *testing.T) {
		// Test that the embedded ChainSpecificAddressCodec works
		result, err := codec.ChainSpecificAddressCodec.AddressBytesToString([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, "test-address", result)

		bytes, err := codec.ChainSpecificAddressCodec.AddressStringToBytes("test")
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-address"), bytes)

		oracleBytes, err := codec.ChainSpecificAddressCodec.OracleIDAsAddressBytes(42)
		assert.NoError(t, err)
		assert.Equal(t, []byte{42}, oracleBytes)

		transmitter, err := codec.ChainSpecificAddressCodec.TransmitterBytesToString([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, "test-transmitter", transmitter)
	})

	t.Run("CommitPluginCodec", func(t *testing.T) {
		// Test that the embedded CommitPluginCodec works
		report := ccipocr3.CommitPluginReport{}
		encoded, err := codec.CommitPluginCodec.Encode(ctx, report)
		assert.NoError(t, err)
		assert.Equal(t, []byte("encoded-commit-report"), encoded)

		decoded, err := codec.CommitPluginCodec.Decode(ctx, encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
	})

	t.Run("ExecutePluginCodec", func(t *testing.T) {
		// Test that the embedded ExecutePluginCodec works
		report := ccipocr3.ExecutePluginReport{}
		encoded, err := codec.ExecutePluginCodec.Encode(ctx, report)
		assert.NoError(t, err)
		assert.Equal(t, []byte("encoded-execute-report"), encoded)

		decoded, err := codec.ExecutePluginCodec.Decode(ctx, encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
	})

	t.Run("TokenDataEncoder", func(t *testing.T) {
		// Test that the embedded TokenDataEncoder works
		message := ccipocr3.Bytes("test-message")
		attestation := ccipocr3.Bytes("test-attestation")

		encoded, err := codec.TokenDataEncoder.EncodeUSDC(ctx, message, attestation)
		assert.NoError(t, err)
		assert.Equal(t, ccipocr3.Bytes("encoded-usdc"), encoded)
	})

	t.Run("SourceChainExtraDataCodec", func(t *testing.T) {
		// Test that the embedded SourceChainExtraDataCodec works
		extraArgs := ccipocr3.Bytes("test-extra-args")
		result, err := codec.SourceChainExtraDataCodec.DecodeExtraArgsToMap(extraArgs)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "gasLimit")

		destExecData := ccipocr3.Bytes("test-dest-exec-data")
		result2, err := codec.SourceChainExtraDataCodec.DecodeDestExecDataToMap(destExecData)
		assert.NoError(t, err)
		assert.NotNil(t, result2)
		assert.Contains(t, result2, "data")
	})
}

func TestCodecEvaluator(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	evaluator := CodecEvaluator(lggr)
	codec1 := Codec(lggr)
	codec2 := Codec(lggr)

	t.Run("Evaluate", func(t *testing.T) {
		err := evaluator.Evaluate(ctx, codec1)
		assert.NoError(t, err)

		err = evaluator.Evaluate(ctx, codec2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		evaluator.AssertEqual(ctx, t, codec1)
		evaluator.AssertEqual(ctx, t, codec2)
	})
}

func TestCodecIntegration(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := Codec(lggr)

	t.Run("FullWorkflow", func(t *testing.T) {
		// Test a full encode/decode workflow for commit reports
		originalCommitReport := ccipocr3.CommitPluginReport{
			PriceUpdates: ccipocr3.PriceUpdates{
				TokenPriceUpdates: []ccipocr3.TokenPrice{},
				GasPriceUpdates:   []ccipocr3.GasPriceChain{},
			},
			BlessedMerkleRoots:   []ccipocr3.MerkleRootChain{},
			UnblessedMerkleRoots: []ccipocr3.MerkleRootChain{},
			RMNSignatures:        []ccipocr3.RMNECDSASignature{},
		}

		encoded, err := codec.CommitPluginCodec.Encode(ctx, originalCommitReport)
		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		decoded, err := codec.CommitPluginCodec.Decode(ctx, encoded)
		require.NoError(t, err)
		assert.NotNil(t, decoded)

		// Test a full encode/decode workflow for execute reports
		originalExecuteReport := ccipocr3.ExecutePluginReport{
			ChainReports: []ccipocr3.ExecutePluginReportSingleChain{},
		}

		encodedExecute, err := codec.ExecutePluginCodec.Encode(ctx, originalExecuteReport)
		require.NoError(t, err)
		assert.NotEmpty(t, encodedExecute)

		decodedExecute, err := codec.ExecutePluginCodec.Decode(ctx, encodedExecute)
		require.NoError(t, err)
		assert.NotNil(t, decodedExecute)
	})

	t.Run("AddressCodecWorkflow", func(t *testing.T) {
		// Test address conversion workflow
		testAddress := "0x1234567890abcdef"

		addressBytes, err := codec.ChainSpecificAddressCodec.AddressStringToBytes(testAddress)
		require.NoError(t, err)
		assert.NotEmpty(t, addressBytes)

		addressString, err := codec.ChainSpecificAddressCodec.AddressBytesToString(addressBytes)
		require.NoError(t, err)
		assert.NotEmpty(t, addressString)

		// Test oracle ID to address conversion
		oracleID := uint8(42)
		oracleAddress, err := codec.ChainSpecificAddressCodec.OracleIDAsAddressBytes(oracleID)
		require.NoError(t, err)
		assert.Equal(t, []byte{42}, oracleAddress)
	})

	t.Run("TokenDataWorkflow", func(t *testing.T) {
		// Test USDC encoding workflow
		message := ccipocr3.Bytes("test-usdc-message")
		attestation := ccipocr3.Bytes("test-usdc-attestation")

		encodedUSDC, err := codec.TokenDataEncoder.EncodeUSDC(ctx, message, attestation)
		require.NoError(t, err)
		assert.NotEmpty(t, encodedUSDC)
		assert.Equal(t, ccipocr3.Bytes("encoded-usdc"), encodedUSDC)
	})

	t.Run("ExtraDataWorkflow", func(t *testing.T) {
		// Test extra data decoding workflow
		extraArgs := ccipocr3.Bytes("{\"gasLimit\": 100000}")
		argsMap, err := codec.SourceChainExtraDataCodec.DecodeExtraArgsToMap(extraArgs)
		require.NoError(t, err)
		assert.NotEmpty(t, argsMap)

		destExecData := ccipocr3.Bytes("{\"data\": \"test\"}")
		execMap, err := codec.SourceChainExtraDataCodec.DecodeDestExecDataToMap(destExecData)
		require.NoError(t, err)
		assert.NotEmpty(t, execMap)
	})
}
