package values

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestValueEvent struct {
	TriggerType string       `json:"triggerType"`
	ID          string       `json:"id"`
	Timestamp   string       `json:"timestamp"`
	Payload     []TestReport `json:"payload"`
}

type TestReport struct {
	FeedID     int64  `json:"feedId"`
	FullReport string `json:"fullreport"`
}

func Test_Value(t *testing.T) {
	testCases := []struct {
		name     string
		newValue func() (any, Value, error)
		equal    func(t *testing.T, expected any, unwrapped any)
	}{
		{
			name: "map",
			newValue: func() (any, Value, error) {
				m := map[string]any{
					"hello": "world",
				}
				mv, err := NewMap(m)
				return m, mv, err
			},
		},
		{
			name: "list",
			newValue: func() (any, Value, error) {
				l := []any{
					1,
					"2",
					decimal.NewFromFloat(1.0),
				}
				lv, err := NewList(l)
				return l, lv, err
			},
			equal: func(t *testing.T, expected any, unwrapped any) {
				e, u := expected.([]any), unwrapped.([]any)
				assert.Equal(t, int64(e[0].(int)), u[0])
				assert.Equal(t, e[1], u[1])
				assert.Equal(t, e[2].(decimal.Decimal).String(), u[2].(decimal.Decimal).String())
			},
		},
		{
			name: "decimal",
			newValue: func() (any, Value, error) {
				dec, err := decimal.NewFromString("1.03")
				if err != nil {
					return nil, nil, err
				}
				decv := NewDecimal(dec)
				return dec, decv, err
			},
		},
		{
			name: "string",
			newValue: func() (any, Value, error) {
				s := "hello"
				sv := NewString(s)
				return s, sv, nil
			},
		},
		{
			name: "bytes",
			newValue: func() (any, Value, error) {
				b := []byte("hello")
				bv := NewBytes(b)
				return b, bv, nil
			},
		},
		{
			name: "bool",
			newValue: func() (any, Value, error) {
				b := true
				bv := NewBool(b)
				return b, bv, nil
			},
		},
		{
			name: "bigInt",
			newValue: func() (any, Value, error) {
				b := big.NewInt(math.MaxInt64)
				bv := NewBigInt(b)
				return b, bv, nil
			},
		},
		{
			name: "recursive map",
			newValue: func() (any, Value, error) {
				m := map[string]any{
					"hello": map[string]any{
						"world": "foo",
					},
					"baz": []any{
						int64(1), int64(2), int64(3),
					},
				}
				mv, err := NewMap(m)
				return m, mv, err
			},
		},
		{
			name: "struct",
			newValue: func() (any, Value, error) {
				var v TestReport
				m := map[string]any{
					"FeedID":     int64(2),
					"FullReport": "hello",
				}
				err := mapstructure.Decode(m, &v)
				if err != nil {
					return nil, nil, err
				}
				vv, err := Wrap(v)
				return m, vv, err
			},
		},
		{
			name: "structPointer",
			newValue: func() (any, Value, error) {
				var v TestReport
				m := map[string]any{
					"FeedID":     int64(3),
					"FullReport": "world",
				}
				err := mapstructure.Decode(m, &v)
				if err != nil {
					return nil, nil, err
				}
				vv, err := Wrap(&v)
				return m, vv, err
			},
		},
		{
			name: "nestedStruct",
			newValue: func() (any, Value, error) {
				var v TestValueEvent
				m := map[string]any{
					"TriggerType": "mercury",
					"ID":          "123",
					"Timestamp":   "123",
					"Payload": []any{
						map[string]any{
							"FeedID":     int64(4),
							"FullReport": "hello",
						},
						map[string]any{
							"FeedID":     int64(5),
							"FullReport": "world",
						},
					},
				}
				err := mapstructure.Decode(m, &v)
				if err != nil {
					return nil, nil, err
				}
				vv, err := Wrap(v)
				return m, vv, err
			},
		},
		{
			name: "map of values",
			newValue: func() (any, Value, error) {
				bar := "bar"
				str := &String{Underlying: bar}
				l, err := NewList([]any{1, 2, 3})
				if err != nil {
					return nil, nil, err
				}
				m := map[string]any{
					"hello": map[string]any{
						"string": str,
						"nil":    nil,
						"list":   l,
					},
				}
				mv, err := NewMap(m)

				list := []any{int64(1), int64(2), int64(3)}
				expectedUnwrapped := map[string]any{
					"hello": map[string]any{
						"string": bar,
						"nil":    nil,
						"list":   list,
					},
				}

				return expectedUnwrapped, mv, err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			originalValue, wrapped, err := tc.newValue()
			require.NoError(st, err)

			pb := Proto(wrapped)

			rehydratedValue, err := FromProto(pb)
			require.NoError(t, err)
			assert.Equal(st, wrapped, rehydratedValue)

			unwrapped, err := Unwrap(rehydratedValue)
			require.NoError(st, err)
			if tc.equal != nil {
				tc.equal(st, originalValue, unwrapped)
			} else {
				assert.Equal(st, originalValue, unwrapped)
			}
		})
	}
}

