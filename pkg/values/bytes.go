package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Bytes struct {
	Value []byte
}

func NewBytes(b []byte) (*Bytes, error) {
	return &Bytes{Value: b}, nil
}

func (b *Bytes) Proto() (*pb.Value, error) {
	return pb.NewBytesValue(b.Value)
}

func (b *Bytes) Unwrap() (any, error) {
	return b.Value, nil
}
