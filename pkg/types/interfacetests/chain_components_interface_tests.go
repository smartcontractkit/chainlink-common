package interfacetests

import (
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
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// GetLatestValue method
const (
	ContractReaderGetLatestValueAsValuesDotValue                              = "Gets the latest value as a values.Value"
	ContractReaderGetLatestValueNoArgumentsAndPrimitiveReturnAsValuesDotValue = "Get latest value without arguments and with primitive return as a values.Value"
	ContractReaderGetLatestValueNoArgumentsAndSliceReturnAsValueDotValue      = "Get latest value without arguments and with slice return as a values.Value"
	ContractReaderGetLatestValue                                              = "Gets the latest value"
	ContractReaderGetLatestValueWithHeadData                                  = "Gets the latest value with head data"
	ContractReaderGetLatestValueWithPrimitiveReturn                           = "Get latest value without arguments and with primitive return"
	ContractReaderGetLatestValueBasedOnConfidenceLevel                        = "Get latest value based on confidence level"
	ContractReaderGetLatestValueFromMultipleContractsNamesSameFunction        = "Get latest value allows multiple contract names to have the same function "
	ContractReaderGetLatestValueWithModifiersUsingOwnMapstrctureOverrides     = "Get latest value wraps config with modifiers using its own mapstructure overrides"
	ContractReaderGetLatestValueNoArgumentsAndSliceReturn                     = "Get latest value without arguments and with slice return"
)

// GetLatestValue event
const (
	ContractReaderGetLatestValueGetsLatestForEvent                      = "Get latest value gets latest event"
	ContractReaderGetLatestValueBasedOnConfidenceLevelForEvent          = "Get latest event based on provided confidence level"
	ContractReaderGetLatestValueReturnsNotFoundWhenNotTriggeredForEvent = "Get latest value returns not found if event was never triggered"
	ContractReaderGetLatestValueWithFilteringForEvent                   = "Get latest value gets latest event with filtering"
)

// BatchGet
const (
	ContractReaderBatchGetLatestValue                                                   = "BatchGetLatestValues works"
	ContractReaderBatchGetLatestValueNoArgumentsPrimitiveReturn                         = "BatchGetLatestValues works without arguments and with primitive return"
	ContractReaderBatchGetLatestValueMultipleContractNamesSameFunction                  = "BatchGetLatestValues allows multiple contract names to have the same function Name"
	ContractReaderBatchGetLatestValueNoArgumentsWithSliceReturn                         = "BatchGetLatestValue without arguments and with slice return"
	ContractReaderBatchGetLatestValueWithModifiersOwnMapstructureOverride               = "BatchGetLatestValues wraps config with modifiers using its own mapstructure overrides"
	ContractReaderBatchGetLatestValueDifferentParamsResultsRetainOrder                  = "BatchGetLatestValues supports same read with different params and results retain order from request"
	ContractReaderBatchGetLatestValueDifferentParamsResultsRetainOrderMultipleContracts = "BatchGetLatestValues supports same read with different params and results retain order from request even with multiple contracts"
	ContractReaderBatchGetLatestValueSetsErrorsProperly                                 = "BatchGetLatestValues sets errors properly"
)

// Query key
const (
	ContractReaderQueryKeyNotFound                     = "QueryKey returns not found if sequence never happened"
	ContractReaderQueryKeyReturnsData                  = "QueryKey returns sequence data properly"
	ContractReaderQueryKeyReturnsDataAsValuesDotValue  = "QueryKey returns sequence data properly as values.Value"
	ContractReaderQueryKeyCanFilterWithValueComparator = "QueryKey can filter data with value comparator"
	ContractReaderQueryKeyCanLimitResultsWithCursor    = "QueryKey can limit results with cursor"
)

type ChainComponentsInterfaceTester[T TestingT[T]] interface {
	BasicTester[T]
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
	EventNameField                              = EventName + ".Field"
	ProtoTest                                   = "ProtoTest"
	ProtoTestIntComparator                      = ProtoTest + ".IntComparator"
	ProtoTestStringComparator                   = ProtoTest + ".StringComparator"
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

	tests := []Testcase[T]{
		{
			Name: ContractReaderGetLatestValueAsValuesDotValue,
			Test: func(t T) {
				ctx := tests.Context(t)
				firstItem := CreateTestStruct(0, tester)

				contracts := tester.GetBindings(t)
				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, firstItem, contracts[0], types.Unconfirmed)

				secondItem := CreateTestStruct(1, tester)

				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, secondItem, contracts[0], types.Unconfirmed)

				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0] // minimum of one bound contract expected, otherwise panics

				require.NoError(t, cr.Bind(ctx, bindings))

				params := &LatestParams{I: 1}
				var value values.Value

				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, &value)
				require.NoError(t, err)

				actual := TestStruct{}
				err = value.UnwrapTo(&actual)
				require.NoError(t, err)
				assert.Equal(t, &firstItem, &actual)

				params = &LatestParams{I: 2}
				err = cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodTakingLatestParamsReturningTestStruct), primitives.Unconfirmed, params, &value)
				require.NoError(t, err)

				actual = TestStruct{}
				err = value.UnwrapTo(&actual)
				require.NoError(t, err)
				assert.Equal(t, &secondItem, &actual)
			},
		},

		{
			Name: ContractReaderGetLatestValueNoArgumentsAndPrimitiveReturnAsValuesDotValue,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var value values.Value
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &value)
				require.NoError(t, err)

				var prim uint64
				err = value.UnwrapTo(&prim)
				require.NoError(t, err)

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
			},
		},
		{
			Name: ContractReaderGetLatestValueNoArgumentsAndSliceReturnAsValueDotValue,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var value values.Value
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64Slice), primitives.Unconfirmed, nil, &value)
				require.NoError(t, err)

				var slice []uint64
				err = value.UnwrapTo(&slice)
				require.NoError(t, err)
				assert.Equal(t, AnySliceToReadWithoutAnArgument, slice)
			},
		},
		{
			Name: ContractReaderGetLatestValue,
			Test: func(t T) {
				ctx := tests.Context(t)
				firstItem := CreateTestStruct(0, tester)

				contracts := tester.GetBindings(t)
				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, firstItem, contracts[0], types.Unconfirmed)

				secondItem := CreateTestStruct(1, tester)

				_ = SubmitTransactionToCW(t, tester, MethodSettingStruct, secondItem, contracts[0], types.Unconfirmed)

				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0] // minimum of one bound contract expected, otherwise panics

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
			Name: ContractReaderGetLatestValueWithPrimitiveReturn,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
			},
		},
		{
			Name: ContractReaderGetLatestValueBasedOnConfidenceLevel,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
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
				bound := BindingsByName(bindings, callArgs.ContractName)[0]

				require.Error(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), primitives.Finalized, callArgs.Params, &prim1))

				err := WaitForTransactionStatus(t, tester, txID, types.Finalized, mockRun)
				require.NoError(t, err)

				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningAlterableUint64), primitives.Finalized, nil, &prim1))
				assert.Equal(t, uint64(10), prim1)

				_ = SubmitTransactionToCW(t, tester, MethodSettingUint64, PrimitiveArgs{Value: 20}, contracts[0], types.Unconfirmed)

				var prim2 uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(callArgs.ReadName), callArgs.ConfidenceLevel, callArgs.Params, &prim2))
				assert.Equal(t, uint64(20), prim2)
			},
		},
		{
			Name: ContractReaderGetLatestValueFromMultipleContractsNamesSameFunction,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnySecondContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64), primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, prim)
			},
		},
		{
			Name: ContractReaderGetLatestValueNoArgumentsAndSliceReturn,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				var slice []uint64
				require.NoError(t, cr.GetLatestValue(ctx, bound.ReadIdentifier(MethodReturningUint64Slice), primitives.Unconfirmed, nil, &slice))

				assert.Equal(t, AnySliceToReadWithoutAnArgument, slice)
			},
		},
		{
			Name: ContractReaderGetLatestValueWithModifiersUsingOwnMapstrctureOverrides,
			Test: func(t T) {
				ctx := tests.Context(t)
				testStruct := CreateTestStruct(0, tester)
				testStruct.BigField = nil
				testStruct.AccountStruct.Account = nil

				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				cr := tester.GetContractReader(t)
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
			Name: ContractReaderGetLatestValueGetsLatestForEvent,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))
				contracts := tester.GetBindings(t)

				ts := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts, contracts[0], types.Unconfirmed)

				ts = CreateTestStruct[T](1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts, contracts[0], types.Unconfirmed)

				result := &TestStruct{}
				require.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			Name: ContractReaderGetLatestValueBasedOnConfidenceLevelForEvent,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))
				ts1 := CreateTestStruct[T](2, tester)

				txID := SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, bindings[0], types.Unconfirmed)

				result := &TestStruct{}
				require.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err != nil && assert.ErrorContains(t, err, types.ErrNotFound.Error())
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				err := WaitForTransactionStatus(t, tester, txID, types.Finalized, mockRun)
				require.NoError(t, err)

				ts2 := CreateTestStruct[T](3, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, bindings[0], types.Unconfirmed)

				require.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Finalized, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				require.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts2)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			Name: ContractReaderGetLatestValueReturnsNotFoundWhenNotTriggeredForEvent,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, bindings))

				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, bound.ReadIdentifier(EventName), primitives.Unconfirmed, nil, &result)
				assert.True(t, errors.Is(err, types.ErrNotFound))
			},
		},
		{
			Name: ContractReaderGetLatestValueWithFilteringForEvent,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

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
			},
		},
	}
	RunTests(t, tester, tests)
}

func runContractReaderBatchGetLatestValuesInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], mockRun bool) {

	testCases := []Testcase[T]{
		{
			Name: ContractReaderBatchGetLatestValue,
			Test: func(t T) {
				// setup test data
				firstItem := CreateTestStruct(1, tester)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

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

				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				assert.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodTakingLatestParamsReturningTestStruct)
				assert.Equal(t, &firstItem, returnValue)
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueNoArgumentsPrimitiveReturn,
			Test: func(t T) {
				// setup call data
				var primitiveReturnValue uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValuesRequest[bound] = []types.BatchRead{
					{
						ReadName:  MethodReturningUint64,
						Params:    nil,
						ReturnVal: &primitiveReturnValue,
					},
				}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))

				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64)
				assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValue.(*uint64))
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueMultipleContractNamesSameFunction,
			Test: func(t T) {
				var primitiveReturnValueAnyContract, primitiveReturnValueAnySecondContract uint64
				batchGetLatestValuesRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := BindingsByName(bindings, AnyContractName)[0]
				bound2 := BindingsByName(bindings, AnySecondContractName)[0]

				batchGetLatestValuesRequest[bound1] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnyContract}}
				batchGetLatestValuesRequest[bound2] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
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
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueNoArgumentsWithSliceReturn,
			Test: func(t T) {
				// setup call data
				var sliceReturnValue []uint64
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValueRequest[bound] = []types.BatchRead{{ReadName: MethodReturningUint64Slice, Params: nil, ReturnVal: &sliceReturnValue}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				require.NoError(t, cr.Bind(ctx, bindings))
				result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValueRequest)
				require.NoError(t, err)

				anyContractBatch := result[bound]
				returnValue, err := anyContractBatch[0].GetResult()
				require.NoError(t, err)
				assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64Slice)
				assert.Equal(t, AnySliceToReadWithoutAnArgument, *returnValue.(*[]uint64))
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueWithModifiersOwnMapstructureOverride,
			Test: func(t T) {
				// setup call data
				testStruct := CreateTestStruct(0, tester)
				testStruct.BigField = nil
				testStruct.AccountStruct.Account = nil
				actual := &TestStructWithExtraField{}
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				batchGetLatestValueRequest[bound] = []types.BatchRead{{ReadName: MethodReturningSeenStruct, Params: testStruct, ReturnVal: actual}}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
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
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueDifferentParamsResultsRetainOrder,
			Test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

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
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueDifferentParamsResultsRetainOrderMultipleContracts,
			Test: func(t T) {
				batchCallEntry := make(BatchCallEntry)
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := BindingsByName(bindings, AnyContractName)[0]
				bound2 := BindingsByName(bindings, AnySecondContractName)[0]

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
			},
		},
		{
			Name: ContractReaderBatchGetLatestValueSetsErrorsProperly,
			Test: func(t T) {
				batchGetLatestValueRequest := make(types.BatchGetLatestValuesRequest)
				bindings := tester.GetBindings(t)
				bound1 := BindingsByName(bindings, AnyContractName)[0]
				bound2 := BindingsByName(bindings, AnySecondContractName)[0]

				for i := 0; i < 10; i++ {
					// setup call data and set invalid params that cause an error
					batchGetLatestValueRequest[bound1] = append(batchGetLatestValueRequest[bound1], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
					batchGetLatestValueRequest[bound2] = append(batchGetLatestValueRequest[bound2], types.BatchRead{ReadName: MethodTakingLatestParamsReturningTestStruct, Params: &LatestParams{I: 0}, ReturnVal: &TestStruct{}})
				}

				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
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
			},
		},
	}
	RunTests(t, tester, testCases)
}

