package interfacetests

import (
	"errors"
	"reflect"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type ChainReaderInterfaceTester[T TestingT[T]] interface {
	BasicTester[T]
	GetChainReader(t T) types.ContractReader
	// SetTestStructLatestValue is expected to return the same bound contract and method in the same test
	// Any setup required for this should be done in Setup.
	// The contract should take a LatestParams as the params and return the nth TestStruct set
	SetTestStructLatestValue(t T, testStruct *TestStruct)
	// SetUintLatestValue is expected to return the same bound contract and method in the same test
	// Any setup required for this should be done in Setup.
	// The contract should take a uint64 as the params and returns the same.
	// forCall is used to attach value to a call, this is useful in chain specific test since in chain agnostic tests we can just use hard coded readIdentifier constants.
	SetUintLatestValue(t T, val uint64, forCall ExpectedGetLatestValueArgs)
	SetBatchLatestValues(t T, batchCallEntry BatchCallEntry)
	TriggerEvent(t T, testStruct *TestStruct)
	GetBindings(t T) []types.BoundContract
	// GenerateBlocksTillConfidenceLevel raises confidence level to the provided level for a specific read.
	GenerateBlocksTillConfidenceLevel(t T, contractName, readName string, confidenceLevel primitives.ConfidenceLevel)
	MaxWaitTimeForEvents() time.Duration
}

const (
	AnyValueToReadWithoutAnArgument             = uint64(3)
	AnyDifferentValueToReadWithoutAnArgument    = uint64(1990)
	MethodTakingLatestParamsReturningTestStruct = "GetLatestValues"
	MethodReturningUint64                       = "GetPrimitiveValue"
	MethodReturningAlterableUint64              = "GetAlterablePrimitiveValue"
	MethodReturningUint64Slice                  = "GetSliceValue"
	MethodReturningSeenStruct                   = "GetSeenStruct"
	EventName                                   = "SomeEvent"
	EventWithFilterName                         = "SomeEventToFilter"
	AnyContractName                             = "TestContract"
	AnySecondContractName                       = "Not" + AnyContractName
)

var AnySliceToReadWithoutAnArgument = []uint64{3, 4}

const AnyExtraValue = 3

func RunChainReaderInterfaceTests[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T]) {
	t.Run("GetLatestValue for "+tester.Name(), func(t T) { runChainReaderGetLatestValueInterfaceTests(t, tester) })
	t.Run("BatchGetLatestValues for "+tester.Name(), func(t T) { runChainReaderBatchGetLatestValuesInterfaceTests(t, tester) })
	t.Run("QueryKey for "+tester.Name(), func(t T) { runQueryKeyInterfaceTests(t, tester) })
}

