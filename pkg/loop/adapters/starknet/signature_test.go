package starknet

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestSignature(t *testing.T) {

	s, err := SignatureFromBigInts(big.NewInt(7),
		big.NewInt(11))

	require.NoError(t, err)

	x, y, err := s.Ints()
	require.Equal(t, big.NewInt(7), x)
	require.Equal(t, big.NewInt(11), y)

	b, err := s.Bytes()
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Len(t, b, signatureLen)

	roundTrip, err := SignatureFromBytes(b)
	require.NoError(t, err)
	require.Equal(t, s.x, roundTrip.x)
	require.Equal(t, s.y, roundTrip.y)

	// no negative allowed
	s, err = SignatureFromBigInts(big.NewInt(-7),
		big.NewInt(11))
	require.Error(t, err)
	require.Nil(t, s)
}
