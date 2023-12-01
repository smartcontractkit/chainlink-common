package codec

import (
	"fmt"
	"reflect"

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
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}
	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ ElementExtractorLocation) error {
		field.Type = reflect.SliceOf(field.Type)
		return nil
	}

	return m
}

type elementExtractor struct {
	modifierBase[ElementExtractorLocation]
}

func (e *elementExtractor) TransformForOnChain(input any) (any, error) {
	return transformWithMaps(input, e.offToOnChainType, e.fields, extractMap)
}

func (e *elementExtractor) TransformForOffChain(output any) (any, error) {
	return transformWithMaps(output, e.onToOffChainType, e.fields, expandMap)
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
