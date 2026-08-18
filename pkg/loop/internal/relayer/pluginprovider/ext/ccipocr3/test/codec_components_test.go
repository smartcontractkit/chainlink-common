package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

func TestChainSpecificAddressCodec(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := ChainSpecificAddressCodec(lggr)

	t.Run("AddressBytesToString", func(t *testing.T) {
		result, err := codec.AddressBytesToString([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, "test-address", result)
	})

	t.Run("AddressStringToBytes", func(t *testing.T) {
		result, err := codec.AddressStringToBytes("test")
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-address"), result)
	})

	t.Run("OracleIDAsAddressBytes", func(t *testing.T) {
		result, err := codec.OracleIDAsAddressBytes(42)
		assert.NoError(t, err)
		assert.Equal(t, []byte{42}, result)
	})

	t.Run("TransmitterBytesToString", func(t *testing.T) {
		result, err := codec.TransmitterBytesToString([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, "test-transmitter", result)
	})

	t.Run("Evaluate", func(t *testing.T) {
		codec2 := ChainSpecificAddressCodec(lggr)
		err := codec.Evaluate(ctx, codec2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		codec2 := ChainSpecificAddressCodec(lggr)
		codec.AssertEqual(ctx, t, codec2)
	})
}

func TestCommitPluginCodec(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := CommitPluginCodec(lggr)

	t.Run("Encode", func(t *testing.T) {
		report := ccipocr3.CommitPluginReport{}
		encoded, err := codec.Encode(ctx, report)
		assert.NoError(t, err)
		assert.Equal(t, []byte("encoded-commit-report"), encoded)
	})

	t.Run("Decode", func(t *testing.T) {
		data := []byte("test-data")
		report, err := codec.Decode(ctx, data)
		assert.NoError(t, err)
		assert.NotNil(t, report)
	})

	t.Run("Encode_Decode_Roundtrip", func(t *testing.T) {
		originalReport := ccipocr3.CommitPluginReport{}
		encoded, err := codec.Encode(ctx, originalReport)
		require.NoError(t, err)

		decoded, err := codec.Decode(ctx, encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
	})

	t.Run("Evaluate", func(t *testing.T) {
		codec2 := CommitPluginCodec(lggr)
		err := codec.Evaluate(ctx, codec2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		codec2 := CommitPluginCodec(lggr)
		codec.AssertEqual(ctx, t, codec2)
	})
}

func TestExecutePluginCodec(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := ExecutePluginCodec(lggr)

	t.Run("Encode", func(t *testing.T) {
		report := ccipocr3.ExecutePluginReport{}
		encoded, err := codec.Encode(ctx, report)
		assert.NoError(t, err)
		assert.Equal(t, []byte("encoded-execute-report"), encoded)
	})

	t.Run("Decode", func(t *testing.T) {
		data := []byte("test-data")
		report, err := codec.Decode(ctx, data)
		assert.NoError(t, err)
		assert.NotNil(t, report)
	})

	t.Run("Encode_Decode_Roundtrip", func(t *testing.T) {
		originalReport := ccipocr3.ExecutePluginReport{}
		encoded, err := codec.Encode(ctx, originalReport)
		require.NoError(t, err)

		decoded, err := codec.Decode(ctx, encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
	})

	t.Run("Evaluate", func(t *testing.T) {
		codec2 := ExecutePluginCodec(lggr)
		err := codec.Evaluate(ctx, codec2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		codec2 := ExecutePluginCodec(lggr)
		codec.AssertEqual(ctx, t, codec2)
	})
}

func TestTokenDataEncoder(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	encoder := TokenDataEncoder(lggr)

	t.Run("EncodeUSDC", func(t *testing.T) {
		message := ccipocr3.Bytes("test-message")
		attestation := ccipocr3.Bytes("test-attestation")

		encoded, err := encoder.EncodeUSDC(ctx, message, attestation)
		assert.NoError(t, err)
		assert.Equal(t, ccipocr3.Bytes("encoded-usdc"), encoded)
	})

	t.Run("Evaluate", func(t *testing.T) {
		encoder2 := TokenDataEncoder(lggr)
		err := encoder.Evaluate(ctx, encoder2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		encoder2 := TokenDataEncoder(lggr)
		encoder.AssertEqual(ctx, t, encoder2)
	})
}

func TestSourceChainExtraDataCodec(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	codec := SourceChainExtraDataCodec(lggr)

	t.Run("DecodeExtraArgsToMap", func(t *testing.T) {
		extraArgs := ccipocr3.Bytes("test-extra-args")
		result, err := codec.DecodeExtraArgsToMap(extraArgs)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "gasLimit")
		assert.Equal(t, uint64(100000), result["gasLimit"])
	})

	t.Run("DecodeDestExecDataToMap", func(t *testing.T) {
		destExecData := ccipocr3.Bytes("test-dest-exec-data")
		result, err := codec.DecodeDestExecDataToMap(destExecData)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result, "data")
		assert.Equal(t, "test-data", result["data"])
	})

	t.Run("Evaluate", func(t *testing.T) {
		codec2 := SourceChainExtraDataCodec(lggr)
		err := codec.Evaluate(ctx, codec2)
		assert.NoError(t, err)
	})

	t.Run("AssertEqual", func(t *testing.T) {
		codec2 := SourceChainExtraDataCodec(lggr)
		codec.AssertEqual(ctx, t, codec2)
	})
}
