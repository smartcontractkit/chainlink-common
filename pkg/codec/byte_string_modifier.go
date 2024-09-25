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
	return transformWithMaps(offChainValue, t.offToOnChainType, t.fields, noop, addressToStringHook(t.length, t.checksum))
}

func (t *bytesToStringModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return transformWithMaps(onChainValue, t.onToOffChainType, t.fields, noop, addressToStringHook(t.length, t.checksum))
}

func convertBytesToString(t reflect.Type, field string) (reflect.Type, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return reflect.PointerTo(reflect.TypeOf("")), nil
	case reflect.Array:
		return reflect.TypeOf(""), nil
	default:
		return nil, fmt.Errorf("%w: cannot convert bytes for field %s", types.ErrInvalidType, field)
	}
}
