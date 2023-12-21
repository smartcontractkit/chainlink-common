package bigendian

import (
	"fmt"
	"math"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// Float64 follows IEEE-754 convention for float64, the specification says what the "high" and "low" bits represent
// Leaving big endian to specify the byte ordering.
type Float64 struct{}

func (Float64) Encode(value any, into []byte) ([]byte, error) {
	f64, ok := value.(float64)
	if ok {
		return Uint64{}.Encode(math.Float64bits(f64), into)
	}

	return nil, fmt.Errorf("%w expected float64, got %T", types.ErrInvalidType, value)
}

func (Float64) Decode(encoded []byte) (any, []byte, error) {
	u64, remaining, err := Uint64{}.Decode(encoded)
	return math.Float64frombits(u64.(uint64)), remaining, err
}

func (Float64) GetType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

func (Float64) Size(_ int) (int, error) {
	return 4, nil
}

func (Float64) FixedSize() (int, error) {
	return 4, nil
}

var _ encodings.TypeCodec = Float64{}
