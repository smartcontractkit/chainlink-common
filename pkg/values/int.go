package values

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Int64 struct {
	Underlying int64
}

func NewInt64(i int64) *Int64 {
	return &Int64{Underlying: i}
}

func (i *Int64) proto() *pb.Value {
	return pb.NewInt64Value(i.Underlying)
}

func (i *Int64) Unwrap() (any, error) {
	var u int64
	return u, i.UnwrapTo(&u)
}

func (i *Int64) copy() Value {
	if i == nil {
		return nil
	}
	return &Int64{Underlying: i.Underlying}
}

func (i *Int64) UnwrapTo(to any) error {
	if i == nil {
		return errors.New("cannot unwrap nil values.Int64")
	}

	if to == nil {
		return fmt.Errorf("cannot unwrap to nil pointer: %+v", to)
	}

	if reflect.ValueOf(to).Kind() != reflect.Pointer {
		return fmt.Errorf("cannot unwrap to non-pointer value: %+v", to)
	}

	rToVal := reflect.Indirect(reflect.ValueOf(to))
	switch rToVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rToVal.OverflowInt(i.Underlying) {
			return fmt.Errorf("cannot unwrap int64 to %T: overflow", to)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if rToVal.OverflowUint(uint64(i.Underlying)) {
			return fmt.Errorf("cannot unwrap int64 to %T: overflow", to)
		}
	}

	return unwrapTo(i.Underlying, to)
}
