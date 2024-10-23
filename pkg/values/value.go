package values

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Unwrappable interface {
	Unwrap() (any, error)
	UnwrapTo(any) error
}

type Value interface {
	proto() *pb.Value

	copy() Value
	Unwrappable
}

func Copy(v Value) Value {
	if v == nil {
		return v
	}

	return v.copy()
}

func Wrap(v any) (Value, error) {
	switch tv := v.(type) {
	case map[string]any:
		return NewMap(tv)
	case string:
		return NewString(tv), nil
	case bool:
		return NewBool(tv), nil
	case []byte:
		return NewBytes(tv), nil
	case []any:
		return NewList(tv)
	case decimal.Decimal:
		return NewDecimal(tv), nil
	case int64:
		return NewInt64(tv), nil
	case int32:
		return NewInt64(int64(tv)), nil
	case int16:
		return NewInt64(int64(tv)), nil
	case int8:
		return NewInt64(int64(tv)), nil
	case int:
		return NewInt64(int64(tv)), nil
	case uint64:
		if tv > math.MaxInt64 {
			return NewBigInt(new(big.Int).SetUint64(tv)), nil
		}
		return NewInt64(int64(tv)), nil
	case uint32:
		return NewInt64(int64(tv)), nil
	case uint16:
		return NewInt64(int64(tv)), nil
	case uint8:
		return NewInt64(int64(tv)), nil
	case uint:
		return NewInt64(int64(tv)), nil
	case float64:
		return NewFloat64(tv), nil
	case float32:
		return NewFloat64(float64(tv)), nil
	case *big.Int:
		return NewBigInt(tv), nil
	case time.Time:
		return NewTime(tv), nil
	case nil:
		return nil, nil

	// Transparently wrap values.
	// This is helpful for recursive wrapping of values.
	case *Map:
		return tv, nil
	case *List:
		return tv, nil
	case *String:
		return tv, nil
	case *Bytes:
		return tv, nil
	case *Decimal:
		return tv, nil
	case *Int64:
		return tv, nil
	case *Float64:
		return tv, nil
	case *Bool:
		return tv, nil
	case *BigInt:
		return tv, nil
	case *Time:
		return tv, nil
	}

	// Handle slices, structs, and pointers to structs
	val := reflect.ValueOf(v)

	if val.CanConvert(reflect.TypeOf(decimal.Decimal{})) {
		return Wrap(val.Convert(reflect.TypeOf(decimal.Decimal{})).Interface())
	} else if val.CanConvert(reflect.TypeOf(new(big.Int))) {
		return Wrap(val.Convert(reflect.TypeOf(new(big.Int))).Interface())
	}

	// nolint
	switch val.Kind() {
	// Better complex type support for maps
	case reflect.Map:
		m := make(map[string]any, val.Len())
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			ks, ok := k.Interface().(string)
			if !ok {
				return nil, fmt.Errorf("could not wrap into value %+v", v)
			}
			v := iter.Value()
			m[ks] = v.Interface()
		}
		return NewMap(m)
	// Better complex type support for slices
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return NewBytes(val.Bytes()), nil
		}
		return createListFromSlice(val)
	case reflect.Array:
		arrayLen := val.Len()
		slice := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), arrayLen, arrayLen)
		for i := 0; i < arrayLen; i++ {
			slice.Index(i).Set(val.Index(i))
		}
		return Wrap(slice.Interface())
	case reflect.Struct:
		return CreateMapFromStruct(v)
	case reflect.Pointer:
		// pointer can't be null or the switch statement above would catch it.
		return Wrap(val.Elem().Interface())
	case reflect.String:
		return Wrap(val.Convert(reflect.TypeOf("")).Interface())

	case reflect.Bool:
		return Wrap(val.Convert(reflect.TypeOf(true)).Interface())
	case reflect.Uint64:
		return Wrap(val.Convert(reflect.TypeOf(uint64(0))).Interface())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return Wrap(val.Convert(reflect.TypeOf(int64(0))).Interface())
	case reflect.Float32, reflect.Float64:
		return Wrap(val.Convert(reflect.TypeOf(float64(0))).Interface())
	}

	return nil, fmt.Errorf("could not wrap into value: %+v", v)
}

func createListFromSlice(val reflect.Value) (Value, error) {
	s := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		s[i] = item
	}
	return NewList(s)
}

func WrapMap(a any) (*Map, error) {
	v, err := Wrap(a)
	if err != nil {
		return nil, err
	}

	vm, ok := v.(*Map)
	if !ok {
		return nil, fmt.Errorf("could not wrap %+v to map", a)
	}

	return vm, nil
}

func Unwrap(v Value) (any, error) {
	if v == nil {
		return nil, nil
	}

	return v.Unwrap()
}

func Proto(v Value) *pb.Value {
	if v == nil {
		return &pb.Value{}
	}

	return v.proto()
}

func ProtoMap(v *Map) *pb.Map {
	return Proto(v).GetMapValue()
}

