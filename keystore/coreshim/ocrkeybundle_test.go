package coreshim

import (
	"crypto/ed25519"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/curve25519"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/ocr2offchain"
)

func TestOCRKeyBundleRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"
	chainType := ChainType("EVM")

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewKeystore(ks)

	encryptedBundle, err := coreshimKs.GenerateEncryptedOCRKeyBundle(ctx, chainType, password)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedBundle)

	signingKeyPath := keystore.NewKeyPath(ocr2offchain.PrefixOCR2Offchain, keyNameDefault, ocr2offchain.OCR2OffchainSigning)
	encryptionKeyPath := keystore.NewKeyPath(ocr2offchain.PrefixOCR2Offchain, keyNameDefault, ocr2offchain.OCR2OffchainEncryption)
	onchainKeyPath := keystore.NewKeyPath(PrefixOCR2Onchain, keyNameDefault, string(chainType))

	getKeysResp, err := ks.GetKeys(ctx, keystore.GetKeysRequest{
		KeyNames: []string{
			signingKeyPath.String(),
			encryptionKeyPath.String(),
			onchainKeyPath.String(),
		},
	})
	require.NoError(t, err)
	require.Len(t, getKeysResp.Keys, 3)

	var storedSigningPubKey, storedEncryptionPubKey, storedOnchainPubKey []byte
	for _, key := range getKeysResp.Keys {
		switch key.KeyInfo.Name {
		case signingKeyPath.String():
			storedSigningPubKey = key.KeyInfo.PublicKey
		case encryptionKeyPath.String():
			storedEncryptionPubKey = key.KeyInfo.PublicKey
		case onchainKeyPath.String():
			storedOnchainPubKey = key.KeyInfo.PublicKey
		}
	}

	require.NotEmpty(t, storedSigningPubKey)
	require.NotEmpty(t, storedEncryptionPubKey)
	require.NotEmpty(t, storedOnchainPubKey)

	bundle, err := FromEncryptedOCRKeyBundle(encryptedBundle, password)
	require.NoError(t, err)
	require.NotNil(t, bundle)

	require.NotEmpty(t, bundle.OffchainSigningKey)
	derivedSigningPubKey := ed25519.PrivateKey(bundle.OffchainSigningKey).Public().(ed25519.PublicKey)
	require.Equal(t, storedSigningPubKey, []byte(derivedSigningPubKey))

	var derivedEncryptionPubKey [32]byte
	curve25519.ScalarBaseMult(&derivedEncryptionPubKey, (*[32]byte)(bundle.OffchainEncryptionKey))
	require.Equal(t, storedEncryptionPubKey, derivedEncryptionPubKey[:])

	onchainPrivKey, err := crypto.ToECDSA(bundle.OnchainSigningKey)
	require.NoError(t, err)
	derivedOnchainPubKey := crypto.FromECDSAPub(&onchainPrivKey.PublicKey)
	require.Equal(t, storedOnchainPubKey, derivedOnchainPubKey)
}

func TestOCRKeyBundleImportWithWrongPassword(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"
	wrongPassword := "wrong-password"
	chainType := ChainType("EVM")

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewKeystore(ks)

	encryptedBundle, err := coreshimKs.GenerateEncryptedOCRKeyBundle(ctx, chainType, password)
	require.NoError(t, err)
	require.NotNil(t, encryptedBundle)

	_, err = FromEncryptedOCRKeyBundle(encryptedBundle, wrongPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not decrypt data")
}

func TestOCRKeyBundleImportInvalidFormat(t *testing.T) {
	t.Parallel()

	_, err := FromEncryptedOCRKeyBundle([]byte("invalid json"), "password")
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not unmarshal import data")
}

func TestOCRKeyBundleInvalidKeyType(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	password := "test-password"

	st := keystore.NewMemoryStorage()
	ks, err := keystore.LoadKeystore(ctx, st, "test",
		keystore.WithScryptParams(keystore.FastScryptParams),
	)
	require.NoError(t, err)

	coreshimKs := NewKeystore(ks)

	// Generate a CSA key and try to import it as an OCR key bundle
	encryptedCSAKey, err := coreshimKs.GenerateEncryptedCSAKey(ctx, password)
	require.NoError(t, err)

	_, err = FromEncryptedOCRKeyBundle(encryptedCSAKey, password)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key type")
}
