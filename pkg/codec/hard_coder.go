package codec

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewHardCoder(onChain map[string]any, offChain map[string]any) Modifier {
	m := &onChainHardCoder{
		modifierBase: modifierBase[any]{
			fields:           offChain,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
		onChain: onChain,
	}
	m.modifyFieldForInput = func(field *reflect.StructField, key string, v any) error {
		// if we are typing it differently, we need to make sure it's hard-coded the other way
		newType := reflect.TypeOf(v)
		if _, ok := m.onChain[key]; !ok && field.Type != newType {
			return fmt.Errorf(
				"%w: cannot change field type without hard-coding its onchain value for key %s",
				types.ErrInvalidType,
				key)
		}
		field.Type = newType
		return nil
	}
	m.addFieldForInput = func(key string, value any) reflect.StructField {
		return reflect.StructField{
			Name: key,
			Type: reflect.TypeOf(value),
		}
	}
	return m
}

type onChainHardCoder struct {
	modifierBase[any]
	onChain map[string]any
}

func (o *onChainHardCoder) TransformForOnChain(offChainValue any) (any, error) {
	return transformWithMaps(offChainValue, o.offToOnChainType, o.onChain, hardCode, hardCodeManyHook)
}

func (o *onChainHardCoder) TransformForOffChain(onChainValue any) (any, error) {
	return transformWithMaps(onChainValue, o.onToOffChainType, o.fields, hardCode, hardCodeManyHook)
}

func hardCode(extractMap map[string]any, key string, item any) error {
	if val, ok := extractMap[key]; ok {
		if m, ok := val.(map[string]any); ok && len(m) != 0 {
			fmt.Println("HERE WE ARE")
		} else if ms, ok := val.([]map[string]any); ok {
			fmt.Println(ms)
		} else if reflect.Zero(reflect.TypeOf(val)).Interface() != val {
			fmt.Println("HERE WE ARE")
		}
	}
	extractMap[key] = item
	return nil
}

func hardCodeManyHook(from reflect.Value, to reflect.Value) (any, error) {
	switch to.Kind() {
	case reflect.Slice, reflect.Array:
		switch from.Kind() {
		case reflect.Slice, reflect.Array:
			return from.Interface(), nil
		}
	default:
		return from.Interface(), nil
	}

	length := to.Len()
	array := reflect.MakeSlice(reflect.SliceOf(from.Type()), length, length)
	for i := 0; i < length; i++ {
		array.Index(i).Set(from)
	}
	return array.Interface(), nil
}

func flattenMap(m map[string]any) error {
	for k, v := range m {
		iv := reflect.ValueOf(v)
		for ; iv.Kind() == reflect.Pointer; iv = iv.Elem() {
		}
		switch iv.Kind() {
		case reflect.Map:
			structMap, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: cannot flatten map with key %s", types.ErrInvalidType, k)
			}
			if err := flattenMap(structMap); err != nil {
				return err
			}
		case reflect.Struct:
			var innerMap map[string]any
			if err := mapstructure.Decode(v, &innerMap); err != nil {
				return fmt.Errorf("%w: cannot flatten map with key %s: %w", types.ErrInvalidType, k, err)
			}
			m[k] = innerMap
		case reflect.Array, reflect.Slice:
			skipTypes := map[reflect.Kind]bool{
				reflect.Array:   true,
				reflect.Slice:   true,
				reflect.Pointer: true,
			}
			tmp := iv.Type()
			for ; skipTypes[tmp.Kind()]; tmp = tmp.Elem() {
			}
			if tmp.Kind() != reflect.Struct {
				return fmt.Errorf("%w: cannot flatten map with key %s", types.ErrInvalidType, k)
			}
			length := iv.Len()
			results := make([]map[string]any, length)
			for i := 0; i < length; i++ {
				if err := mapstructure.Decode(iv.Index(i).Interface(), &results[i]); err != nil {
					return fmt.Errorf("%w: cannot flatten map with key %s: %w", types.ErrInvalidType, k, err)
				}
				if err := flattenMap(results[i]); err != nil {
					return err
				}
			}
			m[k] = results
		default:
			return fmt.Errorf("%w: cannot flatten map with key %s", types.ErrInvalidType, k)
		}
	}
	return nil
}

func flattenObject(item any) (map[string]any, error) {
	itemT := reflect.TypeOf(item)
	for ; itemT.Kind() == reflect.Ptr; itemT = itemT.Elem() {
	}

	switch itemT.Kind() {

	}

	m := map[string]any{}
	if err := mapstructure.Decode(item, &m); err != nil {
		return nil, fmt.Errorf("%w: cannot flatten object: %w", types.ErrInvalidType, err)
	}
	if err := flattenMap(m); err != nil {
		return nil, err
	}
	return m, nil
}
