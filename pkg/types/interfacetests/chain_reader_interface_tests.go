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
	SetUintLatestValue(t T, val uint64)
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
	DifferentMethodReturningUint64              = "GetDifferentPrimitiveValue"
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
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				actual := &TestStruct{}
				params := &LatestParams{I: 1}
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodTakingLatestParamsReturningTestStruct, primitives.Unconfirmed, params, actual))
				assert.Equal(t, &firstItem, actual)

				params.I = 2
				actual = &TestStruct{}
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodTakingLatestParamsReturningTestStruct, primitives.Unconfirmed, params, actual))
				assert.Equal(t, &secondItem, actual)
			},
		},
		{
			name: "Get latest value without arguments and with primitive return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningUint64, primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyValueToReadWithoutAnArgument, prim)
			},
		},
		{ // skip test for chains with immediate finality?
			name: "Get latest value based on confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				var prim1 uint64
				tester.SetUintLatestValue(t, 10)
				require.Error(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningAlterableUint64, primitives.Finalized, nil, &prim1))

				tester.GenerateBlocksTillConfidenceLevel(t, AnyContractName, EventName, primitives.Finalized)
				tester.SetUintLatestValue(t, 20)

				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningAlterableUint64, primitives.Finalized, nil, &prim1))
				assert.Equal(t, uint64(10), prim1)

				var prim2 uint64
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningAlterableUint64, primitives.Unconfirmed, nil, &prim2))
				assert.Equal(t, uint64(20), prim2)
			},
		},
		{
			name: "Get latest value allows a contract name to resolve different contracts internally",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, DifferentMethodReturningUint64, primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, prim)
			},
		},
		{
			name: "Get latest value allows multiple contract names to have the same function name",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				bindings := tester.GetBindings(t)
				seenAddrs := map[string]bool{}
				for _, binding := range bindings {
					assert.False(t, seenAddrs[binding.Address])
					seenAddrs[binding.Address] = true
				}

				require.NoError(t, cr.Bind(ctx, bindings))

				var prim uint64
				require.NoError(t, cr.GetLatestValue(ctx, AnySecondContractName, MethodReturningUint64, primitives.Unconfirmed, nil, &prim))

				assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, prim)
			},
		},
		{
			name: "Get latest value without arguments and with slice return",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				var slice []uint64
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningUint64Slice, primitives.Unconfirmed, nil, &slice))

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
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				actual := &TestStructWithExtraField{}
				require.NoError(t, cr.GetLatestValue(ctx, AnyContractName, MethodReturningSeenStruct, primitives.Unconfirmed, testStruct, actual))

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
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				ts := CreateTestStruct[T](0, tester)
				tester.TriggerEvent(t, &ts)
				ts = CreateTestStruct[T](1, tester)
				tester.TriggerEvent(t, &ts)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, AnyContractName, EventName, primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{ // skip test for chains with immediate finality?
			name: "Get latest event based on provided confidence level",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				ts1 := CreateTestStruct[T](2, tester)
				tester.TriggerEvent(t, &ts1)

				result := &TestStruct{}
				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, AnyContractName, EventName, primitives.Finalized, nil, &result)
					return err != nil && assert.ErrorContains(t, err, types.ErrNotFound.Error())
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				tester.GenerateBlocksTillConfidenceLevel(t, AnyContractName, EventName, primitives.Finalized)
				ts2 := CreateTestStruct[T](3, tester)
				tester.TriggerEvent(t, &ts2)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, AnyContractName, EventName, primitives.Finalized, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

				assert.Eventually(t, func() bool {
					err := cr.GetLatestValue(ctx, AnyContractName, EventName, primitives.Unconfirmed, nil, &result)
					return err == nil && reflect.DeepEqual(result, &ts2)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
		{
			name: "Get latest value returns not found if event was never triggered",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, AnyContractName, EventName, primitives.Unconfirmed, nil, &result)
				assert.True(t, errors.Is(err, types.ErrNotFound))
			},
		},
		{
			name: "Get latest value gets latest event with filtering",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				ts0 := CreateTestStruct(0, tester)
				tester.TriggerEvent(t, &ts0)
				ts1 := CreateTestStruct(1, tester)
				tester.TriggerEvent(t, &ts1)

				filterParams := &FilterEventParams{Field: *ts0.Field}
				assert.Never(t, func() bool {
					result := &TestStruct{}
					err := cr.GetLatestValue(ctx, AnyContractName, EventWithFilterName, primitives.Unconfirmed, filterParams, &result)
					return err == nil && reflect.DeepEqual(result, &ts1)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
				// get the result one more time to verify it.
				// Using the result from the Never statement by creating result outside the block is a data race
				result := &TestStruct{}
				err := cr.GetLatestValue(ctx, AnyContractName, EventWithFilterName, primitives.Unconfirmed, filterParams, &result)
				require.NoError(t, err)
				assert.Equal(t, &ts0, result)
			},
		},
	}
	runTests(t, tester, tests)
}

func runQueryKeyInterfaceTests[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T]) {
	tests := []testcase[T]{
		{
			name: "QueryKey returns not found if sequence never happened",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)

				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))

				logs, err := cr.QueryKey(ctx, AnyContractName, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, &TestStruct{})

				require.NoError(t, err)
				assert.Len(t, logs, 0)
			},
		},
		{
			name: "QueryKey returns sequence data properly",
			test: func(t T) {
				ctx := tests.Context(t)
				cr := tester.GetChainReader(t)
				require.NoError(t, cr.Bind(ctx, tester.GetBindings(t)))
				ts1 := CreateTestStruct[T](0, tester)
				tester.TriggerEvent(t, &ts1)
				ts2 := CreateTestStruct[T](1, tester)
				tester.TriggerEvent(t, &ts2)

				ts := &TestStruct{}
				assert.Eventually(t, func() bool {
					// sequences from queryKey without limit and sort should be in descending order
					sequences, err := cr.QueryKey(ctx, AnyContractName, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, ts)
					return err == nil && len(sequences) == 2 && reflect.DeepEqual(&ts1, sequences[1].Data) && reflect.DeepEqual(&ts2, sequences[0].Data)
				}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			},
		},
	}

	runTests(t, tester, tests)
}
