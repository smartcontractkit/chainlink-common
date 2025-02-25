package codec

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type modifierBase[T any] struct {
	enablePathTraverse  bool
	fields              map[string]T
	onToOffChainType    map[reflect.Type]reflect.Type
	offToOnChainType    map[reflect.Type]reflect.Type
	modifyFieldForInput func(pkgPath string, outputField *reflect.StructField, fullPath string, change T) error
	addFieldForInput    func(pkgPath, name string, change T) reflect.StructField
	onChainStructType   reflect.Type
	offChainStructType  reflect.Type
}

// RetypeToOffChain sets the on-chain and off-chain types for modifications. If itemType is empty, the type returned
// will be the full off-chain type and all type mappings will be reset. If itemType is not empty, retyping assumes a
// sub-field is expected and the off-chain type of the sub-field is returned with no modifications to internal type
// mappings.
func (m *modifierBase[T]) RetypeToOffChain(onChainType reflect.Type, itemType string) (tpe reflect.Type, err error) {
	// onChainType could be the entire struct or a sub-field type
	defer func() {
		// StructOf can panic if the fields are not valid
		if r := recover(); r != nil {
			tpe = nil
			err = fmt.Errorf("%w: %v", types.ErrInvalidType, r)
		}
	}()

	// path traverse allows an item type of Struct.FieldA.NestedField to isolate modifiers
	// associated with the nested field `NestedField`.
	if !m.enablePathTraverse {
		itemType = ""
	}

	// if itemType is empty, store the type mappings
	// if itemType is not empty, assume a sub-field property is expected to be extracted
	onChainStructType := onChainType
	if itemType != "" {
		onChainStructType = m.onChainStructType
	}

	// this will only work for the full on-chain struct type unless we cache the individual
	// field types too.
	if cached, ok := m.onToOffChainType[onChainStructType]; ok {
		return typeForPath(cached, itemType)
	}

	if len(m.fields) == 0 {
		m.offToOnChainType[onChainType] = onChainType
		m.onToOffChainType[onChainType] = onChainType
		m.onChainStructType = onChainType
		m.offChainStructType = onChainType

		return typeForPath(onChainType, itemType)
	}

	var offChainType reflect.Type

	// the onChainStructType here should always reference the full on-chain struct type
	switch onChainStructType.Kind() {
	case reflect.Pointer:
		var elm reflect.Type

		if elm, err = m.RetypeToOffChain(onChainStructType.Elem(), itemType); err != nil {
			return nil, err
		}

		offChainType = reflect.PointerTo(elm)
	case reflect.Slice:
		var elm reflect.Type

		if elm, err = m.RetypeToOffChain(onChainStructType.Elem(), ""); err != nil {
			return nil, err
		}

		offChainType = reflect.SliceOf(elm)
	case reflect.Array:
		var elm reflect.Type

		if elm, err = m.RetypeToOffChain(onChainStructType.Elem(), ""); err != nil {
			return nil, err
		}

		offChainType = reflect.ArrayOf(onChainStructType.Len(), elm)
	case reflect.Struct:
		if offChainType, err = m.getStructType(onChainStructType); err != nil {
			return nil, err
		}
	default:
		// if the types don't match, it means we are attempting to traverse the main struct
		if onChainType != m.onChainStructType {
			return onChainType, nil
		}

		return nil, fmt.Errorf("%w: cannot retype the kind %v", types.ErrInvalidType, onChainStructType.Kind())
	}

	m.onToOffChainType[onChainStructType] = offChainType
	m.offToOnChainType[offChainType] = onChainStructType

	if m.onChainStructType == nil {
		m.onChainStructType = onChainType
		m.offChainStructType = offChainType
	}

	return typeForPath(offChainType, itemType)
}

