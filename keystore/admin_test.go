package keystore_test

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
)

func TestKeystore_CreateDeleteReadKeys(t *testing.T) {
	ctx := t.Context()
	type key struct {
		name     string
		metadata []byte
		keyType  keystore.KeyType
	}
	type keyOp struct {
		key           key
		expectedError error
		op            string
	}
	var testKeysEd25519 = []key{
		{name: "test-key-X25519", keyType: keystore.X25519, metadata: []byte{}},
		{name: "test-key-X25519-2", keyType: keystore.X25519, metadata: []byte{}},
	}
	var tt = []struct {
		name         string
		keyOps       []keyOp
		expectedKeys []key
	}{
		{
			name: "Create key",
			keyOps: []keyOp{
				{key: testKeysEd25519[0], op: "create", expectedError: nil},
			},
			expectedKeys: []key{testKeysEd25519[0]},
		},
		{
			name: "Delete only key",
			keyOps: []keyOp{
				{key: testKeysEd25519[0], op: "create", expectedError: nil},
				{key: testKeysEd25519[0], op: "delete", expectedError: nil},
			},
			expectedKeys: []key{},
		},
		{
			name: "No duplicate names",
			keyOps: []keyOp{
				{key: testKeysEd25519[0], op: "create", expectedError: nil},
				{key: testKeysEd25519[0], op: "create", expectedError: keystore.ErrKeyAlreadyExists},
			},
			expectedKeys: []key{testKeysEd25519[0]},
		},
		{
			name: "Invalid key name",
			keyOps: []keyOp{
				{key: key{name: "", keyType: keystore.X25519, metadata: []byte{}}, op: "create", expectedError: keystore.ErrInvalidKeyName},
			},
			expectedKeys: []key{},
		},
		{
			name: "Delete non-existent key",
			keyOps: []keyOp{
				{key: testKeysEd25519[0], op: "delete", expectedError: keystore.ErrKeyNotFound},
			},
			expectedKeys: []key{},
		},
		{
			name: "Create key with unsupported type",
			keyOps: []keyOp{
				{key: key{name: "blah", keyType: "unsupported", metadata: []byte{}}, op: "create", expectedError: keystore.ErrUnsupportedKeyType},
			},
			expectedKeys: []key{},
		},
		{
			name: "Create multiple instances of same type",
			keyOps: []keyOp{
				{key: testKeysEd25519[0], op: "create", expectedError: nil},
				{key: testKeysEd25519[1], op: "create", expectedError: nil},
			},
			// Note key 0 is lexicographically less than key 1
			// So we assert in order to ensure deterministic ordering
			expectedKeys: []key{testKeysEd25519[0], testKeysEd25519[1]},
		},
		{
			name: "Create one of each type",
			keyOps: func() []keyOp {
				var keyOps []keyOp
				for _, keyType := range keystore.AllKeyTypes {
					keyOps = append(keyOps, keyOp{key: key{name: fmt.Sprintf("test-key-%s", keyType), keyType: keyType, metadata: []byte{}}, op: "create", expectedError: nil})
				}
				return keyOps
			}(),
			expectedKeys: func() []key {
				var expectedKeys []key
				sort.Slice(keystore.AllKeyTypes, func(i, j int) bool { return keystore.AllKeyTypes[i] < keystore.AllKeyTypes[j] })
				for _, keyType := range keystore.AllKeyTypes {
					expectedKeys = append(expectedKeys, key{name: fmt.Sprintf("test-key-%s", keyType), keyType: keyType, metadata: []byte{}})
				}
				return expectedKeys
			}(),
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			storage := keystore.NewMemoryStorage()
			ks, err := keystore.LoadKeystore(ctx, storage, "test-password", keystore.WithScryptParams(scrypt.FastScryptParams))
			require.NoError(t, err)
			for _, op := range tt.keyOps {
				switch op.op {
				case "create":
					_, err = ks.CreateKeys(ctx, keystore.CreateKeysRequest{
						Keys: []keystore.CreateKeyRequest{
							{KeyName: op.key.name, KeyType: op.key.keyType},
						},
					})
					if op.expectedError != nil {
						require.ErrorIs(t, err, op.expectedError)
						continue
					}
					require.NoError(t, err)
				case "delete":
					_, err = ks.DeleteKeys(ctx, keystore.DeleteKeysRequest{
						KeyNames: []string{op.key.name},
					})
					if op.expectedError != nil {
						require.ErrorIs(t, err, op.expectedError)
						continue
					}
					require.NoError(t, op.expectedError)
				}
			}
			resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
			require.NoError(t, err)
			var haveKeys []key
			for _, respKey := range resp.Keys {
				// No crypto without a public key yet so lets assert that its present.
				assert.NotEmpty(t, respKey.KeyInfo.PublicKey)
				assert.NotEmpty(t, respKey.KeyInfo.CreatedAt)
				haveKeys = append(haveKeys, key{name: respKey.KeyInfo.Name, keyType: respKey.KeyInfo.KeyType, metadata: respKey.KeyInfo.Metadata})
			}
			for i, expectedKey := range tt.expectedKeys {
				assert.Equal(t, expectedKey, haveKeys[i])
			}
		})
	}
}

