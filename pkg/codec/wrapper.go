package codec

import (
	"fmt"
	"reflect"
)

func NewWrapperModifier(fields map[string]string) Modifier {
	m := &wrapperModifier{
		modifierBase: modifierBase[string]{
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
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

func (t *wrapperModifier) TransformToOnChain(offChainValue any, _ string) (any, error) {
	return transformWithMaps(offChainValue, t.offToOnChainType, t.fields, unwrapFieldMapAction)
}

func (t *wrapperModifier) TransformToOffChain(onChainValue any, _ string) (any, error) {
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields, wrapFieldMapAction)
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
