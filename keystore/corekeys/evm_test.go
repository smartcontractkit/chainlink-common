package corekeys

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func TestEVMKeyRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewStore(ks)

	encryptedKey, err := coreshimKs.GenerateEncryptedEVMKey(ctx, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedKey)

	evmKeyPath := keystore.NewKeyPath(TypeEVM, nameDefault)
	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{evmKeyPath.String()},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 1)

	storedPublicKey := getKeysResp.Keys[0].KeyInfo.PublicKey
	require.NotEmpty(t, storedPublicKey)

	privateKey, err := FromEncryptedEVMKey(encryptedKey, password)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)

	require.Len(t, privateKey, 32)

	ecdsaPK, err := crypto.ToECDSA(privateKey)
	require.NoError(t, err)

	derivedPublicKey := crypto.FromECDSAPub(&ecdsaPK.PublicKey)
	require.Equal(t, storedPublicKey, derivedPublicKey)
}

func TestEVMKeyImportWithWrongPassword(t *testing.T) {
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

	encryptedKey, err := coreshimKs.GenerateEncryptedEVMKey(ctx, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedKey)

	_, err = FromEncryptedEVMKey(encryptedKey, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestEVMKeyImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedEVMKey([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}
