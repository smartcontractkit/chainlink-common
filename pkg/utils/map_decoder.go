package utils

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

func DecoderFromMapDecoder(decoder types.MapDecoder, extraHooks ...mapstructure.DecodeHookFunc) (types.Decoder, error) {
	if decoder == nil {
		return nil, errors.New("decoder must not be nil")
	}

	numExtraHooks := len(extraHooks)
	hooks := make([]mapstructure.DecodeHookFunc, numExtraHooks+2)
	copy(hooks, extraHooks)
	hooks[numExtraHooks] = SliceToArrayVerifySizeHook
	hooks[numExtraHooks+1] = BigIntHook
	return &mapDecoder{decoder: decoder, hooks: hooks}, nil
}

type mapDecoder struct {
	decoder types.MapDecoder
	hooks   []mapstructure.DecodeHookFunc
}

func (m *mapDecoder) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return m.decoder.GetMaxDecodingSize(ctx, n, itemType)
}

func (m *mapDecoder) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	rInto := reflect.ValueOf(into)
	if rInto.Kind() != reflect.Pointer {
		return types.InvalidTypeError{}
	}

	elm := reflect.Indirect(rInto)
	switch elm.Kind() {
	case reflect.Array:
		return m.decodeMultiple(ctx, raw, arrayProvider(elm), itemType)
	case reflect.Slice:
		return m.decodeMultiple(ctx, raw, sliceProvider(elm), itemType)
	default:
		rawMap, err := m.decoder.DecodeSingle(ctx, raw, itemType)
		if err != nil {
			return err
		}
		return m.mapToItem(rawMap, into)
	}
}

// VerifyFieldMaps is a utility for verifying the keys exactly match the fields
// it is not done in Decode directly as it can often be more efficiently by MapDecoders
// in the case of DecodeMany
func VerifyFieldMaps(fields []string, result map[string]any) error {
	for _, field := range fields {
		if _, ok := result[field]; !ok {
			return types.InvalidEncodingError{}
		}
	}

	if len(fields) != len(result) {
		return types.InvalidEncodingError{}
	}

	return nil
}

func arrayProvider(rInto reflect.Value) func(size int) (reflect.Value, error) {
	return func(size int) (reflect.Value, error) {
		if rInto.Len() != size {
			return reflect.Value{}, types.WrongNumberOfElements{}
		}
		return rInto, nil
	}
}

func sliceProvider(rInto reflect.Value) func(size int) (reflect.Value, error) {
	return func(size int) (reflect.Value, error) {
		element := reflect.MakeSlice(rInto.Type(), size, size)
		rInto.Set(element)
		return rInto, nil
	}
}

func (m *mapDecoder) decodeMultiple(ctx context.Context, raw []byte, init func(size int) (reflect.Value, error), itemType string) error {
	decoded, err := m.decoder.DecodeMany(ctx, raw, itemType)
	if err != nil {
		return err
	}

	rInto, err := init(len(decoded))
	if err != nil {
		return err
	}

	for i, singleDecode := range decoded {
		if err = m.mapToItem(singleDecode, rInto.Index(i).Addr().Interface()); err != nil {
			return err
		}
	}

	return nil
}

func BigIntHook(_, to reflect.Type, data any) (any, error) {
	if to == reflect.TypeOf(&big.Int{}) {
		bigInt := big.NewInt(0)

		switch v := data.(type) {
		case float64:
			bigInt.SetInt64(int64(v))
		case float32:
			bigInt.SetInt64(int64(v))
		case int:
			bigInt.SetInt64(int64(v))
		case int8:
			bigInt.SetInt64(int64(v))
		case int16:
			bigInt.SetInt64(int64(v))
		case int32:
			bigInt.SetInt64(int64(v))
		case int64:
			bigInt.SetInt64(v)
		case uint:
			bigInt.SetUint64(uint64(v))
		case uint8:
			bigInt.SetUint64(uint64(v))
		case uint16:
			bigInt.SetUint64(uint64(v))
		case uint32:
			bigInt.SetUint64(uint64(v))
		case uint64:
			bigInt.SetUint64(v)
		case string:
			_, ok := bigInt.SetString(v, 10)
			if !ok {
				return nil, types.InvalidTypeError{}
			}
		default:
			return data, nil
		}

		return bigInt, nil
	}

	if bi, ok := data.(*big.Int); ok {
		switch to {
		case reflect.TypeOf(0):
			if FitsInNBitsSigned(strconv.IntSize, bi) {
				return int(bi.Int64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(int8(0)):
			if FitsInNBitsSigned(8, bi) {
				return int8(bi.Int64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(int16(0)):
			if FitsInNBitsSigned(16, bi) {
				return int16(bi.Int64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(int32(0)):
			if FitsInNBitsSigned(32, bi) {
				return int32(bi.Int64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(int64(0)):
			if FitsInNBitsSigned(64, bi) {
				return bi.Int64(), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(uint(0)):
			if bi.Sign() >= 0 && bi.BitLen() <= strconv.IntSize {
				return uint(bi.Uint64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(uint8(0)):
			if bi.Sign() >= 0 && bi.BitLen() <= 8 {
				return uint8(bi.Uint64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(uint16(0)):
			if bi.Sign() >= 0 && bi.BitLen() <= 16 {
				return uint16(bi.Uint64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(uint32(0)):
			if bi.Sign() >= 0 && bi.BitLen() <= 32 {
				return uint32(bi.Uint64()), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(uint64(0)):
			if bi.Sign() >= 0 && bi.BitLen() <= 64 {
				return bi.Uint64(), nil
			}
			return nil, types.InvalidTypeError{}
		case reflect.TypeOf(""):
			return bi.String(), nil
		default:
			return data, nil
		}
	}

	return data, nil
}

func SliceToArrayVerifySizeHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	// By default, if the array is bigger it'll still work. (ie []int{1, 2, 3} => [4]int{} works with 0 at the end
	// [2]int{} would not work. This seems to lenient, but can be discussed.
	if from.Kind() == reflect.Slice && to.Kind() == reflect.Array {
		slice := reflect.ValueOf(data)
		if slice.Len() != to.Len() {
			return nil, types.WrongNumberOfElements{}
		}
	}

	return data, nil
}

func (m *mapDecoder) mapToItem(rawMap map[string]any, into any) error {
	md := &mapstructure.Metadata{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		// TODO add hook to verify number sizes. mapstructure seems to check -ve values for unsigned, but not other boundaries
		DecodeHook: mapstructure.ComposeDecodeHookFunc(m.hooks...),
		Metadata:   md,
		Result:     into,
	})

	if err != nil {
		return types.InvalidTypeError{}
	}

	if err = decoder.Decode(rawMap); err != nil {
		return types.InvalidTypeError{}
	}

	return nil
}
