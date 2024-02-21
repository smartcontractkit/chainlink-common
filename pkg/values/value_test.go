package values

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Value(t *testing.T) {
	testCases := []struct {
		name         string
		newValue     func() (any, Value, error)
		mustNewValue func() (any, Value)
		equal        func(t *testing.T, expected any, unwrapped any)
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
			mustNewValue: func() (any, Value) {
				m := map[string]any{
					"hello": "world",
				}
				mv := MustNewMap(m)
				return m, mv
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
			mustNewValue: func() (any, Value) {
				l := []any{
					1,
					"2",
					decimal.NewFromFloat(1.0),
				}
				lv := MustNewList(l)
				return l, lv
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
				decv, err := NewDecimal(dec)
				return dec, decv, err
			},
			mustNewValue: func() (any, Value) {
				dec := decimal.NewFromFloat(1.03)
				decv := MustNewDecimal(dec)
				return dec, decv
			},
		},
		{
			name: "string",
			newValue: func() (any, Value, error) {
				s := "hello"
				sv, err := NewString(s)
				return s, sv, err
			},
			mustNewValue: func() (any, Value) {
				s := "hello"
				sv := MustNewString(s)
				return s, sv
			},
		},
		{
			name: "bytes",
			newValue: func() (any, Value, error) {
				b := []byte("hello")
				bv, err := NewBytes(b)
				return b, bv, err
			},
			mustNewValue: func() (any, Value) {
				b := []byte("hello")
				bv := MustNewBytes(b)
				return b, bv
			},
		},
		{
			name: "bool",
			newValue: func() (any, Value, error) {
				b := true
				bv, err := NewBool(b)
				return b, bv, err
			},
			mustNewValue: func() (any, Value) {
				b := true
				bv := MustNewBool(b)
				return b, bv
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
			mustNewValue: func() (any, Value) {
				m := map[string]any{
					"hello": map[string]any{
						"world": "foo",
					},
					"baz": []any{
						int64(1), int64(2), int64(3),
					},
				}
				mv := MustNewMap(m)
				return m, mv
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			originalValue, wrapped, err := tc.newValue()
			require.NoError(t, err)

			assert.NotPanics(t, func() {
				tc.mustNewValue()
			})

			pb, err := wrapped.Proto()
			require.NoError(t, err)

			rehydratedValue, err := FromProto(pb)
			require.NoError(t, err)
			assert.Equal(t, wrapped, rehydratedValue)

			unwrapped, err := rehydratedValue.Unwrap()
			require.NoError(t, err)
			if tc.equal != nil {
				tc.equal(t, originalValue, unwrapped)
			} else {
				assert.Equal(t, originalValue, unwrapped)
			}
		})
	}
}
