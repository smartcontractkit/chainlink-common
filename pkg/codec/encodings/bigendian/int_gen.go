// DO NOT MODIFY: automatically generated from pkg/codec/raw/types/main.go using the template int_gen.go

package bigendian

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type Int8 struct{}

var _ encodings.TypeCodec = Int8{}

func (Int8) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(int8)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an int8", types.ErrInvalidType, value)
	}
	return append(into, byte(v)), nil
}

func (Int8) Decode(encoded []byte) (any, []byte, error) {
	ui, remaining, err := encodings.SafeDecode[uint8](encoded, 1, func(encoded []byte) byte { return encoded[0] })
	return int8(ui), remaining, err
}

func (Int8) GetType() reflect.Type {
	return reflect.TypeOf(int8(0))
}

func (Int8) Size(int) (int, error) {
	return 8 / 8, nil
}

func (Int8) FixedSize() (int, error) {
	return 8 / 8, nil
}

type Uint8 struct{}

var _ encodings.TypeCodec = Uint8{}

func (Uint8) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(uint8)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an uint8", types.ErrInvalidType, value)
	}
	return append(into, v), nil
}

func (Uint8) Decode(encoded []byte) (any, []byte, error) {
	return encodings.SafeDecode[uint8](encoded, 1, func(encoded []byte) byte { return encoded[0] })
}

func (Uint8) GetType() reflect.Type {
	return reflect.TypeOf(uint8(0))
}

func (Uint8) Size(int) (int, error) {
	return 8 / 8, nil
}

func (Uint8) FixedSize() (int, error) {
	return 8 / 8, nil
}

type Int16 struct{}

var _ encodings.TypeCodec = Int16{}

func (Int16) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(int16)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an int16", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint16(into, uint16(v)), nil
}

func (Int16) Decode(encoded []byte) (any, []byte, error) {
	ui, remaining, err := encodings.SafeDecode[uint16](encoded, 16/8, binary.BigEndian.Uint16)
	return int16(ui), remaining, err
}

func (Int16) GetType() reflect.Type {
	return reflect.TypeOf(int16(0))
}

func (Int16) Size(int) (int, error) {
	return 16 / 8, nil
}

func (Int16) FixedSize() (int, error) {
	return 16 / 8, nil
}

type Uint16 struct{}

var _ encodings.TypeCodec = Uint16{}

func (Uint16) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(uint16)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an uint16", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint16(into, v), nil
}

func (Uint16) Decode(encoded []byte) (any, []byte, error) {
	return encodings.SafeDecode[uint16](encoded, 16/8, binary.BigEndian.Uint16)
}

func (Uint16) GetType() reflect.Type {
	return reflect.TypeOf(uint16(0))
}

func (Uint16) Size(int) (int, error) {
	return 16 / 8, nil
}

func (Uint16) FixedSize() (int, error) {
	return 16 / 8, nil
}

type Int32 struct{}

var _ encodings.TypeCodec = Int32{}

func (Int32) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(int32)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an int32", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint32(into, uint32(v)), nil
}

func (Int32) Decode(encoded []byte) (any, []byte, error) {
	ui, remaining, err := encodings.SafeDecode[uint32](encoded, 32/8, binary.BigEndian.Uint32)
	return int32(ui), remaining, err
}

func (Int32) GetType() reflect.Type {
	return reflect.TypeOf(int32(0))
}

func (Int32) Size(int) (int, error) {
	return 32 / 8, nil
}

func (Int32) FixedSize() (int, error) {
	return 32 / 8, nil
}

type Uint32 struct{}

var _ encodings.TypeCodec = Uint32{}

func (Uint32) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(uint32)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an uint32", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint32(into, v), nil
}

func (Uint32) Decode(encoded []byte) (any, []byte, error) {
	return encodings.SafeDecode[uint32](encoded, 32/8, binary.BigEndian.Uint32)
}

func (Uint32) GetType() reflect.Type {
	return reflect.TypeOf(uint32(0))
}

func (Uint32) Size(int) (int, error) {
	return 32 / 8, nil
}

