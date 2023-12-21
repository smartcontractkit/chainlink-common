package bigendian

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/smartcontractkit/libocr/bigbigendian"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewBigInt(numBytes int, signed bool) (encodings.TypeCodec, error) {
	if numBytes > bigbigendian.MaxSize || numBytes < 0 {
		return nil, fmt.Errorf(
			"%w: numBytes is %v, but must be between 1 and %v", types.ErrInvalidConfig, numBytes, bigbigendian.MaxSize)
	}
	return &bigInt{
		NumBytes: numBytes,
		Signed:   signed,
	}, nil
}

type bigInt struct {
	NumBytes int
	Signed   bool
}

var _ encodings.TypeCodec = &bigInt{}

func (i *bigInt) Encode(value any, into []byte) ([]byte, error) {
	bi, ok := value.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("%w: expected big.Int, got %T", types.ErrInvalidType, value)
	}

	if i.Signed {
		bytes, err := bigbigendian.SerializeSigned(i.NumBytes, bi)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", types.ErrInvalidType, err)
		}

		return append(into, bytes...), nil
	}

	if bi.Sign() < 0 {
		return nil, fmt.Errorf("%w: cannot encode %v as unsigned", types.ErrInvalidType, bi)
	}

	if bi.BitLen() > i.NumBytes*8 {
		return nil, fmt.Errorf("%w: %v doesn't fit into a %v-bytes", types.ErrInvalidType, bi, i.NumBytes)
	}

	bts := make([]byte, i.NumBytes)
	bi.FillBytes(bts)
	return append(into, bts...), nil
}

func (i *bigInt) Decode(encoded []byte) (any, []byte, error) {
	if i.Signed {
		return encodings.SafeDecode[*big.Int](encoded, i.NumBytes, func(bytes []byte) *big.Int {
			b, _ := bigbigendian.DeserializeSigned(i.NumBytes, bytes)
			return b
		})
	}

	return encodings.SafeDecode[*big.Int](encoded, i.NumBytes, func(bytes []byte) *big.Int {
		return new(big.Int).SetBytes(bytes)
	})
}

func (i *bigInt) GetType() reflect.Type {
	return reflect.TypeOf((*big.Int)(nil))
}

func (i *bigInt) Size(_ int) (int, error) {
	return i.NumBytes, nil
}

func (i *bigInt) FixedSize() (int, error) {
	return i.NumBytes, nil
}
