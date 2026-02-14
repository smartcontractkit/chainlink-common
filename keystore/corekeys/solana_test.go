package corekeys

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestSolanaKeyRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedSolanaKey(ctx, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	solanaKeyPath := keystore.NewKeyPath(TypeSolana, nameDefault)
	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{solanaKeyPath.String()},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 1)

	storedPublicKey := getKeysResp.Keys[0].KeyInfo.PublicKey
	require.NotEmpty(t, storedPublicKey)

	privateKey, err := FromEncryptedSolanaKey(encryptedKey, password)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)

	require.Len(t, privateKey, 64)

	derivedPublicKey := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	require.Equal(t, storedPublicKey, []byte(derivedPublicKey))
}

func TestSolanaKeyImportWithWrongPassword(t *testing.T) {
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

	encryptedKey, err := coreshimKs.GenerateEncryptedSolanaKey(ctx, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	_, err = FromEncryptedSolanaKey(encryptedKey, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestSolanaKeyImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedSolanaKey([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}
