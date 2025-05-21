package pb

import (
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewBoolValue(b bool) *Value {
	return &Value{
		Value: &Value_BoolValue{
			BoolValue: b,
		},
	}
}

func NewBytesValue(b []byte) *Value {
	return &Value{
		Value: &Value_BytesValue{
			BytesValue: b,
		},
	}
}

func NewDecimalValue(d decimal.Decimal) *Value {
	return &Value{
		Value: &Value_DecimalValue{
			DecimalValue: &Decimal{
				Coefficient: &BigInt{
					AbsVal: d.Coefficient().Bytes(),
					Sign:   int64(d.Coefficient().Sign()),
				},
				Exponent: d.Exponent(),
			},
		},
	}
}

func NewStringValue(s string) *Value {
	return &Value{
		Value: &Value_StringValue{
			StringValue: s,
		},
	}
}

func NewMapValue(m map[string]*Value) *Value {
	return &Value{
		Value: &Value_MapValue{
			MapValue: &Map{
				Fields: m,
			},
		},
	}
}

func NewListValue(m []*Value) *Value {
	return &Value{
		Value: &Value_ListValue{
			ListValue: &List{
				Fields: m,
			},
		},
	}
}

func NewInt64Value(i int64) *Value {
	return &Value{
		Value: &Value_Int64Value{
			Int64Value: i,
		},
	}
}

func NewBigIntValue(sign int, bib []byte) *Value {
	return &Value{
		Value: &Value_BigintValue{
			BigintValue: &BigInt{
				AbsVal: bib,
				Sign:   int64(sign),
			},
		},
	}
}

func NewTime(t time.Time) *Value {
	return &Value{
		Value: &Value_TimeValue{
			TimeValue: timestamppb.New(t),
		},
	}
}

func NewFloat64(f float64) *Value {
	return &Value{
		Value: &Value_Float64Value{
			Float64Value: f,
		},
	}
}
