package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Bytes struct {
	Underlying []byte
}

func NewBytes(b []byte) (*Bytes, error) {
	return &Bytes{Underlying: b}, nil
}

func MustNewBytes(b []byte) *Bytes {
	value, _ := NewBytes(b)
	return value
}

func (b *Bytes) Proto() (*pb.Value, error) {
	return pb.NewBytesValue(b.Underlying)
}

func (b *Bytes) Unwrap() (any, error) {
	return b.Underlying, nil
}
