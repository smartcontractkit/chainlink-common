package codec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewPropertyExtractor creates a modifier that will extract a single property from a struct.
// This modifier is lossy, as TransformToOffchain will discard unwanted struct properties and
// return a single element. Calling TransformToOnchain will result in unset properties.
// Extracting a field which is nested under a slice of structs will return a slice containing the extracted property for every element of the slice, this is completely lossy and cannot be transformed back.
func NewPropertyExtractor(fieldName string) Modifier {
	return NewPathTraversePropertyExtractor(fieldName, false)
}

func NewPathTraversePropertyExtractor(fieldName string, enablePathTraverse bool) Modifier {
	return &propertyExtractor{
		onToOffChainType:   map[reflect.Type]reflect.Type{},
		offToOnChainType:   map[reflect.Type]reflect.Type{},
		fieldName:          fieldName,
		enablePathTraverse: enablePathTraverse,
	}
}

type propertyExtractor struct {
	fieldName          string
	enablePathTraverse bool
	onToOffChainType   map[reflect.Type]reflect.Type
	offToOnChainType   map[reflect.Type]reflect.Type
	onChainStructType  reflect.Type
	offChainStructType reflect.Type
}

func (e *propertyExtractor) RetypeToOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	if e.fieldName == "" {
		return nil, fmt.Errorf("%w: field name required for extraction", types.ErrInvalidConfig)
	}

	// path traverse allows an item type of Struct.FieldA.NestedField to isolate modifiers
	// associated with the nested field `NestedField`.
	if !e.enablePathTraverse {
		itemType = ""
	}

	// if itemType is empty, store the type mappings
	// if itemType is not empty, assume a sub-field property is expected to be extracted
	onChainStructType := onChainType
	if itemType != "" {
		onChainStructType = e.onChainStructType
	}

	if cached, ok := e.onToOffChainType[onChainStructType]; ok {
		return cached, nil
	}

	var (
		offChainType reflect.Type
		err          error
	)

	switch onChainType.Kind() {
	case reflect.Pointer:
		var elm reflect.Type

		if elm, err = e.RetypeToOffChain(onChainStructType.Elem(), ""); err != nil {
			return nil, err
		}

		offChainType = reflect.PointerTo(elm)
	case reflect.Slice:
		var elm reflect.Type

		if elm, err = e.RetypeToOffChain(onChainStructType.Elem(), ""); err != nil {
			return nil, err
		}

		offChainType = reflect.SliceOf(elm)
	case reflect.Array:
		var elm reflect.Type

		if elm, err = e.RetypeToOffChain(onChainStructType.Elem(), ""); err != nil {
			return nil, err
		}

		offChainType = reflect.ArrayOf(onChainStructType.Len(), elm)
	case reflect.Struct:
		if offChainType, err = e.getPropTypeFromStruct(onChainStructType); err != nil {
			return nil, err
		}
	default:
		// if the types don't match, it means we are attempting to traverse the main struct
		if onChainType != e.onChainStructType {
			return onChainType, nil
		}

		return nil, fmt.Errorf("%w: cannot retype the kind %v", types.ErrInvalidType, onChainType.Kind())
	}

	e.onToOffChainType[onChainStructType] = offChainType
	e.offToOnChainType[offChainType] = onChainStructType

	if e.onChainStructType == nil {
		e.onChainStructType = onChainType
		e.offChainStructType = offChainType
	}

	return typeForPath(offChainType, itemType)
}

func (e *propertyExtractor) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	offChainValue, itemType, err := e.selectType(offChainValue, e.offChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := extractOrExpandWithMaps(offChainValue, e.offToOnChainType, e.fieldName, expandWithMapsHelper)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		// add the field name because the offChainType was nested into a new struct
		itemType = fmt.Sprintf("%s.%s", e.fieldName, itemType)

		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

func (e *propertyExtractor) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	onChainValue, itemType, err := e.selectType(onChainValue, e.onChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := extractOrExpandWithMaps(onChainValue, e.onToOffChainType, e.fieldName, extractWithMapsHelper)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		// remove the head from the itemType because a field was extracted
		_, tail := ItemTyper(itemType).Next()

		return ValueForPath(reflect.ValueOf(modified), tail)
	}

	return modified, nil
}

