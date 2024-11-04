package values

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type Map struct {
	Underlying map[string]Value
}

// TODO: this is temporary until the gateway can correctly unmarshal Web API trigger requests.
// func (m Map) MarshalJSON() ([]byte, error) {
// 	tempMap := make(map[string]interface{})
// 	var err error
// 	for k, v := range m.Underlying {
// 		tempMap[k], err = v.Unwrap()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return json.Marshal(tempMap)
// }

// func (m *Map) UnmarshalJSON(data []byte) error {
// 	tempMap := make(map[string]interface{})
// 	if err := json.Unmarshal(data, &tempMap); err != nil {
// 		return err
// 	}

// 	m.Underlying = make(map[string]Value)
// 	for k, v := range tempMap {
// 		switch tv := v.(type) {
// 		case int64:
// 			m.Underlying[k] = NewInt64(tv)
// 		case string:
// 			m.Underlying[k] = NewString(tv)
// 		default:
// 			m.Underlying[k] = nil
// 		}
// 	}

// 	return nil
// }

func EmptyMap() *Map {
	return &Map{
		Underlying: map[string]Value{},
	}
}

func NewMap[T any](m map[string]T) (*Map, error) {
	mv := map[string]Value{}
	for k, v := range m {
		val, err := Wrap(v)
		if err != nil {
			return nil, err
		}

		mv[k] = val
	}

	return &Map{
		Underlying: mv,
	}, nil
}

func (m *Map) proto() *pb.Value {
	if m == nil {
		return pb.NewMapValue(map[string]*pb.Value{})
	}

	pm := map[string]*pb.Value{}
	for k, v := range m.Underlying {
		pm[k] = Proto(v)
	}

	return pb.NewMapValue(pm)
}

func (m *Map) Unwrap() (any, error) {
	nm := map[string]any{}
	return nm, m.UnwrapTo(&nm)
}

func (m *Map) copy() Value {
	return m.CopyMap()
}

func (m *Map) CopyMap() *Map {
	if m == nil {
		return nil
	}

	dest := map[string]Value{}
	for k, v := range m.Underlying {
		dest[k] = Copy(v)
	}

	return &Map{Underlying: dest}
}

func (m *Map) UnwrapTo(to any) error {
	if m == nil {
		return errors.New("cannot unwrap nil values.Map")
	}

	c := &mapstructure.DecoderConfig{
		Result: to,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapValueToMap,
			unwrapsValues,
		),
	}

	d, err := mapstructure.NewDecoder(c)
	if err != nil {
		return err
	}

	return d.Decode(m.Underlying)
}

// DeleteAtPath deletes a value from a map at a given dot separated path.  Returns true if an element at the given
// path was found and deleted, false otherwise.
func (m *Map) DeleteAtPath(path string) bool {
	pathSegments := strings.Split(path, ".")
	underlying := m.Underlying
	for segmentIdx, pathSegment := range pathSegments {
		if segmentIdx == len(pathSegments)-1 {
			_, ok := underlying[pathSegment]
			if !ok {
				return false
			}

			delete(underlying, pathSegment)
			return true
		}

		value := underlying[pathSegment]
		mv, ok := value.(*Map)
		if !ok {
			return false
		}
		underlying = mv.Underlying
	}

	return false
}

func mapValueToMap(f reflect.Type, t reflect.Type, data any) (any, error) {
	if f != reflect.TypeOf(map[string]Value{}) {
		return data, nil
	}
	switch t {
	// If the destination type is `map[string]any` or `any`,
	// fully unwrap the values.Map.
	// We have to handle the `any` case here as otherwise UnwrapTo won't work on
	// maps recursively
	case reflect.TypeOf(map[string]any{}), reflect.TypeOf((*any)(nil)).Elem():
		dv := data.(map[string]Value)
		d := map[string]any{}
		for k, v := range dv {
			unw, err := Unwrap(v)
			if err != nil {
				return nil, err
			}

			d[k] = unw
		}

		return d, nil
	}

	return data, nil
}

func baseTypesEqual(f reflect.Type, t reflect.Type) bool {
	if f.Kind() == reflect.Pointer {
		f = f.Elem()
	}

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return f == t
}

// unwrapsValues takes a value of type `f` and tries to convert it to a value of type `t`
func unwrapsValues(f reflect.Type, t reflect.Type, data any) (any, error) {
	// First, check if t and f have the same base type. If they do,
	// we don't need to do anything further. `mapstructure` will
	// automatically convert between values and pointers to get the right result.
	if baseTypesEqual(f, t) {
		return data, nil
	}

	valueType := reflect.TypeOf((*Value)(nil)).Elem()

	// Next, if f is a `Value`, we'll try to transform it to `t`,
	// but only if `t` is not itself a `Value`.
	// This avoids the following cases which we handle differently:
	// - f and t are the same concrete value type -- handled above.
	// - data is a concrete value and t represents the `Value` interface type.
	//   This is compatible and we'll handle it on line 137 by returning `data`.
	// - f and t are different concrete value types -- we can't handle that
	//   here, so we'll just return data untransformed.
	// In all other cases, we want to rely on data's UnwrapTo implementation
	// to try to get the right result.
	if f.Implements(valueType) && !t.Implements(valueType) {
		dv := data.(Value)

		n := reflect.New(t).Interface()
		err := dv.UnwrapTo(n)
		if err != nil {
			// Do not return the error to allow mapstructure to retry with different types
			// Eg: mapstructure will attempt **big.Int before *big.Int if the field is a *big.Int.
			return data, nil
		}

		if reflect.TypeOf(n).Elem() == t {
			return n, nil
		}
	}

	return data, nil
}