// TestKeystore_ConcurrentCreateAndRead tests that the keystore can be used concurrently to create and read keys.
// go test -race -run TestKeystore_ConcurrentCreateAndRead to check for race conditions.
func TestKeystore_ConcurrentCreateAndRead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test", keystore.WithScryptParams(scrypt.FastScryptParams))
	require.NoError(t, err)

	const (
		numWriters     = 8
		keysPerWriter  = 25
		numReaders     = 6
		readsPerReader = 40
	)

	var wg sync.WaitGroup
	wg.Add(numWriters + numReaders)

	for w := 0; w < numWriters; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < keysPerWriter; i++ {
				name := fmt.Sprintf("k-%d-%d", w, i)
				_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
					Keys: []keystore.CreateKeyRequest{
						{KeyName: name, KeyType: keystore.Ed25519},
					},
				})
				require.NoError(t, err)
			}
		}()
	}

	for r := 0; r < numReaders; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < readsPerReader; i++ {
				_, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
				require.NoError(t, err)
			}
		}()
	}

	wg.Wait()
	resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
	require.NoError(t, err)
	require.Equal(t, numWriters*keysPerWriter, len(resp.Keys))
}

func TestKeystore_ExportImport(t *testing.T) {
	ks1, err := keystore.LoadKeystore(t.Context(), keystore.NewMemoryStorage(), "ks1")
	require.NoError(t, err)
	ks2, err := keystore.LoadKeystore(t.Context(), keystore.NewMemoryStorage(), "ks2")
	require.NoError(t, err)

	t.Run("export and import", func(t *testing.T) {
		exportParams := keystore.EncryptionParams{
			Password:     "export-pass",
			ScryptParams: scrypt.FastScryptParams,
		}
		_, err = ks1.CreateKeys(t.Context(), keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{KeyName: "key1", KeyType: keystore.Ed25519},
			},
		})
		require.NoError(t, err)
		exportResponse, err := ks1.ExportKeys(t.Context(), keystore.ExportKeysRequest{
			Keys: []keystore.ExportKeyParam{
				{KeyName: "key1", Enc: exportParams},
			},
		})
		require.NoError(t, err)
		require.Len(t, exportResponse.Keys, 1)
		_, err = ks2.ImportKeys(t.Context(), keystore.ImportKeysRequest{
			Keys: []keystore.ImportKeyRequest{
				{Password: exportParams.Password, Data: exportResponse.Keys[0].Data},
			},
		})
		require.NoError(t, err)

		// Importing a key with the same name again fails.
		_, err = ks2.ImportKeys(t.Context(), keystore.ImportKeysRequest{
			Keys: []keystore.ImportKeyRequest{
				{Password: exportParams.Password, Data: exportResponse.Keys[0].Data},
			},
		})
		require.ErrorIs(t, err, keystore.ErrKeyAlreadyExists)

		// Importing a key with a new name is allowed.
		_, err = ks2.ImportKeys(t.Context(), keystore.ImportKeysRequest{
			Keys: []keystore.ImportKeyRequest{
				{NewKeyName: "new-name", Password: exportParams.Password, Data: exportResponse.Keys[0].Data},
			},
		})
		require.NoError(t, err)

		// Verify that imported key matches exported key.
		// We cannot compare private keys directly, so we test that signing with key1 from ks1 and verifying
		// with key1 from ks2 works as if the two keys are the same.
		key1ks1, err := ks1.GetKeys(t.Context(), keystore.GetKeysRequest{KeyNames: []string{"key1"}})
		require.NoError(t, err)
		key1ks2, err := ks2.GetKeys(t.Context(), keystore.GetKeysRequest{KeyNames: []string{"key1"}})
		require.NoError(t, err)
		// Test equality of the keys except of the CreatedAt field.
		require.Len(t, key1ks1.Keys, 1)
		require.Len(t, key1ks2.Keys, 1)
		key1ks1Info := key1ks1.Keys[0].KeyInfo
		key1ks2Info := key1ks2.Keys[0].KeyInfo
		require.Equal(t, key1ks1Info.Name, key1ks2Info.Name)
		require.Equal(t, key1ks1Info.PublicKey, key1ks2Info.PublicKey)
		require.Equal(t, key1ks1Info.KeyType, key1ks2Info.KeyType)
		require.Equal(t, key1ks1Info.Metadata, key1ks2Info.Metadata)

		testData := []byte("hello world")
		signature, err := ks2.Sign(t.Context(), keystore.SignRequest{
			KeyName: "key1",
			Data:    testData,
		})
		require.NoError(t, err)
		verifyResp, err := ks1.Verify(t.Context(), keystore.VerifyRequest{
			KeyType:   keystore.Ed25519,
			PublicKey: key1ks1.Keys[0].KeyInfo.PublicKey,
			Data:      testData,
			Signature: signature.Signature,
		})
		require.NoError(t, err)
		require.True(t, verifyResp.Valid)
	})

	t.Run("export non-existent key", func(t *testing.T) {
		_, err = ks1.ExportKeys(t.Context(), keystore.ExportKeysRequest{
			Keys: []keystore.ExportKeyParam{
				{KeyName: "key2", Enc: keystore.EncryptionParams{}},
			},
		})
		require.ErrorIs(t, err, keystore.ErrKeyNotFound)
	})
}

