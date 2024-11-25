package interfacetests

import (
	"errors"
	"testing"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type EncodeRequest struct {
	TestStructs  []TestStruct
	ExtraField   bool
	MissingField bool
	TestOn       string
}

type CodecInterfaceTester interface {
	BasicTester[*testing.T]
	EncodeFields(t *testing.T, request *EncodeRequest) []byte
	GetCodec(t *testing.T) types.Codec

	// IncludeArrayEncodingSizeEnforcement is here in case there's no way to have fixed arrays in the encoded values
	IncludeArrayEncodingSizeEnforcement() bool
}

const (
	TestItemType            = "TestItem"
	TestItemSliceType       = "TestItemSliceType"
	TestItemArray1Type      = "TestItemArray1Type"
	TestItemArray2Type      = "TestItemArray2Type"
	TestItemWithConfigExtra = "TestItemWithConfigExtra"
	NilType                 = "NilType"
)

func RunCodecInterfaceTests(t *testing.T, tester CodecInterfaceTester) {
	tests := []Testcase[*testing.T]{
		{
			Name: "Encodes and decodes a single item",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
				resp := tester.EncodeFields(t, req)

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, item, TestItemType)
				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				into := TestStruct{}
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
				assert.Equal(t, item, into)
			},
		},
		{
			Name: "Encodes compatible types",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
				resp := tester.EncodeFields(t, req)
				compatibleItem := compatibleTestStruct{
					AccountStruct:       item.AccountStruct,
					Accounts:            item.Accounts,
					BigField:            item.BigField,
					DifferentField:      item.DifferentField,
					Field:               *item.Field,
					NestedDynamicStruct: item.NestedDynamicStruct,
					NestedStaticStruct:  item.NestedStaticStruct,
					OracleID:            item.OracleID,
					OracleIDs:           item.OracleIDs,
				}

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, compatibleItem, TestItemType)
				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				into := TestStruct{}
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
				assert.Equal(t, item, into)
			},
		},
		{
			Name: "Encodes compatible maps",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
				resp := tester.EncodeFields(t, req)
				compatibleMap := map[string]any{
					"AccountStruct": map[string]any{
						"Account":    item.AccountStruct.Account,
						"AccountStr": item.AccountStruct.AccountStr,
					},
					"Accounts":       item.Accounts,
					"BigField":       item.BigField,
					"DifferentField": item.DifferentField,
					"Field":          item.Field,
					"NestedDynamicStruct": map[string]any{
						// since we're testing compatibility, also use slice instead of array
						"FixedBytes": item.NestedDynamicStruct.FixedBytes[:],
						"Inner": map[string]any{
							"I": item.NestedDynamicStruct.Inner.I,
							"S": item.NestedDynamicStruct.Inner.S,
						},
					},
					"NestedStaticStruct": map[string]any{
						// since we're testing compatibility, also use slice instead of array
						"FixedBytes": item.NestedStaticStruct.FixedBytes[:],
						"Inner": map[string]any{
							"I": item.NestedStaticStruct.Inner.I,
							"A": item.NestedStaticStruct.Inner.A,
						},
					},
					"OracleID":  item.OracleID,
					"OracleIDs": item.OracleIDs,
				}

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, compatibleMap, TestItemType)
				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				into := TestStruct{}
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
				assert.Equal(t, item, into)
			},
		},
		{
			Name: "Encode returns an error if a field is not provided",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				ts := CreateTestStruct[*testing.T](0, tester)
				item := &TestStructMissingField{
					DifferentField:      ts.DifferentField,
					OracleID:            ts.OracleID,
					OracleIDs:           ts.OracleIDs,
					AccountStruct:       ts.AccountStruct,
					Accounts:            ts.Accounts,
					BigField:            ts.BigField,
					NestedDynamicStruct: ts.NestedDynamicStruct,
					NestedStaticStruct:  ts.NestedStaticStruct,
				}

				codec := tester.GetCodec(t)
				_, err := codec.Encode(ctx, item, TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			},
		},
		{
			Name: "Encodes and decodes a slice",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item1 := CreateTestStruct[*testing.T](0, tester)
				item2 := CreateTestStruct[*testing.T](1, tester)
				items := []TestStruct{item1, item2}
				req := &EncodeRequest{TestStructs: items, TestOn: TestItemSliceType}
				resp := tester.EncodeFields(t, req)

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, items, TestItemSliceType)
				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				var into []TestStruct
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemSliceType))
				assert.Equal(t, items, into)
			},
		},
		{
			Name: "Encodes and decodes a slices with one element",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item1 := CreateTestStruct[*testing.T](0, tester)
				items := []TestStruct{item1}
				req := &EncodeRequest{TestStructs: items, TestOn: TestItemSliceType}
				resp := tester.EncodeFields(t, req)

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, items, TestItemSliceType)

				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				var into []TestStruct
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemSliceType))
				assert.Equal(t, items, into)
			},
		},
		{
			Name: "Encodes and decodes an array",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item1 := CreateTestStruct[*testing.T](0, tester)
				item2 := CreateTestStruct[*testing.T](1, tester)
				items := [2]TestStruct{item1, item2}
				req := &EncodeRequest{TestStructs: items[:], TestOn: TestItemArray2Type}
				resp := tester.EncodeFields(t, req)

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, items, TestItemArray2Type)

				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				var into [2]TestStruct
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemArray2Type))
				assert.Equal(t, items, into)
			},
		},
		{
			Name: "Encodes and decodes an arrays with one element",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item1 := CreateTestStruct[*testing.T](0, tester)
				items := [1]TestStruct{item1}
				req := &EncodeRequest{TestStructs: items[:], TestOn: TestItemArray1Type}
				resp := tester.EncodeFields(t, req)

				codec := tester.GetCodec(t)
				actualEncoding, err := codec.Encode(ctx, items, TestItemArray1Type)

				require.NoError(t, err)
				assert.Equal(t, resp, actualEncoding)

				var into [1]TestStruct
				require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemArray1Type))
				assert.Equal(t, items, into)
			},
		},
		{
			Name: "Returns an error if type is undefined",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, item, "NOT"+TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))

				err = codec.Decode(ctx, []byte(""), item, "NOT"+TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			},
		},
		{
			Name: "Returns an error encoding if arrays are the too small to encode",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				if !tester.IncludeArrayEncodingSizeEnforcement() {
					return
				}

				item1 := CreateTestStruct[*testing.T](0, tester)
				items := [1]TestStruct{item1}
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, items, TestItemArray2Type)
				assert.True(t, errors.Is(err, types.ErrSliceWrongLen))
			},
		},
		{
			Name: "Returns an error encoding if arrays are the too large to encode",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				if !tester.IncludeArrayEncodingSizeEnforcement() {
					return
				}

				item1 := CreateTestStruct[*testing.T](0, tester)
				item2 := CreateTestStruct[*testing.T](1, tester)
				items := [2]TestStruct{item1, item2}
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, items, TestItemArray1Type)
				assert.True(t, errors.Is(err, types.ErrSliceWrongLen))
			},
		},
		{
			Name: "GetMaxEncodingSize returns errors for unknown types",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				_, err := cr.GetMaxEncodingSize(ctx, 10, "not"+TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			},
		},
		{
			Name: "GetMaxDecodingSize returns errors for unknown types",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				_, err := cr.GetMaxDecodingSize(ctx, 10, "not"+TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			},
		},
		{
			Name: "Decode respects config",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				original := CreateTestStruct[*testing.T](0, tester)
				bytes, err := cr.Encode(ctx, original, TestItemType)
				require.NoError(t, err)

				decoded := &TestStructWithExtraField{}
				require.NoError(t, cr.Decode(ctx, bytes, decoded, TestItemWithConfigExtra))

				expected := &TestStructWithExtraField{
					ExtraField: AnyExtraValue,
					TestStruct: original,
				}
				assert.Equal(t, expected, decoded)
			},
		},
		{
			Name: "Encode respects config",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				modified := CreateTestStruct[*testing.T](0, tester)
				modified.BigField = nil
				modified.AccountStruct.Account = nil
				actual, err := cr.Encode(ctx, modified, TestItemWithConfigExtra)
				require.NoError(t, err)

				original := CreateTestStruct[*testing.T](0, tester)
				expected, err := cr.Encode(ctx, original, TestItemType)
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			},
		},
		{
			Name: "Encode allows nil params to be encoded, either as empty encoding or with prefix",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				_, err := cr.Encode(ctx, nil, NilType)
				require.NoError(t, err)
			},
		},
		{
			Name: "Encode does not panic on nil field",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				nilArgs := &TestStruct{
					Field:               nil,
					DifferentField:      "",
					OracleID:            0,
					OracleIDs:           [32]commontypes.OracleID{},
					AccountStruct:       AccountStruct{},
					Accounts:            nil,
					BigField:            nil,
					NestedDynamicStruct: MidLevelDynamicTestStruct{},
					NestedStaticStruct:  MidLevelStaticTestStruct{},
				}
				// Assure no panic, use _,_ to tell the compiler we don't care about the error
				_, _ = cr.Encode(ctx, nilArgs, TestItemType)
			},
		},
		{
			Name: "Encode returns an error if the item isn't compatible",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				cr := tester.GetCodec(t)
				notTestStruct := &MidLevelDynamicTestStruct{}
				_, err := cr.Encode(ctx, notTestStruct, TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			},
		},
	}
	RunTests(t, tester, tests)
}

