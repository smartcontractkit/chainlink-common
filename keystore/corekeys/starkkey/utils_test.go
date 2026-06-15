package starkkey

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

func TestGenerateKeyNonZero(t *testing.T) {
	t.Parallel()

	for i := 0; i < 20; i++ {
		key, err := GenerateKey(rand.Reader)
		require.NoError(t, err)

		priv := new(big.Int).SetBytes(internal.Bytes(key.Raw()))
		require.NotZero(t, priv.Sign(), "private key must not be zero")
		require.Negative(t, priv.Cmp(curveOrder), "private key must be below curve order")
	}
}
