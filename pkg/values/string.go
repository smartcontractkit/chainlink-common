package values

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type String struct {
	Underlying string
}

func NewString(s string) *String {
	return &String{Underlying: s}
}

func (s *String) proto() *pb.Value {
	return pb.NewStringValue(s.Underlying)
}

func (s *String) Unwrap() (any, error) {
	var to string
	return to, s.UnwrapTo(&to)
}

func (s *String) UnwrapTo(to any) error {
	if s == nil {
		return errors.New("cannot unwrap nil values.String")
	}
	return unwrapTo(s.Underlying, to)
}

func (s *String) copy() Value {
	if s == nil {
		return s
	}
	return &String{Underlying: s.Underlying}
}
