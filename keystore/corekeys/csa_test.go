package corekeys

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestCSAKeyRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedCSAKey(ctx, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	csaKeyPath := keystore.NewKeyPath(TypeCSA, nameDefault)
	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{csaKeyPath.String()},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 1)

	storedPublicKey := getKeysResp.Keys[0].KeyInfo.PublicKey
	require.NotEmpty(t, storedPublicKey)

	privateKey, err := FromEncryptedCSAKey(encryptedKey, password)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)

	require.Len(t, privateKey, 64)

	derivedPublicKey := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	require.Equal(t, storedPublicKey, []byte(derivedPublicKey))
}

func TestCSAKeyImportWithWrongPassword(t *testing.T) {
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

	encryptedKey, err := coreshimKs.GenerateEncryptedCSAKey(ctx, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	_, err = FromEncryptedCSAKey(encryptedKey, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestCSAKeyImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedCSAKey([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}
