package codec

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ExampleAddressModifier struct{}

func (m *ExampleAddressModifier) EncodeAddress(bts []byte) (string, error) {
	if len(bts) > 32 {
		return "", errors.New("upexpected address byte length")
	}

	normalized := make([]byte, 32)

	// apply byts as big endian
	copy(normalized[:], bts[:])

	return base64.StdEncoding.EncodeToString(normalized), nil
}

func (m *ExampleAddressModifier) DecodeAddress(str string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	if len(decodedBytes) != 32 {
		return nil, errors.New("unexpected address byte length")
	}

	return decodedBytes, nil
}

func (m *ExampleAddressModifier) Length() int {
	return 32
}

// AddressModifier defines the interface for encoding, decoding, and handling addresses.
// This interface allows for chain-specific logic to be injected into the modifier without
// modifying the common repository.
type AddressModifier interface {
	// EncodeAddress converts byte array representing an address into its string form using chain-specific logic.
	EncodeAddress([]byte) (string, error)
	// DecodeAddress converts a string representation of an address back into its byte array form using chain-specific logic.
	DecodeAddress(string) ([]byte, error)
	// Length returns the expected byte length of the address for the specific chain.
	Length() int
}

// NewAddressBytesToStringModifier creates and returns a new modifier that transforms address byte
// arrays to their corresponding string representation (or vice versa) based on the provided
// AddressModifier.
//
// The fields parameter specifies which fields within a struct should be modified. The AddressModifier
// is injected into the modifier to handle chain-specific logic during the contractReader relayer configuration.
func NewAddressBytesToStringModifier(
	fields []string,
	modifier AddressModifier,
) Modifier {
	return NewPathTraverseAddressBytesToStringModifier(fields, modifier, false)
}

func NewPathTraverseAddressBytesToStringModifier(
	fields []string,
	modifier AddressModifier,
	enablePathTraverse bool,
) Modifier {
	// bool is a placeholder value
	fieldMap := map[string]bool{}
	for _, field := range fields {
		fieldMap[field] = true
	}

	m := &bytesToStringModifier{
		modifier: modifier,
		modifierBase: modifierBase[bool]{
			enablePathTraverse: enablePathTraverse,
			fields:             fieldMap,
			onToOffChainType:   map[reflect.Type]reflect.Type{},
			offToOnChainType:   map[reflect.Type]reflect.Type{},
		},
	}

	// Modify field for input using the modifier to convert the byte array to string
	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ bool) error {
		t, err := createStringTypeForBytes(field.Type, field.Name, modifier.Length())
		if err != nil {
			return err
		}
		field.Type = t
		return nil
	}

	return m
}

func NewConstrainedLengthBytesToStringModifier(
	fields []string,
	maxLen int,
) Modifier {
	return NewPathTraverseAddressBytesToStringModifier(fields, &constrainedLengthBytesToStringModifier{maxLen: maxLen}, false)
}

func NewPathTraverseConstrainedLengthBytesToStringModifier(
	fields []string,
	maxLen int,
	enablePathTraverse bool,
) Modifier {
	return NewPathTraverseAddressBytesToStringModifier(fields, &constrainedLengthBytesToStringModifier{maxLen: maxLen}, enablePathTraverse)
}

type constrainedLengthBytesToStringModifier struct {
	maxLen int
}

func (m constrainedLengthBytesToStringModifier) EncodeAddress(bts []byte) (string, error) {
	return string(bytes.Trim(bts, "\x00")), nil
}

func (m constrainedLengthBytesToStringModifier) DecodeAddress(str string) ([]byte, error) {
	output := make([]byte, m.maxLen)

	copy(output, []byte(str)[:])

	return output, nil
}

func (m constrainedLengthBytesToStringModifier) Length() int {
	return m.maxLen
}

type bytesToStringModifier struct {
	// Injected modifier that contains chain-specific logic
	modifier AddressModifier
	modifierBase[bool]
}

