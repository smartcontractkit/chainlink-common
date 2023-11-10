package interfacetests

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrTypes "github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type EncodeRequest struct {
	TestStructs  []TestStruct
	ExtraField   bool
	MissingField bool
	TestOn       string
}

type ChainReaderInterfaceTester interface {
	Setup(t *testing.T)
	Teardown(t *testing.T)
	Name() string
	EncodeFields(t *testing.T, request *EncodeRequest) ocrTypes.Report
	GetAccountBytes(i int) []byte
	GetChainReader(t *testing.T) types.ChainReader

	// IncludeArrayEncodingSizeEnforcement is here in case there's no way to have fixed arrays in the encoded values
	IncludeArrayEncodingSizeEnforcement() bool

	// SetLatestValue is expected to return the same bound contract and method in the same test
	// Any setup required for this should be done in Setup.
	// The contract should take a LatestParams as the params and return the nth TestStruct set
	SetLatestValue(t *testing.T, ctx context.Context, testStruct *TestStruct) (types.BoundContract, string)
	GetPrimitiveContract(t *testing.T, ctx context.Context) (types.BoundContract, string)
}

const (
	TestItemType                    = "TestItem"
	TestItemSliceType               = "TestItemSliceType"
	TestItemArray1Type              = "TestItemArray1Type"
	TestItemArray2Type              = "TestItemArray2Type"
	AnyValueToReadWithoutAnArgument = uint64(3)
)

// RunChainReaderInterfaceTests uses TestStruct and TestStructWithSpecialFields
func RunChainReaderInterfaceTests(t *testing.T, tester ChainReaderInterfaceTester) {
	ctx := context.Background()
	tests := map[string]func(t *testing.T){
		"Encodes and decodes a single item": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
			req := &EncodeRequest{TestStructs: []TestStruct{item}, TestOn: TestItemType}
			resp := tester.EncodeFields(t, req)

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, item, TestItemType)
			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			into := TestStruct{}
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
			assert.Equal(t, item, into)
		},
		"Encodes compatible types": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
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

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, compatibleItem, TestItemType)
			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			into := TestStruct{}
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
			assert.Equal(t, item, into)
		},
		"Encodes compatible maps": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
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

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, compatibleMap, TestItemType)
			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			into := TestStruct{}
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemType))
			assert.Equal(t, item, into)
		},
		"Encodes and decodes a slice": func(t *testing.T) {
			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			item2 := CreateTestStruct(1, tester.GetAccountBytes)
			items := []TestStruct{item1, item2}
			req := &EncodeRequest{TestStructs: items, TestOn: TestItemSliceType}
			resp := tester.EncodeFields(t, req)

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, items, TestItemSliceType)
			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			var into []TestStruct
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemSliceType))
			assert.Equal(t, items, into)
		},
		"Encodes and decodes a slices with one element": func(t *testing.T) {
			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			items := []TestStruct{item1}
			req := &EncodeRequest{TestStructs: items, TestOn: TestItemSliceType}
			resp := tester.EncodeFields(t, req)

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, items, TestItemSliceType)

			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			var into []TestStruct
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemSliceType))
			assert.Equal(t, items, into)
		},
		"Encodes and decodes an array": func(t *testing.T) {
			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			item2 := CreateTestStruct(1, tester.GetAccountBytes)
			items := [2]TestStruct{item1, item2}
			req := &EncodeRequest{TestStructs: items[:], TestOn: TestItemArray2Type}
			resp := tester.EncodeFields(t, req)

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, items, TestItemArray2Type)

			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			var into [2]TestStruct
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemArray2Type))
			assert.Equal(t, items, into)
		},
		"Encodes and decodes a arrays with one element": func(t *testing.T) {
			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			items := [1]TestStruct{item1}
			req := &EncodeRequest{TestStructs: items[:], TestOn: TestItemArray1Type}
			resp := tester.EncodeFields(t, req)

			codec := tester.GetChainReader(t)
			actualEncoding, err := codec.Encode(ctx, items, TestItemArray1Type)

			require.NoError(t, err)
			assert.Equal(t, resp, actualEncoding)

			var into [1]TestStruct
			require.NoError(t, codec.Decode(ctx, actualEncoding, &into, TestItemArray1Type))
			assert.Equal(t, items, into)
		},
		"Returns an error if type is undefined": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
			codec := tester.GetChainReader(t)

			_, err := codec.Encode(ctx, item, "NOT"+TestItemType)
			assert.IsType(t, types.InvalidTypeError{}, err)

			err = codec.Decode(ctx, []byte(""), item, "NOT"+TestItemType)
			assert.IsType(t, types.InvalidTypeError{}, err)
		},
		"Returns an error encoding if arrays are the too small to encode": func(t *testing.T) {
			if !tester.IncludeArrayEncodingSizeEnforcement() {
				return
			}

			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			items := [1]TestStruct{item1}
			codec := tester.GetChainReader(t)

			_, err := codec.Encode(ctx, items, TestItemArray2Type)
			assert.IsType(t, types.InvalidTypeError{}, err)
		},
		"Returns an error encoding if arrays are the too large to encode": func(t *testing.T) {
			if !tester.IncludeArrayEncodingSizeEnforcement() {
				return
			}

			item1 := CreateTestStruct(0, tester.GetAccountBytes)
			item2 := CreateTestStruct(1, tester.GetAccountBytes)
			items := [2]TestStruct{item1, item2}
			codec := tester.GetChainReader(t)

			_, err := codec.Encode(ctx, items, TestItemArray1Type)
			assert.IsType(t, types.InvalidTypeError{}, err)
		},
		"Gets the latest value": func(t *testing.T) {
			firstItem := CreateTestStruct(0, tester.GetAccountBytes)
			bc, method := tester.SetLatestValue(t, ctx, &firstItem)
			secondItem := CreateTestStruct(1, tester.GetAccountBytes)
			tester.SetLatestValue(t, ctx, &secondItem)

			cr := tester.GetChainReader(t)
			actual := &TestStruct{}
			params := &LatestParams{I: 1}

			require.NoError(t, cr.GetLatestValue(ctx, bc, method, params, actual))
			assert.Equal(t, &firstItem, actual)

			params.I = 2
			actual = &TestStruct{}
			require.NoError(t, cr.GetLatestValue(ctx, bc, method, params, actual))
			assert.Equal(t, &secondItem, actual)
		},
		"Get latest value without arguments and with primitive return": func(t *testing.T) {
			bc, method := tester.GetPrimitiveContract(t, ctx)

			cr := tester.GetChainReader(t)

			var prim uint64
			require.NoError(t, cr.GetLatestValue(ctx, bc, method, nil, &prim))

			assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
		},
		"GetMaxEncodingSize returns errors for unknown types": func(t *testing.T) {
			cr := tester.GetChainReader(t)
			_, err := cr.GetMaxEncodingSize(ctx, 10, "not"+TestItemType)
			assert.IsType(t, types.InvalidTypeError{}, err)
		},
	}

	runTests(t, tester, tests)
}

