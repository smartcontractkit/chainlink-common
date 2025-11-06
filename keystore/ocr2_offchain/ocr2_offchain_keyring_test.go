package ocr2_offchain_test

import (
	"testing"

	commonks "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/ocr2_offchain"
	"github.com/stretchr/testify/require"
)

func TestOCR2OffchainKeyring(t *testing.T) {
	storage := commonks.NewMemoryStorage()
	ctx := t.Context()
	ks, err := commonks.LoadKeystore(ctx, storage, "test-password")
	require.NoError(t, err)
	keyring, err := ocr2_offchain.CreateOCR2OffchainKeyring(ctx, ks, "test-ocr2-offchain-keyring")
	require.NoError(t, err)
	require.NotNil(t, keyring)

	msg := []byte("test-message")
	signature, err := keyring.OffchainSign(msg)
	require.NoError(t, err)
	require.NotNil(t, signature)

	keyrings, err := ocr2_offchain.GetOCR2OffchainKeyrings(ctx, ks, []string{"test-ocr2-offchain-keyring"})
	require.NoError(t, err)
	require.Equal(t, 1, len(keyrings))
	require.Equal(t, keyring.OffchainPublicKey(), keyrings[0].OffchainPublicKey())
	require.Equal(t, keyring.ConfigEncryptionPublicKey(), keyrings[0].ConfigEncryptionPublicKey())

	// List all works
	allKeyrings, err := ocr2_offchain.GetOCR2OffchainKeyrings(ctx, ks, []string{})
	require.NoError(t, err)
	require.Equal(t, 1, len(allKeyrings))

	// List non-existent errors.
	nonExistentKeyrings, err := ocr2_offchain.GetOCR2OffchainKeyrings(ctx, ks, []string{"non-existent-ocr2-offchain-keyring"})
	require.Error(t, err)
	require.Nil(t, nonExistentKeyrings)

	// Can create multiple.
	keyring2, err := ocr2_offchain.CreateOCR2OffchainKeyring(ctx, ks, "test-ocr2-offchain-keyring-2")
	require.NoError(t, err)
	require.NotNil(t, keyring2)
	msg2 := []byte("test-message-2")
	signature2, err := keyring2.OffchainSign(msg2)
	require.NoError(t, err)
	require.NotNil(t, signature2)

	// List by name works.
	keyrings2, err := ocr2_offchain.GetOCR2OffchainKeyrings(ctx, ks, []string{"test-ocr2-offchain-keyring-2"})
	require.NoError(t, err)
	require.Equal(t, 1, len(keyrings2))
	require.Equal(t, keyring2.OffchainPublicKey(), keyrings2[0].OffchainPublicKey())
	require.Equal(t, keyring2.ConfigEncryptionPublicKey(), keyrings2[0].ConfigEncryptionPublicKey())

	// List all works with multiple.
	allKeyrings2, err := ocr2_offchain.GetOCR2OffchainKeyrings(ctx, ks, []string{})
	require.NoError(t, err)
	require.Equal(t, 2, len(allKeyrings2))

	sig, err := allKeyrings[0].OffchainSign([]byte("test-message"))
	require.NoError(t, err)
	require.NotNil(t, sig)

	pubkey := allKeyrings[0].OffchainPublicKey()
	valid, err := ks.Verify(ctx, commonks.VerifyRequest{
		KeyType:   commonks.Ed25519,
		PublicKey: pubkey[:],
		Data:      []byte("test-message"),
		Signature: sig,
	})
	require.NoError(t, err)
	require.True(t, valid.Valid)
}