func (e *propertyExtractor) selectType(inputValue any, savedType reflect.Type, itemType string) (any, string, error) {
	// set itemType to an ignore value if path traversal is not enabled
	if !e.enablePathTraverse {
		return inputValue, "", nil
	}

	// the offChainValue might be a subfield value; get the true offChainStruct type already stored and set the value
	baseStructValue := inputValue

	// path traversal is expected, but offChainValue is the value of a field, not the actual struct
	// create a new struct from the stored offChainStruct with the provided value applied and all other fields set to
	// their zero value.
	if itemType != "" {
		into := reflect.New(savedType)

		if err := SetValueAtPath(into, reflect.ValueOf(inputValue), itemType); err != nil {
			return nil, itemType, err
		}

		baseStructValue = reflect.Indirect(into).Interface()
	}

	return baseStructValue, itemType, nil
}

func (e *propertyExtractor) getPropTypeFromStruct(onChainType reflect.Type) (reflect.Type, error) {
	filedLocations, err := getFieldIndices(onChainType)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(e.fieldName, ".")
	fieldName := parts[len(parts)-1]
	parts = parts[:len(parts)-1]

	curLocations := filedLocations
	prevLocations := curLocations
	for _, part := range parts {
		prevLocations = curLocations
		if curLocations, err = curLocations.populateSubFields(part); err != nil {
			return nil, err
		}
	}

	// if value for extraction is nested under a slice return [][] with type of the value to be extracted
	var prevIsSlice bool
	if prevLocations != nil && len(parts) > 0 {
		prevLocation, ok := prevLocations.fieldByName(parts[len(parts)-1])
		if !ok {
			return nil, fmt.Errorf("%w: field not found in on-chain type %s", types.ErrInvalidType, e.fieldName)
		}

		if prevLocation.Type.Kind() == reflect.Ptr {
			if prevLocation.Type.Elem().Kind() == reflect.Slice {
				prevIsSlice = true
			}
		} else if prevLocation.Type.Kind() == reflect.Slice {
			prevIsSlice = true
		}
	}

	curLocations.updateTypeFromSubkeyMods(fieldName)
	field, ok := curLocations.fieldByName(fieldName)
	if !ok {
		return nil, fmt.Errorf("%w: field not found in on-chain type %s", types.ErrInvalidType, e.fieldName)
	}

	if prevIsSlice {
		field.Type = reflect.SliceOf(field.Type)
	}

	e.onToOffChainType[onChainType] = field.Type
	e.offToOnChainType[field.Type] = onChainType

	return field.Type, nil
}

type transformHelperFunc func(reflect.Value, reflect.Type, string) (reflect.Value, error)

func extractOrExpandWithMaps(input any, typeMap map[reflect.Type]reflect.Type, field string, fn transformHelperFunc) (any, error) {
	rItem := reflect.ValueOf(input)
	toType, ok := typeMap[rItem.Type()]
	if !ok {
		if rItem.Kind() == reflect.Struct && rItem.NumField() == 1 {
			toType, ok = typeMap[rItem.Field(0).Type()]
			if !ok {
				return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
			}
		} else {
			return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
		}
	}

	output, err := fn(rItem, toType, field)
	if err != nil {
		return reflect.Value{}, err
	}

	return output.Interface(), err
}

func expandWithMapsHelper(rItem reflect.Value, toType reflect.Type, field string) (reflect.Value, error) {
	switch toType.Kind() {
	case reflect.Pointer:
		if rItem.Kind() != reflect.Pointer {
			return reflect.Value{}, fmt.Errorf("%w: value to expand should be pointer", types.ErrInvalidType)
		}

		if toType.Elem().Kind() == reflect.Struct {
			into := reflect.New(toType.Elem())
			err := setFieldValue(rItem.Elem().Interface(), into.Interface(), field)
			return into, err
		}

		tmp, err := expandWithMapsHelper(rItem.Elem(), toType.Elem(), field)
		result := reflect.New(toType.Elem())
		reflect.Indirect(result).Set(tmp)

		return result, err
	case reflect.Struct:
		into := reflect.New(toType)
		err := setFieldValue(rItem.Interface(), into.Interface(), field)

		return into.Elem(), err
	case reflect.Slice:
		if rItem.Kind() != reflect.Slice {
			return reflect.Value{}, fmt.Errorf("%w: value to expand should be slice", types.ErrInvalidType)
		}

		length := rItem.Len()
		into := reflect.MakeSlice(toType, length, length)
		err := extractOrExpandMany(rItem, into, field, expandWithMapsHelper)

		return into, err
	case reflect.Array:
		if rItem.Kind() != reflect.Array {
			return reflect.Value{}, fmt.Errorf("%w: value to expand should be array", types.ErrInvalidType)
		}

		into := reflect.New(toType).Elem()
		err := extractOrExpandMany(rItem, into, field, expandWithMapsHelper)

		return into, err
	default:
		return reflect.Value{}, fmt.Errorf("%w: cannot retype", types.ErrInvalidType)
	}
}