func TestKeystore_SetMetadata(t *testing.T) {
	ks, err := keystore.LoadKeystore(t.Context(), keystore.NewMemoryStorage(), "ks")
	require.NoError(t, err)

	t.Run("update existing key", func(t *testing.T) {
		_, err = ks.CreateKeys(t.Context(), keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{KeyName: "key1", KeyType: keystore.Ed25519},
			},
		})
		require.NoError(t, err)

		_, err = ks.SetMetadata(t.Context(), keystore.SetMetadataRequest{
			[]keystore.SetMetadataUpdate{
				{KeyName: "key1", Metadata: []byte("my-metadata")},
			},
		})
		require.NoError(t, err)

		keysResp, err := ks.GetKeys(t.Context(), keystore.GetKeysRequest{KeyNames: []string{"key1"}})
		require.NoError(t, err)
		require.Len(t, keysResp.Keys, 1)
		assert.Equal(t, []byte("my-metadata"), keysResp.Keys[0].KeyInfo.Metadata)
	})

	t.Run("update non-existent key", func(t *testing.T) {
		_, err = ks.SetMetadata(t.Context(), keystore.SetMetadataRequest{
			[]keystore.SetMetadataUpdate{
				{KeyName: "key2", Metadata: []byte("")},
			},
		})
		require.ErrorIs(t, err, keystore.ErrKeyNotFound)
	})
}

