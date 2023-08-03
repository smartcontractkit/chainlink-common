package fee

import (
	"math/big"

	"github.com/shopspring/decimal"
)

func ApplyMultiplier(feeLimit uint32, multiplier float32) uint32 {
	return uint32(decimal.NewFromBigInt(big.NewInt(0).SetUint64(uint64(feeLimit)), 0).Mul(decimal.NewFromFloat32(multiplier)).IntPart())
}

// Returns the fee in its chain specific unit.
type feeUnitToChainUnit func(fee *big.Int) string
