package keystore_test

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestKeystore_AdminReader(t *testing.T) {
	storage := storage.NewMemoryStorage()
	ks, err := keystore.NewKeystore(storage, "test-password")
	require.NoError(t, err)
	ctx := context.Background()

	req := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{Name: "test-ed25519", KeyType: keystore.Ed25519},
			{Name: "test-secp256k1", KeyType: keystore.EcdsaSecp256k1},
			{Name: "test-x25519", KeyType: keystore.X25519},
		},
	}

	resp, err := ks.CreateKeys(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 3)

	expectedTypes := []keystore.KeyType{keystore.Ed25519, keystore.EcdsaSecp256k1, keystore.X25519}
	for i, key := range resp.Keys {
		require.Equal(t, expectedTypes[i], key.KeyInfo.KeyType)
		require.Equal(t, req.Keys[i].Name, key.KeyInfo.Name)
		require.NotEmpty(t, key.KeyInfo.PublicKey, "Expected non-empty public key for %s", key.KeyInfo.Name)
	}

	getReq := keystore.GetKeysRequest{
		Names: []string{"test-ed25519", "test-secp256k1"},
	}

	getResp, err := ks.GetKeys(ctx, getReq)
	require.NoError(t, err)
	require.Len(t, getResp.Keys, 2)

	allKeysReq := keystore.GetKeysRequest{}
	allKeysResp, err := ks.GetKeys(ctx, allKeysReq)
	require.NoError(t, err)
	require.Len(t, allKeysResp.Keys, 3)

	deleteReq := keystore.DeleteKeysRequest{
		Names: []string{"test-x25519"},
	}

	_, err = ks.DeleteKeys(ctx, deleteReq)
	require.NoError(t, err)

	deleteVerifyReq := keystore.GetKeysRequest{Names: []string{"test-x25519"}}
	_, err = ks.GetKeys(ctx, deleteVerifyReq)
	require.Error(t, err)

	finalKeysReq := keystore.GetKeysRequest{}
	finalKeysResp, err := ks.GetKeys(ctx, finalKeysReq)
	require.NoError(t, err)
	require.Len(t, finalKeysResp.Keys, 2)
}
