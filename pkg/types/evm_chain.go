package types

import (
	"math/big"
)

type TransactionFee struct {
	TransactionFee *big.Int // Cost of transaction in wei
}
