package mercury

import (
	"github.com/pkg/errors"
	"math/big"
)

// PriceScalingFactor indicates the multiplier applied to token prices.
// e.g. for a 1e8 multiplier, a LINK/USD value of 7.42 will be represented as 742000000
// This is what we expect from our data source.
var PRICE_SCALING_FACTOR = big.NewFloat(1e8)

// FeeScalingFactor indicates the multiplier applied to fees.
// e.g. for a 1e18 multiplier, a LINK fee of 7.42 will be represented as 7.42e18
// This is what will be baked into the report for use on-chain.
var FEE_SCALING_FACTOR = big.NewFloat(1e18)

var CENTS_PER_DOLLAR = big.NewFloat(100)

// CalculateFee outputs a fee in wei
func CalculateFee(tokenPriceInUSD *big.Int, baseUSDFeeCents uint32) (*big.Int, error) {
	if tokenPriceInUSD.Cmp(big.NewInt(0)) == 0 || baseUSDFeeCents == 0 {
		return nil, errors.Errorf("token price and base fee must be non-zero")
	}

	// big.Float base fee in USD
	baseFee := new(big.Float).Quo(new(big.Float).SetInt64(int64(baseUSDFeeCents)), CENTS_PER_DOLLAR)

	// big.Float descaled token price in USD
	tokenPrice := new(big.Float).Quo(new(big.Float).SetInt(tokenPriceInUSD), PRICE_SCALING_FACTOR)

	// big.Float fee denominated in token
	fee := new(big.Float).Quo(baseFee, tokenPrice)

	// scale fee to the expected format
	fee.Mul(fee, FEE_SCALING_FACTOR)

	// convert to big.Int
	finalFee, _ := fee.Int(nil)
	return finalFee, nil
}
