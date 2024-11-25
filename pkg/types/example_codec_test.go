package types_test

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings"
	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings/binary"
)

// ExampleCodec provides a minimal example of constructing and using a codec.
func ExampleCodec() {
	ctx := context.Background()
	typeCodec, _ := binary.BigEndian().BigInt(32, true)

	// start with empty encoded bytes
	encodedBytes := []byte{}
	originalValue := big.NewInt(42)

	encodedBytes, _ = typeCodec.Encode(originalValue, encodedBytes) // new encoded bytes are appended to existing
	value, _, _ := typeCodec.Decode(encodedBytes)

	// originalValue is the same as value
	fmt.Printf("%+v == %+v\n", originalValue, value)

	// TopLevelCodec is a TypeCodec that has a total size of all encoded elements
	tlCodec, _ := encodings.NewStructCodec([]encodings.NamedTypeCodec{{Name: "Value", Codec: typeCodec}})
	codec := encodings.CodecFromTypeCodec{"SomeStruct": tlCodec}

	type SomeStruct struct {
		Value *big.Int
	}

	originalStruct := SomeStruct{Value: big.NewInt(42)}
	encodedStructBytes, _ := codec.Encode(ctx, originalStruct, "SomeStruct")

	var someStruct SomeStruct
	_ = codec.Decode(ctx, encodedStructBytes, &someStruct, "SomeStruct")

	decodedStruct, _ := codec.CreateType("SomeStruct", false)
	_ = codec.Decode(ctx, encodedStructBytes, &decodedStruct, "SomeStruct")

	// encoded struct is equal to decoded struct using defined type and/or CreateType
	fmt.Printf("%+v == %+v == %+v\n", originalStruct, someStruct, decodedStruct)
}
