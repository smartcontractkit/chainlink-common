package ocr2key

import (
	"testing"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonkeystore "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestExport(t *testing.T) {
	var tt = []struct {
		chain corekeys.ChainType
	}{
		{chain: corekeys.EVM},
		{chain: corekeys.Cosmos},
		{chain: corekeys.Solana},
		{chain: corekeys.StarkNet},
		{chain: corekeys.Aptos},
		{chain: corekeys.Tron},
		{chain: corekeys.TON},
		{chain: corekeys.Sui},
		{chain: corekeys.Stellar},
	}
	for _, tc := range tt {
		t.Run(string(tc.chain), func(t *testing.T) {
			kb, err := New(tc.chain)
			require.NoError(t, err)
			ej, err := ToEncryptedJSON(kb, "blah", commonkeystore.FastScryptParams)
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

			signature, err := kb.Sign(ocrtypes.ReportContext{}, make(ocrtypes.Report, 32))
			require.NoError(t, err)
			restoredSignature, err := kbAfter.Sign(ocrtypes.ReportContext{}, make(ocrtypes.Report, 32))
			require.NoError(t, err)
			if tc.chain == corekeys.StarkNet {
				assert.True(t, kb.Verify(kb.PublicKey(), ocrtypes.ReportContext{}, make(ocrtypes.Report, 32), signature))
				assert.True(t, kbAfter.Verify(kbAfter.PublicKey(), ocrtypes.ReportContext{}, make(ocrtypes.Report, 32), restoredSignature))
			} else {
				assert.Equal(t, signature, restoredSignature)
			}
		})
	}
}
