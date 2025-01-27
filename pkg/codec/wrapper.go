package codec

import (
	"fmt"
	"reflect"
)

// NewWrapperModifier creates a modifier that will wrap specified on-chain fields in a struct.
func NewWrapperModifier(fields map[string]string) Modifier {
	return NewPathTraverseWrapperModifier(fields, false)
}

func NewPathTraverseWrapperModifier(fields map[string]string, enablePathTraverse bool) Modifier {
	m := &wrapperModifier{
		modifierBase: modifierBase[string]{
			enablePathTraverse: enablePathTraverse,
			fields:             fields,
			onToOffChainType:   map[reflect.Type]reflect.Type{},
			offToOnChainType:   map[reflect.Type]reflect.Type{},
		},
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, fieldName string) error {
		field.Type = reflect.StructOf([]reflect.StructField{{
			Name: fieldName,
			Type: field.Type,
		}})
		return nil
	}

	return m
}

type wrapperModifier struct {
	modifierBase[string]
}

func (m *wrapperModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	offChainValue, itemType, err := m.modifierBase.selectType(offChainValue, m.offChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := transformWithMaps(offChainValue, m.offToOnChainType, m.fields, unwrapFieldMapAction)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return valueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

func (m *wrapperModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	onChainValue, itemType, err := m.modifierBase.selectType(onChainValue, m.onChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := transformWithMaps(onChainValue, m.onToOffChainType, m.fields, wrapFieldMapAction)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return valueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

func wrapFieldMapAction(typesMap map[string]any, fieldName string, wrappedFieldName string) error {
	field, exists := typesMap[fieldName]
	if !exists {
		return fmt.Errorf("field %s does not exist", fieldName)
	}

	typesMap[fieldName] = map[string]any{wrappedFieldName: field}
	return nil
}

func unwrapFieldMapAction(typesMap map[string]any, fieldName string, wrappedFieldName string) error {
	_, exists := typesMap[fieldName]
	if !exists {
		return fmt.Errorf("field %s does not exist", fieldName)
	}
	val, isOk := typesMap[fieldName].(map[string]any)[wrappedFieldName]
	if !isOk {
		return fmt.Errorf("field %s.%s does not exist", fieldName, wrappedFieldName)
	}

	typesMap[fieldName] = val
	return nil
}
