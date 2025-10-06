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

func TestEncryptDecrypt_SharedSecret(t *testing.T) {
	ctx := context.Background()
	ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)

	for _, keyType := range keystore.AllEncryptionKeyTypes {
		t.Run(fmt.Sprintf("keyType_%s", keyType), func(t *testing.T) {
			keyName := fmt.Sprintf("test-key-%s", keyType)
			keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
				Keys: []keystore.CreateKeyRequest{
					{KeyName: keyName, KeyType: keyType},
				},
			})
			require.NoError(t, err)
			_, err = ks.DeriveSharedSecret(ctx, keystore.DeriveSharedSecretRequest{
				LocalKeyName: keyName,
				RemotePubKey: keys.Keys[0].KeyInfo.PublicKey,
			})
			require.NoError(t, err)
		})
	}
}

func TestEncryptDecrypt_PayloadSizeLimit(t *testing.T) {
	ctx := context.Background()
	ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)

	for _, keyType := range keystore.AllEncryptionKeyTypes {
		t.Run(fmt.Sprintf("keyType_%s", keyType), func(t *testing.T) {
			keyName := fmt.Sprintf("test-key-%s", keyType)
			keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
				Keys: []keystore.CreateKeyRequest{
					{KeyName: keyName, KeyType: keyType},
				},
			})
			require.NoError(t, err)
			// Test encrypting at the limit
			maxPayload := make([]byte, keystore.MaxEncryptionPayloadSize)
			maxEncryptResp, err := ks.Encrypt(ctx, keystore.EncryptRequest{
				KeyName:      keyName,
				RemotePubKey: keys.Keys[0].KeyInfo.PublicKey,
				Data:         maxPayload,
			})
			require.NoError(t, err)

			// Test decrypting at max (confirm overhead sufficient)
			maxDecryptResp, err := ks.Decrypt(ctx, keystore.DecryptRequest{
				KeyName:       keyName,
				EncryptedData: maxEncryptResp.EncryptedData,
			})
			require.NoError(t, err)
			require.Equal(t, len(maxDecryptResp.Data), len(maxPayload))

			// Test encrypting above the limit
			_, err = ks.Encrypt(ctx, keystore.EncryptRequest{
				KeyName:      keyName,
				RemotePubKey: keys.Keys[0].KeyInfo.PublicKey,
				Data:         make([]byte, keystore.MaxEncryptionPayloadSize+1),
			})
			require.Error(t, err)

			// Test decrypting above the limit
			_, err = ks.Decrypt(ctx, keystore.DecryptRequest{
				KeyName:       keyName,
				EncryptedData: make([]byte, keystore.MaxEncryptionPayloadSize+1025),
			})
			require.Error(t, err)
		})
	}
}

func FuzzEncryptDecryptRoundtrip(f *testing.F) {
	// Add seed corpus with various input sizes and patterns
	seedCorpus := [][]byte{
		{0x00},                // Single null byte
		{0xFF},                // Single 0xFF byte
		{0x00, 0xFF},          // Two bytes
		[]byte("hello"),       // Short string
		[]byte("hello world"), // Medium string
		[]byte("The quick brown fox jumps over the lazy dog"), // Longer string
		make([]byte, 100),       // 100 null bytes
		make([]byte, 1000),      // 1000 null bytes
		make([]byte, 10000),     // 10KB of null bytes
		make([]byte, 100000),    // 100KB of null bytes
		make([]byte, 1024*1024), // Exactly 1MB (at limit)
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > keystore.MaxEncryptionPayloadSize || len(data) == 0 {
			t.Skip("Invalid data size for fuzz test")
		}

		ctx := context.Background()
		ks, err := keystore.LoadKeystore(ctx, storage.NewMemoryStorage(), keystore.EncryptionParams{
			Password:     "test-password",
			ScryptParams: keystore.FastScryptParams,
		})
		require.NoError(t, err)

		// Test each encryption key type
		for i, keyType := range keystore.AllEncryptionKeyTypes {
			t.Run(fmt.Sprintf("keyType_%s", keyType), func(t *testing.T) {
				// Create two keys of the same type for encryption/decryption
				senderName := fmt.Sprintf("sender-%d", i)
				receiverName := fmt.Sprintf("receiver-%d", i)
				keys, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
					Keys: []keystore.CreateKeyRequest{
						{KeyName: senderName, KeyType: keyType},
						{KeyName: receiverName, KeyType: keyType},
					},
				})
				require.NoError(t, err)

				// Encrypt data using sender key to receiver's public key
				encryptResp, err := ks.Encrypt(ctx, keystore.EncryptRequest{
					KeyName:      senderName,
					RemotePubKey: keys.Keys[1].KeyInfo.PublicKey, // receiver's public key
					Data:         data,
				})
				require.NoError(t, err, "Encryption should succeed for keyType %s with data length %d", keyType, len(data))

				// Decrypt using receiver key
				decryptResp, err := ks.Decrypt(ctx, keystore.DecryptRequest{
					KeyName:       receiverName,
					EncryptedData: encryptResp.EncryptedData,
				})
				require.NoError(t, err, "Decryption should succeed for keyType %s with data length %d", keyType, len(data))

				// Verify roundtrip integrity
				require.Equal(t, data, decryptResp.Data,
					"Roundtrip failed for keyType %s: original data length %d, decrypted data length %d",
					keyType, len(data), len(decryptResp.Data))
			})
		}
	})
}
