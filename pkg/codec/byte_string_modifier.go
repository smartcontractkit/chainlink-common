package codec

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/crypto/sha3"

	"github.com/mr-tron/base58"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type AddressLength int
type ChecksumType string
type EncodingType string

const (
	// EVM
	Byte20Address AddressLength = 20
	EIP55         ChecksumType  = "eip55"
	HexEncoding   EncodingType  = "hex"
	// Solana
	Byte32Address  AddressLength = 32
	Base58Encoding EncodingType  = "base58"
	// General
	NoneChecksum ChecksumType = "none"
)

type AddressChecksum func([]byte) []byte

// EIP55AddressChecksum Applies EIP55 checksum logic to the address
// assuming input 'a' is already in hex form without the "0x" prefix
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

// NoChecksum is a checksum function that returns the input bytes unchanged.
func NoChecksum(a []byte) []byte {
	return a
}

// NewAddressBytesToStringModifier creates a new modifier that converts byte arrays to strings.
// It uses the specified address length, checksum, fields, and encoding.
func NewAddressBytesToStringModifier(length AddressLength, checksum AddressChecksum, fields []string, encoding EncodingType) Modifier {
	fieldMap := map[string]bool{}
	for _, field := range fields {
		fieldMap[field] = true
	}

	m := &bytesToStringModifier{
		length:   length,
		checksum: checksum,
		encoding: encoding,
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
	encoding EncodingType
	modifierBase[bool]
}

// getChecksumFunction returns the checksum function based on the ChecksumType.
func getChecksumFunction(checksumTy ChecksumType) (AddressChecksum, error) {
	switch checksumTy {
	case EIP55:
		return EIP55AddressChecksum, nil
	case NoneChecksum:
		return NoChecksum, nil
	default:
		return nil, fmt.Errorf("checksum type unavailable: %s", checksumTy)
	}
}

// getAddressLength returns the AddressLength based on the input value.
func getAddressLength(length AddressLength) (AddressLength, error) {
	switch length {
	case Byte20Address:
		return Byte20Address, nil
	case Byte32Address:
		return Byte32Address, nil
	default:
		return 0, fmt.Errorf("address length unavailable: %d", length)
	}
}

// getEncodingType returns the EncodingType based on the input value.
func getEncodingType(encodingTy EncodingType) (EncodingType, error) {
	switch encodingTy {
	case HexEncoding:
		return HexEncoding, nil
	case Base58Encoding:
		return Base58Encoding, nil
	default:
		return "", fmt.Errorf("unsupported encoding type: %s", encodingTy)
	}
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

	addrType, err := typeFromAddressLength(t.length)
	if err != nil {
		return nil, err
	}

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
	return transformWithMaps(offChainValue, t.offToOnChainType, t.fields, noop, stringToAddressHookForOnChain(t.length, t.encoding))
}

func (t *bytesToStringModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields,
		addressTransformationAction[bool](t.length),
		addressToStringHookForOffChain(t.length, t.encoding, t.checksum),
	)
}

// addressTransformationAction performs conversions over the fields we want to modify.
// It handles byte arrays, ensuring they are convertible to the expected custom type AddressLength.
// It then replaces the field in the map with the transformed value.
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

			// Handle type alias that are convertible to the expected type
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
				} else if rVal.Type() == expectedType {
					// Handle a single array (e.g., [length]byte)
					addressVal := reflect.New(expectedType).Elem()
					reflect.Copy(addressVal, rVal)

					// Replace the field in the map with the converted array
					em[fieldName] = addressVal.Interface()
				} else {
					return fmt.Errorf("expected [%v]byte or array of [%v]byte but got %v for field %s", length, length, rVal.Type(), fieldName)
				}

			case reflect.Slice:
				// Handle slices of byte arrays (e.g., [][length]byte)
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

// convertBytesToString converts a byte array, pointer, or slice type to a string type for a given field.
// This function inspects the kind of the input type (array, pointer, slice) and performs the conversion
// if the element type matches the specified byte array length. Returns an error if the conversion is not possible.
func convertBytesToString(t reflect.Type, field string, length int) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return convertBytesToString(t.Elem(), field, length)

	case reflect.Array:
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			return reflect.ArrayOf(t.Len(), reflect.TypeOf("")), nil
		}
		return reflect.TypeOf(""), nil

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == length {
			return reflect.SliceOf(reflect.TypeOf("")), nil
		}
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	default:
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)
	}
}

