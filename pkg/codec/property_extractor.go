package codec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewPropertyExtractor creates a modifier that will extract a single property from a struct.
// This modifier is lossy, as TransformToOffchain will discard unwanted struct properties and
// return a single element. Calling TransformToOnchain will result in unset properties.
func NewPropertyExtractor(fieldName string) Modifier {
	m := &propertyExtractor{
		modifierBase: modifierBase[bool]{
			fields:           map[string]bool{fieldName: false},
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
		fieldName: fieldName,
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ bool) error {
		return nil
	}

	return m
}

type propertyExtractor struct {
	modifierBase[bool]
	fieldName string
}

func (e *propertyExtractor) TransformToOnChain(offChainValue any, _ string) (any, error) {
	parts := strings.Split(e.fieldName, ".")
	fieldName := parts[len(parts)-1]

	fields := []reflect.StructField{
		{Name: fieldName, Type: reflect.TypeOf(offChainValue)},
	}

	onChainStruct := reflect.StructOf(fields)
	output := reflect.New(onChainStruct)
	iOutput := reflect.Indirect(output)
	iOutput.FieldByName(fieldName).Set(reflect.ValueOf(offChainValue))

	return iOutput.Interface(), nil
}

func (e *propertyExtractor) TransformToOffChain(onChainValue any, _ string) (any, error) {
	rItem := reflect.ValueOf(onChainValue)

	_, ok := e.onToOffChainType[rItem.Type()]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
	}

	switch rItem.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		rItem = rItem.Elem()
		if rItem.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("%w: must be struct or pointer to struct to extract value", types.ErrInvalidType)
		}
	default:
		return reflect.Value{}, fmt.Errorf("%w: must be struct or pointer to struct to extract value", types.ErrInvalidType)
	}

	valueMapping := map[string]any{}
	if err := mapstructure.Decode(rItem.Interface(), &valueMapping); err != nil {
		return reflect.Value{}, err
	}

	path := strings.Split(e.fieldName, ".")
	name := path[len(path)-1]
	path = path[:len(path)-1]

	extractMaps, err := getMapsFromPath(valueMapping, path)
	if err != nil {
		return reflect.Value{}, err
	}

	if len(extractMaps) != 1 {
		return reflect.Value{}, fmt.Errorf("%w: cannot find %s", types.ErrInvalidType, e.fieldName)
	}

	em := extractMaps[0]

	item, ok := em[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: cannot find %s", types.ErrInvalidType, e.fieldName)
	}

	return item, nil
}
