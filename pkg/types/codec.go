/*
Codec is an interface that provides encoding and decoding functionality for a specific type identified by a name.
Because there are many types that a ContractReader or ContractWriter can either accept or return, all encoding
instructions provided by the codec are based on the type name.

Starting from the lowest level, take for instance a big.Int encoder where we want the output to be big endian binary
encoded.

	typeCodec, _ := binary.NewBigInt(32, true, binary.BigEndian())

This allows us to encode and decode big.Int values using the big endian encoding using the encodings.TypeCodec interface.

	encodedBytes := []byte{}

	originalValue := big.NewInt(42)
	encodedBytes, _ = typeCodec.Encode(originalValue, encodedBytes) // new encoded bytes are appended to existing

	value := new(big.Int)
	_ = typeCodec.Decode(encodedBytes, value)

The additional encodings.TypeCodec methods such as `GetType() reflect.Type` allow composition. This is useful for
creating a struct codec such as the one defined in encodings/struct.go.

	tlCodec, _ := NewStructCodec([]NamedTypeCodec{{Name: "Value", Codec: typeCodec}})

This provides a `TopLevelCodec` which a `TypeCodec` with a total size of all encoded elements. Going up another level,
we create a `Codec` from a map of `TypeCodec` instances using `CodecFromTypeCodec`.

	codec := types.CodecFromTypeCodec{"SomeStruct": tlCodec}

	type SomeStruct struct {
		Value *big.Int
	}

	encodedStructBytes, _ := codec.Encode(SomeStruct{Value: big.NewInt(42)}, "SomeStruct")

	var someStruct SomeStruct
	_ = codec.Decode(encodedStructBytes, &someStruct, "SomeStruct")

Therefore `itemType` passed to `Encode` and `Decode` references the key in the map of `TypeCodec` instances. Also worth
noting that a `TopLevelCodec` can also be added to a `CodecFromTypeCodec` map. This allows for the `SizeAtTopLevel`
method to be referenced when `GetMaxEncodingSize` is called on the `Codec`.

Also, when the type is unknown to the caller, the decoded type for an `itemName` can be retrieved from the codec to be
used for decoding. The `CreateType` method returns an instance of the expected type using reflection under the hood and
the overall composition of `TypeCodec` instances.

	decodedStruct, _ := codec.CreateType("SomeStruct", false)
	_ = codec.Decode(encodedStructBytes, &decodedStruct, "SomeStruct")

The `encodings` package provides a `Builder` interface that allows for the creation of any encoding type. This is useful
for creating custom encodings such as the EVM ABI encoding. An encoder implements the `Builder` interface and plugs
directly into `TypeCodec`.

From the perspective of a `ContractReader` instance, the `itemType` at the top level is the `readIdentifier` which
can be imagined as `contractName + methodName` given that a contract method call returns some configured value that
would need its own codec. Each implementation of `ContractReader` maps the names to codecs differently on the inside,
but from the level of the interface, the `itemType` is the `readIdentifier`.
*/
package types

import (
	"context"
)

const (
	ErrFieldNotFound   = InvalidArgumentError("field not found")
	ErrInvalidEncoding = InvalidArgumentError("invalid encoding")
	ErrSliceWrongLen   = InvalidArgumentError("slice is wrong length")
	ErrNotASlice       = InvalidArgumentError("element is not a slice")
)

type Encoder interface {
	Encode(ctx context.Context, item any, itemType string) ([]byte, error)
	// GetMaxEncodingSize returns the max size in bytes if n elements are supplied for all top level dynamically sized elements.
	// If no elements are dynamically sized, the returned value will be the same for all n.
	// If there are multiple levels of dynamically sized elements, or itemType cannot be found,
	// ErrInvalidType will be returned.
	GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error)
}

type Decoder interface {
	Decode(ctx context.Context, raw []byte, into any, itemType string) error
	// GetMaxDecodingSize returns the max size in bytes if n elements are supplied for all top level dynamically sized elements.
	// If no elements are dynamically sized, the returned value will be the same for all n.
	// If there are multiple levels of dynamically sized elements, or itemType cannot be found,
	// ErrInvalidType will be returned.
	GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error)
}

type Codec interface {
	Encoder
	Decoder
}

type TypeProvider interface {
	CreateType(itemType string, forEncoding bool) (any, error)
}

type ContractTypeProvider interface {
	CreateContractType(readName string, forEncoding bool) (any, error)
}

type RemoteCodec interface {
	Codec
	TypeProvider
}