func TestKeystore_RenameKey(t *testing.T) {
	ctx := t.Context()
	ks, err := keystore.LoadKeystore(ctx, keystore.NewMemoryStorage(), "ks")
	require.NoError(t, err)
	_, err = ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "key1", KeyType: keystore.Ed25519},
		},
	})
	require.NoError(t, err)
	originalKey, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{"key1"},
	})
	require.NoError(t, err)
	require.Len(t, originalKey.Keys, 1)

	t.Run("rename non-existent key", func(t *testing.T) {
		_, err = ks.RenameKey(ctx, keystore.RenameKeyRequest{
			OldName: "key2",
			NewName: "new-name",
		})
		require.ErrorIs(t, err, keystore.ErrKeyNotFound)
	})

	t.Run("rename to invalid name", func(t *testing.T) {
		_, err = ks.RenameKey(ctx, keystore.RenameKeyRequest{
			OldName: "key1",
			NewName: "", // Empty name is invalid
		})
		require.ErrorIs(t, err, keystore.ErrInvalidKeyName)
	})

	t.Run("rename to existing name", func(t *testing.T) {
		// Create another key
		_, err = ks.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{KeyName: "another", KeyType: keystore.Ed25519},
			},
		})
		require.NoError(t, err)

		_, err = ks.RenameKey(ctx, keystore.RenameKeyRequest{
			OldName: "key1",
			NewName: "another", // Name already exists
		})
		require.ErrorIs(t, err, keystore.ErrKeyAlreadyExists)
	})

	t.Run("rename to same name", func(t *testing.T) {
		_, err = ks.RenameKey(ctx, keystore.RenameKeyRequest{
			OldName: "key1",
			NewName: "key1",
		})
		require.NoError(t, err)

		// Verify the key still exists
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{"key1"},
		})
		require.NoError(t, err)
		require.Equal(t, resp, originalKey)
	})

	t.Run("successful rename", func(t *testing.T) {
		// Rename the key
		_, err = ks.RenameKey(ctx, keystore.RenameKeyRequest{
			OldName: "key1",
			NewName: "renamed",
		})
		require.NoError(t, err)

		// Verify the key exists under new name
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{"renamed"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Keys, 1)
		require.Equal(t, resp.Keys[0].KeyInfo.Name, "renamed")

		// set name to the old one for easier comparison
		resp.Keys[0].KeyInfo.Name = "key1"
		assert.Equal(t, resp, originalKey)

		// Verify the old name no longer exists
		resp, err = ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{"key1"},
		})
		require.EqualError(t, err, "key not found: key1")
	})
}

func TestECDSA_Serialization_WithPadding(t *testing.T) {
	// This test ensures that ECDSA private keys that serialize to less than 32 bytes
	// are correctly padded with leading zeros during serialization and deserialization.
	// This is important for compatibility with Ethereum's crypto library which expects
	// 32-byte private keys.

	// The example key has been found randomly such that it has 2 leading zero bytes when serialized.
	key, ok := big.NewInt(0).SetString("57269542458293433845411819226400606954116463824740942170224417652371448", 10)
	require.True(t, ok)
	privateKeyBytes := make([]byte, 32)
	key.FillBytes(privateKeyBytes)
	require.Equal(t, []byte{0, 0, 8, 76, 62, 209, 247, 104, 97, 108, 141, 217, 255, 150, 114, 196, 223, 66, 254, 101, 209, 14, 233, 174, 149, 89, 207, 141, 2, 188, 111, 248}, privateKeyBytes)
	deserializedKey, err := gethcrypto.ToECDSA(privateKeyBytes)
	require.NoError(t, err)
	require.Equal(t, key, deserializedKey.D)
}
