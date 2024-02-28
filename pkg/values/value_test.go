package values

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestValueEvent struct {
	TriggerType string       `json:"triggerType"`
	Id          string       `json:"id"`
	Timestamp   string       `json:"timestamp"`
	Payload     []TestReport `json:"payload"`
}

type TestReport struct {
	FeedId     int64  `json:"feedId"`
	Fullreport string `json:"fullreport"`
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
				decv, err := NewDecimal(dec)
				return dec, decv, err
			},
		},
		{
			name: "string",
			newValue: func() (any, Value, error) {
				s := "hello"
				sv, err := NewString(s)
				return s, sv, err
			},
		},
		{
			name: "bytes",
			newValue: func() (any, Value, error) {
				b := []byte("hello")
				bv, err := NewBytes(b)
				return b, bv, err
			},
		},
		{
			name: "bool",
			newValue: func() (any, Value, error) {
				b := true
				bv, err := NewBool(b)
				return b, bv, err
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
				v := TestReport{
					FeedId:     2,
					Fullreport: "hello",
				}
				m := map[string]any{
					"feedId":     int64(2),
					"fullreport": "hello",
				}
				vv, err := Wrap(v)
				return m, vv, err
			},
		},
		{
			name: "structPointer",
			newValue: func() (any, Value, error) {
				v := &TestReport{
					FeedId:     2,
					Fullreport: "hello",
				}
				m := map[string]any{
					"feedId":     int64(2),
					"fullreport": "hello",
				}
				vv, err := Wrap(v)
				return m, vv, err
			},
		},
		{
			name: "nestedStruct",
			newValue: func() (any, Value, error) {
				v := TestValueEvent{
					TriggerType: "mercury",
					Id:          "123",
					Timestamp:   "123",
					Payload: []TestReport{
						{
							FeedId:     2,
							Fullreport: "hello",
						},
						{
							FeedId:     3,
							Fullreport: "world",
						},
					},
				}
				m := map[string]any{
					"triggerType": "mercury",
					"id":          "123",
					"timestamp":   "123",
					"payload": []any{
						map[string]any{
							"feedId":     int64(2),
							"fullreport": "hello",
						},
						map[string]any{
							"feedId":     int64(3),
							"fullreport": "world",
						},
					},
				}
				vv, err := Wrap(v)
				return m, vv, err
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
						"nil":    &Nil{},
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
			require.NoError(t, err)

			pb, err := wrapped.Proto()
			require.NoError(t, err)

			rehydratedValue, err := FromProto(pb)
			require.NoError(t, err)
			assert.Equal(t, wrapped, rehydratedValue)

			unwrapped, err := Unwrap(rehydratedValue)
			require.NoError(t, err)
			if tc.equal != nil {
				tc.equal(t, originalValue, unwrapped)
			} else {
				assert.Equal(t, originalValue, unwrapped)
			}
		})
	}
}
