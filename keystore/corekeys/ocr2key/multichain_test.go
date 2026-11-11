package ocr2key

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestUnmarshalMultichainPublicKey(t *testing.T) {
	var encoded []byte
	for _, entry := range []struct {
		family corekeys.ChainType
		key    []byte
	}{{corekeys.EVM, []byte{1, 2}}, {corekeys.Aptos, []byte{3}}} {
		typ, err := entry.family.Type()
		require.NoError(t, err)
		encoded = append(encoded, typ)
		encoded = binary.LittleEndian.AppendUint16(encoded, uint16(len(entry.key))) //nolint:gosec // test keys are small
		encoded = append(encoded, entry.key...)
	}

	keys, err := UnmarshalMultichainPublicKey(encoded)
	require.NoError(t, err)
	assert.Equal(t, map[string]ocrtypes.OnchainPublicKey{"evm": {1, 2}, "aptos": {3}}, keys)

	_, err = UnmarshalMultichainPublicKey(encoded[:len(encoded)-1])
	require.Error(t, err)
}
