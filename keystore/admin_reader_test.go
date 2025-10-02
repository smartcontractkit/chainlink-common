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
	var (
		testKeyEd25519        = "test-ed25519"
		testKeyEcdsaSecp256k1 = "test-ecdsa-secp256k1"
		testKeyX25519         = "test-x25519"
	)

	req := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: testKeyEd25519, KeyType: keystore.Ed25519},
			{KeyName: testKeyEcdsaSecp256k1, KeyType: keystore.EcdsaSecp256k1},
			{KeyName: testKeyX25519, KeyType: keystore.X25519},
		},
	}

	resp, err := ks.CreateKeys(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 3)

	expectedTypes := []keystore.KeyType{keystore.Ed25519, keystore.EcdsaSecp256k1, keystore.X25519}
	for i, key := range resp.Keys {
		require.Equal(t, expectedTypes[i], key.KeyInfo.KeyType)
		require.Equal(t, req.Keys[i].KeyName, key.KeyInfo.Name)
		require.NotEmpty(t, key.KeyInfo.PublicKey, "Expected non-empty public key for %s", key.KeyInfo.Name)
	}

	getReq := keystore.GetKeysRequest{
		KeyNames: []string{testKeyEd25519, testKeyEcdsaSecp256k1},
	}

	getResp, err := ks.GetKeys(ctx, getReq)
	require.NoError(t, err)
	require.Len(t, getResp.Keys, 2)

	allKeysReq := keystore.GetKeysRequest{}
	allKeysResp, err := ks.GetKeys(ctx, allKeysReq)
	require.NoError(t, err)
	require.Len(t, allKeysResp.Keys, 3)

	deleteReq := keystore.DeleteKeysRequest{
		KeyNames: []string{testKeyX25519},
	}

	_, err = ks.DeleteKeys(ctx, deleteReq)
	require.NoError(t, err)

	deleteVerifyReq := keystore.GetKeysRequest{KeyNames: []string{testKeyX25519}}
	_, err = ks.GetKeys(ctx, deleteVerifyReq)
	require.Error(t, err)

	finalKeysReq := keystore.GetKeysRequest{}
	finalKeysResp, err := ks.GetKeys(ctx, finalKeysReq)
	require.NoError(t, err)
	require.Len(t, finalKeysResp.Keys, 2)
}