func (m *modifierBase[T]) getStructType(outputType reflect.Type) (reflect.Type, error) {
	filedLocations, err := getFieldIndices(outputType)
	if err != nil {
		return nil, err
	}

	for _, key := range m.subkeysFirst() {
		curLocations := filedLocations
		parts := strings.Split(key, ".")
		fieldName := parts[len(parts)-1]

		parts = parts[:len(parts)-1]
		for _, part := range parts {
			if curLocations, err = curLocations.populateSubFields(part); err != nil {
				return nil, err
			}
		}

		// If a subkey has been modified, update the underlying types first
		curLocations.updateTypeFromSubkeyMods(fieldName)
		if field, ok := curLocations.fieldByName(fieldName); ok {
			if err = m.modifyFieldForInput(curLocations.pkgPath, field, key, m.fields[key]); err != nil {
				return nil, err
			}
		} else {
			if m.addFieldForInput == nil {
				return nil, fmt.Errorf("%w: cannot find %s", types.ErrInvalidType, key)
			}
			curLocations.addNewField(m.addFieldForInput(curLocations.pkgPath, fieldName, m.fields[key]))
		}
	}

	return filedLocations.makeNewType(), nil
}

// subkeysFirst returns a list of keys that will always have a sub-key before the key if both are present
func (m *modifierBase[T]) subkeysFirst() []string {
	orderedKeys := make([]string, 0, len(m.fields))
	for k := range m.fields {
		orderedKeys = append(orderedKeys, k)
	}

	sort.Slice(orderedKeys, func(i, j int) bool {
		return orderedKeys[i] > orderedKeys[j]
	})

	return orderedKeys
}

func (m *modifierBase[T]) onToOffChainTyper(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	onChainRefType := onChainType
	if itemType != "" {
		onChainRefType = m.onChainStructType
	}

	offChainType, ok := m.onToOffChainType[onChainRefType]
	if !ok {
		return nil, fmt.Errorf("%w: cannot rename unknown type %v", types.ErrInvalidType, onChainType)
	}

	return typeForPath(offChainType, itemType)
}

func (m *modifierBase[T]) offToOnChainTyper(offChainType reflect.Type, itemType string) (reflect.Type, error) {
	offChainRefType := offChainType
	if itemType != "" {
		offChainRefType = m.offChainStructType
	}

	onChainType, ok := m.offToOnChainType[offChainRefType]
	if !ok {
		return nil, fmt.Errorf("%w: cannot rename unknown type %v", types.ErrInvalidType, offChainType)
	}

	return typeForPath(onChainType, itemType)
}

