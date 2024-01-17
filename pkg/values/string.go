package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type String struct {
	Value string
}

func NewString(s string) (*String, error) {
	return &String{Value: s}, nil
}

func (s *String) Proto() (*pb.Value, error) {
	return pb.NewStringValue(s.Value)
}

func (s *String) Unwrap() (any, error) {
	return s.Value, nil
}
