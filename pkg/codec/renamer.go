package codec

import (
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type Renamer struct {
	Fields            map[string]string
	outputToInputType map[reflect.Type]reflect.Type
	inputToOutputType map[reflect.Type]reflect.Type
}

func (r *Renamer) RetypeForInput(outputType reflect.Type) (reflect.Type, error) {
	if r.Fields == nil || len(r.Fields) == 0 {
		return outputType, nil
	}

	if r.outputToInputType == nil {
		r.outputToInputType = map[reflect.Type]reflect.Type{}
		r.inputToOutputType = map[reflect.Type]reflect.Type{}
	}

	if cached, ok := r.outputToInputType[outputType]; ok {
		return cached, nil
	}

	switch outputType.Kind() {
	case reflect.Pointer:
		if elm, err := r.RetypeForInput(outputType.Elem()); err == nil {
			ptr := reflect.PointerTo(elm)
			r.outputToInputType[outputType] = ptr
			r.inputToOutputType[ptr] = outputType
			return ptr, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Slice:
		if elm, err := r.RetypeForInput(outputType.Elem()); err == nil {
			sliceType := reflect.SliceOf(elm)
			r.outputToInputType[outputType] = sliceType
			r.inputToOutputType[sliceType] = outputType
			return sliceType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Array:
		if elm, err := r.RetypeForInput(outputType.Elem()); err == nil {
			arrayType := reflect.ArrayOf(outputType.Len(), elm)
			r.outputToInputType[outputType] = arrayType
			r.inputToOutputType[arrayType] = outputType
			return arrayType, nil
		}
		return nil, types.ErrInvalidType
	case reflect.Struct:
		return r.getRenameType(outputType)
	}

	return nil, types.ErrInvalidType
}

func (r *Renamer) TransformOutput(output any) (any, error) {
	rOutput, err := transform(r.outputToInputType, reflect.ValueOf(output))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
}

func (r *Renamer) TransformInput(input any) (any, error) {
	rOutput, err := transform(r.inputToOutputType, reflect.ValueOf(input))
	if err != nil {
		return nil, err
	}
	return rOutput.Interface(), nil
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
	r.outputToInputType[outputType] = newStruct
	r.inputToOutputType[newStruct] = outputType
	return newStruct, nil
}

func transform(typeMap map[reflect.Type]reflect.Type, rInput reflect.Value) (reflect.Value, error) {
	toType, ok := typeMap[rInput.Type()]
	if !ok {
		return reflect.Value{}, types.ErrInvalidType
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

var _ Modifier = &Renamer{}