func runChainReaderGetLatestValueInterfaceTests[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T]) {
	tests := []testcase[T]{
		{
			name: "Gets the latest value",
			test: func(t T) {
				ctx := tests.Context(t)
				firstItem := CreateTestStruct(0, tester)
				tester.SetTestStructLatestValue(t, &firstItem)
				secondItem := CreateTestStruct(1, tester)
				tester.SetTestStructLatestValue(t, &secondItem)

				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				actual := &TestStruct{}
				params := &LatestParams{I: 1}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, actual))
				assert.Equal(t, &firstItem, actual)

				params.I = 2
				actual = &TestStruct{}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, actual))
				assert.Equal(t, &secondItem, actual)
			},
		},
		{
			name: "Get latest value without arguments and with primitive return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
			},
		},
		{
			name: "Get latest value based on confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))

				var returnVal1 uint64
				callArgs := ExpectedGetLatestValueArgs{
					ContractName:    AnyContractName,
					ReadName:        MethodReturningAlterableUint64,
					ConfidenceLevel: primitives.Unconfirmed,
					Params:          nil,
					ReturnVal:       &returnVal1,
				}

				var prim1 uint64
				tester.SetUintLatestValue(t, 10, callArgs)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == callArgs.ContractName {
						bound = bindings[idx]
					}
				}

				require.Error(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), primitives.Finalized, callArgs.Params, &prim1))

				tester.GenerateBlocksTillConfidenceLevel(t, AnyContractName, MethodReturningAlterableUint64, primitives.Finalized)
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), primitives.Finalized, nil, &prim1))
				assert.Equal(t, uint64(10), prim1)

				var returnVal2 uint64
				callArgs2 := ExpectedGetLatestValueArgs{
					ContractName:    AnyContractName,
					ReadName:        MethodReturningAlterableUint64,
					ConfidenceLevel: primitives.Unconfirmed,
					Params:          nil,
					ReturnVal:       returnVal2,
				}

				var prim2 uint64
				tester.SetUintLatestValue(t, 20, callArgs2)
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), callArgs.ConfidenceLevel, callArgs.Params, &prim2))
				assert.Equal(t, uint64(20), prim2)
			},
		},
		{
			name: "Get latest value allows multiple contract names to have the same function name",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)
				seenAddrs := map[string]bool{}

				var bound types.BoundContract

				for idx, binding := range bindings {
					assert.False(t, seenAddrs[binding.Address])
					seenAddrs[binding.Address] = true

					if binding.Name == AnySecondContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, prim)
			},
		},
		{
			name: "Get latest value without arguments and with slice return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				var slice []uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64Slice), primitives.Unconfirmed, nil, &slice))

				assert.Equal(t, AnySliceToReadWithoutAnArgument, slice)
			},
		},
		{
			name: "Get latest value wraps config with modifiers using its own mapstructure overrides",
			test: func(t T) {
				ctx := tests.Context(t)
				testStruct := CreateTestStruct(0, tester)
				testStruct.BigField = nil
				testStruct.Account = nil

				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				actual := &TestStructWithExtraField{}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningSeenStruct), primitives.Unconfirmed, testStruct, actual))

				expected := &TestStructWithExtraField{
					ExtraField: AnyExtraValue,
					TestStruct: CreateTestStruct(0, tester),
				}

				assert.Equal(t, expected, actual)
			},
		},
		{
			name: "Get latest value gets latest event",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))
				ts := CreateTestStruct(0, tester)
				tester.TriggerEvent(t, &ts)
				ts = CreateTestStruct(1, tester)
				tester.TriggerEvent(t, &ts)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			name: "Get latest event based on provided confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))
				ts1 := CreateTestStruct[T](2, tester)
				tester.TriggerEvent(t, &ts1)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err != nil && assert.ErrorContains(t, err, types.ErrNotFound.Error())
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				tester.GenerateBlocksTillConfidenceLevel(t, AnyContractName, EventName, primitives.Finalized)
				ts2 := CreateTestStruct[T](3, tester)
				tester.TriggerEvent(t, &ts2)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts2)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			name: "Get latest value returns not found if event was never triggered",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
				assert.True(t, errors.Is(err, types.ErrNotFound))
			},
		},
		{
			name: "Get latest value gets latest event with filtering",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))
				ts0 := CreateTestStruct(0, tester)
				tester.TriggerEvent(t, &ts0)
				ts1 := CreateTestStruct(1, tester)
				tester.TriggerEvent(t, &ts1)

				filterParams := &FilterEventParams{Field: *ts0.Field}
				assert.Never(t, func() bool {
					result := &TestStruct{}
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventWithFilterName), primitives.Unconfirmed, filterParams, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				// get the result one more time to verify it.
				// Using the result from the Never statement by creating result outside the block is a data race
				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventWithFilterName), primitives.Unconfirmed, filterParams, &result)
				require.NoError(t, err)
				assert.Equal(t, &ts0, result)
			},
		},
	}
	runTests(t, tester, tests)
}

