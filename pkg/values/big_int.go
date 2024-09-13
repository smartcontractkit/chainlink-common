package values

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

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
	if b == nil || b.Underlying == nil {
		return errors.New("could not unwrap nil")
	}

	// check any here because unwrap to will make the *any point to a big.Int instead of *big.Int
	switch tb := to.(type) {
	case *big.Int:
		if tb == nil {
			return errors.New("cannot unwrap to nil pointer")
		}
		*tb = *b.Underlying
	case *uint64:
		*tb = b.Underlying.Uint64()
	case *any:
		if tb == nil {
			return errors.New("cannot unwrap to nil pointer")
		}

		*tb = b.Underlying
		return nil
	default:
		rto := reflect.ValueOf(to)
		if rto.CanConvert(reflect.TypeOf(new(big.Int))) {
			return b.UnwrapTo(rto.Convert(reflect.TypeOf(new(big.Int))).Interface())
		}
		return fmt.Errorf("cannot unwrap to value of type: %T", to)
	}

	return nil
}

func (b *BigInt) copy() Value {
	if b == nil {
		return nil
	}

	nw := new(big.Int)
	nw.Set(b.Underlying)
	return &BigInt{Underlying: nw}
}
