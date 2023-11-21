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
		return reflect.NewAt(toType.Elem(), rInput.UnsafePointer()), nil
	case reflect.Struct, reflect.Slice, reflect.Array:
		return transformNonPointer(toType, rInput)
	}

	return reflect.Value{}, types.ErrInvalidType
}

func transformNonPointer(toType reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	// make sure the input is addressable
	ptr := reflect.New(rInput.Type())
	reflect.Indirect(ptr).Set(rInput)
	changed := reflect.NewAt(toType, ptr.UnsafePointer()).Elem()
	return changed, nil
}
