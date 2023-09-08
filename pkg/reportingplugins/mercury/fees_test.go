package mercury

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func scaleTokenPrice(usdPrice float64) *big.Int {
	scaledPrice := new(big.Float).Mul(big.NewFloat(usdPrice), PRICE_SCALING_FACTOR)
	scaledPriceInt, _ := scaledPrice.Int(nil)
	return scaledPriceInt
}

func Test_Fees(t *testing.T) {
	var baseUSDFeeCents uint32 = 70
	t.Run("with token price > 1", func(t *testing.T) {
		tokenPriceInUSD := scaleTokenPrice(1630)
		fee, err := CalculateFee(tokenPriceInUSD, baseUSDFeeCents)
		assert.NoError(t, err)
		expectedFee := big.NewInt(429447852760736) // 0.000429447852760736 18 decimals
		if fee.Cmp(expectedFee) != 0 {
			t.Errorf("Expected fee to be %v, got %v", expectedFee, fee)
		}
	})

	t.Run("with token price < 1", func(t *testing.T) {
		tokenPriceInUSD := scaleTokenPrice(0.4)
		fee, err := CalculateFee(tokenPriceInUSD, baseUSDFeeCents)
		assert.NoError(t, err)
		expectedFee := big.NewInt(1750000000000000000) // 1.75 18 decimals
		if fee.Cmp(expectedFee) != 0 {
			t.Errorf("Expected fee to be %v, got %v", expectedFee, fee)
		}
	})

	t.Run("with token price == 0", func(t *testing.T) {
		tokenPriceInUSD := scaleTokenPrice(0)
		_, err := CalculateFee(tokenPriceInUSD, baseUSDFeeCents)
		assert.EqualError(t, err, "token price and base fee must be non-zero")
	})
}
