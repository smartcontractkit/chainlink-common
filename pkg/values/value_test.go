package values

import (
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

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
			name: "time",
			newValue: func() (any, Value, error) {
				t, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				tv := NewTime(t)
				return t, tv, err
			},
		},
		{
			name: "float64",
			newValue: func() (any, Value, error) {
				f := 1.0
				fv := NewFloat64(f)
				return f, fv, nil
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

func Test_WrapPointers(t *testing.T) {
	underlying := "foo"
	actual, err := Wrap(&underlying)
	require.NoError(t, err)

	expected, err := Wrap("foo")
	require.NoError(t, err)

	assert.True(t, reflect.DeepEqual(expected, actual))
}

func Test_IntTypes(t *testing.T) {
	anyValue := int64(100)
	testCases := []struct {
		name string
		test func(tt *testing.T)
	}{
		{name: "int32", test: func(tt *testing.T) { wrappableTest[int64, int32](tt, anyValue) }},
		{name: "int16", test: func(tt *testing.T) { wrappableTest[int64, int16](tt, anyValue) }},
		{name: "int8", test: func(tt *testing.T) { wrappableTest[int64, int8](tt, anyValue) }},
		{name: "int", test: func(tt *testing.T) { wrappableTest[int64, int](tt, anyValue) }},
		{name: "uint64", test: func(tt *testing.T) { wrappableTest[int64, uint64](tt, anyValue) }},
		{name: "uint32", test: func(tt *testing.T) { wrappableTest[int64, uint32](tt, anyValue) }},
		{name: "uint16", test: func(tt *testing.T) { wrappableTest[int64, uint16](tt, anyValue) }},
		{name: "uint8", test: func(tt *testing.T) { wrappableTest[int64, uint8](tt, anyValue) }},
		{name: "uint", test: func(tt *testing.T) { wrappableTest[int64, uint](tt, anyValue) }},
		{name: "uint64 small enough for int64", test: func(tt *testing.T) {
			u64, err := Wrap(uint64(math.MaxInt64))
			require.NoError(tt, err)

			expected, err := Wrap(int64(math.MaxInt64))
			require.NoError(tt, err)

			assert.Equal(tt, expected, u64)

			unwrapped := uint64(0)
			err = u64.UnwrapTo(&unwrapped)
			require.NoError(tt, err)
			assert.Equal(tt, uint64(math.MaxInt64), unwrapped)
		}},
		{name: "uint64 too large for int64", test: func(tt *testing.T) {
			u64, err := Wrap(uint64(math.MaxInt64 + 1))
			require.NoError(tt, err)

			expected, err := Wrap(new(big.Int).SetUint64(math.MaxInt64 + 1))
			require.NoError(tt, err)

			assert.Equal(tt, expected, u64)

			unwrapped := uint64(0)
			err = u64.UnwrapTo(&unwrapped)
			require.NoError(tt, err)
			assert.Equal(tt, uint64(math.MaxInt64+1), unwrapped)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			tc.test(st)
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
			value: NewTime(time.Time{}),
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
		{
			value: (*Time)(nil),
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

type aliasBytes []byte
type aliasString string
type aliasInt int
type aliasMap map[string]any
type aliasByte uint8
type decimalAlias decimal.Decimal
type bigIntAlias big.Int
type bigIntPtrAlias *big.Int
type aliasUint64 uint64

func Test_Aliases(t *testing.T) {
	testCases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "[]byte alias",
			test: func(t *testing.T) { wrappableTest[[]byte, aliasBytes](t, []byte("hello")) },
		},
		{
			name: "byte alias in slice",
			test: func(tt *testing.T) {
				wrappableTestWithConversion[[]byte, []aliasByte](tt, []byte("hello"), func(native []aliasByte) []byte {
					converted := make([]byte, len(native))
					for i, b := range native {
						converted[i] = byte(b)
					}
					return converted
				})
			},
		},
		{
			name: "basic alias",
			test: func(tt *testing.T) { wrappableTest[string, aliasString](tt, "hello") },
		},
		{
			name: "integer",
			test: func(tt *testing.T) { wrappableTest[int, aliasInt](tt, 1) },
		},
		{
			name: "uint64 fits in int64",
			test: func(tt *testing.T) { wrappableTest[uint64, aliasUint64](tt, uint64(math.MaxInt64)) },
		},
		{
			name: "uint64 too large for int64",
			test: func(tt *testing.T) { wrappableTest[uint64, aliasUint64](tt, uint64(math.MaxInt64+1)) },
		},
		{
			name: "map",
			test: func(tt *testing.T) { wrappableTest[map[string]any, aliasMap](tt, map[string]any{"hello": "world"}) },
		},
		{
			name: "decimal alias",
			test: func(tt *testing.T) { wrappableTest[decimal.Decimal, decimalAlias](tt, decimal.NewFromFloat(1.0)) },
		},
		{
			name: "big int alias",
			test: func(tt *testing.T) {
				testBigIntType[*bigIntAlias](tt, big.NewInt(1), func() *bigIntAlias {
					return new(bigIntAlias)
				})
			},
		},
		{
			name: "big int pointer alias",
			test: func(tt *testing.T) {
				testBigIntType[bigIntPtrAlias](tt, big.NewInt(1), func() bigIntPtrAlias {
					return new(big.Int)
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			tc.test(st)
		})
	}
}

func wrappableTest[Native, Alias any](t *testing.T, native Native) {
	wrappableTestWithConversion(t, native, func(alias Alias) Native {
		return reflect.ValueOf(alias).Convert(reflect.TypeOf(native)).Interface().(Native)
	})
}

func testBigIntType[Alias any](t *testing.T, native *big.Int, create func() Alias) {
	wv, err := Wrap(native)
	require.NoError(t, err)

	a := create()

	err = wv.UnwrapTo(a)
	require.NoError(t, err)

	assert.Equal(t, native, reflect.ValueOf(a).Convert(reflect.TypeOf(native)).Interface())

	aliasWrapped, err := Wrap(a)
	require.NoError(t, err)
	assert.Equal(t, wv, aliasWrapped)
}

func wrappableTestWithConversion[Native, Alias any](t *testing.T, native Native, convert func(native Alias) Native) {
	wv, err := Wrap(native)
	require.NoError(t, err)

	var a Alias
	err = wv.UnwrapTo(&a)
	require.NoError(t, err)

	assert.Equal(t, native, convert(a))

	aliasWrapped, err := Wrap(a)
	require.NoError(t, err)
	assert.Equal(t, wv, aliasWrapped)
}
