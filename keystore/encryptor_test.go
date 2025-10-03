package keystore_test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestEncryptor_X25519(t *testing.T) {
	ctx := context.Background()
	ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)

	keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "A", KeyType: keystore.X25519},
			{KeyName: "B", KeyType: keystore.X25519},
		},
	})
	require.NoError(t, err)
	encryptResp, err := ks.Encrypt(ctx, keystore.EncryptRequest{
		KeyName: "A",
		// Encrypt to B
		RemotePubKey: keys.Keys[1].KeyInfo.PublicKey,
		Data:         []byte("hello world"),
	})
	require.NoError(t, err)
	decryptResp, err := ks.Decrypt(ctx, keystore.DecryptRequest{
		KeyName:       "B",
		EncryptedData: encryptResp.EncryptedData,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("hello world"), decryptResp.Data)
}

func TestEncryptor_EcdhP256(t *testing.T) {
	ctx := context.Background()
	ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)
	keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "A", KeyType: keystore.EcdhP256},
			{KeyName: "B", KeyType: keystore.EcdhP256},
		},
	})
	require.NoError(t, err)
	encryptResp, err := ks.Encrypt(ctx, keystore.EncryptRequest{
		KeyName:      "A",
		RemotePubKey: keys.Keys[1].KeyInfo.PublicKey,
		Data:         []byte("hello world"),
	})
	require.NoError(t, err)
	decryptResp, err := ks.Decrypt(ctx, keystore.DecryptRequest{
		KeyName:       "B",
		EncryptedData: encryptResp.EncryptedData,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("hello world"), decryptResp.Data)
}