func (Uint32) FixedSize() (int, error) {
	return 32 / 8, nil
}

type Int64 struct{}

var _ encodings.TypeCodec = Int64{}

func (Int64) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an int64", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint64(into, uint64(v)), nil
}

func (Int64) Decode(encoded []byte) (any, []byte, error) {
	ui, remaining, err := encodings.SafeDecode[uint64](encoded, 64/8, binary.BigEndian.Uint64)
	return int64(ui), remaining, err
}

func (Int64) GetType() reflect.Type {
	return reflect.TypeOf(int64(0))
}

func (Int64) Size(int) (int, error) {
	return 64 / 8, nil
}

func (Int64) FixedSize() (int, error) {
	return 64 / 8, nil
}

type Uint64 struct{}

var _ encodings.TypeCodec = Uint64{}

func (Uint64) Encode(value any, into []byte) ([]byte, error) {
	v, ok := value.(uint64)
	if !ok {
		return nil, fmt.Errorf("%w: %T is not an uint64", types.ErrInvalidType, value)
	}
	return binary.BigEndian.AppendUint64(into, v), nil
}

func (Uint64) Decode(encoded []byte) (any, []byte, error) {
	return encodings.SafeDecode[uint64](encoded, 64/8, binary.BigEndian.Uint64)
}

func (Uint64) GetType() reflect.Type {
	return reflect.TypeOf(uint64(0))
}

func (Uint64) Size(int) (int, error) {
	return 64 / 8, nil
}

func (Uint64) FixedSize() (int, error) {
	return 64 / 8, nil
}

func GetIntTypeCodecByByteSize(size int) (encodings.TypeCodec, error) {
	switch size {
	case 8 / 8:
		return &intCodec{
			codec:   Int8{},
			toInt:   func(v any) int { return int(v.(int8)) },
			fromInt: func(v int) any { return int8(v) },
		}, nil
	case 16 / 8:
		return &intCodec{
			codec:   Int16{},
			toInt:   func(v any) int { return int(v.(int16)) },
			fromInt: func(v int) any { return int16(v) },
		}, nil
	case 32 / 8:
		return &intCodec{
			codec:   Int32{},
			toInt:   func(v any) int { return int(v.(int32)) },
			fromInt: func(v int) any { return int32(v) },
		}, nil
	case 64 / 8:
		return &intCodec{
			codec:   Int64{},
			toInt:   func(v any) int { return int(v.(int64)) },
			fromInt: func(v int) any { return int64(v) },
		}, nil
	default:
		c, err := NewBigInt(size, true)
		return &intCodec{
			codec:   c,
			toInt:   func(v any) int { return int(v.(*big.Int).Int64()) },
			fromInt: func(v int) any { return big.NewInt(int64(v)) },
		}, err
	}
}

func GetUintTypeCodecByByteSize(size int) (encodings.TypeCodec, error) {
	switch size {
	case 8 / 8:
		return &uintCodec{
			codec:    Uint8{},
			toUint:   func(v any) uint { return uint(v.(uint8)) },
			fromUint: func(v uint) any { return uint8(v) },
		}, nil
	case 16 / 8:
		return &uintCodec{
			codec:    Uint16{},
			toUint:   func(v any) uint { return uint(v.(uint16)) },
			fromUint: func(v uint) any { return uint16(v) },
		}, nil
	case 32 / 8:
		return &uintCodec{
			codec:    Uint32{},
			toUint:   func(v any) uint { return uint(v.(uint32)) },
			fromUint: func(v uint) any { return uint32(v) },
		}, nil
	case 64 / 8:
		return &uintCodec{
			codec:    Uint64{},
			toUint:   func(v any) uint { return uint(v.(uint64)) },
			fromUint: func(v uint) any { return uint64(v) },
		}, nil
	default:
		c, err := NewBigInt(size, false)
		return &uintCodec{
			codec:    c,
			toUint:   func(v any) uint { return uint(v.(*big.Int).Uint64()) },
			fromUint: func(v uint) any { return new(big.Int).SetUint64(uint64(v)) },
		}, err
	}
}
