package codec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewHardCoder creates a modifier that will hard-code values for on-chain and off-chain types. The modifier will
// override any values of the same name, if you need an overwritten value to be used in a different field. NewRenamer
// must be used before NewHardCoder.
func NewHardCoder(
	onChain map[string]any,
	offChain map[string]any,
	hooks ...mapstructure.DecodeHookFunc,
) (Modifier, error) {
	return NewPathTraverseHardCoder(onChain, offChain, false, hooks...)
}

func NewPathTraverseHardCoder(
	onChain map[string]any,
	offChain map[string]any,
	enablePathTraverse bool,
	hooks ...mapstructure.DecodeHookFunc,
) (Modifier, error) {
	if err := verifyHardCodeKeys(onChain); err != nil {
		return nil, err
	} else if err = verifyHardCodeKeys(offChain); err != nil {
		return nil, err
	}

	myHooks := make([]mapstructure.DecodeHookFunc, len(hooks)+1)
	copy(myHooks, hooks)
	myHooks[len(hooks)] = hardCodeManyHook

	m := &onChainHardCoder{
		modifierBase: modifierBase[any]{
			enablePathTraverse: enablePathTraverse,
			fields:             offChain,
			onToOffChainType:   map[reflect.Type]reflect.Type{},
			offToOnChainType:   map[reflect.Type]reflect.Type{},
		},
		onChain: onChain,
		hooks:   myHooks,
	}
	m.modifyFieldForInput = func(_ string, field *reflect.StructField, key string, v any) error {
		// if we are typing it differently, we need to make sure it's hard-coded the other way
		newType := reflect.TypeOf(v)
		if _, ok := m.onChain[key]; !ok && field.Type != newType {
			return fmt.Errorf(
				"%w: cannot change field type without hard-coding its onchain value for key %s",
				types.ErrInvalidType,
				key)
		}
		field.Type = newType
		return nil
	}
	m.addFieldForInput = func(_, key string, value any) reflect.StructField {
		return reflect.StructField{
			Name: key,
			Type: reflect.TypeOf(value),
		}
	}
	return m, nil
}

type onChainHardCoder struct {
	modifierBase[any]
	onChain map[string]any
	hooks   []mapstructure.DecodeHookFunc
}

// verifyHardCodeKeys checks that no key is a prefix of another key
// This is important because if you hard code "A" : {"B" : 10}, and "A.C" : 20
// A key will override all A values and the A.C will add to existing values, which is inconsistent.
// instead the user should do "A" : {"B" : 10, "C" : 20} if they want to override or
// "A.B" : 10, "A.C" : 20 if they want to add
func verifyHardCodeKeys(values map[string]any) error {
	seen := map[string]bool{}
	for _, k := range subkeysLast(values) {
		parts := strings.Split(k, ".")
		on := ""
		for _, part := range parts {
			on += part
			if seen[on] {
				return fmt.Errorf("%w: key %s and %s cannot both be present", types.ErrInvalidConfig, on, k)
			}
		}
		seen[k] = true
	}
	return nil
}

// TransformToOnChain will apply the hard-code modifier and hooks on the value identified by itemType. If path traverse
// is not enabled, itemType is ignored.
//
// For path-traversal, the itemType may reference a field that does not exist in the off-chain type, but is being added
// by the hard-code modifier. Ex. offChain A.B (does not have 'C') and onChain A.B.C ('C' gets hard-coded); if 'C' is
// intended to be the result of the transformation, itemType must be A.B.C even though the off-chain type does not have
// field 'C'.
func (m *onChainHardCoder) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	offChainValue, itemType, err := m.modifierBase.selectType(offChainValue, m.offChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := transformWithMaps(offChainValue, m.offToOnChainType, m.onChain, hardCode, m.hooks...)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

// TransformToOffChain will apply the hard-code modifier and hooks on the value identified by itemType. If path traverse
// is not enabled, itemType is ignored.
//
// For path-traversal, the itemType may reference a field that does not exist in the on-chain type, but is being added
// by the hard-code modifier. Ex. on-chain A.B (does not have 'C') and off-chain A.B.C ('C' gets hard-coded); if 'C' is
// intended to be the result of the transformation, itemType must be A.B.C even though the on-chain type does not have
// field 'C'.
func (m *onChainHardCoder) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	onChainValue, itemType, err := m.modifierBase.selectType(onChainValue, m.onChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	allHooks := make([]mapstructure.DecodeHookFunc, len(m.hooks)+1)
	copy(allHooks, m.hooks)
	allHooks[len(m.hooks)] = hardCodeManyHook

	// if there is only one field with unset name, then a primitive variable is being hardcoded
	if len(m.fields) == 1 {
		for k, v := range m.fields {
			if k == "" {
				return v, nil
			}
		}
	}

	modified, err := transformWithMaps(onChainValue, m.onToOffChainType, m.fields, hardCode, allHooks...)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

func hardCode(extractMap map[string]any, key string, item any) error {
	extractMap[key] = item
	return nil
}

// hardCodeManyHook allows a user to specify a single value for a slice or array
// This is useful because users may not know how many values are in an array ahead of time (e.g. number of reports)
// Instead, a user can specify A.C = 10 and if A is an array, all A.C values will be set to 10
func hardCodeManyHook(from reflect.Value, to reflect.Value) (any, error) {
	// A slice or array could be behind pointers. mapstructure could add an extra pointer level too.
	for to.Kind() == reflect.Pointer {
		to = to.Elem()
	}

	for from.Kind() == reflect.Pointer {
		from = from.Elem()
	}

	switch to.Kind() {
	case reflect.Slice, reflect.Array:
		switch from.Kind() {
		case reflect.Slice, reflect.Array:
			return from.Interface(), nil
		default:
		}
	default:
		return from.Interface(), nil
	}

	length := to.Len()
	array := reflect.MakeSlice(reflect.SliceOf(from.Type()), length, length)
	for i := range length {
		array.Index(i).Set(from)
	}
	return array.Interface(), nil
}