func runChainReaderBatchGetLatestValuesInterfaceTests[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T]) {
	testCases := []testcase[T]{
		{
			name: "BatchGetLatestValues works",
			test: func(t T) {
				// setup test data
				firstItem := CreateTestStruct(1, tester)
				batchCallEntry := make(BatchCallEntry)
				batchCallEntry[AnyContractName] = ContractBatchEntry{{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &firstItem}}
				tester.SetBatchLatestValues(t, batchCallEntry)

				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				// setup call data
				params, actual := &LatestParams{I: 1}, &TestStruct{}
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				batchGetLatestValueRequest[AnyContractName] = []types.BatchRead{
					{
						ReadIdentifier: bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct),
						Params:         params,
						ReturnVal:      actual,
					},
				}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[AnyContractName]
				returnValue, err := anyContractBatch[0].GetResult()
				assert.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
				assert.Equal(t, &firstItem, returnValue)
			},
		},
		{
			name: "BatchGetLatestValues works without arguments and with primitive return",
			test: func(t T) {
				// setup call data
				var primitiveReturnValue uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				batchGetLatestValuesRequest[AnyContractName] = []types.BatchRead{
					{
						ReadIdentifier: bound.ReadIdentifier(MethodReturningUint64),
						Params:         nil,
						ReturnVal:      &primitiveReturnValue,
					},
				}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
				require.NoError(t, err)

				anyContractBatch := result[AnyContractName]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadIdentifier, MethodReturningUint64)
				assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValue.(*uint64))
			},
		},
		{
			name: "BatchGetLatestValues allows multiple contract names to have the same function Name",
			test: func(t T) {
				var primitiveReturnValueAnyContract, primitiveReturnValueAnySecondContract uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound1 types.BoundContract
				var bound2 types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound1 = bindings[idx]
					} else if bindings[idx].Name == AnySecondContractName {
						bound2 = bindings[idx]
					}
				}

				batchGetLatestValuesRequest[AnyContractName] = []types.BatchRead{{ReadIdentifier: bound1.ReadIdentifier(MethodReturningUint64), Params: nil, ReturnVal: &primitiveReturnValueAnyContract}}
				batchGetLatestValuesRequest[AnySecondContractName] = []types.BatchRead{{ReadIdentifier: bound2.ReadIdentifier(MethodReturningUint64), Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract}}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
				require.NoError(t, err)

				anyContractBatch, anySecondContractBatch := result[AnyContractName], result[AnySecondContractName]
				returnValueAnyContract, errAnyContract := anyContractBatch[0].GetResult()
				returnValueAnySecondContract, errAnySecondContract := anySecondContractBatch[0].GetResult()
				require.NoError(t, errAnyContract)
				require.NoError(t, errAnySecondContract)
				assert.Contains(t, anyContractBatch[0].ReadIdentifier, MethodReturningUint64)
				assert.Contains(t, anySecondContractBatch[0].ReadIdentifier, MethodReturningUint64)
				assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValueAnyContract.(*uint64))
				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, *returnValueAnySecondContract.(*uint64))
			},
		},
		{
			name: "BatchGetLatestValue without arguments and with slice return",
			test: func(t T) {
				// setup call data
				var sliceReturnValue []uint64
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				batchGetLatestValueRequest[AnyContractName] = []types.BatchRead{{ReadIdentifier: bound.ReadIdentifier(MethodReturningUint64Slice), Params: nil, ReturnVal: &sliceReturnValue}}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[AnyContractName]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadIdentifier, MethodReturningUint64Slice)
				assert.Equal(t, AnySliceToReadWithoutAnArgument, *returnValue.(*[]uint64))
			},
		},
		{
			name: "BatchGetLatestValues wraps config with modifiers using its own mapstructure overrides",
			test: func(t T) {
				// setup call data
				testStruct := CreateTestStruct(0, tester)
				testStruct.BigField = nil
				testStruct.Account = nil
				actual := &TestStructWithExtraField{}
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				batchGetLatestValueRequest[AnyContractName] = []types.BatchRead{{ReadIdentifier: bound.ReadIdentifier(MethodReturningSeenStruct), Params: testStruct, ReturnVal: actual}}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[AnyContractName]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadIdentifier, MethodReturningSeenStruct)
				assert.Equal(t,
					&TestStructWithExtraField{
						ExtraField: AnyExtraValue,
						TestStruct: CreateTestStruct(0, tester),
					},
					returnValue)
			},
		},
		{
			name: "BatchGetLatestValues supports same read with different params and results retain order from request",
			test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				for i := 0; i < 10; i++ {
					// setup test data
					ts := CreateTestStruct(i, tester)
					batchCallEntry[AnyContractName] = append(batchCallEntry[AnyContractName], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts})
					// setup call data
					batchGetLatestValueRequest[AnyContractName] = append(
						batchGetLatestValueRequest[AnyContractName],
						types.BatchRead{ReadIdentifier: bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}},
					)
				}
				tester.SetBatchLatestValues(t, batchCallEntry)

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					resultAnyContract, testDataAnyContract := result[AnyContractName], batchCallEntry[AnyContractName]
					returnValue, err := resultAnyContract[i].GetResult()
					assert.NoError(t, err)
					assert.Contains(t, resultAnyContract[i].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, testDataAnyContract[i].ReturnValue, returnValue)
				}
			},
		},
		{
			name: "BatchGetLatestValues supports same read with different params and results retain order from request even with multiple contracts",
			test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound1 types.BoundContract
				var bound2 types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound1 = bindings[idx]
					} else if bindings[idx].Name == AnySecondContractName {
						bound2 = bindings[idx]
					}
				}

				for i := 0; i < 10; i++ {
					// setup test data
					ts1, ts2 := CreateTestStruct(i, tester), CreateTestStruct(i+10, tester)
					batchCallEntry[AnyContractName] = append(batchCallEntry[AnyContractName], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts1})
					batchCallEntry[AnySecondContractName] = append(batchCallEntry[AnySecondContractName], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts2})
					// setup call data
					batchGetLatestValueRequest[AnyContractName] = append(batchGetLatestValueRequest[AnyContractName], types.BatchRead{ReadIdentifier: bound1.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}})
					batchGetLatestValueRequest[AnySecondContractName] = append(batchGetLatestValueRequest[AnySecondContractName], types.BatchRead{ReadIdentifier: bound2.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}})
				}
				tester.SetBatchLatestValues(t, batchCallEntry)

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					testDataAnyContract, testDataAnySecondContract := batchCallEntry[AnyContractName], batchCallEntry[AnySecondContractName]
					resultAnyContract, resultAnySecondContract := result[AnyContractName], result[AnySecondContractName]
					returnValueAnyContract, errAnyContract := resultAnyContract[i].GetResult()
					returnValueAnySecondContract, errAnySecondContract := resultAnySecondContract[i].GetResult()
					assert.NoError(t, errAnyContract)
					assert.NoError(t, errAnySecondContract)
					assert.Contains(t, resultAnyContract[i].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
					assert.Contains(t, resultAnySecondContract[i].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, testDataAnyContract[i].ReturnValue, returnValueAnyContract)
					assert.Equal(t, testDataAnySecondContract[i].ReturnValue, returnValueAnySecondContract)
				}
			},
		},
		{
			name: "BatchGetLatestValues sets errors properly",
			test: func(t T) {
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)

				var bound1 types.BoundContract
				var bound2 types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound1 = bindings[idx]
					} else if bindings[idx].Name == AnySecondContractName {
						bound2 = bindings[idx]
					}
				}

				for i := 0; i < 10; i++ {
					// setup call data and set invalid params that cause an error
					batchGetLatestValueRequest[AnyContractName] = append(batchGetLatestValueRequest[AnyContractName], types.BatchRead{ReadIdentifier: bound1.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
					batchGetLatestValueRequest[AnySecondContractName] = append(batchGetLatestValueRequest[AnySecondContractName], types.BatchRead{ReadIdentifier: bound2.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
				}

				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					resultAnyContract, resultAnySecondContract := result[AnyContractName], result[AnySecondContractName]
					returnValueAnyContract, errAnyContract := resultAnyContract[i].GetResult()
					returnValueAnySecondContract, errAnySecondContract := resultAnySecondContract[i].GetResult()
					assert.Error(t, errAnyContract)
					assert.Error(t, errAnySecondContract)
					assert.Contains(t, resultAnyContract[i].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
					assert.Contains(t, resultAnySecondContract[i].ReadIdentifier, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, &TestStruct{}, returnValueAnyContract)
					assert.Equal(t, &TestStruct{}, returnValueAnySecondContract)
				}
			},
		},
	}

	runTests(t, tester, testCases)
}

func runQueryKeyInterfaceTests[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T]) {
	tests := []testcase[T]{
		{
			name: "QueryKey returns not found if sequence never happened",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				logs, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &TestStruct{})

				require.NoError(t, err)
				assert.Len(t, logs, 0)
			},
		},
		{
			name: "QueryKey returns sequence data properly",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)

				var bound types.BoundContract
				for idx := range bindings {
					if bindings[idx].Name == AnyContractName {
						bound = bindings[idx]
					}
				}

				require.NoError(t, cr.Bind(ctx, bindings))
				ts1 := CreateTestStruct(0, tester)
				tester.TriggerEvent(t, &ts1)
				ts2 := CreateTestStruct(1, tester)
				tester.TriggerEvent(t, &ts2)

				ts := &TestStruct{}
				assert.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, ts)
					return err == nil && len(sequences) == 2 && reflect.DeepEqual(&ts1, sequences[1].Data) && reflect.DeepEqual(&ts2, sequences[0].Data)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
	}

	runTests(t, tester, tests)
}
