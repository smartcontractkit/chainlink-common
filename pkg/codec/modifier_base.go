package codec

import (
	"reflect"

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
		return m.getRenameType(outputType)
	}

	return nil, types.ErrInvalidType
}

func (m *modifierBase[T]) getRenameType(outputType reflect.Type) (reflect.Type, error) {
	numFields := outputType.NumField()
	allFields := make([]reflect.StructField, numFields)
	numModified := 0
	for i := 0; i < numFields; i++ {
		tmp := outputType.Field(i)
		if change, ok := m.Fields[tmp.Name]; ok {
			numModified++
			m.modifyFieldForInput(&tmp, change)
		}
		allFields[i] = tmp
	}

	if numModified != len(m.Fields) {
		return nil, types.ErrInvalidType
	}

	newStruct := reflect.StructOf(allFields)
	m.outputToInputType[outputType] = newStruct
	m.inputToOutputType[newStruct] = outputType
	return newStruct, nil
}
