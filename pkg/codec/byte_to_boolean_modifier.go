package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewBoolToByteModifier converts on-chain uint8 to off-chain bool and vice versa.
func NewBoolToByteModifier(fields []string) Modifier {
	fieldMap := map[string]any{}
	for _, field := range fields {
		fieldMap[field] = true
	}

	m := &boolToByteModifier{
		modifierBase[any]{
			fields:           fieldMap,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ any) error {
		t, err := convertUint8InTypeToBool(field.Type, field.Name)
		if err != nil {
			return err
		}
		field.Type = t
		return nil
	}

	return m
}

// boolToByteModifier implements the two-phase transform logic using a helper hook.
type boolToByteModifier struct {
	modifierBase[any]
}

func (m *boolToByteModifier) TransformToOnChain(offChainValue any, _ string) (any, error) {
	return transformWithMaps(offChainValue, m.offToOnChainType, m.fields, boolToByteHook)
}

func (m *boolToByteModifier) TransformToOffChain(onChainValue any, _ string) (any, error) {
	return transformWithMaps(onChainValue, m.onToOffChainType, m.fields, byteToBoolHook)
}

// convertUint8InTypeToBool examines the type `t`. If it is (or can be unwrapped to) a uint8,
// we convert it to bool. If it's a slice/array of uint8, convert to a slice/array of bool, etc.
func convertUint8InTypeToBool(t reflect.Type, field string) (reflect.Type, error) {
	if t.Kind() == reflect.Uint8 {
		return reflect.TypeOf(true), nil
	}

	switch t.Kind() {
	case reflect.Pointer:
		t = t.Elem()
		if t.Kind() == reflect.Uint8 {
			return reflect.PointerTo(reflect.TypeOf(true)), nil
		}
		fallthrough
	default:
		return nil, fmt.Errorf("%w: cannot convert bool for field %s", types.ErrInvalidType, field)
	}
}

// boolToByteHook is the transform hook used in `transformWithMaps`. It
// transforms a `bool` into `uint8`. If you find a bool and the field is one of
// your target fields, you convert it to 0 or 1. If not, you pass through unchanged.
func boolToByteHook(m map[string]any, key string, _ any) error {
	var toReturn uint8
	val := m[key]
	if b, ok := val.(bool); ok {
		if b {
			toReturn = 1
			m[key] = toReturn
		} else {
			toReturn = 0
			m[key] = toReturn
		}
		return nil
	} else if bptr, ok := val.(*bool); ok && bptr != nil {
		if *bptr {
			toReturn = 1
			m[key] = &toReturn
		} else {
			toReturn = 0
			m[key] = &toReturn
		}
		return nil
	}

	return fmt.Errorf("%w: cannot convert bool to byte for field", types.ErrInvalidType)
}

// byteToBoolHook is the transform hook used in `transformWithMaps`. It
// transforms a `uint8` into a `bool`. If the byte is nonzero → true, else false.
func byteToBoolHook(m map[string]any, key string, _ any) error {
	val := m[key]
	switch x := val.(type) {
	case uint8:
		m[key] = x != 0
	case *uint8:
		if x != nil && *x != 0 {
			truth := true
			m[key] = &truth
		} else {
			falsity := false
			m[key] = &falsity
		}
	default:
		return fmt.Errorf("%w: cannot convert byte to bool for field %s", types.ErrInvalidType, key)
	}

	return nil
}
