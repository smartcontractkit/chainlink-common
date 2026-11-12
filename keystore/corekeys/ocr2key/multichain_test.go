package ocr2key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestMultichainPublicKey(t *testing.T) {
	want := MultichainPublicKey{"evm": {1, 2}, "aptos": {3}}

	encoded, err := want.MarshalBinary()
	require.NoError(t, err)

	var got MultichainPublicKey
	require.NoError(t, got.UnmarshalBinary(encoded))
	assert.Equal(t, want, got)

	again, err := got.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, encoded, again, "the encoding is deterministic")

	require.Error(t, got.UnmarshalBinary(encoded[:len(encoded)-1]))

	unknown, err := MultichainPublicKey{"evm": {1}, "not-a-chain": {2}}.MarshalBinary()
	require.NoError(t, err)
	require.NoError(t, got.UnmarshalBinary(unknown))
	assert.Equal(t, MultichainPublicKey{"evm": ocrtypes.OnchainPublicKey{1}}, got, "unknown chain types are skipped")
}
