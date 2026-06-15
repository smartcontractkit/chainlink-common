package starkkey

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

func TestGenerateKeyScalarRange(t *testing.T) {
	t.Parallel()

	// GenerateKey documents sampling in [1, curveOrder-1]; assert that contract.
	key, err := GenerateKey(rand.Reader)
	require.NoError(t, err)

	priv := new(big.Int).SetBytes(internal.Bytes(key.Raw()))
	require.Equal(t, 1, priv.Sign(), "private key must be positive")
	require.Negative(t, priv.Cmp(curveOrder), "private key must be below curve order")
}
