package keystore_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
)

func TestCoreKeystore(t *testing.T) {
	ctx := t.Context()

	ks, err := keystore.LoadKeystore(t.Context(), keystore.NewMemoryStorage(), "test-password", keystore.WithScryptParams(scrypt.FastScryptParams))
	require.NoError(t, err)
	coreKs := keystore.NewCoreKeystore(ks)

	keysResp, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "encrypt", KeyType: keystore.X25519},
			{KeyName: "sign", KeyType: keystore.ECDSA_S256},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(keysResp.Keys))

	accounts, err := coreKs.Accounts(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"encrypt", "sign"}, accounts)

	signature, err := coreKs.Sign(ctx, "sign", crypto.Keccak256([]byte("test-data-to-sign")))
	require.NoError(t, err)
	require.NotEmpty(t, signature)
	verifyResp, err := ks.Verify(ctx, keystore.VerifyRequest{
		KeyType:   keysResp.Keys[1].KeyInfo.KeyType,
		PublicKey: keysResp.Keys[1].KeyInfo.PublicKey,
		Data:      crypto.Keccak256([]byte("test-data-to-sign")),
		Signature: signature,
	})
	require.NoError(t, err)
	require.True(t, verifyResp.Valid)

	encryptedData, err := ks.Encrypt(ctx, keystore.EncryptRequest{
		RemoteKeyType: keysResp.Keys[0].KeyInfo.KeyType,
		RemotePubKey:  keysResp.Keys[0].KeyInfo.PublicKey,
		Data:          []byte("test-data-to-encrypt"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, encryptedData)

	decryptedData, err := coreKs.Decrypt(ctx, "encrypt", encryptedData.EncryptedData)
	require.NoError(t, err)
	require.Equal(t, []byte("test-data-to-encrypt"), decryptedData)
}
