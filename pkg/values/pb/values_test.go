package pb

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Value(t *testing.T) {
	type successTestCases struct {
		valType        string
		unwrappedValue any
		wrappedValue   *Value
	}

	wrappedStringValue := &Value{
		Value: &Value_StringValue{
			StringValue: "hello world",
		},
	}
	wrappedInt64Value := &Value{
		Value: &Value_Int64Value{
			Int64Value: 42,
		},
	}
	wrappedDecimalValue := &Value{
		Value: &Value_DecimalValue{
			DecimalValue: &DecimalValue{
				Integral: 13,
				Scale:    -1,
			},
		},
	}

	testCases := []successTestCases{
		{
			valType:        "bool",
			unwrappedValue: false,
			wrappedValue: &Value{
				Value: &Value_BoolValue{
					BoolValue: false,
				},
			},
		},
		{
			valType:        "string",
			unwrappedValue: "hello world",
			wrappedValue:   wrappedStringValue,
		},
		{
			valType:        "int64",
			unwrappedValue: int64(42),
			wrappedValue:   wrappedInt64Value,
		},
		{
			valType:        "decimal",
			unwrappedValue: decimal.NewFromFloat(1.3),
			wrappedValue:   wrappedDecimalValue,
		},
		{
			valType:        "nil",
			unwrappedValue: nil,
			wrappedValue: &Value{
				Value: &Value_NilValue{
					NilValue: &Nil{},
				},
			},
		},
		{
			valType:        "bytes",
			unwrappedValue: []byte("hello world"),
			wrappedValue: &Value{
				Value: &Value_BytesValue{
					BytesValue: []byte("hello world"),
				},
			},
		},
		{
			valType: "map[string]any",
			unwrappedValue: map[string]any{
				"foo": int64(42),
				"bar": "hello world",
				"baz": decimal.NewFromFloat(1.3),
			},
			wrappedValue: &Value{
				Value: &Value_MapValue{
					MapValue: &Map{
						Fields: map[string]*Value{
							"foo": wrappedInt64Value,
							"bar": wrappedStringValue,
							"baz": wrappedDecimalValue,
						},
					},
				},
			},
		},
		{
			valType: "[]any",
			unwrappedValue: []any{
				int64(42),
				"hello world",
				decimal.NewFromFloat(1.3),
			},
			wrappedValue: &Value{
				Value: &Value_ListValue{
					ListValue: &List{
						Fields: []*Value{
							wrappedInt64Value,
							wrappedStringValue,
							wrappedDecimalValue,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Wrap(%s)", tc.valType), func(t *testing.T) {
			newWrappedValue, err := Wrap(tc.unwrappedValue)
			require.NoError(t, err)

			assert.Equal(t, tc.wrappedValue, newWrappedValue)
		})

		t.Run(fmt.Sprintf("(%s).Unwrap()", tc.valType), func(t *testing.T) {
			newUnwrappedValue := tc.wrappedValue.Unwrap()

			assert.Equal(t, tc.unwrappedValue, newUnwrappedValue)
		})
	}

	t.Run("errors when wrapping unsupported value", func(t *testing.T) {
		type StructWithUnsupportedValue struct {
			Field chan int
		}
		_, err := Wrap(&StructWithUnsupportedValue{})
		assert.Contains(t, err.Error(), "could not wrap into value: &{Field:<nil>}; kind *pb.StructWithUnsupportedValue")
	})
}