func (m *modifierBase[T]) selectType(inputValue any, savedType reflect.Type, itemType string) (any, string, error) {
	// set itemType to an ignore value if path traversal is not enabled
	if !m.enablePathTraverse {
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

// subkeysLast returns a list of keys that will always have a sub-key after the key if both are present
func subkeysLast[T any](fields map[string]T) []string {
	orderedKeys := make([]string, 0, len(fields))
	for k := range fields {
		orderedKeys = append(orderedKeys, k)
	}

	sort.Strings(orderedKeys)

	return orderedKeys
}

type mapAction[T any] func(extractMap map[string]any, key string, element T) error

func transformWithMaps[T any](
	item any,
	typeMap map[reflect.Type]reflect.Type,
	fields map[string]T,
	fn mapAction[T],
	hooks ...mapstructure.DecodeHookFunc) (any, error) {
	rItem := reflect.ValueOf(item)

	toType, ok := typeMap[rItem.Type()]
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
	}

	rOutput, err := transformWithMapsHelper(rItem, toType, fields, fn, hooks)
	if err != nil {
		return reflect.Value{}, err
	}

	return rOutput.Interface(), nil
}

func transformWithMapsHelper[T any](
	rItem reflect.Value,
	toType reflect.Type,
	fields map[string]T,
	fn mapAction[T],
	hooks []mapstructure.DecodeHookFunc) (reflect.Value, error) {
	switch rItem.Kind() {
	case reflect.Pointer:
		elm := rItem.Elem()
		if elm.Kind() == reflect.Struct {
			into := reflect.New(toType.Elem())
			err := changeElements(rItem.Interface(), into.Interface(), fields, fn, hooks)

			return into, err
		}

		tmp, err := transformWithMapsHelper(elm, toType.Elem(), fields, fn, hooks)
		result := reflect.New(toType.Elem())
		reflect.Indirect(result).Set(tmp)

		return result, err
	case reflect.Struct:
		into := reflect.New(toType)
		err := changeElements(rItem.Interface(), into.Interface(), fields, fn, hooks)

		return into.Elem(), err
	case reflect.Slice:
		length := rItem.Len()
		into := reflect.MakeSlice(toType, length, length)
		err := doMany(rItem, into, fields, fn, hooks)

		return into, err
	case reflect.Array:
		into := reflect.New(toType).Elem()
		err := doMany(rItem, into, fields, fn, hooks)

		return into, err
	default:
		return reflect.Value{}, fmt.Errorf("%w: cannot retype %v", types.ErrInvalidType, rItem.Type())
	}
}

func doMany[T any](rInput, rOutput reflect.Value, fields map[string]T, fn mapAction[T], hooks []mapstructure.DecodeHookFunc) error {
	length := rInput.Len()
	for i := 0; i < length; i++ {
		// Make sure the items are addressable
		inTmp := rInput.Index(i)
		outTmp := rOutput.Index(i)

		output, err := transformWithMapsHelper(inTmp, outTmp.Type(), fields, fn, hooks)
		if err != nil {
			return err
		}

		outTmp.Set(output)
	}

	return nil
}

func changeElements[T any](src, dest any, fields map[string]T, fn mapAction[T], hooks []mapstructure.DecodeHookFunc) error {
	valueMapping := map[string]any{}
	if err := mapstructure.Decode(src, &valueMapping); err != nil {
		return fmt.Errorf("%w: failed to decode source type: %w", types.ErrInvalidType, err)
	}

	if err := doForMapElements(valueMapping, fields, fn); err != nil {
		return err
	}

	conf := &mapstructure.DecoderConfig{Result: &dest, Squash: true}
	if len(hooks) != 0 {
		conf.DecodeHook = mapstructure.ComposeDecodeHookFunc(hooks...)
	}

	hookedDecoder, err := mapstructure.NewDecoder(conf)
	if err != nil {
		return fmt.Errorf("%w: failed to create configured decoder: %w", types.ErrInvalidType, err)
	}

	if err = hookedDecoder.Decode(valueMapping); err != nil {
		return fmt.Errorf("%w: failed to decode destination type: %w", types.ErrInvalidType, err)
	}

	return nil
}

func doForMapElements[T any](valueMapping map[string]any, fields map[string]T, fn mapAction[T]) error {
	for key, value := range fields {
		path := strings.Split(key, ".")
		name := path[len(path)-1]
		path = path[:len(path)-1]

		extractMaps, err := getMapsFromPath(valueMapping, path)
		if err != nil {
			return PathMappingError{Err: err, Path: key}
		}

		for _, em := range extractMaps {
			if err = fn(em, name, value); err != nil {
				return PathMappingError{Err: err, Path: key}
			}
		}
	}

	return nil
}

func typeForPath(from reflect.Type, itemType string) (reflect.Type, error) {
	if itemType == "" {
		return from, nil
	}

	switch from.Kind() {
	case reflect.Pointer:
		elem, err := typeForPath(from.Elem(), itemType)
		if err != nil {
			return nil, err
		}

		return elem, nil
	case reflect.Array, reflect.Slice:
		return nil, fmt.Errorf("%w: cannot extract a field from an array or slice", types.ErrInvalidType)
	case reflect.Struct:
		head, tail := ItemTyper(itemType).Next()

		field, ok := from.FieldByName(head)
		if !ok {
			return nil, fmt.Errorf("%w: field not found for path %s and itemType %s", types.ErrInvalidType, from, itemType)
		}

		if tail == "" {
			return field.Type, nil
		}

		return typeForPath(field.Type, tail)
	default:
		return nil, fmt.Errorf("%w: cannot extract a field from kind %s", types.ErrInvalidType, from.Kind())
	}
}

type PathMappingError struct {
	Err  error
	Path string
}

func (e PathMappingError) Error() string {
	return fmt.Sprintf("mapping error for path (%s): %s", e.Path, e.Err)
}

func (e PathMappingError) Cause() error {
	return e.Err
}

type ItemTyper string

func (t ItemTyper) Next() (string, string) {
	if string(t) == "" {
		return "", ""
	}

	path := strings.Split(string(t), ".")
	if len(path) == 1 {
		return path[0], ""
	}

	return path[0], strings.Join(path[1:], ".")
}
