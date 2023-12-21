// DO NOT MODIFY: automatically generated from pkg/codec/raw/types/main.go using the template int_gen_test.go

package bigendian_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings/bigendian"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestInt8(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Int8{}
		value := int8(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 8/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Int8{}
		value := int8(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 8/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Int8{}
		value := int8(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Int8{}
		bytes := make([]byte, 8/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Int8{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(int8(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Int8{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Int8{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("returns an error if the input is not an uint8", func(t *testing.T) {
		i := bigendian.Int8{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestUint8(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Uint8{}
		value := uint8(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 8/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Uint8{}
		value := uint8(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 8/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Uint8{}
		value := uint8(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Uint8{}
		bytes := make([]byte, 8/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Uint8{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(uint8(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint8{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint8{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("returns an error if the input is not an uint8", func(t *testing.T) {
		i := bigendian.Uint8{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestInt16(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Int16{}
		value := int16(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 16/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Int16{}
		value := int16(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 16/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Int16{}
		value := int16(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Int16{}
		bytes := make([]byte, 16/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Int16{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(int16(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Int16{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Int16{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("returns an error if the input is not an uint16", func(t *testing.T) {
		i := bigendian.Int16{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestUint16(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Uint16{}
		value := uint16(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 16/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Uint16{}
		value := uint16(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 16/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Uint16{}
		value := uint16(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Uint16{}
		bytes := make([]byte, 16/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Uint16{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(uint16(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint16{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint16{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("returns an error if the input is not an uint16", func(t *testing.T) {
		i := bigendian.Uint16{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestInt32(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Int32{}
		value := int32(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 32/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Int32{}
		value := int32(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 32/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Int32{}
		value := int32(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Int32{}
		bytes := make([]byte, 32/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Int32{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(int32(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Int32{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Int32{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("returns an error if the input is not an uint32", func(t *testing.T) {
		i := bigendian.Int32{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestUint32(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Uint32{}
		value := uint32(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 32/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Uint32{}
		value := uint32(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 32/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Uint32{}
		value := uint32(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Uint32{}
		bytes := make([]byte, 32/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Uint32{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(uint32(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint32{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint32{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("returns an error if the input is not an uint32", func(t *testing.T) {
		i := bigendian.Uint32{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestInt64(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Int64{}
		value := int64(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 64/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Int64{}
		value := int64(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 64/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Int64{}
		value := int64(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Int64{}
		bytes := make([]byte, 64/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Int64{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(int64(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Int64{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Int64{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("returns an error if the input is not an uint64", func(t *testing.T) {
		i := bigendian.Int64{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestUint64(t *testing.T) {
	t.Parallel()
	t.Run("Encodes and decodes to the same value with correct encoding length", func(t *testing.T) {
		i := bigendian.Uint64{}
		value := uint64(123)

		encoded, err := i.Encode(value, nil)

		require.NoError(t, err)
		assert.Equal(t, 64/8, len(encoded))

		decoded, remaining, err := i.Decode(encoded)

		require.NoError(t, err)
		assert.Equal(t, 0, len(remaining))
		assert.Equal(t, value, decoded)
	})

	t.Run("Encodes appends to prefix", func(t *testing.T) {
		i := bigendian.Uint64{}
		value := uint64(123)
		prefix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, prefix)

		require.NoError(t, err)
		assert.Equal(t, 64/8+3, len(encoded))
		expected, err := i.Encode(value, nil)
		assert.Equal(t, expected, encoded[3:])
	})

	t.Run("Decode leaves a suffix", func(t *testing.T) {
		i := bigendian.Uint64{}
		value := uint64(123)
		suffix := []byte{1, 2, 3}

		encoded, err := i.Encode(value, nil)
		require.NoError(t, err)
		encoded = append(encoded, suffix...)

		decoded, remaining, err := i.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, suffix, remaining)
		assert.Equal(t, value, decoded)
	})

	t.Run("Decode returns an error if there are not enough bytes", func(t *testing.T) {
		i := bigendian.Uint64{}
		bytes := make([]byte, 64/8-1)
		_, _, err := i.Decode(bytes)
		require.True(t, errors.Is(err, types.ErrInvalidEncoding))
	})

	t.Run("GetType returns correct type", func(t *testing.T) {
		i := bigendian.Uint64{}
		assert.Equal(t, i.GetType(), reflect.TypeOf(uint64(0)))
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint64{}.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		size, err := bigendian.Uint64{}.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("returns an error if the input is not an uint64", func(t *testing.T) {
		i := bigendian.Uint64{}

		_, err := i.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestGetIntTypeCodecByByteSize(t *testing.T) {

	t.Run("Wraps encoding and decoding for 8 bytes", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		anyValue := 123

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 8/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("Wraps encoding and decoding for 16 bytes", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		anyValue := 123

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 16/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("Wraps encoding and decoding for 32 bytes", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		anyValue := 123

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 32/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("Wraps encoding and decoding for 64 bytes", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		anyValue := 123

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 64/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("Wraps encoding and decoding for other sized bytes", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(10)
		require.NoError(t, err)
		anyValue := 123

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 10, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("returns an error if the input is not an int", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(4)
		require.NoError(t, err)

		_, err = codec.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("GetType returns int", func(t *testing.T) {
		codec, err := bigendian.GetIntTypeCodecByByteSize(4)
		require.NoError(t, err)

		assert.Equal(t, reflect.TypeOf(0), codec.GetType())
	})
}

func TestGetUintTypeCodecByByteSize(t *testing.T) {

	t.Run("Wraps encoding and decoding for 8 bytes", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		anyValue := uint(123)

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 8/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(8 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 8/8)
	})

	t.Run("Wraps encoding and decoding for 16 bytes", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		anyValue := uint(123)

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 16/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(16 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 16/8)
	})

	t.Run("Wraps encoding and decoding for 32 bytes", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		anyValue := uint(123)

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 32/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(32 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 32/8)
	})

	t.Run("Wraps encoding and decoding for 64 bytes", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		anyValue := uint(123)

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 64/8, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("Size returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		size, err := codec.Size(100)
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("FixedSize returns correct size", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(64 / 8)
		require.NoError(t, err)
		size, err := codec.FixedSize()
		require.NoError(t, err)
		assert.Equal(t, size, 64/8)
	})

	t.Run("Wraps encoding and decoding for other sized bytes", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(10)
		require.NoError(t, err)
		anyValue := uint(123)

		encoded, err := codec.Encode(anyValue, nil)
		require.NoError(t, err)
		require.Equal(t, 10, len(encoded))

		decoded, remaining, err := codec.Decode(encoded)
		require.NoError(t, err)
		require.Empty(t, remaining)
		require.Equal(t, anyValue, decoded)
	})

	t.Run("returns an error if the input is not an int", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(4)
		require.NoError(t, err)

		_, err = codec.Encode("foo", nil)
		require.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("GetType returns uint", func(t *testing.T) {
		codec, err := bigendian.GetUintTypeCodecByByteSize(4)
		require.NoError(t, err)

		assert.Equal(t, reflect.TypeOf(uint(0)), codec.GetType())
	})
}
