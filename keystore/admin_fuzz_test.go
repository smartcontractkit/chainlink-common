package keystore_test

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"testing"

	ks "github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/require"
)

func FuzzKeystore_AdminModel(f *testing.F) {
	// Here we fuzz a sequence of operations on the keystore
	// and check that the invariants are maintained.
	testKeyNames := []string{"test-key-1", "test-key-2", "test-key-3"}
	//testMetadata := [][]byte{[]byte("test-metadata-1"), []byte("test-metadata-2")}
	type key struct {
		keyType  ks.KeyType
		metadata []byte
	}

	f.Fuzz(func(t *testing.T, seed []byte) {
		t.Parallel()
		ctx := context.Background()
		mem := storage.NewMemoryStorage()
		enc := ks.EncryptionParams{Password: "pw", ScryptParams: ks.FastScryptParams}
		kstore, err := ks.NewKeystore(mem, enc)
		require.NoError(t, err)

		// Generate a random sequence of operations
		// from the seed.
		h := sha256.Sum256(seed)
		r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(h[:]))))
		n := 1 + r.Intn(5)

		expected := make(map[string]key)
		for i := 0; i < n; i++ {
			// TODO: extend to other operations.
			switch r.Intn(2) {
			case 0:
				// create keys
				numKeys := r.Intn(len(testKeyNames) - 1)
				keys := make([]ks.CreateKeyRequest, numKeys)
				for i := 0; i < numKeys; i++ {
					kIndex := r.Intn(len(ks.AllKeyTypes) - 1)
					kType := ks.AllKeyTypes[kIndex]
					kNameIndex := r.Intn(len(testKeyNames) - 1)
					kName := testKeyNames[kNameIndex]
					keys[i] = ks.CreateKeyRequest{KeyName: kName, KeyType: kType}
				}
				_, _ = kstore.CreateKeys(ctx, ks.CreateKeysRequest{Keys: keys})
				for _, k := range keys {
					expected[k.KeyName] = key{keyType: k.KeyType, metadata: []byte{}}
				}
			case 1:
				// delete keys
				numKeys := r.Intn(len(testKeyNames) - 1)
				var req ks.DeleteKeysRequest
				for i := 0; i < numKeys; i++ {
					kNameIndex := r.Intn(len(testKeyNames) - 1)
					kName := testKeyNames[kNameIndex]
					req.KeyNames = append(req.KeyNames, kName)
				}
				_, _ = kstore.DeleteKeys(ctx, req)
				for _, name := range req.KeyNames {
					delete(expected, name)
				}
			}
		}

		// Check that the invariants are maintained:
		// 1. No duplicate names.
		// 2. Net set keys is as expected (including metadata).
		resp, err := kstore.GetKeys(ctx, ks.GetKeysRequest{})
		require.NoError(t, err)

		have := make(map[string]key)
		for _, k := range resp.Keys {
			have[k.KeyInfo.Name] = key{
				keyType:  k.KeyInfo.KeyType,
				metadata: k.KeyInfo.Metadata,
			}
		}
		require.Equal(t, len(expected), len(have))
		for expectedKeyName, expectedKey := range expected {
			haveKey, ok := have[expectedKeyName]
			require.True(t, ok)
			require.Equal(t, expectedKey.keyType, haveKey.keyType)
			require.Equal(t, expectedKey.metadata, haveKey.metadata)
		}
	})
}