func Test_StructWrapUnwrap(t *testing.T) {
	// TODO: https://smartcontract-it.atlassian.net/browse/KS-439 decimal.Decimal is broken when encoded.
	type sStruct struct {
		Str string
		I   int
		Bi  *big.Int
		// D  decimal.Decimal
	}
	expected := sStruct{
		Str: "hi",
		I:   10,
		Bi:  big.NewInt(1),
		// D:  decimal.NewFromFloat(24.3),
	}

	wrapped, err := Wrap(expected)
	require.NoError(t, err)

	unwrapped := sStruct{}
	err = wrapped.UnwrapTo(&unwrapped)
	require.NoError(t, err)

	assert.Equal(t, expected, unwrapped)
}

func Test_SameUnderlyingTypes(t *testing.T) {
	type str string
	type i int
	type bi big.Int
	// TODO https://smartcontract-it.atlassian.net/browse/KS-439 decimal.Decimal is broken when encoded.
	// type d decimal.Decimal
	type sStruct struct {
		Str str
		I   i
		Bi  *bi
		// D   d
	}
	expected := sStruct{
		Str: "hi",
		I:   10,
		Bi:  (*bi)(big.NewInt(1)),
		// D:   d(decimal.NewFromFloat(24.3)),
	}

	wrapped, err := Wrap(expected)
	require.NoError(t, err)

	unwrapped := sStruct{}
	err = wrapped.UnwrapTo(&unwrapped)
	require.NoError(t, err)

	// big ints don't pass assert equal because pointer isn't the same
	assert.Equal(t, 0, (*big.Int)(expected.Bi).Cmp((*big.Int)(unwrapped.Bi)))
	expected.Bi = unwrapped.Bi
	assert.Equal(t, expected, unwrapped)
}

func Test_WrapMap(t *testing.T) {
	a := struct{ A string }{A: "foo"}
	am, err := WrapMap(a)
	require.NoError(t, err)

	assert.Len(t, am.Underlying, 1)
	assert.Equal(t, am.Underlying["A"], NewString("foo"))

	_, err = WrapMap("foo")
	require.ErrorContains(t, err, "could not wrap")
}

func Test_Copy(t *testing.T) {
	dec, err := decimal.NewFromString("1.01")
	require.NoError(t, err)

	list, err := NewList([]any{"hello", int64(1.00)})
	require.NoError(t, err)

	mp, err := NewMap(map[string]any{
		"hello": 1,
		"world": map[string]any{
			"a": "b",
			"c": 10,
		},
		"foo": big.NewInt(100),
		"bar": decimal.NewFromFloat(1.00),
	})
	require.NoError(t, err)

	tcs := []struct {
		value Value
		isNil bool
	}{
		{
			value: NewString("hello"),
		},
		{
			value: NewBytes([]byte("hello")),
		},
		{
			value: NewInt64(int64(100)),
		},
		{
			value: NewDecimal(dec),
		},
		{
			value: NewBigInt(big.NewInt(101)),
		},
		{
			value: NewBool(true),
		},
		{
			value: list,
		},
		{
			value: mp,
		},
		{
			value: (*String)(nil),
			isNil: true,
		},
		{
			value: (*Bytes)(nil),
			isNil: true,
		},
		{
			value: (*Int64)(nil),
			isNil: true,
		},
		{
			value: (*BigInt)(nil),
			isNil: true,
		},
		{
			value: (*Bool)(nil),
			isNil: true,
		},
		{
			value: (*List)(nil),
			isNil: true,
		},
		{
			value: (*Map)(nil),
			isNil: true,
		},
	}

	for _, tc := range tcs {
		copied := Copy(tc.value)
		if tc.isNil {
			assert.Nil(t, Copy(tc.value))
		} else {
			assert.Equal(t, tc.value, copied)
		}
	}
}

type aliasByte []byte
type aliasString string
type aliasInt int
type aliasMap map[string]any
type aliasSingleByte uint8

func Test_Aliases(t *testing.T) {
	testCases := []struct {
		name    string
		val     func() any
		alias   func() any
		convert func(any) any
	}{
		{
			name:  "alias to []byte",
			val:   func() any { return []byte("string") },
			alias: func() any { return aliasByte([]byte{}) },
		},
		{
			name:  "simple aliases",
			val:   func() any { return "string" },
			alias: func() any { return aliasString("") },
		},
		{
			name:  "aliasByte -> []byte",
			val:   func() any { return []byte("string") },
			alias: func() any { return aliasByte([]byte{}) },
		},
		{
			name:  "[]aliasSingleByte -> []byte",
			val:   func() any { return []byte("string") },
			alias: func() any { return []aliasSingleByte{} },
		},
		{
			name:    "int",
			val:     func() any { return 2 },
			alias:   func() any { return aliasInt(0) },
			convert: func(a any) any { return int(a.(int64)) },
		},
		{
			name:  "[][]byte -> []aliasByte",
			val:   func() any { return [][]byte{[]byte("hello")} },
			alias: func() any { return []aliasByte{} },
			convert: func(a any) any {
				to := [][]byte{}
				for _, v := range a.([]interface{}) {
					to = append(to, v.([]byte))
				}

				return to
			},
		},
		{
			name:    "aliasMap -> map[string]any",
			val:     func() any { return map[string]any{} },
			alias:   func() any { return aliasMap{} },
			convert: func(a any) any { return map[string]any(a.(aliasMap)) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			v := tc.val()
			wv, err := Wrap(v)
			require.NoError(t, err)

			a := tc.alias()
			err = wv.UnwrapTo(&a)
			require.NoError(t, err)

			if tc.convert != nil {
				assert.Equal(t, tc.convert(a), v)
			} else {
				assert.Equal(t, a, v)
			}
		})
	}
}
