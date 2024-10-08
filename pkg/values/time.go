package values

import (
	"errors"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Time struct {
	Underlying time.Time
}

func NewTime(t time.Time) *Time {
	return &Time{Underlying: t}
}

func (t *Time) UnwrapTo(to any) error {
	if t == nil {
		return errors.New("could not unwrap nil values.Time")
	}

	return unwrapTo(t.Underlying, to)
}

func (t *Time) Unwrap() (any, error) {
	tt := new(time.Time)
	return *tt, t.UnwrapTo(tt)
}

func (t *Time) copy() Value {
	if t == nil {
		return nil
	}

	return NewTime(t.Underlying)
}

func (t *Time) proto() *pb.Value {
	return pb.NewTime(t.Underlying)
}
