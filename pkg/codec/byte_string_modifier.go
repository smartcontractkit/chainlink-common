package codec

import (
	"fmt"
	"log"
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
		t, err := convertBytesToString(field.Type, field.Name)
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

	log.Printf("RetypeToOffChain called with type: %v, itemType: %T", onChainType, itemType)

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
		log.Printf("Pointer detected: %v, element type: %v", onChainType, onChainType.Elem())

		// Recursively call RetypeToOffChain on the element type
		elm, err := t.RetypeToOffChain(onChainType.Elem(), "")
		if err != nil {
			log.Printf("Error while processing pointer element type: %v", err)
			return nil, err
		}

		ptr := reflect.PointerTo(elm)
		log.Printf("Pointer transformed: %v -> %v", onChainType, ptr)

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
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields, noop, addressToStringHook(t.length, t.checksum))
}

func convertBytesToString(t reflect.Type, field string) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Array:
		log.Printf("Array detected: %v, field: %s", t, field)
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == 20 {
			log.Printf("Converting array of [20]byte to array of strings: %v", t)
			return reflect.ArrayOf(t.Len(), reflect.TypeOf("")), nil
		}
		log.Printf("Converting [20]byte to string: %v", t)
		return reflect.TypeOf(""), nil

	case reflect.Pointer:
		log.Printf("Pointer detected: %v, field: %s", t, field)
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == 20 {
			log.Printf("Converting pointer to [20]byte to string: %v", t)
			return reflect.TypeOf(""), nil
		}
		log.Printf("Pointer is not pointing to an array of [20]byte, returning error")
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	case reflect.Slice:
		log.Printf("Slice detected: %v, field: %s", t, field)
		// Handle slices of [20]byte
		if t.Elem().Kind() == reflect.Array && t.Elem().Len() == 20 {
			log.Printf("Converting slice of [20]byte to slice of strings: %v", t)
			return reflect.SliceOf(reflect.TypeOf("")), nil
		}
		log.Printf("Slice does not contain [20]byte elements, returning error")
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)

	default:
		log.Printf("Unsupported type detected: %v, field: %s", t, field)
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)
	}
}