// RunChainReaderWithStrictArgsInterfaceTest is meant to be used by codecs that don't pad
// They can assure that the right argument size is verified.
// Padding makes that harder/impossible to verify for come codecs.
// However, the extra verification is nice to have when possible.
func RunChainReaderWithStrictArgsInterfaceTest(t *testing.T, tester ChainReaderInterfaceTester) {
	ctx := context.Background()
	RunChainReaderInterfaceTests(t, tester)

	tests := map[string]func(t *testing.T){
		"Gives an error decoding extra fields on an item": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
			req := &EncodeRequest{
				TestStructs: []TestStruct{item},
				ExtraField:  true,
				TestOn:      TestItemType,
			}
			resp := tester.EncodeFields(t, req)
			codec := tester.GetChainReader(t)
			err := codec.Decode(ctx, resp, &item, TestItemType)
			assert.IsType(t, types.InvalidEncodingError{}, err)
		},
		"Gives an error decoding missing fields on an item": func(t *testing.T) {
			item := CreateTestStruct(0, tester.GetAccountBytes)
			req := &EncodeRequest{
				TestStructs:  []TestStruct{item},
				MissingField: true,
				TestOn:       TestItemType,
			}
			resp := tester.EncodeFields(t, req)
			codec := tester.GetChainReader(t)
			err := codec.Decode(ctx, resp, &item, TestItemType)
			assert.IsType(t, types.InvalidEncodingError{}, err)
		},
		"Gives an error decoding extra fields on a slice": func(t *testing.T) {
			items := []TestStruct{CreateTestStruct(0, tester.GetAccountBytes)}
			req := &EncodeRequest{
				TestStructs: items,
				ExtraField:  true,
				TestOn:      TestItemSliceType,
			}
			resp := tester.EncodeFields(t, req)
			codec := tester.GetChainReader(t)
			err := codec.Decode(ctx, resp, &items, TestItemSliceType)
			assert.IsType(t, types.InvalidEncodingError{}, err)
		},
		"Gives an error decoding missing fields on an slice": func(t *testing.T) {
			items := []TestStruct{CreateTestStruct(0, tester.GetAccountBytes)}
			req := &EncodeRequest{
				TestStructs:  items,
				MissingField: true,
				TestOn:       TestItemSliceType,
			}
			resp := tester.EncodeFields(t, req)
			codec := tester.GetChainReader(t)
			err := codec.Decode(ctx, resp, &items, TestItemSliceType)
			assert.IsType(t, types.InvalidEncodingError{}, err)
		},
	}

	runTests(t, tester, tests)
}

func runTests(t *testing.T, tester ChainReaderInterfaceTester, tests map[string]func(t *testing.T)) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tester.Setup(t)
			defer func() { tester.Teardown(t) }()
			test(t)
		})
	}
}

type InnerTestStruct struct {
	I int
	S string
}

type MidLevelTestStruct struct {
	FixedBytes [2]byte
	Inner      InnerTestStruct
}

type TestStruct struct {
	Field          int32
	DifferentField string
	OracleId       commontypes.OracleID
	OracleIds      [32]commontypes.OracleID
	Account        []byte
	Accounts       [][]byte
	BigField       *big.Int
	NestedStruct   MidLevelTestStruct
}

// compatibleTestStruct has fields in a different order
type compatibleTestStruct struct {
	Account        []byte
	Accounts       [][]byte
	BigField       *big.Int
	DifferentField string
	Field          int32
	NestedStruct   MidLevelTestStruct
	OracleId       commontypes.OracleID
	OracleIds      [32]commontypes.OracleID
}

type LatestParams struct {
	I int
}

func CreateTestStruct(i int, accGen func(int) []byte) TestStruct {
	s := fmt.Sprintf("field%v", i)
	return TestStruct{
		Field:          int32(i),
		DifferentField: s,
		OracleId:       commontypes.OracleID(i + 1),
		OracleIds:      [32]commontypes.OracleID{commontypes.OracleID(i + 2), commontypes.OracleID(i + 3)},
		Account:        accGen(i + 3),
		Accounts:       [][]byte{accGen(i + 4), accGen(i + 5)},
		BigField:       big.NewInt(int64((i + 1) * (i + 2))),
		NestedStruct: MidLevelTestStruct{
			FixedBytes: [2]byte{uint8(i), uint8(i + 1)},
			Inner: InnerTestStruct{
				I: i,
				S: s,
			},
		},
	}
}
