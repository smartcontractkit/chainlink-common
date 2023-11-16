package codec_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-relay/pkg/codec"
)

func TestFitsInNBitsSigned(t *testing.T) {
	t.Parallel()
	t.Run("Fits", func(t *testing.T) {
		bi := big.NewInt(math.MaxInt16)
		assert.True(t, codec.FitsInNBitsSigned(16, bi))
	})

	t.Run("Too large", func(t *testing.T) {
		bi := big.NewInt(math.MaxInt16 + 1)
		assert.False(t, codec.FitsInNBitsSigned(16, bi))
	})

	t.Run("Too small", func(t *testing.T) {
		bi := big.NewInt(math.MinInt16 - 1)
		assert.False(t, codec.FitsInNBitsSigned(16, bi))
	})
}
