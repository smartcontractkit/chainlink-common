package codec

import (
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewRenamer(fields map[string]string) Modifier {
	m := &renamer{
		modifierBase: modifierBase[string]{
			Fields:            fields,
			outputToInputType: map[reflect.Type]reflect.Type{},
			inputToOutputType: map[reflect.Type]reflect.Type{},
		},
	}
	m.modifyFieldForInput = func(field *reflect.StructField, newName string) { field.Name = newName }
	return m
}

type renamer struct {
	modifierBase[string]
}

func (r *renamer) TransformOutput(output any) (any, error) {
	rOutput, err := transform(r.outputToInputType, reflect.ValueOf(output))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
}

func (r *renamer) TransformInput(input any) (any, error) {
	rOutput, err := transform(r.inputToOutputType, reflect.ValueOf(input))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
}

func transform(typeMap map[reflect.Type]reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	toType, ok := typeMap[rInput.Type()]
	if !ok {
		return reflect.Value{}, types.ErrInvalidType
	}

	if toType == rInput.Type() {
		return rInput, nil
	}

	switch rInput.Kind() {
	case reflect.Pointer:
		return transformPointer(toType, rInput)
	case reflect.Struct:
		return transformStruct(toType, rInput)
	case reflect.Slice:
		return transformSlice(toType, typeMap, rInput)
	case reflect.Array:
		return transformArray(toType, typeMap, rInput)
	}

	return reflect.Value{}, types.ErrInvalidType
}

func transformSlice(toType reflect.Type, typeMap map[reflect.Type]reflect.Type, input reflect.Value) (reflect.Value, error) {
	length := input.Len()
	converted := reflect.MakeSlice(toType, length, length)
	return transformSliceOrArray(typeMap, input, converted)
}

func transformArray(
	toType reflect.Type, typeMap map[reflect.Type]reflect.Type, input reflect.Value) (reflect.Value, error) {
	converted := reflect.New(toType).Elem()
	return transformSliceOrArray(typeMap, input, converted)
}

func transformSliceOrArray(
	typeMap map[reflect.Type]reflect.Type, input reflect.Value, converted reflect.Value) (reflect.Value, error) {
	length := input.Len()
	for i := 0; i < length; i++ {
		elm, err := transform(typeMap, input.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		converted.Index(i).Set(elm)
	}
	return converted, nil
}

func transformPointer(toType reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	changed := reflect.NewAt(toType.Elem(), rInput.UnsafePointer())
	return changed, nil
}

func transformStruct(toType reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	// make sure the input is addressable
	ptr := reflect.New(rInput.Type())
	reflect.Indirect(ptr).Set(rInput)
	changed := reflect.NewAt(toType, ptr.UnsafePointer()).Elem()
	return changed, nil
}
