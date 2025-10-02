package keystore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeystore_AdminCreateDeleteKeys(t *testing.T) {
	storage := storage.NewMemoryStorage()
	ks, err := keystore.NewKeystore(storage, keystore.EncryptionParams{
		Password:     "test-password",
		ScryptParams: keystore.FastScryptParams,
	})
	require.NoError(t, err)
	ctx := context.Background()
	// Can create keys of all types.
	var keys []keystore.GetKeyResponse
	for _, kType := range keystore.AllKeyTypes {
		name := fmt.Sprintf("test-key-%s", kType)
		_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{
					KeyName: name,
					KeyType: kType,
				},
			},
		})
		require.NoError(t, err)
		// Should be able to read back the key we created.
		resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{name},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Keys))
		require.Equal(t, kType, resp.Keys[0].KeyInfo.KeyType)
		require.Equal(t, name, resp.Keys[0].KeyInfo.Name)
		require.NotEmpty(t, resp.Keys[0].KeyInfo.PublicKey)
		require.Empty(t, resp.Keys[0].KeyInfo.Metadata)
		keys = append(keys, resp.Keys[0])
	}
	// Get all should return same keys we created.
	resp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{})
	require.NoError(t, err)
	require.Equal(t, len(keystore.AllKeyTypes), len(resp.Keys))
	assert.ElementsMatch(t, keys, resp.Keys)

	// Can't create keys with duplicate names.
	for _, kType := range keystore.AllKeyTypes {
		_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{
					KeyName: fmt.Sprintf("test-key-%s", kType),
					KeyType: kType,
				},
			},
		})
		require.Error(t, err)
	}

	// Delete each key we created.
	for _, k := range keys {
		_, err := ks.DeleteKeys(ctx, keystore.DeleteKeysRequest{
			KeyNames: []string{k.KeyInfo.Name},
		})
		require.NoError(t, err)
		// Key no longer exists
		_, err = ks.GetKeys(ctx, keystore.GetKeysRequest{
			KeyNames: []string{k.KeyInfo.Name},
		})
		require.Error(t, err)
	}
	// Get all should return no keys.
	resp, err = ks.GetKeys(ctx, keystore.GetKeysRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Keys)

	// Can't delete keys that don't exist.
	_, err = ks.DeleteKeys(ctx, keystore.DeleteKeysRequest{
		KeyNames: []string{fmt.Sprintf("test-key-%s", keystore.AllKeyTypes[0])},
	})
	require.Error(t, err)
}
