package codec

import (
	"reflect"
	"sort"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type modifierBase[T any] struct {
	Fields              map[string]T
	onToOffChainType    map[reflect.Type]reflect.Type
	offToOneChainType   map[reflect.Type]reflect.Type
	modifyFieldForInput func(outputField *reflect.StructField, change T)
}

func (m *modifierBase[T]) RetypeForOffChain(onChainType reflect.Type) (reflect.Type, error) {
	if m.Fields == nil || len(m.Fields) == 0 {
		m.offToOneChainType[onChainType] = onChainType
		m.onToOffChainType[onChainType] = onChainType
		return onChainType, nil
	}

	if cached, ok := m.onToOffChainType[onChainType]; ok {
		return cached, nil
	}

	switch onChainType.Kind() {
	case reflect.Pointer:
		if elm, err := m.RetypeForOffChain(onChainType.Elem()); err == nil {
			ptr := reflect.PointerTo(elm)
			m.onToOffChainType[onChainType] = ptr
			m.offToOneChainType[ptr] = onChainType
			return ptr, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Slice:
		if elm, err := m.RetypeForOffChain(onChainType.Elem()); err == nil {
			sliceType := reflect.SliceOf(elm)
			m.onToOffChainType[onChainType] = sliceType
			m.offToOneChainType[sliceType] = onChainType
			return sliceType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Array:
		if elm, err := m.RetypeForOffChain(onChainType.Elem()); err == nil {
			arrayType := reflect.ArrayOf(onChainType.Len(), elm)
			m.onToOffChainType[onChainType] = arrayType
			m.offToOneChainType[arrayType] = onChainType
			return arrayType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Struct:
		return m.getStructType(onChainType)
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
	m.onToOffChainType[outputType] = newStruct
	m.offToOneChainType[newStruct] = outputType
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
