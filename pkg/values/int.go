package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Int64 struct {
	Value int64
}

func NewInt64(i int64) (*Int64, error) {
	return &Int64{Value: i}, nil
}

func (i *Int64) Proto() (*pb.Value, error) {
	return pb.NewInt64Value(i.Value)
}

func (i *Int64) Unwrap() (any, error) {
	return i.Value, nil
}