func extractWithMapsHelper(rItem reflect.Value, toType reflect.Type, field string) (reflect.Value, error) {
	switch rItem.Kind() {
	case reflect.Pointer:
		elm := rItem.Elem()
		if elm.Kind() == reflect.Struct {
			var (
				tmp reflect.Value
				err error
			)

			if tmp, err = extractElement(rItem.Interface(), field); err != nil {
				return rItem, err
			}

			result := reflect.New(toType.Elem())
			err = mapstructure.Decode(tmp.Interface(), result.Interface())

			return result, err
		}

		tmp, err := extractWithMapsHelper(elm, toType.Elem(), field)
		result := reflect.New(toType.Elem())
		reflect.Indirect(result).Set(tmp)

		return result, err
	case reflect.Struct:
		return extractElement(rItem.Interface(), field)
	case reflect.Slice:
		length := rItem.Len()
		into := reflect.MakeSlice(toType, length, length)
		err := extractOrExpandMany(rItem, into, field, extractWithMapsHelper)
		return into, err
	case reflect.Array:
		into := reflect.New(toType).Elem()
		err := extractOrExpandMany(rItem, into, field, extractWithMapsHelper)
		return into, err
	default:
		return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
	}
}

type extractOrExpandHelperFunc func(reflect.Value, reflect.Type, string) (reflect.Value, error)

func extractOrExpandMany(rInput, rOutput reflect.Value, field string, fn extractOrExpandHelperFunc) error {
	length := rInput.Len()

	for i := 0; i < length; i++ {
		inTmp := rInput.Index(i)
		outTmp := rOutput.Index(i)

		output, err := fn(inTmp, outTmp.Type(), field)
		if err != nil {
			return err
		}

		outTmp.Set(output)
	}

	return nil
}

func extractElement(src any, field string) (reflect.Value, error) {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(src, &valueMapping); err != nil {
		return reflect.Value{}, err
	}

	path, name := pathAndName(field)

	extractMaps, err := getMapsFromPath(valueMapping, path)
	if err != nil {
		return reflect.Value{}, err
	}

	if len(extractMaps) != 1 {
		var sliceValue reflect.Value
		var sliceInitialized bool
		for _, fields := range extractMaps {
			val, ok := fields[name]
			if !ok {
				continue
			}

			rv := reflect.ValueOf(val)
			// If this is the first item found, initialize the typed slice
			if !sliceInitialized {
				sliceType := reflect.SliceOf(rv.Type())
				sliceValue = reflect.MakeSlice(sliceType, 0, 0)
				sliceInitialized = true
			}

			sliceValue = reflect.Append(sliceValue, rv)
		}

		if !sliceInitialized || sliceValue.Len() == 0 {
			return reflect.Value{}, fmt.Errorf("%w: cannot find %q", types.ErrInvalidType, field)
		}

		return sliceValue, nil
	}

	em := extractMaps[0]

	item, ok := em[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: cannot find %s", types.ErrInvalidType, field)
	}

	return reflect.ValueOf(item), nil
}

func setFieldValue(src, dest any, field string) error {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(dest, &valueMapping); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	path, name := pathAndName(field)

	extractMaps, err := getMapsFromPath(valueMapping, path)
	if err != nil {
		return err
	}

	if len(extractMaps) != 1 {
		return fmt.Errorf("%w: only 1 extract map expected", types.ErrInvalidType)
	}

	extractMaps[0][name] = src

	conf := &mapstructure.DecoderConfig{Result: &dest, Squash: true}
	decoder, err := mapstructure.NewDecoder(conf)
	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	if err = decoder.Decode(valueMapping); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}

	return nil
}

func pathAndName(field string) ([]string, string) {
	path := strings.Split(field, ".")
	name := path[len(path)-1]
	path = path[:len(path)-1]

	return path, name
}
