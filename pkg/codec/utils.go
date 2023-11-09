package codec

import (
	"encoding/base64"
	"math/big"
	"reflect"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

func FitsInNBitsSigned(n int, bi *big.Int) bool {
	if bi.Sign() < 0 {
		bi = new(big.Int).Neg(bi)
		bi.Sub(bi, big.NewInt(1))
	}
	return bi.BitLen() <= n-1
}

func MergeValueFields(valueFields []map[string]any) (map[string]any, error) {
	numItems := len(valueFields)

	switch numItems {
	case 0:
		return map[string]any{}, nil
	default:
		mergedReflect := map[string]reflect.Value{}
		for k, v := range valueFields[0] {
			rv := reflect.ValueOf(v)
			slice := reflect.MakeSlice(reflect.SliceOf(rv.Type()), numItems, numItems)
			slice.Index(0).Set(rv)
			mergedReflect[k] = slice
		}

		for i, valueField := range valueFields[1:] {
			if len(valueField) != len(mergedReflect) {
				return nil, types.InvalidTypeError{}
			}

			for k, slice := range mergedReflect {
				if value, ok := valueField[k]; ok {
					sliceElm := slice.Index(i + 1)
					rv := reflect.ValueOf(value)
					if !rv.Type().AssignableTo(sliceElm.Type()) {
						return nil, types.InvalidTypeError{}
					}
					sliceElm.Set(rv)
				} else {
					return nil, types.InvalidTypeError{}
				}
			}
		}

		merged := map[string]any{}

		for k, v := range mergedReflect {
			merged[k] = v.Interface()
		}

		return merged, nil
	}
}

func SplitValueFields(decoded map[string]any) ([]map[string]any, error) {
	var result []map[string]any

	for k, v := range decoded {
		iv := reflect.ValueOf(v)
		kind := iv.Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			if kind != reflect.String {
				return nil, types.NotASliceError{}
			}
			rawBytes, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				return nil, types.InvalidTypeError{}
			}
			iv = reflect.ValueOf(rawBytes)
		}

		length := iv.Len()
		if result == nil {
			result = make([]map[string]any, length)
			for i := 0; i < length; i++ {
				result[i] = map[string]any{}
			}
		}

		if len(result) != length {
			return nil, types.InvalidTypeError{}
		}

		for i := 0; i < length; i++ {
			result[i][k] = iv.Index(i).Interface()
		}
	}

	return result, nil
}