func (m *bytesToStringModifier) RetypeToOffChain(onChainType reflect.Type, _ string) (tpe reflect.Type, err error) {
	defer func() {
		// StructOf can panic if the fields are not valid
		if r := recover(); r != nil {
			tpe = nil
			err = fmt.Errorf("%w: %v", types.ErrInvalidType, r)
		}
	}()

	// Attempt to retype using the shared functionality in modifierBase
	offChainType, err := m.modifierBase.RetypeToOffChain(onChainType, "")
	if err != nil {
		// Handle additional cases specific to bytesToStringModifier
		if onChainType.Kind() == reflect.Array {
			addrType := reflect.ArrayOf(m.modifier.Length(), reflect.TypeOf(byte(0)))
			// Check for nested byte arrays (e.g., [n][20]byte)
			if onChainType.Elem() == addrType.Elem() {
				return reflect.ArrayOf(onChainType.Len(), reflect.TypeOf("")), nil
			}
		}
	}

	return offChainType, err
}

// TransformToOnChain uses the AddressModifier for string-to-address conversion.
func (m *bytesToStringModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	offChainValue, itemType, err := m.modifierBase.selectType(offChainValue, m.offChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := transformWithMaps(offChainValue, m.offToOnChainType, m.fields, noop, stringToAddressHookForOnChain(m.modifier))
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

// TransformToOffChain uses the AddressModifier for address-to-string conversion.
func (m *bytesToStringModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	onChainValue, itemType, err := m.modifierBase.selectType(onChainValue, m.onChainStructType, itemType)
	if err != nil {
		return nil, err
	}

	modified, err := transformWithMaps(onChainValue, m.onToOffChainType, m.fields,
		addressTransformationAction(m.modifier.Length()),
		addressToStringHookForOffChain(m.modifier),
	)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return ValueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
}

// addressTransformationAction performs conversions over the fields we want to modify.
// It handles byte arrays, ensuring they are convertible to the expected length.
// It then replaces the field in the map with the transformed value.
func addressTransformationAction(length int) func(extractMap map[string]any, key string, _ bool) error {
	return func(em map[string]any, fieldName string, _ bool) error {
		if val, ok := em[fieldName]; ok {
			rVal := reflect.ValueOf(val)

			if !rVal.IsValid() {
				return fmt.Errorf("invalid value for field %s", fieldName)
			}

			if rVal.Kind() == reflect.Ptr && !rVal.IsNil() {
				rVal = reflect.Indirect(rVal)
			}

			expectedType := reflect.ArrayOf(length, reflect.TypeOf(byte(0)))
			if rVal.Type().ConvertibleTo(expectedType) {
				if !rVal.CanConvert(expectedType) {
					return fmt.Errorf("cannot convert type %v to expected type %v for field %s", rVal.Type(), expectedType, fieldName)
				}
				rVal = rVal.Convert(expectedType)
			}

			switch rVal.Kind() {
			case reflect.Array:
				// Handle outer arrays (e.g., [n][length]byte)
				if rVal.Type().Elem().Kind() == reflect.Array && rVal.Type().Elem().Len() == length {
					addressArray := reflect.New(reflect.ArrayOf(rVal.Len(), expectedType)).Elem()
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != length {
							return fmt.Errorf("expected [%d]byte but got length %d for element %d in field %s", length, elem.Len(), i, fieldName)
						}
						reflect.Copy(addressArray.Index(i), elem)
					}
					em[fieldName] = addressArray.Interface()
				} else if rVal.Type() == expectedType {
					// Handle a single array (e.g., [length]byte)
					addressVal := reflect.New(expectedType).Elem()
					reflect.Copy(addressVal, rVal)
					em[fieldName] = addressVal.Interface()
				} else {
					return fmt.Errorf("expected [%d]byte but got %v for field %s", length, rVal.Type(), fieldName)
				}
			case reflect.Slice:
				// Handle slices of byte arrays (e.g., [][length]byte)
				if rVal.Len() > 0 && rVal.Index(0).Type() == expectedType {
					addressSlice := reflect.MakeSlice(reflect.SliceOf(expectedType), rVal.Len(), rVal.Len())
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != length {
							return fmt.Errorf("expected element of [%d]byte but got length %d at index %d for field %s", length, elem.Len(), i, fieldName)
						}
						reflect.Copy(addressSlice.Index(i), elem)
					}
					em[fieldName] = addressSlice.Interface()
				} else {
					return fmt.Errorf("expected slice of [%d]byte but got %v for field %s", length, rVal.Type(), fieldName)
				}
			default:
				return fmt.Errorf("unexpected type %v for field %s", rVal.Kind(), fieldName)
			}
		}
		return nil
	}
}

