package interfacetests

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type ChainComponentsInterfaceTester[T TestingT[T]] interface {
	BasicTester[T]
	StartServices(context.Context, T)
	CloseServices(T)
	GetContractReader(t T) types.ContractReader
	GetChainWriter(t T) types.ChainWriter
	GetBindings(t T) []types.BoundContract
	// DirtyContracts signals to the underlying tester than the test contracts are dirty, i.e. the state has been changed such that
	// new, fresh contracts should be deployed. This usually happens after a value is written to the contract via
	// the ChainWriter.
	DirtyContracts()
	MaxWaitTimeForEvents() time.Duration
	// GenerateBlocksTillConfidenceLevel is only used by the internal common tests, all other tests can/should
	// rely on the ChainWriter waiting for actual blocks to be mined.
	GenerateBlocksTillConfidenceLevel(t T, contractName, readName string, confidenceLevel primitives.ConfidenceLevel)
}

const (
	AnyValueToReadWithoutAnArgument             = uint64(3)
	AnyDifferentValueToReadWithoutAnArgument    = uint64(1990)
	MethodTakingLatestParamsReturningTestStruct = "GetLatestValues"
	MethodReturningUint64                       = "GetPrimitiveValue"
	MethodReturningAlterableUint64              = "GetAlterablePrimitiveValue"
	MethodReturningUint64Slice                  = "GetSliceValue"
	MethodReturningSeenStruct                   = "GetSeenStruct"
	MethodSettingStruct                         = "addTestStruct"
	MethodSettingUint64                         = "setAlterablePrimitiveValue"
	MethodTriggeringEvent                       = "triggerEvent"
	EventName                                   = "SomeEvent"
	EventWithFilterName                         = "SomeEventToFilter"
	AnyContractName                             = "TestContract"
	AnySecondContractName                       = "Not" + AnyContractName
)

var AnySliceToReadWithoutAnArgument = []uint64{3, 4}

const AnyExtraValue = 3

func RunContractReaderInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], mockRun bool) {
	t.Run("GetLatestValue for "+tester.Name(), func(t T) { runContractReaderGetLatestValueInterfaceTests(t, tester, mockRun) })
	t.Run("BatchGetLatestValues for "+tester.Name(), func(t T) { runContractReaderBatchGetLatestValuesInterfaceTests(t, tester, mockRun) })
	t.Run("QueryKey for "+tester.Name(), func(t T) { runQueryKeyInterfaceTests(t, tester) })
}

func runContractReaderGetLatestValueInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], mockRun bool) {
	tests := []testcase[T]{
		{
			name: "Get latest value without starting service returns error",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0] // minimum of one bound contract expected, otherwise panics
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				require.Error(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, nil, nil))
			},
		},
		{
			name: "Gets the latest value",
			test: func(t T) {
				ctx := tests.Context(t)
				firstItem := CreateTestStruct(0, tester)

				contracts := tester.GetBindings(t)
				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, firstItem, contracts[0], types.Unconfirmed)

				secondItem := CreateTestStruct(1, tester)

				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, secondItem, contracts[0], types.Unconfirmed)

				cr := tester.GetContractReader(t)
				fmt.Printf("TYpeof is %T \n", cr)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0] // minimum of one bound contract expected, otherwise panics

				require.NoError(t, cr.Bind(ctx, bindings))

				actual := &TestStruct{}
				params := &LatestParams{I: 1}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, actual))
				assert.Equal(t, &firstItem, actual)

				params.I = 2
				actual = &TestStruct{}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, actual))
				assert.Equal(t, &secondItem, actual)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value without arguments and with primitive return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value based on confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

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

				contracts := tester.GetBindings(t)

				txID := SubmitTransactionToCW(t, tester, MethodSettingUint64, PrimitiveArgs{Value: 10}, contracts[0], types.Unconfirmed)

				var prim1 uint64
				bound := bindingsByName(bindings, callArgs.ContractName)[0]

				require.Error(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), primitives.Finalized, callArgs.Params, &prim1))

				err := WaitForTransactionStatus(t, tester, txID, types.Finalized, mockRun)
				require.NoError(t, err)

				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), primitives.Finalized, callArgs.Params, &prim1))
				assert.Equal(t, uint64(10), prim1)

				_ = SubmitTransactionToCW(t, tester, MethodSettingUint64, PrimitiveArgs{Value: 20}, contracts[0], types.Unconfirmed)

				var prim2 uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), callArgs.ConfidenceLevel, callArgs.Params, &prim2))
				assert.Equal(t, uint64(20), prim2)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value allows multiple contract names to have the same function name",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnySecondContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, prim)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value without arguments and with slice return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var slice []uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64Slice), primitives.Unconfirmed, nil, &slice))

				assert.Equal(t, AnySliceToReadWithoutAnArgument, slice)

				tester.CloseServices(t)
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
				bound := bindingsByName(bindings, AnyContractName)[0]

				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))

				actual := &TestStructWithExtraField{}
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningSeenStruct), primitives.Unconfirmed, testStruct, actual))

				expected := &TestStructWithExtraField{
					ExtraField: AnyExtraValue,
					TestStruct: CreateTestStruct(0, tester),
				}

				assert.Equal(t, expected, actual)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value gets latest event",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))
				contracts := tester.GetBindings(t)

				ts := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts, contracts[0], types.Unconfirmed)

				ts = CreateTestStruct[T](1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts, contracts[0], types.Unconfirmed)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest event based on provided confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))
				ts1 := CreateTestStruct[T](2, tester)

				txID := SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, bindings[0], types.Unconfirmed)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err != nil && assert.ErrorContains(t, err, types.ErrNotFound.Error())
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				err := WaitForTransactionStatus(t, tester, txID, types.Finalized, mockRun)
				require.NoError(t, err)

				ts2 := CreateTestStruct[T](3, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, bindings[0], types.Unconfirmed)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts2)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value returns not found if event was never triggered",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
				assert.True(t, errors.Is(err, types.ErrNotFound))

				tester.CloseServices(t)
			},
		},
		{
			name: "Get latest value gets latest event with filtering",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))
				ts0 := CreateTestStruct(0, tester)

				contracts := tester.GetBindings(t)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts0, contracts[0], types.Unconfirmed)
				ts1 := CreateTestStruct(1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, contracts[0], types.Unconfirmed)

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

				tester.CloseServices(t)
			},
		},
	}
	runTests(t, tester, tests)
}

func runContractReaderBatchGetLatestValuesInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], mockRun bool) {
	testCases := []testcase[T]{
		{
			name: "BatchGetLatestValues without starting service returns error",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)

				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				_, err := cr.BatchGetLatestValues(ctx, make(types.BatchGetLatestValuesRequest))
				require.Error(t, err)
			},
		},
		{
			name: "BatchGetLatestValues works",
			test: func(t T) {
				// setup test data
				firstItem := CreateTestStruct(1, tester)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				batchCallEntry := make(BatchCallEntry)
				batchCallEntry[bound] = ContractBatchEntry{{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &firstItem}}
				batchChainWrite(t, tester, batchCallEntry, mockRun)

				// setup call data
				params, actual := &LatestParams{I: 1}, &TestStruct{}
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				batchGetLatestValueRequest[bound] = []types.BatchRead{
					{
						ReadName:  MethodTakingLatestParamsReturningTestStruct,
						Params:    params,
						ReturnVal: actual,
					},
				}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				assert.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodTakingLatestParamsReturningTestStruct)
				assert.Equal(t, &firstItem, returnValue)

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValues works without arguments and with primitive return",
			test: func(t T) {
				// setup call data
				var primitiveReturnValue uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValuesRequest[bound] = []types.BatchRead{
					{
						ReadName:  MethodReturningUint64,
						Params:    nil,
						ReturnVal: &primitiveReturnValue,
					},
				}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64)
				assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValue.(*uint64))

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValues allows multiple contract names to have the same function Name",
			test: func(t T) {
				var primitiveReturnValueAnyContract, primitiveReturnValueAnySecondContract uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := bindingsByName(bindings, AnyContractName)[0]
				bound2 := bindingsByName(bindings, AnySecondContractName)[0]

				batchGetLatestValuesRequest[bound1] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnyContract}}
				batchGetLatestValuesRequest[bound2] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)

				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
				require.NoError(t, err)

				anyContractBatch, anySecondContractBatch := result[bound1], result[bound2]
				returnValueAnyContract, errAnyContract := anyContractBatch[0].GetResult()
				returnValueAnySecondContract, errAnySecondContract := anySecondContractBatch[0].GetResult()
				require.NoError(t, errAnyContract)
				require.NoError(t, errAnySecondContract)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64)
				assert.Contains(t, anySecondContractBatch[0].ReadName, MethodReturningUint64)
				assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValueAnyContract.(*uint64))
				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, *returnValueAnySecondContract.(*uint64))

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValue without arguments and with slice return",
			test: func(t T) {
				// setup call data
				var sliceReturnValue []uint64
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValueRequest[bound] = []types.BatchRead{{ReadName: MethodReturningUint64Slice, Params: nil, ReturnVal: &sliceReturnValue}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64Slice)
				assert.Equal(t, AnySliceToReadWithoutAnArgument, *returnValue.(*[]uint64))

				tester.CloseServices(t)
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
				bound := bindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValueRequest[bound] = []types.BatchRead{{ReadName: MethodReturningSeenStruct, Params: testStruct, ReturnVal: actual}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningSeenStruct)
				assert.Equal(t,
					&TestStructWithExtraField{
						ExtraField: AnyExtraValue,
						TestStruct: CreateTestStruct(0, tester),
					},
					returnValue)

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValues supports same read with different params and results retain order from request",
			test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				for i := 0; i < 10; i++ {
					// setup test data
					ts := CreateTestStruct(i, tester)
					batchCallEntry[bound] = append(batchCallEntry[bound], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts})
					// setup call data
					batchGetLatestValueRequest[bound] = append(
						batchGetLatestValueRequest[bound],
						types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}},
					)
				}
				batchChainWrite(t, tester, batchCallEntry, mockRun)

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					resultAnyContract, testDataAnyContract := result[bound], batchCallEntry[bound]
					returnValue, err := resultAnyContract[i].GetResult()
					assert.NoError(t, err)
					assert.Contains(t, resultAnyContract[i].ReadName, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, testDataAnyContract[i].ReturnValue, returnValue)
				}

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValues supports same read with different params and results retain order from request even with multiple contracts",
			test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := bindingsByName(bindings, AnyContractName)[0]
				bound2 := bindingsByName(bindings, AnySecondContractName)[0]

				for i := 0; i < 10; i++ {
					// setup test data
					ts1, ts2 := CreateTestStruct(i, tester), CreateTestStruct(i+10, tester)
					batchCallEntry[bound1] = append(batchCallEntry[bound1], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts1})
					batchCallEntry[bound2] = append(batchCallEntry[bound2], ReadEntry{Name: MethodTakingLatestParamsReturningTestStruct, ReturnValue: &ts2})
					// setup call data
					batchGetLatestValueRequest[bound1] = append(batchGetLatestValueRequest[bound1], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}})
					batchGetLatestValueRequest[bound2] = append(batchGetLatestValueRequest[bound2], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 1 + i}, ReturnVal: &TestStruct{}})
				}
				batchChainWrite(t, tester, batchCallEntry, mockRun)

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for idx := 0; idx < 10; idx++ {
					fmt.Printf("expected: %+v\n", batchCallEntry[bound1][idx].ReturnValue)
					if val, err := result[bound1][idx].GetResult(); err == nil {
						fmt.Printf("result: %+v\n", val)
					}
				}

				for i := 0; i < 10; i++ {
					testDataAnyContract, testDataAnySecondContract := batchCallEntry[bound1], batchCallEntry[bound2]
					resultAnyContract, resultAnySecondContract := result[bound1], result[bound2]
					returnValueAnyContract, errAnyContract := resultAnyContract[i].GetResult()
					returnValueAnySecondContract, errAnySecondContract := resultAnySecondContract[i].GetResult()
					assert.NoError(t, errAnyContract)
					assert.NoError(t, errAnySecondContract)
					assert.Contains(t, resultAnyContract[i].ReadName, MethodTakingLatestParamsReturningTestStruct)
					assert.Contains(t, resultAnySecondContract[i].ReadName, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, testDataAnyContract[i].ReturnValue, returnValueAnyContract)
					assert.Equal(t, testDataAnySecondContract[i].ReturnValue, returnValueAnySecondContract)
				}

				tester.CloseServices(t)
			},
		},
		{
			name: "BatchGetLatestValues sets errors properly",
			test: func(t T) {
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := bindingsByName(bindings, AnyContractName)[0]
				bound2 := bindingsByName(bindings, AnySecondContractName)[0]

				for i := 0; i < 10; i++ {
					// setup call data and set invalid params that cause an error
					batchGetLatestValueRequest[bound1] = append(batchGetLatestValueRequest[bound1], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
					batchGetLatestValueRequest[bound2] = append(batchGetLatestValueRequest[bound2], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
				}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				for i := 0; i < 10; i++ {
					resultAnyContract, resultAnySecondContract := result[bound1], result[bound2]
					returnValueAnyContract, errAnyContract := resultAnyContract[i].GetResult()
					returnValueAnySecondContract, errAnySecondContract := resultAnySecondContract[i].GetResult()
					assert.Error(t, errAnyContract)
					assert.Error(t, errAnySecondContract)
					assert.Contains(t, resultAnyContract[i].ReadName, MethodTakingLatestParamsReturningTestStruct)
					assert.Contains(t, resultAnySecondContract[i].ReadName, MethodTakingLatestParamsReturningTestStruct)
					assert.Equal(t, &TestStruct{}, returnValueAnyContract)
					assert.Equal(t, &TestStruct{}, returnValueAnySecondContract)
				}

				tester.CloseServices(t)
			},
		},
	}
	runTests(t, tester, testCases)
}

func runQueryKeyInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T]) {
	tests := []testcase[T]{
		{
			name: "QueryKey without starting service returns error",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				_, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &TestStruct{})
				fmt.Println("err is ", err)
				require.Error(t, err)
			},
		},
		{
			name: "QueryKey returns not found if sequence never happened",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				bindings := tester.GetBindings(t)
				bound := bindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				logs, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &TestStruct{})
				require.NoError(t, err)
				assert.Len(t, logs, 0)

				tester.CloseServices(t)
			},
		},
		{
			name: "QueryKey returns sequence data properly",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				tester.StartServices(ctx, t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				bound := bindingsByName(bindings, AnyContractName)[0]

				ts1 := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, bindings[0], types.Unconfirmed)
				ts2 := CreateTestStruct[T](1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, bindings[0], types.Unconfirmed)

				ts := &TestStruct{}
				assert.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, ts)
					return err == nil && len(sequences) == 2 && reflect.DeepEqual(&ts1, sequences[1].Data) && reflect.DeepEqual(&ts2, sequences[0].Data)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				tester.CloseServices(t)
			},
		},
	}

	runTests(t, tester, tests)
}

func bindingsByName(bindings []types.BoundContract, name string) []types.BoundContract {
	named := make([]types.BoundContract, 0, len(bindings))

	for idx := range bindings {
		if bindings[idx].Name == name {
			named = append(named, bindings[idx])
		}
	}

	return named
}
