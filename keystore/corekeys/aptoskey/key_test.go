package aptoskey

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

func TestAptosKey(t *testing.T) {
	bytes, err := hex.DecodeString("f0d07ab448018b2754475f9a3b580218b0675a1456aad96ad607c7bbd7d9237b")
	require.NoError(t, err)
	k := KeyFor(internal.NewRaw(bytes))
	assert.Equal(t, "2acd605efc181e2af8a0b8c0686a5e12578efa1253d15a235fa5e5ad970c4b29", k.PublicKeyStr())
	assert.Equal(t, "69d8b07f5945185873c622ea66873b0e1fb921de7b94d904d3ef9be80770682e", k.Account())
}

func TestKey_New_StoresAccountAddress(t *testing.T) {
	key, err := New()
	require.NoError(t, err)

	// Account() must return a non-empty 64-char hex string
	assert.Len(t, key.Account(), 64)
	// Account() must NOT equal PublicKeyStr() (they are different values)
	assert.NotEqual(t, key.PublicKeyStr(), key.Account())
}

func TestKey_AccountIsStable(t *testing.T) {
	key, err := New()
	require.NoError(t, err)

	first := key.Account()
	second := key.Account()
	assert.Equal(t, first, second, "Account() must return the same value on repeated calls")
}

func TestKey_WithAccountAddress(t *testing.T) {
	key, err := New()
	require.NoError(t, err)

	customAddr := "000000000000000000000000000000000000000000000000000000000000cafe"
	updated := key.WithAccountAddress(customAddr)

	assert.Equal(t, customAddr, updated.Account())
	// original key is unchanged (value receiver copy)
	assert.NotEqual(t, customAddr, key.Account())
	// public key is unaffected
	assert.Equal(t, key.PublicKeyStr(), updated.PublicKeyStr())
}

func TestKey_MarshalUnmarshal_RoundTrip(t *testing.T) {
	key, err := New()
	require.NoError(t, err)

	customAddr := "000000000000000000000000000000000000000000000000000000000000cafe"
	key = key.WithAccountAddress(customAddr)

	data := key.Marshal()
	restored, err := Unmarshal(data)
	require.NoError(t, err)

	assert.Equal(t, customAddr, restored.Account())
	assert.Equal(t, key.PublicKeyStr(), restored.PublicKeyStr())
	assert.Equal(t, key.ID(), restored.ID())
}

func TestKey_Unmarshal_LegacyRawSeed(t *testing.T) {
	seed, err := hex.DecodeString("f0d07ab448018b2754475f9a3b580218b0675a1456aad96ad607c7bbd7d9237b")
	require.NoError(t, err)

	// Legacy format: just raw seed bytes (not JSON)
	key, err := Unmarshal(seed)
	require.NoError(t, err)

	assert.Equal(t, "2acd605efc181e2af8a0b8c0686a5e12578efa1253d15a235fa5e5ad970c4b29", key.PublicKeyStr())
	assert.Equal(t, "69d8b07f5945185873c622ea66873b0e1fb921de7b94d904d3ef9be80770682e", key.Account())
}
