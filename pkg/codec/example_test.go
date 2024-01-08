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

var _ types.RemoteCodec = &ExampleStructJSONCodec{}

type ExampleStructJSONCodec struct{}

func (ExampleStructJSONCodec) Encode(_ context.Context, item any, _ string) ([]byte, error) {
	return json.Marshal(item)
}

func (ExampleStructJSONCodec) GetMaxEncodingSize(_ context.Context, n int, _ string) (int, error) {
	// not used in the example, and not really valid for json.
	return math.MaxInt32, nil
}

func (ExampleStructJSONCodec) Decode(_ context.Context, raw []byte, into any, _ string) error {
	err := json.Unmarshal(raw, into)
	if err != nil {
		return fmt.Errorf("%w: %s", types.ErrInvalidType, err)
	}
	return nil
}

func (ExampleStructJSONCodec) GetMaxDecodingSize(ctx context.Context, n int, _ string) (int, error) {
	// not used in the example, and not really valid for json.
	return math.MaxInt32, nil
}

func (ExampleStructJSONCodec) CreateType(_ string, _ bool) (any, error) {
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
	mods, err := createModsFromConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := codec.NewModifierCodec(&ExampleStructJSONCodec{}, mods)
	if err != nil {
		fmt.Println(err)
		return
	}

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
		fmt.Println(err)
		return
	}
	fmt.Println(reflect.DeepEqual(input, output))

	anyTimeEpoch := int64(631515600)
	t := time.Unix(anyTimeEpoch, 0)
	modifedInput := &OffChainStruct{
		Bb: 10,
		Cc: true,
		Dd: []string{"terrible example", "great example", "not this one :("},
		Ee: &t,
		Zz: "foo",
	}

	b, err = c.Encode(ctx, modifedInput, anyModifiedStructTypeName)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

	output2 := &OffChainStruct{}
	if err = c.Decode(ctx, b, output2, anyModifiedStructTypeName); err != nil {
		fmt.Println(err)
		return
	}

	expected := *modifedInput

	// Only the middle value was extracted, so decoding can only provide a single value back.
	expected.Dd = []string{"great example"}
	fmt.Println(reflect.DeepEqual(&expected, output2))
	// Output:
	// {"Aa":10,"Bb":"20","Cc":true,"Dd":"great example","Ee":631515600,"Ff":"dog"}
	// true
	// {"Aa":10,"Bb":"","Cc":true,"Dd":"great example","Ee":631515600,"Ff":"dog"}
	// true
}

func createModsFromConfig() (codec.Modifier, error) {
	modifierConfig := &codec.ModifiersConfig{}
	if err := json.Unmarshal([]byte(config), modifierConfig); err != nil {
		return nil, err
	}

	mod, err := modifierConfig.ToModifier()
	if err != nil {
		return nil, err
	}

	modByItemType := map[string]codec.Modifier{
		anyModifiedStructTypeName: mod,
		anyUnmodifiedTypeName:     codec.MultiModifier{},
	}

	return codec.NewByItemTypeModifier(modByItemType)
}
