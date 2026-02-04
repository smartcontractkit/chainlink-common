package corekeys

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestP2PKeyRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedP2PKey(ctx, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	p2pKeyPath := keystore.NewKeyPath(TypeP2P, nameDefault)
	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{p2pKeyPath.String()},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 1)

	storedPublicKey := getKeysResp.Keys[0].KeyInfo.PublicKey
	require.NotEmpty(t, storedPublicKey)

	privateKey, err := FromEncryptedP2PKey(encryptedKey, password)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)

	require.Len(t, privateKey, 64)

	derivedPublicKey := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	require.Equal(t, storedPublicKey, []byte(derivedPublicKey))
}

func TestP2PKeyImportWithWrongPassword(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"
	wrongPassword := "wrong-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedP2PKey(ctx, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	_, err = FromEncryptedP2PKey(encryptedKey, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestP2PKeyImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedP2PKey([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}
