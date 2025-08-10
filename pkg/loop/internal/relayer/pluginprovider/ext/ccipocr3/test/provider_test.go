package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

func TestCCIPProvider(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	// Create the static CCIP provider
	provider := CCIPProvider(lggr)

	// Test that all components are present
	require.NotNil(t, provider.ChainAccessor())
	require.NotNil(t, provider.ContractTransmitter())
	require.NotNil(t, provider.Codec())

	// Test ChainAccessor methods
	chainAccessor := provider.ChainAccessor()
	contractAddr, err := chainAccessor.GetContractAddress("test-contract")
	assert.NoError(t, err)
	assert.NotNil(t, contractAddr)

	// Test that GetAllConfigsLegacy works
	snapshot, configs, err := chainAccessor.GetAllConfigsLegacy(ctx, 1, []ccipocr3.ChainSelector{2, 3})
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.NotNil(t, configs)

	// Test ChainFeeComponents
	feeComponents, err := chainAccessor.GetChainFeeComponents(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, feeComponents)

	// Test Codec components
	codec := provider.Codec()

	// Test ChainSpecificAddressCodec
	addrStr, err := codec.ChainSpecificAddressCodec.AddressBytesToString([]byte("test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, addrStr)

	addrBytes, err := codec.ChainSpecificAddressCodec.AddressStringToBytes("test")
	assert.NoError(t, err)
	assert.NotNil(t, addrBytes)

	// Test CommitPluginCodec
	encodedCommit, err := codec.CommitPluginCodec.Encode(ctx, ccipocr3.CommitPluginReport{})
	assert.NoError(t, err)
	assert.NotNil(t, encodedCommit)

	_, err = codec.CommitPluginCodec.Decode(ctx, encodedCommit)
	assert.NoError(t, err)

	// Test ExecutePluginCodec
	encodedExecute, err := codec.ExecutePluginCodec.Encode(ctx, ccipocr3.ExecutePluginReport{})
	assert.NoError(t, err)
	assert.NotNil(t, encodedExecute)

	_, err = codec.ExecutePluginCodec.Decode(ctx, encodedExecute)
	assert.NoError(t, err)

	// Test TokenDataEncoder
	encodedUSDC, err := codec.TokenDataEncoder.EncodeUSDC(ctx, []byte("message"), []byte("attestation"))
	assert.NoError(t, err)
	assert.NotNil(t, encodedUSDC)

	// Test SourceChainExtraDataCodec
	extraArgs, err := codec.SourceChainExtraDataCodec.DecodeExtraArgsToMap([]byte("test-extra-args"))
	assert.NoError(t, err)
	assert.NotNil(t, extraArgs)

	destExecData, err := codec.SourceChainExtraDataCodec.DecodeDestExecDataToMap([]byte("test-dest-exec-data"))
	assert.NoError(t, err)
	assert.NotNil(t, destExecData)
}

func TestCCIPProviderEvaluate(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	// Create two providers
	provider1 := CCIPProvider(lggr)
	provider2 := CCIPProvider(lggr)

	// Test evaluation
	err := provider1.Evaluate(ctx, provider2)
	assert.NoError(t, err)
}

func TestCCIPProviderAssertEqual(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	// Create two providers
	provider1 := CCIPProvider(lggr)
	provider2 := CCIPProvider(lggr)

	// Test assertion
	provider1.AssertEqual(ctx, t, provider2)
}
