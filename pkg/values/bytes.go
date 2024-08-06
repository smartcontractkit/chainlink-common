package values

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Bytes struct {
	Underlying []byte
}

func NewBytes(b []byte) *Bytes {
	return &Bytes{Underlying: b}
}

func (b *Bytes) proto() *pb.Value {
	return pb.NewBytesValue(b.Underlying)
}

func (b *Bytes) Unwrap() (any, error) {
	if b == nil {
		return nil, errors.New("cannot unwrap nil values.Bytes")
	}
	return b.Underlying, nil
}

func (b *Bytes) UnwrapTo(to any) error {
	if b == nil {
		return errors.New("cannot unwrap nil values.Bytes")
	}
	return unwrapTo(b.Underlying, to)
}

func (b *Bytes) Copy() Value {
	if b == nil {
		return nil
	}

	dest := make([]byte, len(b.Underlying))
	copy(dest, b.Underlying)
	return &Bytes{Underlying: dest}
}
