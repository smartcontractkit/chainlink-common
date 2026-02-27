package keystore_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {

	ctx := t.Context()
	th := NewKeystoreTH(t)
	th.CreateTestKeys(t)

	type testCase struct {
		name                 string
		remoteKeyType        keystore.KeyType
		remotePubKey         []byte
		decryptKey           string
		payload              []byte
		expectedEncryptError error
		expectedDecryptError error
	}

	var tt = []testCase{
		{
			name:                 "Non-existent encrypt key",
			remoteKeyType:        "blah",
			remotePubKey:         th.KeysByType()[keystore.X25519][0].PublicKey,
			decryptKey:           th.KeyName(keystore.X25519, 0),
			payload:              []byte("hello world"),
			expectedEncryptError: keystore.ErrEncryptionFailed,
		},
		{
			name:          "Empty payload x25519",
			remoteKeyType: keystore.X25519,
			remotePubKey:  th.KeysByType()[keystore.X25519][0].PublicKey,
			decryptKey:    th.KeyName(keystore.X25519, 0),
			payload:       []byte{},
		},
		{
			name:          "Empty payload ecdh p256",
			remoteKeyType: keystore.ECDH_P256,
			remotePubKey:  th.KeysByType()[keystore.ECDH_P256][0].PublicKey,
			decryptKey:    th.KeyName(keystore.ECDH_P256, 0),
			payload:       []byte{},
		},
		{
			name:                 "Non-existent decrypt key",
			remoteKeyType:        keystore.X25519,
			remotePubKey:         th.KeysByType()[keystore.X25519][0].PublicKey,
			decryptKey:           "blah",
			payload:              []byte("hello world"),
			expectedDecryptError: keystore.ErrDecryptionFailed,
		},
		{
			name:          "Max payload",
			remoteKeyType: keystore.X25519,
			remotePubKey:  th.KeysByType()[keystore.X25519][0].PublicKey,
			decryptKey:    th.KeyName(keystore.X25519, 0),
			payload:       make([]byte, keystore.MaxEncryptionPayloadSize),
		},
		{
			name:                 "Payload too large",
			remoteKeyType:        keystore.X25519,
			remotePubKey:         th.KeysByType()[keystore.X25519][0].PublicKey,
			decryptKey:           th.KeyName(keystore.X25519, 0),
			payload:              make([]byte, keystore.MaxEncryptionPayloadSize+1),
			expectedEncryptError: keystore.ErrEncryptionFailed,
		}}

	for encName, encKey := range th.KeysByName() {
		testName := fmt.Sprintf("Encrypt to %s", encName)
		var expectedEncryptError error
		if encKey.KeyType.IsEncryptionKeyType() {
			// Same key types should succeed
			expectedEncryptError = nil
		} else {
			// Different key types or non-encryption key types should fail
			expectedEncryptError = keystore.ErrEncryptionFailed
		}

		tt = append(tt, testCase{
			name:                 testName,
			remoteKeyType:        encKey.KeyType,
			remotePubKey:         encKey.PublicKey,
			decryptKey:           encName,
			expectedEncryptError: expectedEncryptError,
			payload:              []byte("hello world"),
		})
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			encryptResp, err := th.Keystore.Encrypt(ctx, keystore.EncryptRequest{
				RemoteKeyType: tt.remoteKeyType,
				RemotePubKey:  tt.remotePubKey,
				Data:          tt.payload,
			})
			if tt.expectedEncryptError != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedEncryptError))
				return
			}
			require.NoError(t, err)
			decryptResp, err := th.Keystore.Decrypt(ctx, keystore.DecryptRequest{
				KeyName:       tt.decryptKey,
				EncryptedData: encryptResp.EncryptedData,
			})
			if tt.expectedDecryptError != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedDecryptError))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.payload, decryptResp.Data)
		})
	}
}

func TestEncryptDecrypt_SharedSecret(t *testing.T) {
	ctx := t.Context()
	th := NewKeystoreTH(t)
	th.CreateTestKeys(t)

	type testCase struct {
		name          string
		keyName       string
		keyType       keystore.KeyType
		expectedError error
	}
	var tt = []testCase{
		{
			name:          "Non-existent key",
			keyName:       "blah",
			keyType:       keystore.X25519,
			expectedError: keystore.ErrSharedSecretFailed,
		},
	}

	for keyType := range th.KeysByType() {
		var expectedError error
		if !keyType.IsEncryptionKeyType() {
			expectedError = keystore.ErrSharedSecretFailed
		}
		tt = append(tt, testCase{
			keyName:       th.KeyName(keyType, 0),
			name:          fmt.Sprintf("keyType_%s", keyType),
			keyType:       keyType,
			expectedError: expectedError,
		})
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			_, err := th.Keystore.DeriveSharedSecret(ctx, keystore.DeriveSharedSecretRequest{
				KeyName:      tt.keyName,
				RemotePubKey: th.KeysByType()[tt.keyType][0].PublicKey,
			})
			if tt.expectedError != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedError))
				return
			}
			require.NoError(t, err)
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

		ctx := t.Context()
		th := NewKeystoreTH(t)
		th.CreateTestKeys(t)
		// Test each encryption key type
		for _, keyType := range keystore.AllEncryptionKeyTypes {
			t.Run(fmt.Sprintf("keyType_%s", keyType), func(t *testing.T) {
				// Encrypt data using sender key to receiver's public key
				encryptResp, err := th.Keystore.Encrypt(ctx, keystore.EncryptRequest{
					RemoteKeyType: keyType,
					RemotePubKey:  th.KeysByType()[keyType][1].PublicKey, // receiver's public key
					Data:          data,
				})
				require.NoError(t, err, "Encryption should succeed for keyType %s with data length %d", keyType, len(data))

				// Decrypt using receiver key
				decryptResp, err := th.Keystore.Decrypt(ctx, keystore.DecryptRequest{
					KeyName:       th.KeyName(keyType, 1),
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
