package interfacetests

import (
	"testing"

	ocrTypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils/tests"
)

type EncodeRequest struct {
	TestStructs  []TestStruct
	ExtraField   bool
	MissingField bool
	TestOn       string
}

type CodecInterfaceTester interface {
	BasicTester
	EncodeFields(t *testing.T, request *EncodeRequest) ocrTypes.Report
	GetCodec(t *testing.T) types.Codec

	// IncludeArrayEncodingSizeEnforcement is here in case there's no way to have fixed arrays in the encoded values
	IncludeArrayEncodingSizeEnforcement() bool
}

const (
	TestItemType       = "TestItem"
	TestItemSliceType  = "TestItemSliceType"
	TestItemArray1Type = "TestItemArray1Type"
	TestItemArray2Type = "TestItemArray2Type"
)

func RunCodecInterfaceTests(t *testing.T, tester CodecInterfaceTester) {
	ctx := tests.Context(t)
	tests := []testcase{
		{
			name: "Encodes and decodes a single item",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
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
			name: "Encodes compatible types",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
				req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
				resp := tester.EncodeFields(t, req)
				compatibleItem := compatibleTestStruct{
					Account:        item.Account,
					Accounts:       item.Accounts,
					BigField:       item.BigField,
					DifferentField: item.DifferentField,
					Field:          item.Field,
					NestedStruct:   item.NestedStruct,
					OracleId:       item.OracleId,
					OracleIds:      item.OracleIds,
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
			name: "Encodes compatible maps",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
				req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
				resp := tester.EncodeFields(t, req)
				compatibleMap := map[string]any{
					"Account":        item.Account,
					"Accounts":       item.Accounts,
					"BigField":       item.BigField,
					"DifferentField": item.DifferentField,
					"Field":          item.Field,
					"NestedStruct": map[string]any{
						// since we're testing compatibility, also use slice instead of array
						"FixedBytes": item.NestedStruct.FixedBytes[:],
						"Inner": map[string]any{
							"I": item.NestedStruct.Inner.I,
							"S": item.NestedStruct.Inner.S,
						},
					},
					"OracleId":  item.OracleId,
					"OracleIds": item.OracleIds,
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
			name: "Encodes and decodes a slice",
			test: func(t *testing.T) {
				item1 := CreateTestStruct(0, tester)
				item2 := CreateTestStruct(1, tester)
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
			name: "Encodes and decodes a slices with one element",
			test: func(t *testing.T) {
				item1 := CreateTestStruct(0, tester)
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
			name: "Encodes and decodes an array",
			test: func(t *testing.T) {
				item1 := CreateTestStruct(0, tester)
				item2 := CreateTestStruct(1, tester)
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
			name: "Encodes and decodes an arrays with one element",
			test: func(t *testing.T) {
				item1 := CreateTestStruct(0, tester)
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
			name: "Returns an error if type is undefined",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, item, "NOT"+TestItemType)
				assert.IsType(t, types.InvalidTypeError{}, err)

				err = codec.Decode(ctx, []byte(""), item, "NOT"+TestItemType)
				assert.IsType(t, types.InvalidTypeError{}, err)
			},
		},
		{
			name: "Returns an error encoding if arrays are the too small to encode",
			test: func(t *testing.T) {
				if !tester.IncludeArrayEncodingSizeEnforcement() {
					return
				}

				item1 := CreateTestStruct(0, tester)
				items := [1]TestStruct{item1}
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, items, TestItemArray2Type)
				assert.IsType(t, types.WrongNumberOfElements{}, err)
			},
		},
		{
			name: "Returns an error encoding if arrays are the too large to encode",
			test: func(t *testing.T) {
				if !tester.IncludeArrayEncodingSizeEnforcement() {
					return
				}

				item1 := CreateTestStruct(0, tester)
				item2 := CreateTestStruct(1, tester)
				items := [2]TestStruct{item1, item2}
				codec := tester.GetCodec(t)

				_, err := codec.Encode(ctx, items, TestItemArray1Type)
				assert.IsType(t, types.WrongNumberOfElements{}, err)
			},
		},
		{
			name: "GetMaxEncodingSize returns errors for unknown types",
			test: func(t *testing.T) {
				cr := tester.GetCodec(t)
				_, err := cr.GetMaxEncodingSize(ctx, 10, "not"+TestItemType)
				assert.IsType(t, types.InvalidTypeError{}, err)
			},
		},
	}
	runTests(t, tester, tests)
}

// RunChainReaderWithStrictArgsInterfaceTest is meant to be used by codecs that don't pad
// They can assure that the right argument size is verified.
// Padding makes that harder/impossible to verify for come codecs.
// However, the extra verification is nice to have when possible.
func RunChainReaderWithStrictArgsInterfaceTest(t *testing.T, tester CodecInterfaceTester) {
	ctx := tests.Context(t)
	RunCodecInterfaceTests(t, tester)

	tests := []testcase{
		{
			name: "Gives an error decoding extra fields on an item",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
				req := &EncodeRequest{
					TestStructs: []TestStruct{item},
					ExtraField:  true,
					TestOn:      TestItemType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &item, TestItemType)
				assert.IsType(t, types.InvalidEncodingError{}, err)
			},
		},
		{
			name: "Gives an error decoding missing fields on an item",
			test: func(t *testing.T) {
				item := CreateTestStruct(0, tester)
				req := &EncodeRequest{
					TestStructs:  []TestStruct{item},
					MissingField: true,
					TestOn:       TestItemType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &item, TestItemType)
				assert.IsType(t, types.InvalidEncodingError{}, err)
			},
		},
		{
			name: "Gives an error decoding extra fields on a slice",
			test: func(t *testing.T) {
				items := []TestStruct{CreateTestStruct(0, tester)}
				req := &EncodeRequest{
					TestStructs: items,
					ExtraField:  true,
					TestOn:      TestItemSliceType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &items, TestItemSliceType)
				assert.IsType(t, types.InvalidEncodingError{}, err)
			},
		},
		{
			name: "Gives an error decoding missing fields on an slice",
			test: func(t *testing.T) {
				items := []TestStruct{CreateTestStruct(0, tester)}
				req := &EncodeRequest{
					TestStructs:  items,
					MissingField: true,
					TestOn:       TestItemSliceType,
				}
				resp := tester.EncodeFields(t, req)
				codec := tester.GetCodec(t)
				err := codec.Decode(ctx, resp, &items, TestItemSliceType)
				assert.IsType(t, types.InvalidEncodingError{}, err)
			},
		},
	}

	runTests(t, tester, tests)
}
