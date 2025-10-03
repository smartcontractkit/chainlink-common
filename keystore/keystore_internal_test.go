package keystore

import (
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
	pubKey, err := publicKeyFromPrivateKey(internal.NewRaw(pk.D.Bytes()), EcdsaSecp256k1)
	require.NoError(t, err)
	pubKeyGeth := gethcrypto.FromECDSAPub(&pk.PublicKey)
	require.Equal(t, pubKeyGeth, pubKey)
}
