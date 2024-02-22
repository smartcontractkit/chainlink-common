package pb

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
)

func Wrap(v any) (*Value, error) {
	switch tv := v.(type) {
	case map[string]any:
		return NewMapValue(tv)
	case string:
		return NewStringValue(tv)
	case bool:
		return NewBoolValue(tv)
	case []byte:
		return NewBytesValue(tv)
	case []any:
		return NewListValue(tv)
	case decimal.Decimal:
		return NewDecimalValue(tv)
	case int64:
		return NewInt64Value(tv)
	case int:
		return NewInt64Value(int64(tv))
	case nil:
		return NewNilValue()
	}

	// This can happen when v is unsupported, e.g., a channel.
	return nil, fmt.Errorf("could not wrap into value: %+v; kind %s", v, reflect.TypeOf(v))
}

func (val *Value) Unwrap() any {
	switch val.Value.(type) {
	case *Value_NilValue:
		return nil
	case *Value_StringValue:
		return val.GetStringValue()
	case *Value_BoolValue:
		return val.GetBoolValue()
	case *Value_DecimalValue:
		return decimal.NewFromInt(val.GetDecimalValue().Integral).Shift(val.GetDecimalValue().Scale)
	case *Value_Int64Value:
		return val.GetInt64Value()
	case *Value_BytesValue:
		return val.GetBytesValue()
	case *Value_ListValue:
		newList := []any{}
		for _, el := range val.GetListValue().Fields {
			newList = append(newList, el.Unwrap())
		}
		return newList
	case *Value_MapValue:
		newMap := map[string]any{}
		for k, field := range val.GetMapValue().Fields {
			newMap[k] = field.Unwrap()
		}
		return newMap
	}

	panic("unreachable")
}

func NewBoolValue(b bool) (*Value, error) {
	return &Value{
		Value: &Value_BoolValue{
			BoolValue: b,
		},
	}, nil
}

func NewBytesValue(b []byte) (*Value, error) {
	return &Value{
		Value: &Value_BytesValue{
			BytesValue: b,
		},
	}, nil
}

func NewDecimalValue(d decimal.Decimal) (*Value, error) {
	return &Value{
		Value: &Value_DecimalValue{
			DecimalValue: &DecimalValue{
				Integral: d.Coefficient().Int64(),
				Scale:    d.Exponent(),
			},
		},
	}, nil
}

func NewStringValue(s string) (*Value, error) {
	return &Value{
		Value: &Value_StringValue{
			StringValue: s,
		},
	}, nil
}

func NewInt64Value(i int64) (*Value, error) {
	return &Value{
		Value: &Value_Int64Value{
			Int64Value: i,
		},
	}, nil
}

func NewMapValue(m map[string]any) (*Value, error) {
	var fields = make(map[string]*Value)

	for k, v := range m {
		wv, err := Wrap(v)
		if err != nil {
			return nil, err
		}

		fields[k] = wv
	}

	return &Value{
		Value: &Value_MapValue{
			MapValue: &Map{
				Fields: fields,
			},
		},
	}, nil
}

func NewListValue(m []any) (*Value, error) {
	var vals []*Value

	for _, v := range m {
		wv, err := Wrap(v)
		if err != nil {
			return nil, err
		}

		vals = append(vals, wv)
	}

	return &Value{
		Value: &Value_ListValue{
			ListValue: &List{
				Fields: vals,
			},
		},
	}, nil
}

func NewNilValue() (*Value, error) {
	return &Value{
		Value: &Value_NilValue{
			NilValue: &Nil{},
		},
	}, nil
}
