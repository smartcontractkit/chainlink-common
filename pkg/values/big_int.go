package values

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type BigInt struct {
	Underlying *big.Int
}

func NewBigInt(b *big.Int) *BigInt {
	return &BigInt{Underlying: b}
}

func (b *BigInt) proto() *pb.Value {
	return pb.NewBigIntValue(
		b.Underlying.Sign(),
		b.Underlying.Bytes(),
	)
}

func (b *BigInt) Unwrap() (any, error) {
	bi := &big.Int{}
	return bi, b.UnwrapTo(bi)
}

func (b *BigInt) UnwrapTo(to any) error {
	if b == nil {
		return errors.New("could not unwrap nil values.BigInt")
	}

	switch tb := to.(type) {
	case *big.Int:
		if tb == nil {
			return fmt.Errorf("cannot unwrap to nil pointer")
		}
		*tb = *b.Underlying
	case *any:
		if tb == nil {
			return fmt.Errorf("cannot unwrap to nil pointer")
		}
		*tb = b.Underlying
	default:
		return fmt.Errorf("cannot unwrap to value of type: %T", to)
	}

	return nil
}

func (b *BigInt) Copy() Value {
	if b == nil {
		return nil
	}

	nw := new(big.Int)
	nw.Set(b.Underlying)
	return &BigInt{Underlying: nw}
}
