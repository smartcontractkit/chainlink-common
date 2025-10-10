package keystore

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyFromPrivateKey(t *testing.T) {
	// Confirm that we correctly convert private keys to public keys.
	// (in particular for secp2561k1 since the stdlib doesn't support it)
	pk, err := gethcrypto.GenerateKey()
	require.NoError(t, err)
	pubKey, err := publicKeyFromPrivateKey(internal.NewRaw(pk.D.Bytes()), ECDSA_S256)
	require.NoError(t, err)
	pubKeyGeth := gethcrypto.FromECDSAPub(&pk.PublicKey)
	require.Equal(t, pubKeyGeth, pubKey)
	// We use SEC1 (uncompressed) format for ECDSA public keys.
	require.Equal(t, 65, len(pubKey))

	ecdhPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)
	pubKey, err = publicKeyFromPrivateKey(internal.NewRaw(ecdhPriv.Bytes()), ECDH_P256)
	require.NoError(t, err)
	require.Equal(t, ecdhPriv.PublicKey().Bytes(), pubKey)
	// We use SEC1 (uncompressed) format for ECDH public keys.
	require.Equal(t, 65, len(pubKey))
}
