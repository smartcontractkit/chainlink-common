package values

import (
	"errors"
	"fmt"
	"math"
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
		if tb == nil {
			return errors.New("cannot unwrap to nil pointer")
		}
		if b.Underlying.Cmp(new(big.Int).SetUint64(math.MaxUint64)) > 0 {
			return errors.New("big.Int value is larger than uint64")
		}
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
		} else if rto.CanConvert(reflect.TypeOf(new(uint64))) {
			return b.UnwrapTo(rto.Convert(reflect.TypeOf(new(uint64))).Interface())
		} else if rto.CanInt() || rto.CanUint() {
			if b.Underlying.Cmp(big.NewInt(math.MaxInt64)) > 0 {
				return fmt.Errorf("big.Int value is larger than int64")
			}

			return NewInt64(b.Underlying.Int64()).UnwrapTo(to)
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
