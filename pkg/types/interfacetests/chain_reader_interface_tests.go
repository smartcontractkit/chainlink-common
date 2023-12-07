package interfacetests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type ChainReaderInterfaceTester interface {
	BasicTester
	GetChainReader(t *testing.T) types.ChainReader

	// SetLatestValue is expected to return the same bound contract and method in the same test
	// Any setup required for this should be done in Setup.
	// The contract should take a LatestParams as the params and return the nth TestStruct set
	SetLatestValue(ctx context.Context, t *testing.T, testStruct *TestStruct) types.BoundContract
	GetPrimitiveContract(ctx context.Context, t *testing.T) types.BoundContract
	GetSliceContract(ctx context.Context, t *testing.T) types.BoundContract
	GetReturnSeenContract(ctx context.Context, t *testing.T) types.BoundContract
}

const (
	AnyValueToReadWithoutAnArgument             = uint64(3)
	MethodTakingLatestParamsReturningTestStruct = "GetLatestValues"
	MethodReturningUint64                       = "GetPrimitiveValue"
	MethodReturningUint64Slice                  = "GetSliceValue"
	MethodReturningSeenStruct                   = "GetSeenStruct"
)

var AnySliceToReadWithoutAnArgument = []uint64{3, 4}

const AnyExtraValue = 3

func RunChainReaderInterfaceTests(t *testing.T, tester ChainReaderInterfaceTester) {
	ctx := tests.Context(t)
	tests := []testcase{
		{
			name: "Gets the latest value",
			test: func(t *testing.T) {
				firstItem := CreateTestStruct(0, tester)
				bc := tester.SetLatestValue(ctx, t, &firstItem)
				secondItem := CreateTestStruct(1, tester)
				tester.SetLatestValue(ctx, t, &secondItem)

				cr := tester.GetChainReader(t)
				actual := &TestStruct{}
				params := &LatestParams{I: 1}

				require.NoError(t, cr.GetLatestValue(ctx, bc, MethodTakingLatestParamsReturningTestStruct, params, actual))
				assert.Equal(t, &firstItem, actual)

				params.I = 2
				actual = &TestStruct{}
				require.NoError(t, cr.GetLatestValue(ctx, bc, MethodTakingLatestParamsReturningTestStruct, params, actual))
				assert.Equal(t, &secondItem, actual)
			},
		},
		{
			name: "Get latest value without arguments and with primitive return",
			test: func(t *testing.T) {
				bc := tester.GetPrimitiveContract(ctx, t)

				cr := tester.GetChainReader(t)

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bc, MethodReturningUint64, nil, &prim))

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
			},
		},
		{
			name: "Get latest value without arguments and with slice return",
			test: func(t *testing.T) {
				bc := tester.GetSliceContract(ctx, t)

				cr := tester.GetChainReader(t)

				var slice []uint64
				require.NoError(t, cr.GetLatestValue(ctx, bc, MethodReturningUint64Slice, nil, &slice))

				assert.Equal(t, AnySliceToReadWithoutAnArgument, slice)
			},
		},
		{
			name: "Get latest value wraps config with modifiers using its own mapstructure overrides",
			test: func(t *testing.T) {
				testStruct := CreateTestStruct(0, tester)
				testStruct.BigField = nil
				testStruct.Account = nil

				tester.Setup(t)

				cr := tester.GetChainReader(t)
				bc := tester.GetReturnSeenContract(ctx, t)

				actual := &TestStructWithExtraField{}
				require.NoError(t, cr.GetLatestValue(ctx, bc, MethodReturningSeenStruct, testStruct, actual))

				expected := &TestStructWithExtraField{
					ExtraField: AnyExtraValue,
					TestStruct: CreateTestStruct(0, tester),
				}

				assert.Equal(t, expected, actual)
			},
		},
	}
	runTests(t, tester, tests)
}