// createStringTypeForBytes converts a byte array, pointer, or slice type to a string type for a given field.
// This function inspects the kind of the input type (array, pointer, slice) and performs the conversion
// if the element type matches the specified byte array length. Returns an error if the conversion is not possible.
func createStringTypeForBytes(t reflect.Type, field string, length int) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return createStringTypeForBytes(t.Elem(), field, length)

	case reflect.Array:
		// Handle arrays, convert array of bytes to array of strings
		if t.Elem().Kind() == reflect.Uint8 && t.Len() == length {
			return reflect.TypeOf(""), nil
		} else if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			// Handle nested arrays (e.g., [2][20]byte to [2]string)
			return reflect.ArrayOf(t.Len(), reflect.TypeOf("")), nil
		}
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	case reflect.Slice:
		// Handle slices of byte arrays, convert to slice of strings
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			return reflect.SliceOf(reflect.TypeOf("")), nil
		}
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	default:
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)
	}
}

// stringToAddressHookForOnChain converts a string representation of an address back into a byte array for on-chain use.
func stringToAddressHookForOnChain(modifier AddressModifier) func(from reflect.Type, to reflect.Type, data any) (any, error) {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		byteArrTyp := reflect.ArrayOf(modifier.Length(), reflect.TypeOf(byte(0)))
		strTyp := reflect.TypeOf("")

		// Convert from string to byte array (e.g., string -> [20]byte)
		if from == strTyp && (to == byteArrTyp || to.ConvertibleTo(byteArrTyp)) {
			addr, ok := data.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type: expected string but got %T", data)
			}

			bts, err := modifier.DecodeAddress(addr)
			if err != nil {
				return nil, err
			}

			if len(bts) != modifier.Length() {
				return nil, fmt.Errorf("length mismatch: expected %d bytes, got %d", modifier.Length(), len(bts))
			}

			val := reflect.New(byteArrTyp).Elem()
			reflect.Copy(val, reflect.ValueOf(bts))
			return val.Interface(), nil
		}
		return data, nil
	}
}

// addressToStringHookForOffChain converts byte arrays to their string representation for off-chain use.
func addressToStringHookForOffChain(modifier AddressModifier) func(from reflect.Type, to reflect.Type, data any) (any, error) {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		byteArrTyp := reflect.ArrayOf(modifier.Length(), reflect.TypeOf(byte(0)))
		strTyp := reflect.TypeOf("")
		rVal := reflect.ValueOf(data)

		if !reflect.ValueOf(data).IsValid() {
			return nil, fmt.Errorf("invalid value for conversion: got %T", data)
		}

		// Convert from byte array to string (e.g., [20]byte -> string)
		if from.ConvertibleTo(byteArrTyp) && to == strTyp {
			bts := make([]byte, rVal.Len())
			for i := 0; i < rVal.Len(); i++ {
				bts[i] = byte(rVal.Index(i).Uint())
			}

			encoded, err := modifier.EncodeAddress(bts)
			if err != nil {
				return nil, fmt.Errorf("failed to encode address: %w", err)
			}

			return encoded, nil
		}
		return data, nil
	}
}
