package values

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Error struct {
	Underlying error
}

func NewError(e error) (*Error, error) {
	return &Error{Underlying: e}, nil
}

func (e *Error) Proto() (*pb.Value, error) {
	return pb.NewErrorValue(e.Underlying)
}

func (e *Error) Unwrap() (any, error) {
	return e.Underlying, nil
}
