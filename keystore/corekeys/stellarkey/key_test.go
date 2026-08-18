package stellarkey

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

func TestStellarKey(t *testing.T) {
	// Same seed used by the aptoskey test vector — the ed25519 public key is
	// therefore identical, which documents that only the address ENCODING
	// differs between ed25519 chains.
	bytes, err := hex.DecodeString("f0d07ab448018b2754475f9a3b580218b0675a1456aad96ad607c7bbd7d9237b")
	require.NoError(t, err)
	k := KeyFor(internal.NewRaw(bytes))

	// Raw ed25519 public key (matches aptoskey's PublicKeyStr for the same seed).
	assert.Equal(t, "2acd605efc181e2af8a0b8c0686a5e12578efa1253d15a235fa5e5ad970c4b29", hex.EncodeToString(k.GetPublic()))

	// StrKey "G..." account address — generated with github.com/stellar/go-stellar-sdk.
	const wantAddr = "GAVM2YC67QMB4KXYUC4MA2DKLYJFPDX2CJJ5CWRDL6S6LLMXBRFSSNPD"
	assert.Equal(t, wantAddr, k.ID())
	assert.Equal(t, wantAddr, k.Account())
	assert.Equal(t, wantAddr, k.PublicKeyStr())
}

func TestStellarKey_SignRoundTrip(t *testing.T) {
	k, err := New()
	require.NoError(t, err)

	msg := []byte("stellar tx hash placeholder")
	sig, err := k.Sign(msg)
	require.NoError(t, err)
	require.Len(t, sig, 64) // ed25519 signatures are 64 bytes

	// The public key verifies the signature produced by the keystore key.
	assert.True(t, ed25519.Verify(k.GetPublic(), msg, sig))
}
