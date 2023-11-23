package codec

import (
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewRenamer(fields map[string]string) Modifier {
	m := &renamer{
		modifierBase: modifierBase[string]{
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}
	m.modifyFieldForInput = func(field *reflect.StructField, _, newName string) error {
		field.Name = newName
		return nil
	}
	return m
}

type renamer struct {
	modifierBase[string]
}

func (r *renamer) TransformForOffChain(output any) (any, error) {
	rOutput, err := renameTransform(r.onToOffChainType, reflect.ValueOf(output))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
}

func (r *renamer) TransformForOnChain(input any) (any, error) {
	rOutput, err := renameTransform(r.offToOnChainType, reflect.ValueOf(input))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
}

func renameTransform(typeMap map[reflect.Type]reflect.Type, rInput reflect.Value) (reflect.Value, error) {
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
