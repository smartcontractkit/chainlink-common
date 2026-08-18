package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewElementExtractorFromOnchain(fields map[string]*ElementExtractorLocation) Modifier {
	m := &elementExtractorFromOnchain{
		modifierBase: modifierBase[*ElementExtractorLocation]{
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ *ElementExtractorLocation) error {
		if field.Type.Kind() == reflect.Pointer {
			field.Type = field.Type.Elem()
		}

		switch field.Type.Kind() {
		case reflect.Array, reflect.Slice:
		default:
			return fmt.Errorf("%w: %q is not a slice or array, but is %q", types.ErrInvalidType, field.Name, field.Type.Kind())
		}

		field.Type = field.Type.Elem()
		return nil
	}

	return m
}

func (e *elementExtractorFromOnchain) TransformToOnChain(offChainValue any, _ string) (any, error) {
	return transformWithMaps(offChainValue, e.offToOnChainType, e.fields, expandMap)
}

func (e *elementExtractorFromOnchain) TransformToOffChain(onChainValue any, _ string) (any, error) {
	return transformWithMaps(onChainValue, e.onToOffChainType, e.fields, extractMap)
}

type elementExtractorFromOnchain struct {
	modifierBase[*ElementExtractorLocation]
}
