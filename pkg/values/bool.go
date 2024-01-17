package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Bool struct {
	Value bool
}

func NewBool(b bool) (*Bool, error) {
	return &Bool{Value: b}, nil
}

func (b *Bool) Proto() (*pb.Value, error) {
	return pb.NewBoolValue(b.Value)
}

func (b *Bool) Unwrap() (any, error) {
	return b.Value, nil
}
