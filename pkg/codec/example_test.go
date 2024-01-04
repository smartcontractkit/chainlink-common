package codec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// This example demonstrates how you can have config for itemTypes with one codec,
// one that is modified before encoding for on-chain and modified after decoding for off-chain
// the other is left unmodified during encoding and decoding.
const (
	anyUnmodifiedTypeName     = "Unmodified"
	anyModifiedStructTypeName = "SecondItem"
)

var _ types.RemoteCodec = &ExampleStructJsonCodec{}

type ExampleStructJsonCodec struct{}

func (j ExampleStructJsonCodec) Encode(_ context.Context, item any, _ string) ([]byte, error) {
	return json.Marshal(item)
}

func (j ExampleStructJsonCodec) GetMaxEncodingSize(_ context.Context, n int, _ string) (int, error) {
	// not used in the example, and not really valid for json.
	return math.MaxInt32, nil
}

func (j ExampleStructJsonCodec) Decode(_ context.Context, raw []byte, into any, _ string) error {
	err := json.Unmarshal(raw, into)
	if err != nil {
		return fmt.Errorf("%w: %s", types.ErrInvalidType, err)
	}
	return nil
}

func (j ExampleStructJsonCodec) GetMaxDecodingSize(ctx context.Context, n int, _ string) (int, error) {
	// not used in the example, and not really valid for json.
	return math.MaxInt32, nil
}

func (j ExampleStructJsonCodec) CreateType(_ string, _ bool) (any, error) {
	// parameters here are unused in the example, but can be used to determine what type to expect.
	// this allows remote execution to know how to decode the incoming message
	// and for [codec.NewModifierCodec] to know what type to expect for intermediate phases.
	return &OnChainStruct{}, nil
}

type OnChainStruct struct {
	Aa int64
	Bb string
	Cc bool
	Dd string
	Ee int64
	Ff string
}

const config = `
[
  { "Type" : "drop", "Fields" :  ["Bb"] },
  { "Type" : "hard code", "OnChainValues" : {"Ff" :  "dog", "Bb" : "bb"}, "OffChainValues" : {"Zz" : "foo"}},
  { "Type" : "rename", "Fields" :  {"Aa" :  "Bb"}},
  { "Type" : "extract element", "Extractions" :  {"Dd" :  "middle"}},
  { "Type" : "epoch to time", "Fields" :  ["Ee"]}
]
`

// config converts the OnChainStruct to this structure
type OffChainStruct struct {
	Bb int64
	Cc bool
	Dd []string
	Ee *time.Time
	Zz string
}

// Example demonstrates how to use the codec package.
// It will make use of each [Modifier] provided in the package, along with their config.
func Example() {
	mods := createModsFromConfig()

	c, err := codec.NewModifierCodec(&ExampleStructJsonCodec{}, mods)
	if err != nil {
		panic(err)
	}
	writeAndReadUnmodified(c)

	writeAndReadModified(c)
	// Output:
	// {"Aa":10,"Bb":"20","Cc":true,"Dd":"great example","Ee":631515600,"Ff":"dog"}
	// true
	// {"Aa":10,"Bb":"","Cc":true,"Dd":"great example","Ee":631515600,"Ff":"dog"}
	// true
}

func writeAndReadUnmodified(c types.Codec) {
	input := &OnChainStruct{
		Aa: 10,
		Bb: "20",
		Cc: true,
		Dd: "great example",
		Ee: 631515600,
		Ff: "dog",
	}

	ctx := context.Background()
	b, err := c.Encode(ctx, input, anyUnmodifiedTypeName)
	fmt.Println(string(b))

	output := &OnChainStruct{}
	if err = c.Decode(ctx, b, output, anyUnmodifiedTypeName); err != nil {
		panic(err)
	}
	fmt.Println(reflect.DeepEqual(input, output))
}

func writeAndReadModified(c types.RemoteCodec) {
	anyTimeEpoch := int64(631515600)
	t := time.Unix(anyTimeEpoch, 0)
	modifedInput := &OffChainStruct{
		Bb: 10,
		Cc: true,
		Dd: []string{"terrible example", "great example", "not this one :("},
		Ee: &t,
		Zz: "foo",
	}

	ctx := context.Background()
	b, err := c.Encode(ctx, modifedInput, anyModifiedStructTypeName)
	if err != nil {
		panic(err)

	}
	fmt.Println(string(b))

	output := &OffChainStruct{}
	if err = c.Decode(ctx, b, output, anyModifiedStructTypeName); err != nil {
		panic(err)
	}

	expected := *modifedInput

	// Only the middle value was extracted, so decoding can only provide a single value back.
	expected.Dd = []string{"great example"}
	fmt.Println(reflect.DeepEqual(&expected, output))
}

func createModsFromConfig() codec.Modifier {
	modifierConfig := &codec.ModifiersConfig{}
	err := json.Unmarshal([]byte(config), modifierConfig)
	if err != nil {
		panic(err)
	}

	mod, err := modifierConfig.ToModifier()
	if err != nil {
		panic(err)
	}

	modByItemType := map[string]codec.Modifier{
		anyModifiedStructTypeName: mod,
		anyUnmodifiedTypeName:     codec.MultiModifier{},
	}

	mods, err := codec.NewByItemTypeModifier(modByItemType)
	if err != nil {
		panic(err)
	}
	return mods
}
