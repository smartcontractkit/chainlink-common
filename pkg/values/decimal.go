package values

import (
	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Decimal struct {
	Value decimal.Decimal
}

func NewDecimal(d decimal.Decimal) (*Decimal, error) {
	return &Decimal{Value: d}, nil
}

func (d *Decimal) Proto() (*pb.Value, error) {
	return pb.NewDecimalValue(d.Value)
}

func (d *Decimal) Unwrap() (any, error) {
	return d.Value, nil
}
