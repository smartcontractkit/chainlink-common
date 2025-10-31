package keystore_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
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
			ks, err := keystore.LoadKeystore(ctx, storage, "test-password", keystore.WithScryptParams(keystore.FastScryptParams))
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
	ks, err := keystore.LoadKeystore(ctx, st, "test", keystore.WithScryptParams(keystore.FastScryptParams))
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
