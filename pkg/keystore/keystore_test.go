package keystore

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore/storage"
	"github.com/stretchr/testify/require"
)

func TestKeystore(t *testing.T) {
	storage := storage.NewMemoryStorage()
	ks, err := NewKeystore(storage, "test-password")
	require.NoError(t, err)
	ctx := context.Background()

	req := CreateKeysRequest{
		Keys: []CreateKeyRequest{
			{Name: "test-ed25519", KeyType: Ed25519},
			{Name: "test-secp256k1", KeyType: Secp256k1},
			{Name: "test-x25519", KeyType: X25519},
		},
	}

	resp, err := ks.CreateKeys(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 3)

	expectedTypes := []KeyType{Ed25519, Secp256k1, X25519}
	for i, key := range resp.Keys {
		require.Equal(t, expectedTypes[i], key.KeyInfo.KeyType)
		require.Equal(t, req.Keys[i].Name, key.KeyInfo.Name)
		require.NotEmpty(t, key.KeyInfo.PublicKey, "Expected non-empty public key for %s", key.KeyInfo.Name)
	}

	getReq := GetKeysRequest{
		Names: []string{"test-ed25519", "test-secp256k1"},
	}

	getResp, err := ks.GetKeys(ctx, getReq)
	require.NoError(t, err)
	require.Len(t, getResp.Keys, 2)

	allKeysReq := GetKeysRequest{}
	allKeysResp, err := ks.GetKeys(ctx, allKeysReq)
	require.NoError(t, err)
	require.Len(t, allKeysResp.Keys, 3)

	deleteReq := DeleteKeysRequest{
		Names: []string{"test-x25519"},
	}

	_, err = ks.DeleteKeys(ctx, deleteReq)
	require.NoError(t, err)

	deleteVerifyReq := GetKeysRequest{Names: []string{"test-x25519"}}
	_, err = ks.GetKeys(ctx, deleteVerifyReq)
	require.Error(t, err)

	finalKeysReq := GetKeysRequest{}
	finalKeysResp, err := ks.GetKeys(ctx, finalKeysReq)
	require.NoError(t, err)
	require.Len(t, finalKeysResp.Keys, 2)
}
