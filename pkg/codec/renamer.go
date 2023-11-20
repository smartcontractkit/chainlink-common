package codec

import (
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type Renamer struct {
	Fields       map[string]string
	typeCache    map[reflect.Type]reflect.Type
	reverseCache map[reflect.Type]reflect.Type
}

func (r *Renamer) AdjustForInput(outputType reflect.Type) (reflect.Type, error) {
	if r.Fields == nil || len(r.Fields) == 0 {
		return outputType, nil
	}

	if r.typeCache == nil {
		r.typeCache = map[reflect.Type]reflect.Type{}
		r.reverseCache = map[reflect.Type]reflect.Type{}
	}

	if cached, ok := r.typeCache[outputType]; ok {
		return cached, nil
	}

	switch outputType.Kind() {
	case reflect.Pointer:
		if elm, err := r.AdjustForInput(outputType.Elem()); err == nil {
			return reflect.PointerTo(elm), nil
		}
		return nil, types.ErrInvalidType
	case reflect.Slice:
		if elm, err := r.AdjustForInput(outputType.Elem()); err == nil {
			return reflect.SliceOf(elm), nil
		}
		return nil, types.ErrInvalidType
	case reflect.Array:
		if elm, err := r.AdjustForInput(outputType.Elem()); err == nil {
			return reflect.ArrayOf(outputType.Len(), elm), nil
		}
		return nil, types.ErrInvalidType
	case reflect.Struct:
		return r.getRenameType(outputType)
	}

	return nil, types.ErrInvalidType
}

func (r *Renamer) TransformInput(input any) (any, error) {
	rInput := reflect.ValueOf(input)
	switch rInput.Kind() {
	case reflect.Pointer:
		adjusted, ok := r.reverseCache[rInput.Elem().Type()]
		if !ok {
			return nil, types.ErrInvalidType
		}
		changed := reflect.NewAt(adjusted, rInput.UnsafePointer()).Interface()
		return changed, nil
	case reflect.Struct:
		// make sure the input is addressable
		adjusted, err := r.AdjustForInput(rInput.Type())
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(rInput.Type())
		reflect.Indirect(ptr).Set(rInput)
		changed := reflect.NewAt(adjusted, ptr.UnsafePointer()).Interface()
		return changed, nil
	case reflect.Slice:
	case reflect.Array:
	}

	return nil, types.ErrInvalidType
}

func (r *Renamer) getRenameType(outputType reflect.Type) (reflect.Type, error) {
	numFields := outputType.NumField()
	allFields := make([]reflect.StructField, numFields)
	numRenamed := 0
	for i := 0; i < numFields; i++ {
		tmp := outputType.Field(i)
		if newName, ok := r.Fields[tmp.Name]; ok {
			numRenamed++
			tmp.Name = newName
		}
		allFields[i] = tmp
	}

	if numRenamed != len(r.Fields) {
		return nil, types.ErrInvalidType
	}

	newStruct := reflect.StructOf(allFields)
	r.typeCache[outputType] = newStruct
	r.reverseCache[newStruct] = outputType
	return newStruct, nil
}

var _ Modifier = &Renamer{}
