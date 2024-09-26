package codec

import (
	"fmt"
	"reflect"

	"golang.org/x/crypto/sha3"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type AddressLength int

const (
	Byte20Address AddressLength = 20
)

type AddressChecksum func([]byte) []byte

func EIP55AddressChecksum(a []byte) []byte {
	buf := a

	// compute checksum
	sha := sha3.NewLegacyKeccak256()
	sha.Write(buf[2:])
	hash := sha.Sum(nil)

	for i := 2; i < len(buf); i++ {
		hashByte := hash[(i-2)/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if buf[i] > '9' && hashByte > 7 {
			buf[i] -= 32
		}
	}

	return buf[:]
}

func NoChecksum(a []byte) []byte {
	return a
}

func NewAddressBytesToStringModifier(length AddressLength, checksum AddressChecksum, fields []string) Modifier {
	fieldMap := map[string]bool{}
	for _, field := range fields {
		fieldMap[field] = true
	}

	m := &bytesToStringModifier{
		length:   length,
		checksum: checksum,
		modifierBase: modifierBase[bool]{
			fields:           fieldMap,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, _ bool) error {
		t, err := convertBytesToString(field.Type, field.Name, int(m.length))
		if err != nil {
			return err
		}
		field.Type = t
		return nil
	}

	return m
}

type bytesToStringModifier struct {
	length   AddressLength
	checksum AddressChecksum
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

	fmt.Printf("\nRetypeToOffChain called with type: %v, itemType: %T", onChainType, itemType)

	if len(t.fields) == 0 {
		t.offToOnChainType[onChainType] = onChainType
		t.onToOffChainType[onChainType] = onChainType
		return onChainType, nil
	}

	if cached, ok := t.onToOffChainType[onChainType]; ok {
		return cached, nil
	}

	addrType, err := typeFromAddressLength(t.length)
	if err != nil {
		return nil, err
	}

	switch onChainType.Kind() {
	case reflect.Pointer:
		fmt.Printf("\nPointer detected: %v, element type: %v", onChainType, onChainType.Elem())

		// Recursively call RetypeToOffChain on the element type
		elm, err := t.RetypeToOffChain(onChainType.Elem(), "")
		if err != nil {
			fmt.Printf("\nError while processing pointer element type: %v", err)
			return nil, err
		}

		ptr := reflect.PointerTo(elm)
		fmt.Printf("\nPointer transformed: %v -> %v", onChainType, ptr)

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
		if onChainType == addrType {
			offChainType = reflect.TypeOf("")
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

func (t *bytesToStringModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	return transformWithMaps(offChainValue, t.offToOnChainType, t.fields, noop, addressToStringHook(t.length, t.checksum))
}

func (t *bytesToStringModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields,
		addressTransformationAction[bool](t.length),
		addressToStringHook(t.length, t.checksum),
	)
}

func addressTransformationAction[T any](length AddressLength) mapAction[T] {
	return func(em map[string]any, fieldName string, value T) error {
		// Check if the field exists in the map
		if val, ok := em[fieldName]; ok {
			// Use reflection to extract the underlying value
			rVal := reflect.ValueOf(val)
			if rVal.Kind() == reflect.Ptr {
				// Dereference the pointer if necessary
				rVal = reflect.Indirect(rVal)
			}

			// Get the expected type from the AddressLength (e.g., [length]byte)
			expectedType, err := typeFromAddressLength(length)
			if err != nil {
				return err
			}

			//	Handle type alias that are convertible to the expected type
			// eg. type addressType [codec.Byte20Address]byte converts to [20]byte
			if rVal.Type().ConvertibleTo(expectedType) {
				rVal = rVal.Convert(expectedType)
			}

			switch rVal.Kind() {
			case reflect.Array:
				// Handle outer arrays (e.g., [n][length]byte)
				if rVal.Type().Elem().Kind() == reflect.Array && rVal.Type().Elem().Len() == int(length) {
					// Create a new array of the correct size to store the converted elements
					addressArray := reflect.New(reflect.ArrayOf(rVal.Len(), expectedType)).Elem()

					// Convert each element from [length]byte to the expected type
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != int(length) {
							return fmt.Errorf("expected [%v]byte but got length %d for element %d", length, elem.Len(), i)
						}
						reflect.Copy(addressArray.Index(i), elem)
					}

					// Replace the field in the map with the converted array
					em[fieldName] = addressArray.Interface()
					fmt.Printf("Converted field '%s' to array of %s: %v\n", fieldName, expectedType, addressArray.Interface())
				} else if rVal.Type() == expectedType {
					// Handle a single array (e.g., [length]byte)
					fmt.Printf("Single array detected: %v\n", rVal.Type())
					addressVal := reflect.New(expectedType).Elem()
					reflect.Copy(addressVal, rVal)

					// Replace the field in the map with the converted array
					em[fieldName] = addressVal.Interface()
					fmt.Printf("Converted field '%s' to %s: %x\n", fieldName, expectedType, addressVal.Interface())
				} else {
					fmt.Printf("Unexpected array type: %v, expected: %v\n", rVal.Type(), expectedType)
					return fmt.Errorf("expected [%v]byte or array of [%v]byte but got %v for field %s", length, length, rVal.Type(), fieldName)
				}

			case reflect.Slice:
				fmt.Printf("Slice detected: %v\n", rVal.Type())

				// Handle slices of byte arrays (e.g., [][][length]byte)
				if rVal.Len() > 0 && rVal.Index(0).Type() == expectedType {
					// Create a slice of the expected type
					addressSlice := reflect.MakeSlice(reflect.SliceOf(expectedType), rVal.Len(), rVal.Len())

					// Convert each element of the slice
					for i := 0; i < rVal.Len(); i++ {
						elem := rVal.Index(i)
						if elem.Len() != int(length) {
							return fmt.Errorf("expected element of [%v]byte but got length %d at index %d", length, elem.Len(), i)
						}

						// Copy the element to the new slice
						reflect.Copy(addressSlice.Index(i), elem)
					}

					// Replace the field in the map with the converted slice
					em[fieldName] = addressSlice.Interface()
					fmt.Printf("Converted field '%s' to slice of %v: %v\n", fieldName, expectedType, addressSlice.Interface())
				} else {
					return fmt.Errorf("expected slice of [%v]byte arrays but got %v for field %s", length, rVal.Type(), fieldName)
				}

			default:
				return fmt.Errorf("unexpected type %v for field %s", rVal.Kind(), fieldName)
			}
		}
		return nil
	}
}

func convertBytesToString(t reflect.Type, field string, length int) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Array:
		fmt.Printf("\nArray detected: %v, field: %s", t, field)
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			fmt.Printf("\nConverting array of [%v]byte to array of strings: %v", length, t)
			return reflect.ArrayOf(t.Len(), reflect.TypeOf("")), nil
		}
		fmt.Printf("\nConverting [20]byte to string: %v", t)
		return reflect.TypeOf(""), nil

	case reflect.Pointer:
		fmt.Printf("\nPointer detected: %v, field: %s", t, field)
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			fmt.Printf("\nConverting pointer to [%v]byte to string: %v", length, t)
			return reflect.TypeOf(""), nil
		}
		fmt.Printf("\nPointer is not pointing to an array of [%v]byte, returning error", length)
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	case reflect.Slice:
		fmt.Printf("\nSlice detected: %v, field: %s", t, field)
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			fmt.Printf("\nConverting slice of [%v]byte to slice of strings: %v", length, t)
			return reflect.SliceOf(reflect.TypeOf("")), nil
		}
		fmt.Printf("\nSlice does not contain [%v]byte elements, returning error", length)
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	default:
		fmt.Printf("\nUnsupported type detected: %v, field: %s", t, field)
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)
	}
}
