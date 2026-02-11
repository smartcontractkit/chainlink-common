package ocr2key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/types"
)

func TestExport(t *testing.T) {
	var tt = []struct {
		chain types.ChainType
	}{
		{chain: types.EVM},
		{chain: types.Cosmos},
		{chain: types.Solana},
		{chain: types.StarkNet},
		{chain: types.Aptos},
		{chain: types.Tron},
	}
	for _, tc := range tt {
		t.Run(string(tc.chain), func(t *testing.T) {
			kb, err := New(tc.chain)
			require.NoError(t, err)
			ej, err := ToEncryptedJSON(kb, "blah", scrypt.FastScryptParams)
			require.NoError(t, err)
			kbAfter, err := FromEncryptedJSON(ej, "blah")
			require.NoError(t, err)
			assert.Equal(t, kbAfter.ID(), kb.ID())
			assert.Equal(t, kbAfter.PublicKey(), kb.PublicKey())
			assert.Equal(t, kbAfter.OffchainPublicKey(), kb.OffchainPublicKey())
			assert.Equal(t, kbAfter.MaxSignatureLength(), kb.MaxSignatureLength())
			assert.Equal(t, kbAfter.Raw(), kb.Raw())
			assert.Equal(t, kbAfter.ConfigEncryptionPublicKey(), kb.ConfigEncryptionPublicKey())
			assert.Equal(t, kbAfter.ChainType(), kb.ChainType())
		})
	}
}
