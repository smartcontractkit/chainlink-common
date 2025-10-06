package keystore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	ctx := context.Background()
	ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)

	// Create 2 keys of each key type.
	testKeysByType := make(map[string]struct {
		keyType   keystore.KeyType
		publicKey []byte
	})
	keyName := func(keyType keystore.KeyType, index int) string {
		return fmt.Sprintf("key-%s-%d", keyType, index)
	}
	for _, keyType := range keystore.AllEncryptionKeyTypes {
		keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{KeyName: keyName(keyType, 0), KeyType: keyType},
				{KeyName: keyName(keyType, 1), KeyType: keyType},
			},
		})
		require.NoError(t, err)
		testKeysByType[keys.Keys[0].KeyInfo.Name] = struct {
			keyType   keystore.KeyType
			publicKey []byte
		}{keyType: keys.Keys[0].KeyInfo.KeyType, publicKey: keys.Keys[0].KeyInfo.PublicKey}
		testKeysByType[keys.Keys[1].KeyInfo.Name] = struct {
			keyType   keystore.KeyType
			publicKey []byte
		}{keyType: keys.Keys[1].KeyInfo.KeyType, publicKey: keys.Keys[1].KeyInfo.PublicKey}
	}

	var tt = []struct {
		name          string
		fromKey       string
		toKey         string
		expectedError error
	}{
		{name: "Encrypt to self x25519", fromKey: keyName(keystore.X25519, 0), toKey: keyName(keystore.X25519, 0), expectedError: nil},
		{name: "Encrypt to other x25519", fromKey: keyName(keystore.X25519, 0), toKey: keyName(keystore.X25519, 1), expectedError: nil},
		{name: "Encrypt to self ecdh-p256", fromKey: keyName(keystore.EcdhP256, 0), toKey: keyName(keystore.EcdhP256, 0), expectedError: nil},
		{name: "Encrypt to other ecdh-p256", fromKey: keyName(keystore.EcdhP256, 0), toKey: keyName(keystore.EcdhP256, 1), expectedError: nil},
		{name: "Encrypt x25519 to ecdh-p256 should fail", fromKey: keyName(keystore.X25519, 0), toKey: keyName(keystore.EcdhP256, 0), expectedError: keystore.ErrEncryptionFailed},
		{name: "Encrypt ecdh-p256 to x25519 should fail", fromKey: keyName(keystore.EcdhP256, 0), toKey: keyName(keystore.X25519, 0), expectedError: keystore.ErrEncryptionFailed},
	}
	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			encryptResp, err := ks.Encrypt(ctx, keystore.EncryptRequest{
				KeyName:      tt.fromKey,
				RemotePubKey: testKeysByType[tt.toKey].publicKey,
				Data:         []byte("hello world"),
			})
			if tt.expectedError != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedError))
				return
			}
			require.NoError(t, err)
			decryptResp, err := ks.Decrypt(ctx, keystore.DecryptRequest{
				KeyName:       tt.toKey,
				EncryptedData: encryptResp.EncryptedData,
			})
			require.NoError(t, err)
			require.Equal(t, []byte("hello world"), decryptResp.Data)
		})
	}
}
