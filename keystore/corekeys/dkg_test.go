package corekeys

import (
	"crypto/ecdh"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestDKGKeyRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedDKGKey(ctx, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	dkgKeyPath := keystore.NewKeyPath(TypeDKG, nameDefault)
	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{dkgKeyPath.String()},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 1)

	storedPublicKey := getKeysResp.Keys[0].KeyInfo.PublicKey
	require.NotEmpty(t, storedPublicKey)

	privateKey, err := FromEncryptedDKGKey(encryptedKey, password)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)

	require.Len(t, privateKey, 32)

	curve := ecdh.P256()
	p256PrivateKey, err := curve.NewPrivateKey(privateKey)
	require.NoError(t, err)
	derivedPublicKey := p256PrivateKey.PublicKey().Bytes()
	require.Equal(t, storedPublicKey, []byte(derivedPublicKey))
}

func TestDKGKeyImportWithWrongPassword(t *testing.T) {
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

	encryptedKey, err := coreshimKs.GenerateEncryptedDKGKey(ctx, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	_, err = FromEncryptedDKGKey(encryptedKey, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestDKGKeyImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedDKGKey([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}
