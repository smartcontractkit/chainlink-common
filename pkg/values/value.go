package values

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Unwrappable interface {
	Unwrap() (any, error)
}

type Value interface {
	Proto() (*pb.Value, error)

	Unwrappable
}

func Wrap(v any) (Value, error) {
	switch tv := v.(type) {
	case map[string]any:
		return NewMap(tv)
	case string:
		return NewString(tv)
	case bool:
		return NewBool(tv)
	case []byte:
		return NewBytes(tv)
	case []any:
		return NewList(tv)
	case decimal.Decimal:
		return NewDecimal(tv)
	case int64:
		return NewInt64(tv)
	case int:
		return NewInt64(int64(tv))
	case nil:
		return NewNil()

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
	case *Nil:
		return tv, nil
	}

	// Handle structs and pointers to structs
	switch reflect.ValueOf(v).Kind() {
	case reflect.Struct:
		return createMapFromStruct(v)
	case reflect.Pointer:
		if reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Struct {
			return createMapFromStruct(&v)
		}
	}

	return nil, fmt.Errorf("could not wrap into value: %+v", v)
}

func Unwrap(v Value) (any, error) {
	if v == nil {
		return nil, nil
	}

	return v.Unwrap()
}

func FromProto(val *pb.Value) (Value, error) {
	if val == nil {
		return nil, nil
	}

	switch val.Value.(type) {
	case *pb.Value_NilValue:
		return NewNil()
	case *pb.Value_StringValue:
		return NewString(val.GetStringValue())
	case *pb.Value_BoolValue:
		return NewBool(val.GetBoolValue())
	case *pb.Value_DecimalValue:
		return FromDecimalValueProto(val.GetDecimalValue())
	case *pb.Value_Int64Value:
		return NewInt64(val.GetInt64Value())
	case *pb.Value_BytesValue:
		return FromBytesValueProto(val.GetBytesValue())
	case *pb.Value_ListValue:
		return FromListValueProto(val.GetListValue())
	case *pb.Value_MapValue:
		return FromMapValueProto(val.GetMapValue())
	}

	return nil, fmt.Errorf("unsupported type %T: %+v", val, val)
}

func FromBytesValueProto(bv string) (*Bytes, error) {
	p, err := base64.StdEncoding.DecodeString(bv)
	if err != nil {
		return nil, err
	}
	return NewBytes(p)
}

func FromMapValueProto(mv *pb.Map) (*Map, error) {
	nm := map[string]Value{}
	for k, v := range mv.Fields {
		val, err := FromProto(v)
		if err != nil {
			return nil, err
		}

		nm[k] = val
	}
	return &Map{Underlying: nm}, nil
}

func FromListValueProto(lv *pb.List) (*List, error) {
	nl := []Value{}
	for _, el := range lv.Fields {
		elv, err := FromProto(el)
		if err != nil {
			return nil, err
		}

		nl = append(nl, elv)
	}
	return &List{Underlying: nl}, nil
}

func FromDecimalValueProto(decStr string) (*Decimal, error) {
	dec := decimal.Decimal{}
	err := json.Unmarshal([]byte(decStr), &dec)
	if err != nil {
		return nil, err
	}
	return NewDecimal(dec)
}

func createMapFromStruct(v any) (Value, error) {
	var resultMap map[string]interface{}
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// convert to a map
	err = json.Unmarshal(jsonData, &resultMap)
	if err != nil {
		return nil, err
	}

	convertedResultMap := convertFloat64ToInt64(resultMap) // TODO: Remove by supporting Float64
	return NewMap(convertedResultMap.(map[string]any))
}

// Recursively converts float64 values to int64 and handles nested maps.
func convertFloat64ToInt64(data any) any {
	switch v := data.(type) {
	case float64:
		return int64(v)
	case map[string]any:
		// If the value is a map, iterate through its keys.
		for key, value := range v {
			v[key] = convertFloat64ToInt64(value)
		}
	case []any:
		// If the value is a slice, iterate through its elements.
		for i, value := range v {
			v[i] = convertFloat64ToInt64(value)
		}
	}
	return data
}