func FromProto(val *pb.Value) (Value, error) {
	if val == nil {
		return nil, nil
	}

	switch val.Value.(type) {
	case nil:
		return nil, nil
	case *pb.Value_StringValue:
		return NewString(val.GetStringValue()), nil
	case *pb.Value_BoolValue:
		return NewBool(val.GetBoolValue()), nil
	case *pb.Value_DecimalValue:
		return fromDecimalValueProto(val.GetDecimalValue()), nil
	case *pb.Value_Int64Value:
		return NewInt64(val.GetInt64Value()), nil
	case *pb.Value_BytesValue:
		return NewBytes(val.GetBytesValue()), nil
	case *pb.Value_ListValue:
		return FromListValueProto(val.GetListValue())
	case *pb.Value_MapValue:
		return FromMapValueProto(val.GetMapValue())
	case *pb.Value_BigintValue:
		return fromBigIntValueProto(val.GetBigintValue()), nil
	case *pb.Value_TimeValue:
		return NewTime(val.GetTimeValue().AsTime()), nil
	case *pb.Value_Float64Value:
		return NewFloat64(val.GetFloat64Value()), nil
	}

	return nil, fmt.Errorf("unsupported type %T: %+v", val, val)
}

func FromMapValueProto(mv *pb.Map) (*Map, error) {
	if mv == nil {
		return nil, nil
	}

	nm := map[string]Value{}
	for k, v := range mv.Fields {
		inner, err := FromProto(v)
		if err != nil {
			return nil, err
		}
		nm[k] = inner
	}
	return &Map{Underlying: nm}, nil
}

func FromListValueProto(lv *pb.List) (*List, error) {
	if lv == nil {
		return nil, nil
	}

	nl := []Value{}
	for _, el := range lv.Fields {
		inner, err := FromProto(el)
		if err != nil {
			return nil, err
		}

		nl = append(nl, inner)
	}
	return &List{Underlying: nl}, nil
}

func fromDecimalValueProto(dec *pb.Decimal) *Decimal {
	if dec == nil {
		return nil
	}

	dc := decimal.NewFromBigInt(protoToBigInt(dec.Coefficient), dec.Exponent)
	return NewDecimal(dc)
}

func protoToBigInt(biv *pb.BigInt) *big.Int {
	if biv == nil {
		return nil
	}

	av := &big.Int{}
	av = av.SetBytes(biv.AbsVal)

	if biv.Sign < 0 {
		av.Neg(av)
	}

	return av
}

func fromBigIntValueProto(biv *pb.BigInt) *BigInt {
	return NewBigInt(protoToBigInt(biv))
}

func CreateMapFromStruct(v any) (*Map, error) {
	var resultMap map[string]interface{}

	err := mapstructure.Decode(v, &resultMap)
	if err != nil {
		return nil, err
	}
	return NewMap(resultMap)
}

func unwrapTo[T any](underlying T, to any) error {
	switch tb := to.(type) {
	case *T:
		if tb == nil {
			return errors.New("cannot unwrap to nil pointer")
		}
		*tb = underlying
	case *any:
		if tb == nil {
			return errors.New("cannot unwrap to nil pointer")
		}
		*tb = underlying
	default:
		// Don't break for custom types that are the same underlying type
		// eg: type FeedId string allows verification of FeedId's shape while unmarshalling
		rTo := reflect.ValueOf(to)
		rUnderlying := reflect.ValueOf(underlying)
		if rTo.Kind() != reflect.Pointer {
			return fmt.Errorf("cannot unwrap to value of type: %T", to)
		}

		if rUnderlying.Type().ConvertibleTo(rTo.Type().Elem()) {
			reflect.Indirect(rTo).Set(rUnderlying.Convert(rTo.Type().Elem()))
			return nil
		}

		rToVal := reflect.Indirect(rTo)
		if rUnderlying.Kind() == reflect.Slice {
			var newList reflect.Value
			if rToVal.Kind() == reflect.Array {
				newListPtr := reflect.New(reflect.ArrayOf(rUnderlying.Len(), rToVal.Type().Elem()))
				newList = reflect.Indirect(newListPtr)
			} else if rToVal.Kind() == reflect.Slice {
				newList = reflect.MakeSlice(rToVal.Type(), rUnderlying.Len(), rUnderlying.Len())
			} else {
				return fmt.Errorf("cannot unwrap slice to value of type: %T", to)
			}

			for i := 0; i < rUnderlying.Len(); i++ {
				el := rUnderlying.Index(i)
				toEl := newList.Index(i)

				if toEl.Kind() == reflect.Ptr {
					err := unwrapTo(el.Interface(), toEl.Interface())
					if err != nil {
						return err
					}
				} else {
					err := unwrapTo(el.Interface(), toEl.Addr().Interface())
					if err != nil {
						return err
					}
				}
			}

			rToVal.Set(newList)
			return nil
		}

		return fmt.Errorf("cannot unwrap to value of type: %T", to)
	}

	return nil
}
