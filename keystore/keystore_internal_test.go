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

func TestJoinKeySegments(t *testing.T) {
	tests := []struct {
		segments []string
		expected string
	}{
		{segments: []string{"evm", "tx", "my-key"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "/tx", "my-key"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx/", "my-key"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "/my-key"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "my-key", ""}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "my-key", "/"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "my-key", "//"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "my-key", "///"}, expected: "evm/tx/my-key"},
		{segments: []string{"evm", "tx", "my-key", "////"}, expected: "evm/tx/my-key"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.expected, joinKeySegments(tt.segments...))
	}
}

func TestKeyPathHasPrefix(t *testing.T) {
	tests := []struct {
		path     KeyPath
		prefix   KeyPath
		expected bool
	}{
		{path: KeyPath{"evm", "tx", "my-key"}, prefix: KeyPath{"evm", "tx"}, expected: true},
		{path: KeyPath{"evm", "tx", "my-key"}, prefix: KeyPath{"evm"}, expected: true},
		{path: KeyPath{"evm", "tx", "my-key"}, prefix: KeyPath{"evm", "tx", "my-key"}, expected: true},
		{path: KeyPath{"evm", "tx", "my-key"}, prefix: KeyPath{"evm", "tx", "my-key", "extra"}, expected: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.path.HasPrefix(tt.prefix))
	}
}
