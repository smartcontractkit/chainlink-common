package values

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Decimal struct {
	Underlying *decimal.Decimal
}

func NewDecimal(d decimal.Decimal) *Decimal {
	return &Decimal{Underlying: &d}
}

func (d *Decimal) proto() *pb.Value {
	return pb.NewDecimalValue(*d.Underlying)
}

func (d *Decimal) Unwrap() (any, error) {
	var dec decimal.Decimal
	return dec, d.UnwrapTo(&dec)
}

func (d *Decimal) UnwrapTo(to any) error {
	if d == nil {
		return errors.New("could not unwrap nil values.Decimal")
	}
	return unwrapTo(*d.Underlying, to)
}

func (d *Decimal) Copy() Value {
	if d == nil {
		return nil
	}
	cpy := d.Underlying.Copy()
	return &Decimal{Underlying: &cpy}
}
