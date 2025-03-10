package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewWrapperModifier creates a modifier that will wrap specified on-chain fields in a struct.
// if key is not provided in the config, the whole value is wrapped with the name of the value from config.
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

func (m *wrapperModifier) RetypeToOffChain(onChainType reflect.Type, _ string) (tpe reflect.Type, err error) {
	defer func() {
		// StructOf can panic if the fields are not valid
		if r := recover(); r != nil {
			tpe = nil
			err = fmt.Errorf("%w: %v", types.ErrInvalidType, r)
		}
	}()

	// custom handling for wrapping primitive value or the whole value
	if m.isWholeValueWrapper() {
		for _, v := range m.fields {
			offChainTyp := reflect.StructOf([]reflect.StructField{{
				Name: v,
				Type: onChainType,
			}})

			m.onToOffChainType[onChainType] = offChainTyp
			m.offToOnChainType[offChainTyp] = onChainType
			return offChainTyp, nil
		}
	}

	return m.modifierBase.RetypeToOffChain(onChainType, "")
}

func (m *wrapperModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	offChainValue, itemType, err := m.modifierBase.selectType(offChainValue, m.offChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	// check if the offChainValue is a wrapper around the whole value
	typ := reflect.TypeOf(offChainValue)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() == reflect.Struct && typ.NumField() == 1 {
		if m.isWholeValueWrapper() {
			return reflect.ValueOf(offChainValue).Field(0).Interface(), nil
		}
	}

	modified, err := transformWithMaps(offChainValue, m.offToOnChainType, m.fields, unwrapFieldMapAction)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

func (m *wrapperModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	onChainValue, itemType, err := m.modifierBase.selectType(onChainValue, m.onChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	if m.isWholeValueWrapper() {
		var name string
		for _, v := range m.fields {
			var ok bool
			name, ok = any(v).(string)
			if !ok {
				return nil, fmt.Errorf("%q invalid type for m.fields value expected string got : %q", types.ErrInternal, reflect.TypeOf(v).String())
			}
		}

		s := reflect.StructOf([]reflect.StructField{{Name: name, Type: reflect.TypeOf(onChainValue)}})
		instance := reflect.New(s).Elem()
		fieldValue := instance.FieldByName(name)
		fieldValue.Set(reflect.ValueOf(onChainValue))

		return instance.Interface(), nil
	}

	modified, err := transformWithMaps(onChainValue, m.onToOffChainType, m.fields, wrapFieldMapAction)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

// isWholeValueWrapper we know that we are wrapping the whole value when the original field name is empty.
func (m *wrapperModifier) isWholeValueWrapper() bool {
	if len(m.fields) == 1 {
		for k := range m.fields {
			return k == ""
		}
	}
	return false
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

	switch v := typesMap[fieldName].(type) {
	case map[string]any:
		val, isOk := v[wrappedFieldName]
		if !isOk {
			return fmt.Errorf("field %s.%s does not exist", fieldName, wrappedFieldName)
		}
		typesMap[fieldName] = val
	default:
		return fmt.Errorf("field %s is not a map", fieldName)
	}

	return nil
}
