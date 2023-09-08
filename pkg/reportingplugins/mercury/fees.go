package mercury

import (
	"github.com/pkg/errors"
	"math/big"
)

// PriceScalingFactor indicates the multiplier applied to token prices that we expect from data source
// e.g. for a 1e8 multiplier, a LINK/USD value of 7.42 will be represented as 742000000
// The factor is decreased 1e8 -> 1e6 to comnpensate for baseUSDFee being in cents not usd
var PRICE_SCALING_FACTOR = big.NewFloat(1e6)

// FeeScalingFactor indicates the multiplier applied to fees.
// e.g. for a 1e18 multiplier, a LINK fee of 7.42 will be represented as 7.42e18
// This is what will be baked into the report for use on-chain.
var FEE_SCALING_FACTOR = big.NewFloat(1e18)

// CalculateFee outputs a fee in wei according to the formula: baseUSDFeeCents * scaleFactor / tokenPriceInUSD
func CalculateFee(tokenPriceInUSD *big.Int, baseUSDFeeCents uint32) (*big.Int, error) {
	if tokenPriceInUSD.Cmp(big.NewInt(0)) == 0 || baseUSDFeeCents == 0 {
		return nil, errors.Errorf("token price and base fee must be non-zero")
	}

	// scale baseFee in USD
	baseFeeScaled := new(big.Float).Mul(new(big.Float).SetInt64(int64(baseUSDFeeCents)), PRICE_SCALING_FACTOR)

	tokenPrice := new(big.Float).SetInt(tokenPriceInUSD)

	// fee denominated in token
	fee := new(big.Float).Quo(baseFeeScaled, tokenPrice)

	// scale fee to the expected format
	fee.Mul(fee, FEE_SCALING_FACTOR)

	// convert to big.Int
	finalFee, _ := fee.Int(nil)
	return finalFee, nil
}