// addressToStringHookForOffChain converts byte arrays to their string representation for off-chain use.
// It handles different encodings (e.g., hex, base58) and applies a checksum if provided.
func addressToStringHookForOffChain(length AddressLength, encoding EncodingType, checksum func([]byte) []byte) func(from reflect.Type, to reflect.Type, data any) (any, error) {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		byteArrTyp, err := typeFromAddressLength(length)
		if err != nil {
			return nil, err
		}

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

			encoded, err := encodeAddressWithChecksum(bts, encoding, checksum)
			if err != nil {
				return nil, err
			}

			return encoded, nil
		}

		return data, nil
	}
}

// stringToAddressHookForOnChain converts a string representation of an address back into a byte array for on-chain use.
// It decodes the address using the specified encoding (e.g., hex, base58) and verifies the length.
func stringToAddressHookForOnChain(length AddressLength, encoding EncodingType) func(from reflect.Type, to reflect.Type, data any) (any, error) {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		byteArrTyp, err := typeFromAddressLength(length)
		if err != nil {
			return nil, err
		}

		strTyp := reflect.TypeOf("")

		// Convert from string to byte array (e.g., string -> [20]byte)
		if from == strTyp && (to == byteArrTyp || to.ConvertibleTo(byteArrTyp)) {
			addr := data.(string)

			// Avoid potential panic when the address is empty
			if len(addr) == 0 {
				return nil, fmt.Errorf("empty address")
			}

			// Decode the address based on encoding type
			bts, err := decodeAddress(addr, encoding, length)
			if err != nil {
				return nil, err
			}

			// Ensure the byte length matches the expected AddressLength
			if len(bts) != int(length) {
				return nil, fmt.Errorf("length mismatch: expected %d bytes, got %d", length, len(bts))
			}

			// Create a new array of the desired type and fill it with the decoded bytes
			val := reflect.New(byteArrTyp).Elem()
			reflect.Copy(val, reflect.ValueOf(bts))

			return val.Interface(), nil
		}

		return data, nil
	}
}

// decodeAddress decodes a string address into a byte array based on the specified encoding (e.g., hex or base58).
// It also checks if the decoded byte array matches the expected address length.
func decodeAddress(addr string, encoding EncodingType, length AddressLength) ([]byte, error) {
	switch encoding {
	case Base58Encoding:
		bts, err := base58.Decode(addr)
		if err != nil {
			return nil, err
		}

		if len(bts) != int(length) {
			return nil, fmt.Errorf("length mismatch: expected %d bytes, got %d", length, len(bts))
		}
		return bts, nil

	case HexEncoding:
		bts, err := hex.DecodeString(addr[2:])
		if err != nil {
			return nil, err
		}

		if len(bts) != int(length) {
			return nil, fmt.Errorf("length mismatch: expected %d bytes, got %d", length, len(bts))
		}
		return bts, nil

	default:
		return nil, fmt.Errorf("unsupported encoding type: %v", encoding)
	}
}

// encodeAddressWithChecksum encodes a byte array into a string based on the specified encoding (e.g., hex or base58).
// It also applies a checksum function to the byte array before returning the encoded result.
func encodeAddressWithChecksum(bts []byte, encoding EncodingType, checksum func([]byte) []byte) (string, error) {
	switch encoding {
	case Base58Encoding:
		return base58.Encode(checksum(bts)), nil
	case HexEncoding:
		encoded := "0x" + hex.EncodeToString(bts)
		return string(checksum([]byte(encoded))), nil
	default:
		return "", fmt.Errorf("unsupported encoding type: %v", encoding)
	}
}

// typeFromAddressLength returns the reflect.Type corresponding to the given AddressLength (e.g., [20]byte for Byte20Address).
func typeFromAddressLength(length AddressLength) (reflect.Type, error) {
	switch length {
	case Byte20Address:
		return reflect.TypeOf([Byte20Address]byte{}), nil
	case Byte32Address:
		return reflect.TypeOf([Byte32Address]byte{}), nil
	default:
		return nil, errors.New("address length not available")
	}
}
