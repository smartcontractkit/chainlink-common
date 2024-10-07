package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// AddressModifier defines the interface for encoding, decoding, and handling addresses.
// This interface allows for chain-specific logic to be injected into the modifier without
// modifying the common repository, making the implementation flexible and adaptable for
// different blockchain ecosystems (e.g., Ethereum, Solana).
type AddressModifier interface {
	// Converts a byte array representing an address into its string form.
	EncodeAddress([]byte) (string, error)
	// Converts a string representation of an address back into its byte array form.
	DecodeAddress(string) ([]byte, error)
	// Returns the expected byte length of the address for the specific chain.
	Length() int
}

// NewAddressBytesToStringModifier creates and returns a new modifier that transforms address byte
// arrays to their corresponding string representation (or vice versa) based on the provided
// AddressModifier.
//
// The fields parameter specifies which fields within a struct should be modified. The AddressModifier
// is injected into the modifier to handle chain-specific logic (e.g., different encoding schemes for
// Ethereum or Solana addresses). This happens during the contractReader relayer configuration.
func NewAddressBytesToStringModifier(fields []string, modifier AddressModifier) Modifier {
	fieldMap := map[string]bool{}
	for _, field := range fields {
		fieldMap[field] = true
	}

	m := &bytesToStringModifier{
		modifier: modifier,
		modifierBase: modifierBase[bool]{
			fields:           fieldMap,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}

	// Modify field for input using the modifier to convert the byte array to string
	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ bool) error {
		t, err := convertBytesToString(field.Type, field.Name, modifier.Length())
		if err != nil {
			return err
		}
		field.Type = t
		return nil
	}

	return m
}

type bytesToStringModifier struct {
	// Injected modifier that contains chain-specific logic
	modifier AddressModifier
	modifierBase[bool]
}

func (t *bytesToStringModifier) RetypeToOffChain(onChainType reflect.Type, itemType string) (tpe reflect.Type, err error) {
	defer func() {
		// StructOf can panic if the fields are not valid
		if r := recover(); r != nil {
			tpe = nil
			err = fmt.Errorf("%w: %v", types.ErrInvalidType, r)
		}
	}()

	if len(t.fields) == 0 {
		t.offToOnChainType[onChainType] = onChainType
		t.onToOffChainType[onChainType] = onChainType
		return onChainType, nil
	}

	if cached, ok := t.onToOffChainType[onChainType]; ok {
		return cached, nil
	}

	addrType := reflect.ArrayOf(t.modifier.Length(), reflect.TypeOf(byte(0)))

	switch onChainType.Kind() {
	case reflect.Pointer:
		// Recursively call RetypeToOffChain on the element type
		elm, err := t.RetypeToOffChain(onChainType.Elem(), "")
		if err != nil {
			return nil, err
		}

		ptr := reflect.PointerTo(elm)
		t.onToOffChainType[onChainType] = ptr
		t.offToOnChainType[ptr] = onChainType
		return ptr, nil
	case reflect.Slice:
		elm, err := t.RetypeToOffChain(onChainType.Elem(), "")
		if err != nil {
			return nil, err
		}

		sliceType := reflect.SliceOf(elm)
		t.onToOffChainType[onChainType] = sliceType
		t.offToOnChainType[sliceType] = onChainType
		return sliceType, nil
	case reflect.Array:
		var offChainType reflect.Type
		// Check for nested byte arrays (e.g., [20]byte)
		if onChainType.Elem() == addrType.Elem() {
			offChainType = reflect.ArrayOf(onChainType.Len(), reflect.TypeOf(""))
		} else {
			elm, err := t.RetypeToOffChain(onChainType.Elem(), "")
			if err != nil {
				return nil, err
			}

			offChainType = reflect.ArrayOf(onChainType.Len(), elm)
		}

		t.onToOffChainType[onChainType] = offChainType
		t.offToOnChainType[offChainType] = onChainType

		return offChainType, nil
	case reflect.Struct:
		return t.getStructType(onChainType)
	default:
		return nil, fmt.Errorf("%w: cannot retype the kind %v", types.ErrInvalidType, onChainType.Kind())
	}
}

// TransformToOnChain uses the AddressModifier for string-to-address conversion.
func (t *bytesToStringModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	return transformWithMaps(offChainValue, t.offToOnChainType, t.fields, noop, stringToAddressHookForOnChain(t.modifier))
}

// TransformToOffChain uses the AddressModifier for address-to-string conversion.
func (t *bytesToStringModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields,
		addressTransformationAction[bool](t.modifier.Length()),
		addressToStringHookForOffChain(t.modifier),
	)
}

