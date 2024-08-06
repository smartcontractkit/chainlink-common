package values

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Bool struct {
	Underlying bool
}

func NewBool(b bool) *Bool {
	return &Bool{Underlying: b}
}

func (b *Bool) proto() *pb.Value {
	return pb.NewBoolValue(b.Underlying)
}

func (b *Bool) Unwrap() (any, error) {
	var bl bool
	return bl, b.UnwrapTo(&bl)
}

func (b *Bool) UnwrapTo(to any) error {
	if b == nil {
		return errors.New("could not unwrap nil values.Bool")
	}
	return unwrapTo(b.Underlying, to)
}

func (b *Bool) Copy() Value {
	if b == nil {
		return nil
	}
	return &Bool{Underlying: b.Underlying}
}