func runQueryKeyInterfaceTests[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T]) {
	tests := []Testcase[T]{
		{
			Name: ContractReaderQueryKeyNotFound,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)
				bound := BindingsByName(bindings, AnyContractName)[0]

				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				logs, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &TestStruct{})

				require.NoError(t, err)
				assert.Len(t, logs, 0)
			},
		},
		{
			Name: ContractReaderQueryKeyReturnsData,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				boundContract := BindingsByName(bindings, AnyContractName)[0]

				ts1 := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, boundContract, types.Unconfirmed)
				ts2 := CreateTestStruct[T](1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, boundContract, types.Unconfirmed)

				ts := &TestStruct{}
				require.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, boundContract, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, ts)
					return err == nil && len(sequences) == 2 && reflect.DeepEqual(&ts1, sequences[1].Data) && reflect.DeepEqual(&ts2, sequences[0].Data)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			Name: ContractReaderQueryKeyReturnsDataAsValuesDotValue,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				bound := BindingsByName(bindings, AnyContractName)[0]

				ts1 := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, bindings[0], types.Unconfirmed)
				ts2 := CreateTestStruct[T](1, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, bindings[0], types.Unconfirmed)

				var value values.Value

				require.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, bound, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &value)
					if err != nil || len(sequences) != 2 {
						return false
					}

					data1 := *sequences[1].Data.(*values.Value)
					ts := TestStruct{}
					err = data1.UnwrapTo(&ts)
					require.NoError(t, err)
					assert.Equal(t, &ts1, &ts)

					data2 := *sequences[0].Data.(*values.Value)
					ts = TestStruct{}
					err = data2.UnwrapTo(&ts)
					require.NoError(t, err)
					assert.Equal(t, &ts2, &ts)

					return true
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			Name: ContractReaderQueryKeyCanFilterWithValueComparator,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				boundContract := BindingsByName(bindings, AnyContractName)[0]

				ts1 := CreateTestStruct[T](0, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1, boundContract, types.Unconfirmed)
				ts2 := CreateTestStruct[T](15, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2, boundContract, types.Unconfirmed)
				ts3 := CreateTestStruct[T](35, tester)
				_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts3, boundContract, types.Unconfirmed)

				ts := &TestStruct{}
				require.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, boundContract, query.KeyFilter{Key: EventName, Expressions: []query.Expression{
						query.Comparator("Field",
							primitives.ValueComparator{
								Value:    *ts2.Field,
								Operator: primitives.Gte,
							},
							primitives.ValueComparator{
								Value:    *ts3.Field,
								Operator: primitives.Lte,
							}),
					},
					}, query.LimitAndSort{}, ts)
					return err == nil && len(sequences) == 2 && reflect.DeepEqual(&ts2, sequences[1].Data) && reflect.DeepEqual(&ts3, sequences[0].Data)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*500)
			},
		},
		{
			Name: ContractReaderQueryKeyCanLimitResultsWithCursor,
			Test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetContractReader(t)
				bindings := tester.GetBindings(t)

				require.NoError(t, cr.Bind(ctx, bindings))
				boundContract := BindingsByName(bindings, AnyContractName)[0]

				// keep this an even number such that the cursor limit can be in batches of 2
				testStructs := make([]TestStruct, 4)

				// create test structs in sequence
				for idx := range testStructs {
					testStructs[idx] = CreateTestStruct(idx*2, tester)

					_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, testStructs[idx], boundContract, types.Unconfirmed)
				}

				require.Eventually(t, func() bool {
					var allSequences []types.Sequence

					filter := query.KeyFilter{Key: EventName, Expressions: []query.Expression{
						query.Confidence(primitives.Finalized),
					}}
					limit := query.LimitAndSort{
						SortBy: []query.SortBy{query.NewSortBySequence(query.Asc)},
						Limit:  query.CountLimit(2),
					}

					for idx := 0; idx < len(testStructs)/2; idx++ {
						// sequences from queryKey without limit and sort should be in descending order
						sequences, err := cr.QueryKey(ctx, boundContract, filter, limit, &TestStruct{})

						require.NoError(t, err)

						if len(sequences) == 0 {
							continue
						}

						limit.Limit = query.CursorLimit(sequences[len(sequences)-1].Cursor, query.CursorFollowing, 2)
						allSequences = append(allSequences, sequences...)
					}

					return len(allSequences) == len(testStructs) &&
						reflect.DeepEqual(&testStructs[0], allSequences[0].Data) &&
						reflect.DeepEqual(&testStructs[len(testStructs)-1], allSequences[len(testStructs)-1].Data)
				}, tester.MaxWaitTimeForEvents(), 500*time.Millisecond)
			},
		},
	}

	RunTests(t, tester, tests)
}

func BindingsByName(bindings []types.BoundContract, name string) []types.BoundContract {
	named := make([]types.BoundContract, 0, len(bindings))

	for idx := range bindings {
		if bindings[idx].Name == name {
			named = append(named, bindings[idx])
		}
	}

	return named
}
