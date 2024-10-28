package types_test

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings"
	"github.com/smartcontractkit/chainlink-common/pkg/codec/encodings/binary"
)

// Example provides a minimal example of constructing and using a codec. Refer to the codec docs in pkg/types/codec.go
// for more explanation. For even more detail, refer to the Example in pkg/codec/example_test.go.
func Example() {
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
	encodedStructBytes, _ := codec.Encode(context.Background(), originalStruct, "SomeStruct")

	var someStruct SomeStruct
	_ = codec.Decode(context.Background(), encodedStructBytes, &someStruct, "SomeStruct")

	decodedStruct, _ := codec.CreateType("SomeStruct", false)
	_ = codec.Decode(context.Background(), encodedStructBytes, &decodedStruct, "SomeStruct")

	// encoded struct is equal to decoded struct using defined type and/or CreateType
	fmt.Printf("%+v == %+v == %+v\n", originalStruct, someStruct, decodedStruct)
}
