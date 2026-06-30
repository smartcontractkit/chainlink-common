package starkkey

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStarkKey_SignRoundTrip(t *testing.T) {
	t.Parallel()

	key, err := New()
	require.NoError(t, err)

	hash := []byte("starknet message hash placeholder")
	sig, err := key.Sign(hash)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	pub := key.PublicKey()
	require.NotNil(t, pub.X)
	require.NotNil(t, pub.Y)
}

func TestMustNewInsecure(t *testing.T) {
	t.Parallel()

	key := MustNewInsecure(rand.Reader)
	require.NotEmpty(t, key.StarkKeyStr())
}
