package values

import (
	"errors"

	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/libocr/commontypes"
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

	t := reflect.TypeOf(to)

	if t.Elem().Kind() == reflect.Array {

		// Don't use the Kind attribute of Elem() to check type here, doing so will cause a panic
		// if the array is an alias of byte type
		var bt byte
		if t.Elem().Elem() == reflect.TypeOf(bt) {
			reflect.Copy(reflect.ValueOf(to).Elem(), reflect.ValueOf(b.Underlying))
			return nil
		}

		// Handle alias types first else
		var oid commontypes.OracleID
		if t.Elem().Elem() == reflect.TypeOf(oid) {
			var oracleIDS []commontypes.OracleID
			// TODO look into better way to convert []underlying to []alias that does not require a loop or unsafe package??
			for _, v := range b.Underlying {
				oracleIDS = append(oracleIDS, commontypes.OracleID(v))
			}

			reflect.Copy(reflect.ValueOf(to).Elem(), reflect.ValueOf(oracleIDS))
			return nil
		}
	}

	return unwrapTo(b.Underlying, to)
}

func (b *Bytes) copy() Value {
	if b == nil {
		return nil
	}

	dest := make([]byte, len(b.Underlying))
	copy(dest, b.Underlying)
	return &Bytes{Underlying: dest}
}
