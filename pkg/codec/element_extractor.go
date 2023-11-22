package codec

import (
	"fmt"
	"reflect"
	"strings"

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
			fields:            fields,
			onToOffChainType:  map[reflect.Type]reflect.Type{},
			offToOneChainType: map[reflect.Type]reflect.Type{},
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

func (e *elementExtractor) TransformForOnChain(input any) (any, error) {
	return e.transform(input, e.offToOneChainType, extractMap)
}

func (e *elementExtractor) TransformForOffChain(output any) (any, error) {
	return e.transform(output, e.onToOffChainType, expandMap)
}

type mapAction func(extractMap map[string]any, key string, elementLocation ElementExtractorLocation) error

func (e *elementExtractor) transform(item any, typeMap map[reflect.Type]reflect.Type, fn mapAction) (any, error) {
	rItem := reflect.ValueOf(item)

	toType, ok := typeMap[rItem.Type()]
	if !ok {
		return reflect.Value{}, types.ErrInvalidType
	}

	switch rItem.Kind() {
	case reflect.Pointer:
		elm := rItem.Elem()
		if elm.Kind() == reflect.Struct {
			into := reflect.New(toType.Elem())
			err := e.changeElements(item, into.Interface(), fn)
			return into.Interface(), err
		}

		tmp, err := e.transform(elm.Interface(), typeMap, fn)
		result := reflect.New(toType.Elem())
		reflect.Indirect(result).Set(reflect.ValueOf(tmp))
		return result.Interface(), err
	case reflect.Struct:
		into := reflect.New(toType)
		err := e.changeElements(item, into.Interface(), fn)
		return into.Elem().Interface(), err
	case reflect.Slice:
		length := rItem.Len()
		into := reflect.MakeSlice(toType, length, length)
		err := e.doMany(rItem, into, fn)
		return into.Interface(), err
	case reflect.Array:
		into := reflect.New(toType).Elem()
		err := e.doMany(rItem, into, fn)
		return into.Interface(), err
	}

	return reflect.Value{}, types.ErrInvalidType
}

func (e *elementExtractor) changeElements(src, dest any, fn mapAction) error {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(src, &valueMapping); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	if err := e.doForMapElements(valueMapping, fn); err != nil {
		return err
	}

	if err := mapstructure.Decode(&valueMapping, &dest); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}
	return nil
}

func (e *elementExtractor) doForMapElements(valueMapping map[string]any, fn mapAction) error {
	for key, elementLocation := range e.fields {
		path := strings.Split(key, ".")
		name := path[len(path)-1]
		path = path[:len(path)-1]
		extractMaps, err := getMapsFromPath(valueMapping, path)
		if err != nil {
			return err
		}

		for _, m := range extractMaps {
			if err = fn(m, name, elementLocation); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractMap(extractMap map[string]any, key string, elementLocation ElementExtractorLocation) error {
	item, ok := extractMap[key]
	if !ok {
		return fmt.Errorf("%w: cannot find %s", types.ErrInvalidType, key)
	} else if item == nil {
		return nil
	}

	rItem := reflect.ValueOf(item)
	switch rItem.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return types.ErrInvalidType
	}

	if rItem.Len() == 0 {
		extractMap[key] = nil
		return nil
	}

	switch elementLocation {
	case FirstElementLocation:
		extractMap[key] = rItem.Index(0).Interface()
	case MiddleElementLocation:
		extractMap[key] = rItem.Index(rItem.Len() / 2).Interface()
	case LastElementLocation:
		extractMap[key] = rItem.Index(rItem.Len() - 1).Interface()
	}

	return nil
}

func expandMap(extractMap map[string]any, key string, _ ElementExtractorLocation) error {
	item, ok := extractMap[key]
	if !ok {
		return types.ErrInvalidType
	} else if item == nil {
		return nil
	}

	rItem := reflect.ValueOf(item)
	slice := reflect.MakeSlice(reflect.SliceOf(rItem.Type()), 1, 1)
	slice.Index(0).Set(rItem)
	extractMap[key] = slice.Interface()
	return nil
}

func (e *elementExtractor) doMany(rInput, rOutput reflect.Value, fn mapAction) error {
	length := rInput.Len()
	for i := 0; i < length; i++ {
		// Make sure the index is addressable
		tmp := rOutput.Index(i).Interface()
		if err := e.changeElements(rInput.Index(i).Interface(), &tmp, fn); err != nil {
			return err
		}
		rOutput.Index(i).Set(reflect.ValueOf(tmp))
	}
	return nil
}
