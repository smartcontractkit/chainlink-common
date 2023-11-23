package codec

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type MapOrMaps struct {
	m  map[string]any
	ms []*MapOrMaps
}

func (m *MapOrMaps) ForEachNestedMap(fn func(map[string]any) error) error {
	if m.m == nil {
		for _, m := range m.ms {
			if err := m.ForEachNestedMap(fn); err != nil {
				return err
			}
		}
	} else {
		return fn(m.m)
	}
	return nil
}

func FlattenToMapOrMaps(input any) (*MapOrMaps, error) {
	if input == nil {
		return &MapOrMaps{}, nil
	}

	var lastPtr reflect.Value
	iInput := reflect.ValueOf(input)
	numPointers := 0
	for iInput.Kind() == reflect.Pointer {
		numPointers++
		lastPtr = iInput
		iInput = reflect.Indirect(iInput)
	}

	if numPointers > 1 {
		input = lastPtr.Interface()
	}

	switch iInput.Kind() {
	case reflect.Map:
		if iInput.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%w: %T", types.ErrInvalidType, input)
		}
		return flattenStruct(input)
	case reflect.Struct:
		return flattenStruct(input)
	case reflect.Array, reflect.Slice:
		result := &MapOrMaps{}
		length := iInput.Len()
		result.ms = make([]*MapOrMaps, length)
		for i := 0; i < length; i++ {
			m, err := FlattenToMapOrMaps(iInput.Index(i).Interface())
			if err != nil {
				return nil, err
			}

			result.ms[i] = m

		}
		return result, nil
	default:
		return nil, fmt.Errorf("%w: %T", types.ErrInvalidType, input)
	}
}

func flattenStruct(input any) (*MapOrMaps, error) {
	iInput := reflect.ValueOf(input)
	result := &MapOrMaps{m: map[string]any{}}
	if err := mapstructure.Decode(input, &result.m); err != nil {
		return nil, fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	for k, v := range result.m {
		iV := reflect.ValueOf(v)
		if iV.Kind() == reflect.Interface && !iV.IsZero() {
			iV = iV.Elem()
		}

		for iV.Kind() == reflect.Pointer {
			iV = reflect.Indirect(iInput)
		}

		switch iV.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			res, err := FlattenToMapOrMaps(v)
			if err != nil {
				return nil, err
			} else if res.m != nil {
				result.m[k] = res.m
			} else {
				result.m[k] = res.ms
			}
		}
	}
	return result, nil
}
