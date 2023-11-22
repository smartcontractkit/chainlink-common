package codec

import (
	"reflect"
	"sort"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type modifierBase[T any] struct {
	Fields              map[string]T
	outputToInputType   map[reflect.Type]reflect.Type
	inputToOutputType   map[reflect.Type]reflect.Type
	modifyFieldForInput func(outputField *reflect.StructField, change T)
}

func (m *modifierBase[T]) RetypeForInput(outputType reflect.Type) (reflect.Type, error) {
	if m.Fields == nil || len(m.Fields) == 0 {
		m.inputToOutputType[outputType] = outputType
		m.outputToInputType[outputType] = outputType
		return outputType, nil
	}

	if cached, ok := m.outputToInputType[outputType]; ok {
		return cached, nil
	}

	switch outputType.Kind() {
	case reflect.Pointer:
		if elm, err := m.RetypeForInput(outputType.Elem()); err == nil {
			ptr := reflect.PointerTo(elm)
			m.outputToInputType[outputType] = ptr
			m.inputToOutputType[ptr] = outputType
			return ptr, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Slice:
		if elm, err := m.RetypeForInput(outputType.Elem()); err == nil {
			sliceType := reflect.SliceOf(elm)
			m.outputToInputType[outputType] = sliceType
			m.inputToOutputType[sliceType] = outputType
			return sliceType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Array:
		if elm, err := m.RetypeForInput(outputType.Elem()); err == nil {
			arrayType := reflect.ArrayOf(outputType.Len(), elm)
			m.outputToInputType[outputType] = arrayType
			m.inputToOutputType[arrayType] = outputType
			return arrayType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Struct:
		return m.getStructType(outputType)
	}

	return nil, types.ErrInvalidType
}

func (m *modifierBase[T]) getStructType(outputType reflect.Type) (reflect.Type, error) {
	filedLocations, err := getFieldIndices(outputType)
	if err != nil {
		return nil, err
	}

	for _, key := range m.subkeysFirst() {
		parts := strings.Split(key, ".")
		fieldName := parts[len(parts)-1]
		parts = parts[:len(parts)-1]
		curLocations := filedLocations
		for _, part := range parts {
			var err error
			if curLocations, err = curLocations.populateSubFields(part); err != nil {
				return nil, err
			}
		}

		// If a subkey has been modified, update the underlying types first
		curLocations.updateTypeFromSubkeyMods(fieldName)
		field, err := curLocations.fieldByName(fieldName)
		if err != nil {
			return nil, err
		}

		m.modifyFieldForInput(field, m.Fields[key])
	}

	newStruct := filedLocations.makeNewType()
	m.outputToInputType[outputType] = newStruct
	m.inputToOutputType[newStruct] = outputType
	return newStruct, nil
}

// subkeysFirst returns a list of keys that will always have a sub-key before the key if both are present
func (m *modifierBase[T]) subkeysFirst() []string {
	orderedKeys := make([]string, 0, len(m.Fields))
	for k, _ := range m.Fields {
		orderedKeys = append(orderedKeys, k)
	}
	sort.Slice(orderedKeys, func(i, j int) bool {
		return orderedKeys[i] > orderedKeys[j]
	})
	return orderedKeys
}