// RunCodecWithStrictArgsInterfaceTest is meant to be used by codecs that don't pad
// They can assure that the right argument size is verified.
// Padding makes that harder/impossible to verify for come codecs.
// However, the extra verification is nice to have when possible.
func RunCodecWithStrictArgsInterfaceTest(t *testing.T, tester CodecInterfaceTester) {
	RunCodecInterfaceTests(t, tester)

	tests := []Testcase[*testing.T]{
		{
			Name: "Gives an error decoding extra fields on an item",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				req := &EncodeRequest{
					TestStructs: []TestStruct{item},
					ExtraField:  true,
					TestOn:      TestItemType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &item, TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidEncoding))
			},
		},
		{
			Name: "Gives an error decoding missing fields on an item",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				item := CreateTestStruct[*testing.T](0, tester)
				req := &EncodeRequest{
					TestStructs:  []TestStruct{item},
					MissingField: true,
					TestOn:       TestItemType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &item, TestItemType)
				assert.True(t, errors.Is(err, types.ErrInvalidEncoding))
			},
		},
		{
			Name: "Gives an error decoding extra fields on a slice",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				items := []TestStruct{CreateTestStruct[*testing.T](0, tester)}
				req := &EncodeRequest{
					TestStructs: items,
					ExtraField:  true,
					TestOn:      TestItemSliceType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &items, TestItemSliceType)
				assert.True(t, errors.Is(err, types.ErrInvalidEncoding))
			},
		},
		{
			Name: "Gives an error decoding missing fields on an slice",
			Test: func(t *testing.T) {
				ctx := tests.Context(t)
				items := []TestStruct{CreateTestStruct[*testing.T](0, tester)}
				req := &EncodeRequest{
					TestStructs:  items,
					MissingField: true,
					TestOn:       TestItemSliceType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &items, TestItemSliceType)
				assert.True(t, errors.Is(err, types.ErrInvalidEncoding))
			},
		},
	}

	RunTests(t, tester, tests)
}