// addressTransformationAction performs conversions over the fields we want to modify.
// It handles byte arrays, ensuring they are convertible to the expected length.
// It then replaces the field in the map with the transformed value.
func addressTransformationAction[T any](length int) mapAction[T] {
	return func(em map[string]any, fieldName string, value T) error {
		// Check if the field exists in the map
		if val, ok := em[fieldName]; ok {
			// Use reflection to extract the underlying value
			rVal := reflect.ValueOf(val)
			if rVal.Kind() == reflect.Ptr {
				// Dereference the pointer if necessary
				rVal = reflect.Indirect(rVal)
			}

			expectedType := reflect.ArrayOf(length, reflect.TypeOf(byte(0)))

			// Handle type alias that are convertible to the expected type
			if rVal.Type().ConvertibleTo(expectedType) {
				rVal = rVal.Convert(expectedType)
			}

			switch rVal.Kind() {
			case reflect.Array:
				// Handle outer arrays (e.g., [n][length]byte)
				if rVal.Type().Elem().Kind() == reflect.Array && rVal.Type().Elem().Len() == length {
					// Create a new array of the correct size to store the converted elements
					addressArray := reflect.New(reflect.ArrayOf(rVal.Len(), expectedType)).Elem()
					// Convert each element from [length]byte to the expected type
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != length {
							return fmt.Errorf("expected [%v]byte but got length %d for element %d", length, elem.Len(), i)
						}
						reflect.Copy(addressArray.Index(i), elem)
					}
					// Replace the field in the map with the converted array
					em[fieldName] = addressArray.Interface()
				} else if rVal.Type() == expectedType {
					// Handle a single array (e.g., [length]byte)
					addressVal := reflect.New(expectedType).Elem()
					reflect.Copy(addressVal, rVal)
					// Replace the field in the map with the converted array
					em[fieldName] = addressVal.Interface()
				} else {
					return fmt.Errorf("expected [%v]byte but got %v for field %s", length, rVal.Type(), fieldName)
				}
			case reflect.Slice:
				// Handle slices of byte arrays (e.g., [][length]byte)
				if rVal.Len() > 0 && rVal.Index(0).Type() == expectedType {
					// Create a slice of the expected type
					addressSlice := reflect.MakeSlice(reflect.SliceOf(expectedType), rVal.Len(), rVal.Len())
					// Convert each element of the slice
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != length {
							return fmt.Errorf("expected element of [%v]byte but got length %d at index %d", length, elem.Len(), i)
						}
						reflect.Copy(addressSlice.Index(i), elem)
					}
					// Replace the field in the map with the converted slice
					em[fieldName] = addressSlice.Interface()
				} else {
					return fmt.Errorf("expected slice of [%v]byte but got %v for field %s", length, rVal.Type(), fieldName)
				}
			default:
				return fmt.Errorf("unexpected type %v for field %s", rVal.Kind(), fieldName)
			}
		}
		return nil
	}
}

// convertBytesToString converts a byte array, pointer, or slice type to a string type for a given field.
// This function inspects the kind of the input type (array, pointer, slice) and performs the conversion
// if the element type matches the specified byte array length. Returns an error if the conversion is not possible.
func convertBytesToString(t reflect.Type, field string, length int) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return convertBytesToString(t.Elem(), field, length)

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
			addr := data.(string)

			// Decode the address based on encoding type
			bts, err := modifier.DecodeAddress(addr)
			if err != nil {
				return nil, err
			}

			// Ensure the byte length matches the expected length
			if len(bts) != modifier.Length() {
				return nil, fmt.Errorf("length mismatch: expected %d bytes, got %d", modifier.Length(), len(bts))
			}

			// Create a new array of the desired type and fill it with the decoded bytes
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

		// if 'from' is a pointer to the byte array (e.g., *[20]byte), dereference it.
		if from.Kind() == reflect.Ptr && from.Elem() == byteArrTyp {
			data = reflect.ValueOf(data).Elem().Interface()
			from = from.Elem()
		}

		// Convert from byte array to string (e.g., [20]byte -> string)
		if from.ConvertibleTo(byteArrTyp) && to == strTyp {
			val := reflect.ValueOf(data)
			bts := make([]byte, val.Len())

			for i := 0; i < val.Len(); i++ {
				bts[i] = byte(val.Index(i).Uint())
			}

			encoded, err := modifier.EncodeAddress(bts)
			if err != nil {
				return nil, err
			}

			return encoded, nil
		}
		return data, nil
	}
}
