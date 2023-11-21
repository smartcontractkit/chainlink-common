package codec

import (
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ElementExtractorLocation string

const (
	FirstElementLocation  ElementExtractorLocation = "first element"
	MiddleElementLocation ElementExtractorLocation = "middle element"
	LastElementLocation   ElementExtractorLocation = "last element"
)

func NewElementExtractor(fields map[string]ElementExtractorLocation) Modifier {
	m := &elementExtractor{
		modifierBase: modifierBase[ElementExtractorLocation]{
			Fields:            fields,
			outputToInputType: map[reflect.Type]reflect.Type{},
			inputToOutputType: map[reflect.Type]reflect.Type{},
		},
	}
	m.modifyFieldForInput = func(field *reflect.StructField, _ ElementExtractorLocation) {
		field.Type = reflect.SliceOf(field.Type)
	}
	return m
}

type elementExtractor struct {
	modifierBase[ElementExtractorLocation]
}

func (e *elementExtractor) TransformInput(input any) (any, error) {
	rInput := reflect.ValueOf(input)

	toType, ok := e.inputToOutputType[rInput.Type()]
	if !ok {
		return reflect.Value{}, types.ErrInvalidType
	}

	switch rInput.Kind() {
	case reflect.Pointer:
		// This doesn't work for non-structs
		into := reflect.New(toType.Elem())
		err := e.extractElements(input, into.Interface())
		return into.Interface(), err
	case reflect.Struct:
		into := reflect.New(toType)
		err := e.extractElements(input, into.Interface())
		return into.Elem().Interface(), err
	case reflect.Slice, reflect.Array:
		return nil, nil
	}

	return reflect.Value{}, types.ErrInvalidType
}

func (e *elementExtractor) TransformOutput(output any) (any, error) {
	rOutput := reflect.ValueOf(output)

	toType, ok := e.outputToInputType[rOutput.Type()]
	if !ok {
		return reflect.Value{}, types.ErrInvalidType
	}

	switch rOutput.Kind() {
	case reflect.Pointer:
		// This doesn't work for non-structs
		into := reflect.New(toType.Elem())
		err := e.expandElements(output, into.Interface())
		return into.Interface(), err
	case reflect.Struct:
		into := reflect.New(toType)
		err := e.expandElements(output, into.Interface())
		return into.Elem().Interface(), err
	case reflect.Slice, reflect.Array:
		return nil, nil
	}

	return reflect.Value{}, types.ErrInvalidType
}

func (e *elementExtractor) extractElements(input, output any) error {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(input, &valueMapping); err != nil {
		return err
	}

	if err := e.extractElementsInMap(valueMapping); err != nil {
		return err
	}

	return mapstructure.Decode(&valueMapping, &output)
}

func (e *elementExtractor) extractElementsInMap(valueMapping map[string]any) error {
	for key, elementLocation := range e.Fields {
		item, ok := valueMapping[key]
		if !ok {
			return types.ErrInvalidType
		} else if item == nil {
			continue
		}

		rItem := reflect.ValueOf(item)
		switch rItem.Kind() {
		case reflect.Array, reflect.Slice:
		default:
			return types.ErrInvalidType
		}

		if rItem.Len() == 0 {
			valueMapping[key] = nil
			continue
		}

		switch elementLocation {
		case FirstElementLocation:
			valueMapping[key] = rItem.Index(0).Interface()
		case MiddleElementLocation:
			valueMapping[key] = rItem.Index(rItem.Len() / 2).Interface()
		case LastElementLocation:
			valueMapping[key] = rItem.Index(rItem.Len() - 1).Interface()
		}
	}
	return nil
}

func (e *elementExtractor) expandElements(input, output any) error {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(input, &valueMapping); err != nil {
		return err
	}

	for key, _ := range e.Fields {
		item, ok := valueMapping[key]
		if !ok {
			return types.ErrInvalidType
		} else if item == nil {
			continue
		}

		rItem := reflect.ValueOf(item)
		slice := reflect.MakeSlice(reflect.SliceOf(rItem.Type()), 1, 1)
		slice.Index(0).Set(rItem)
		valueMapping[key] = slice.Interface()
	}

	return mapstructure.Decode(&valueMapping, output)
}
