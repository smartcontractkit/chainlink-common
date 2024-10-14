package values

import (
	"errors"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Float64 struct {
	Underlying float64
}

func NewFloat64(f float64) *Float64 {
	return &Float64{Underlying: f}
}

func (f *Float64) proto() *pb.Value {
	return pb.NewFloat64(f.Underlying)
}

func (f *Float64) Unwrap() (any, error) {
	var to float64
	return to, f.UnwrapTo(&to)
}

func (f *Float64) UnwrapTo(to any) error {
	if f == nil {
		return errors.New("cannot unwrap nil values.Float64")
	}

	switch t := to.(type) {
	case *decimal.Decimal:
		if t == nil {
			return errors.New("cannot unwrap to nil pointer")
		}
		*t = decimal.NewFromFloat(f.Underlying)
		return nil
	}

	return unwrapTo(f.Underlying, to)
}

func (f *Float64) copy() Value {
	if f == nil {
		return f
	}
	return &Float64{Underlying: f.Underlying}
}
